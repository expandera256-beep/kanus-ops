"""Explicit CPU sandbox adapters; no service discovery, cloud or CUDA access.

These implement LocalBackend's injectable chat/lifecycle interfaces, NOT a live
license/payment bridge. Instantiate only in a sandbox. Nothing auto-binds them.
"""
from __future__ import annotations

import math
import threading
import time
from pathlib import Path

from .models import (DAILY_RUNTIME, MAX_UPTIME, KLEIN, Rejected, Snapshot,
                     select_model)
from .supervisor import Budget


class SandboxChatAdapter:
    """In-memory chat handoff, never systemd/process operations.

    drain_step is a bounded, caller-injected sandbox callback (not a worker
    discovery hook). It must complete existing synthetic chat itself; this class
    never forcibly cuts active chat off. No callback means fail rather than spin.
    """
    live_bound = False

    def __init__(self, *, clock=time.monotonic, drain_step=None):
        self.clock, self.drain_step = clock, drain_step
        self._lock = threading.RLock()
        self._active = self._pending = 0
        self._paused = False
        self._running = True
        self.smoke_ok = True  # Synthetic readiness only, NOT real chat inference.

    def set_work(self, *, active=0, pending=0):
        if any(type(n) is not int or n < 0 for n in (active, pending)):
            raise Rejected('invalid sandbox chat counts')
        with self._lock:
            if active and not self._running:
                raise Rejected('sandbox chat stopped')
            if active > self._active and self._paused:
                raise Rejected('sandbox chat admission paused')
            self._active, self._pending = active, pending

    def chat_pending(self):
        with self._lock:
            return self._pending > 0

    def chat_active(self):
        with self._lock:
            return self._active > 0

    def chat_ready(self):
        with self._lock:
            return self._running

    def pause_chat_admission(self, lease):
        lease.progress()
        with self._lock:
            self._paused = True

    def wait_chat_progress(self, lease):
        lease.progress()
        if self.chat_active():
            if self.drain_step is None:
                raise Rejected('sandbox drain callback required; active chat not interrupted')
            self.drain_step(self, lease)
            lease.progress()

    def stop_chat(self, lease):
        lease.progress()
        with self._lock:
            if not self._paused or self._active or self._pending:
                raise Rejected('pause and complete chat drain required')
            self._running = False

    def _deadline(self, deadline):
        if (type(deadline) not in (int, float) or not math.isfinite(deadline)
                or self.clock() >= deadline):
            raise Rejected('sandbox cleanup deadline exceeded')

    def reload_chat(self, deadline):
        self._deadline(deadline)
        with self._lock:
            if not self._paused:
                raise Rejected('sandbox reload requires paused admission')
            self._running = True

    def resume_chat_admission(self, deadline):
        self._deadline(deadline)
        with self._lock:
            if not self._running:
                raise Rejected('sandbox chat not ready')
            self._paused = False

    def synthetic_smoke(self, deadline):
        self._deadline(deadline)
        return self.chat_ready() and self.smoke_ok is True


class SandboxLifecycleAdapter:
    """Atomic in-memory runtime reservation, NOT money or live SQL.

    Expired/cancelled reservations remain held until fenced cleanup release;
    heartbeat never resurrects them. Release charges elapsed sandbox runtime,
    returns unused reserved time, and cannot charge a second time. A crashed
    process loses this in-memory store: live durability/reconciliation NOT TESTED.
    All calls use the same injected monotonic clock as Supervisor/WorkLease.
    """
    live_bound = False
    _fence_fields = ('generation', 'boot_id', 'pid', 'process_start')

    def __init__(self, *, daily_remaining=DAILY_RUNTIME,
                 uptime_remaining=MAX_UPTIME, clock=time.monotonic):
        Budget(daily_remaining, uptime_remaining).available()
        self.daily_remaining, self.uptime_remaining = daily_remaining, uptime_remaining
        self.clock = clock
        self._lock = threading.RLock()
        self._record = None
        self._completed = {}
        self._cancelled = set()
        self._last_fence = None
        self.alerts = []

    @classmethod
    def _fence(cls, identity):
        if not isinstance(identity, dict):
            raise Rejected('sandbox identity required')
        try:
            values = tuple(identity[k] for k in cls._fence_fields)
        except KeyError as exc:
            raise Rejected('incomplete sandbox fence') from exc
        if (type(values[0]) is not int or values[0] < 1
                or type(values[2]) is not int or values[2] < 1
                or any(not isinstance(v, str) or not v for v in (values[1], values[3]))):
            raise Rejected('invalid sandbox fence')
        return values

    @staticmethod
    def _positive(value):
        return type(value) in (int, float) and math.isfinite(value) and value > 0

    def budget(self):
        with self._lock:
            held = self._record['seconds'] if self._record else 0
            return Budget(self.daily_remaining - held, self.uptime_remaining - held)

    def reserve(self, job, identity, seconds):
        with self._lock:
            try:
                fence = self._fence(identity)
            except Rejected:
                return False
            now, deadline = self.clock(), identity.get('absolute_deadline')
            if (self._record is not None or job.job_id in self._completed
                    or job.job_id in self._cancelled or identity.get('job_id') != job.job_id
                    or not self._positive(seconds) or not self._positive(deadline)
                    or deadline <= now or deadline - now > seconds
                    or seconds > self.budget().available()):
                return False
            if self._last_fence is not None and (fence[0] < self._last_fence[0]
                    or (fence[0] == self._last_fence[0] and fence != self._last_fence)):
                return False
            self._last_fence = fence
            self._record = dict(job_id=job.job_id, fence=fence, seconds=seconds,
                                start=now, deadline=deadline, work_deadline=None,
                                expires=min(deadline, now + 60))
            return True

    def heartbeat(self, identity):
        with self._lock:
            try:
                fence = self._fence(identity)
            except Rejected:
                return False
            r, now = self._record, self.clock()
            deadline, ttl = identity.get('absolute_deadline'), identity.get('ttl')
            if (r is None or fence != r['fence'] or identity.get('job_id') != r['job_id']
                    or r['job_id'] in self._cancelled or now >= r['expires']
                    or not self._positive(ttl) or ttl > 60
                    or not self._positive(deadline) or not now < deadline <= r['deadline']):
                return False
            if r['work_deadline'] is not None and deadline != r['work_deadline']:
                return False
            r['work_deadline'] = deadline
            r['expires'] = min(deadline, now + ttl)
            return True

    def release(self, job_id, identity):
        with self._lock:
            fence = self._fence(identity)
            r = self._record
            # Fenced idempotent cleanup only; no second charge.
            if r is None:
                if identity.get('job_id') == job_id and self._completed.get(job_id) == fence:
                    return
                raise Rejected('unknown sandbox reservation')
            if r['job_id'] != job_id or identity.get('job_id') != job_id or fence != r['fence']:
                raise Rejected('sandbox release fence mismatch')
            elapsed = self.clock() - r['start']
            if elapsed < 0 or elapsed > r['seconds']:
                raise Rejected('sandbox runtime exceeded reservation; operator required')
            self.daily_remaining -= elapsed
            self.uptime_remaining -= elapsed
            self._completed[job_id] = fence
            self._record = None

    def cancel(self, job_id):
        with self._lock:
            self._cancelled.add(job_id)

    def cancelled(self, job_id):
        with self._lock:
            return job_id in self._cancelled

    def alert(self, code):
        # Only fixed operator codes; never store prompts or exception details.
        if not isinstance(code, str) or not code or not all(c.isalnum() or c == '_' for c in code):
            raise Rejected('invalid sandbox alert code')
        with self._lock:
            self.alerts.append(code)


APPROVED_KLEIN_PREFIX = (
    'gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/'
    'e7b7dc27f91deacad38e78976d1f2b499d76a294/'
)


class LocalStagingAdapter:
    """Read-only local integrity, staged location is provenance NOT readiness.

    manifest_sha256 must be a separately reviewed configuration trust anchor.
    The package's klein.snapshot.manifest.json is derived from prior verified
    upload evidence, kept OUTSIDE the snapshot. No network/cloud capability.
    """
    def __init__(self, root, manifest, manifest_sha256):
        if not Path(root).is_absolute() or not Path(manifest).is_absolute():
            raise Rejected('absolute local snapshot and external manifest required')
        if Path(manifest).resolve().is_relative_to(Path(root).resolve()):
            raise Rejected('trust manifest must be external to snapshot')
        self.snapshot = Snapshot(Path(root), Path(manifest), manifest_sha256)

    def staged_location(self, key=KLEIN.key):
        if key != KLEIN.key:
            raise Rejected('schnell unavailable/BLOCKED; no staged snapshot')
        return APPROVED_KLEIN_PREFIX

    def verified_snapshot(self, key=KLEIN.key, *, gated_access=False, fallback_permission=False):
        model = select_model(key, gated_access=gated_access, fallback_permission=fallback_permission)
        if model != KLEIN:
            raise Rejected('schnell unavailable/BLOCKED even with permissions; snapshot absent')
        self.snapshot.verify(model)
        return self.snapshot
