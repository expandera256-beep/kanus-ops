"""Sequential supervisor core. No cloud, boot, SSH or policy-writing capability.

Production adapters must implement the Backend contract; none is auto-discovered.
All changing operations require the OS lock, current fence and bounded work lease.
"""
from __future__ import annotations
import threading
import time
from dataclasses import asdict, dataclass
from enum import Enum
from typing import Callable, Protocol
from .ledger import Ledger
from .models import (DAILY_RUNTIME, MAX_UPTIME, IDLE, KLEIN, SCHNELL,
                     Job, Model, Rejected, Snapshot, select_model)
from .ownership import Ownership, Fence


class State(str, Enum):
    CHAT_READY = 'CHAT_READY'
    CHAT_DRAINING = 'CHAT_DRAINING'
    IMAGE_LOADING = 'IMAGE_LOADING'
    IMAGE_READY = 'IMAGE_READY'
    RUNNING = 'RUNNING'
    IMAGE_DRAINING = 'IMAGE_DRAINING'
    CHAT_RELOADING = 'CHAT_RELOADING'
    FAILED_CLOSED = 'FAILED_CLOSED'


EDGES = {
    State.CHAT_READY: {State.CHAT_DRAINING},
    State.CHAT_DRAINING: {State.CHAT_READY, State.IMAGE_LOADING, State.CHAT_RELOADING},
    State.IMAGE_LOADING: {State.IMAGE_READY, State.IMAGE_DRAINING},
    State.IMAGE_READY: {State.RUNNING, State.IMAGE_DRAINING},
    State.RUNNING: {State.IMAGE_READY, State.IMAGE_DRAINING},
    State.IMAGE_DRAINING: {State.CHAT_RELOADING},
    State.CHAT_RELOADING: {State.CHAT_READY},
    State.FAILED_CLOSED: set(),
}


@dataclass(frozen=True)
class Budget:
    daily_remaining: float
    uptime_remaining: float
    # All values are authoritative snapshots from the existing lifecycle adapter.
    def available(self):
        import math
        if not all(math.isfinite(x) for x in (self.daily_remaining, self.uptime_remaining)):
            raise Rejected('unknown runtime balance')
        if not 0 <= self.daily_remaining <= DAILY_RUNTIME or not 0 <= self.uptime_remaining <= MAX_UPTIME:
            raise Rejected('invalid authoritative runtime balance')
        return min(self.daily_remaining, self.uptime_remaining)


@dataclass(frozen=True)
class Estimate:
    load: float = 120
    per_image: float = 6
    encode: float = 10
    reload: float = 180
    stop: float = 30
    drain: float = 60

    def required(self, n):
        import math
        if not all(math.isfinite(v) and v > 0 for v in asdict(self).values()):
            raise Rejected('invalid budget reserve')
        return self.load + n * self.per_image + self.encode + self.reload + self.stop + self.drain


class WorkLease:
    """Heartbeat only for admitted work; TTL cannot extend the absolute deadline.

    A production lifecycle bridge receives boot/PID/start/fence/job identity and
    must atomically check it on each heartbeat. Expired leases cannot resurrect.
    """
    def __init__(self, owner: Ownership, job_id, deadline, clock, heartbeat, ttl=60):
        self.owner, self.job_id = owner, job_id
        self.deadline, self.clock, self.heartbeat, self.ttl = deadline, clock, heartbeat, ttl
        self.token = owner.assert_current()
        self.expires = min(deadline, clock() + ttl)
        self.closed = False

    def check(self):
        self.owner.assert_current(self.token)
        if self.closed or self.clock() >= min(self.expires, self.deadline):
            raise Rejected('work lease expired; no new work')

    def progress(self):
        self.check()
        identity = dict(asdict(self.token), job_id=self.job_id,
                        absolute_deadline=self.deadline, ttl=self.ttl)
        if self.heartbeat(identity) is not True:
            raise Rejected('lifecycle heartbeat refused')
        self.expires = min(self.deadline, self.clock() + self.ttl)


class Backend(Protocol):
    """Reviewed integration boundary, NOT an implementation of live services.

    Pause blocks NEW chat admission but never interrupts active chat. stop_chat
    must also suppress restart/tunnel races. cuda_empty must use NVML and prove
    no compute/graphics context remains; unknown means False. All blocking
    methods honor the lease deadline and progress at <=20s. Stop/reload use a
    separately reserved cleanup deadline and must not require a valid work TTL.
    reload_chat uses existing local NVMe pack only, no network or new boot.
    """
    def prompt_valid(self, prompt: str, model: str) -> bool: ...
    def reserve(self, job: Job, identity: dict, required_seconds: float) -> bool: ...
    def heartbeat(self, identity: dict) -> bool: ...
    def release(self, job_id: str, identity: dict) -> None: ...
    def chat_pending(self) -> bool: ...
    def chat_active(self) -> bool: ...
    def pause_chat_admission(self, lease: WorkLease) -> None: ...
    def resume_chat_admission(self, deadline: float) -> None: ...
    def wait_chat_progress(self, lease: WorkLease) -> None: ...
    def stop_chat(self, lease: WorkLease) -> None: ...
    def cuda_empty(self) -> bool: ...
    def load_media(self, model: Model, snapshot: Snapshot, lease: WorkLease) -> None: ...
    def generate_one(self, job: Job, index: int, model: Model, lease: WorkLease) -> str: ...
    def stop_media(self, deadline: float) -> None: ...
    def reload_chat(self, deadline: float) -> None: ...
    def chat_ready(self) -> bool: ...
    def synthetic_smoke(self, deadline: float) -> bool: ...
    def cancelled(self, job_id: str) -> bool: ...
    def alert(self, code: str) -> None: ...


class Supervisor:
    def __init__(self, owner: Ownership, ledger: Ledger, backend: Backend,
                 snapshots: dict[str, Snapshot], *, clock: Callable = time.monotonic):
        self.owner, self.ledger, self.backend = owner, ledger, backend
        self.snapshots, self.clock = snapshots, clock
        self.state = State.FAILED_CLOSED if ledger.blocked else State.CHAT_READY
        self.history = [self.state]
        self.started = False
        self._serial = threading.Lock()
        self.results = {}

    def start(self):
        """Acquire authority only; never boot, load, wake or start a service."""
        if self.started:
            raise Rejected('supervisor already started')
        self.owner.acquire()
        self.started = True
        try:
            ready = self.backend.chat_ready()
        except Exception:
            ready = False
        if ready is not True:
            self._fail_closed('initial_chat_not_ready')

    def _transition(self, state):
        self.owner.assert_current()
        if state not in EDGES[self.state]:
            raise Rejected(f'illegal transition: {self.state.value} -> {state.value}')
        self.state = state
        self.history.append(state)

    def _fail_closed(self, reason):
        self.ledger.block(reason)
        self.state = State.FAILED_CLOSED
        self.history.append(self.state)
        try:
            self.backend.alert(reason)
        except Exception:
            pass

    def health(self):
        # Deliberately no backend calls, ownership acquisition or lease refresh.
        return {'state': self.state.value, 'ready': self.started and self.state == State.CHAT_READY,
                'image_admission': self.started and self.state == State.CHAT_READY,
                'policy': {'daily_runtime': DAILY_RUNTIME, 'max_uptime': MAX_UPTIME, 'idle': IDLE}}

    ready = health

    def list_models(self):
        return (KLEIN.key, SCHNELL.key)  # Availability is NOT a download/access claim.

    def poll(self, job_id):
        # Durable READ ONLY lookup; never begins a reservation or takes a lease.
        return self.ledger.poll(job_id)

    def _restore(self, deadline, *, chat_stopped, media_attempted):
        # Never reload while an unknown/stale owner might still be using CUDA.
        self.owner.assert_current()
        if media_attempted:
            if self.state != State.IMAGE_DRAINING:
                self._transition(State.IMAGE_DRAINING)
            self.backend.stop_media(deadline)
            if self.backend.cuda_empty() is not True:
                raise Rejected('media CUDA contexts remain or cannot be proven absent')
        if chat_stopped:
            if self.backend.cuda_empty() is not True:
                raise Rejected('CUDA context release unproven; reload forbidden')
            self._transition(State.CHAT_RELOADING)
            if self.clock() >= deadline:
                raise Rejected('cleanup deadline exceeded')
            self.backend.reload_chat(deadline)
            if (self.backend.chat_ready() is not True
                    or self.backend.synthetic_smoke(deadline) is not True):
                raise Rejected('chat restore readiness/smoke failed')
        self.owner.assert_current()
        if self.clock() >= deadline:
            raise Rejected('restore exceeded absolute deadline')
        self.backend.resume_chat_admission(deadline)
        self._transition(State.CHAT_READY)
        # Clear the crash barrier only after proven safe restore.
        with self.ledger.db:
            self.ledger.db.execute("DELETE FROM control WHERE key='blocked'")

    def run(self, job: Job, budget: Budget, *, estimate=Estimate(),
            fallback_permission=False, gated_access=False, fallback_reason=None):
        if not self._serial.acquire(blocking=False):
            raise Rejected('another admitted image job is active')
        try:
            return self._run(job, budget, estimate, fallback_permission, gated_access, fallback_reason)
        finally:
            self._serial.release()

    def _run(self, job, budget, estimate, fallback_permission, gated_access, fallback_reason):
        self.owner.assert_current()
        if not self.started or self.state != State.CHAT_READY or self.ledger.blocked:
            raise Rejected('image admission closed; no new boot or automatic recovery')
        job.validate(self.backend.prompt_valid)
        model = select_model(job.model, fallback_permission=fallback_permission,
                             gated_access=gated_access, reason=fallback_reason)
        required = estimate.required(job.n)
        available = budget.available()
        if required > available:
            raise Rejected('insufficient runtime for work + encode + restore + stop')
        if model.key not in self.snapshots:
            raise Rejected('verified local snapshot unavailable; network fetch forbidden')
        # Verify before disrupting chat; runtime worker verifies again before load.
        self.snapshots[model.key].verify(model)
        cached = self.ledger.begin(job)
        if cached is not None:
            self.results[job.job_id] = cached
            return dict(cached)
        now = self.clock()
        deadline = now + required  # <= authoritative remaining uptime/day balance
        work_deadline = deadline - estimate.reload - estimate.stop
        token = self.owner.assert_current()
        identity = dict(asdict(token), job_id=job.job_id, absolute_deadline=deadline)
        result = {'job_id': job.job_id, 'status': 'failed', 'model': model.key,
                  'outputs': [], 'seeds': [], 'error': None}
        reserved = False
        chat_stopped = media_attempted = paused = False
        lease = None
        try:
            if self.backend.reserve(job, identity, required) is not True:
                raise Rejected('reservation denied')
            reserved = True
            lease = WorkLease(self.owner, job.job_id, work_deadline, self.clock, self.backend.heartbeat)
            lease.progress()
            # Durable barrier means a crash during a swap never silently returns ready.
            self.ledger.block('swap_in_progress_requires_restore')
            self._transition(State.CHAT_DRAINING)
            paused = True
            self.backend.pause_chat_admission(lease)
            drain_deadline = min(work_deadline, now + estimate.drain)
            while self.backend.chat_active():
                lease.progress()
                if self.clock() >= drain_deadline:
                    raise Rejected('active chat drain timed out; chat not interrupted')
                self.backend.wait_chat_progress(lease)
            lease.progress()
            if self.backend.chat_pending() or self.backend.cancelled(job.job_id):
                result['status'] = 'preempted'
            else:
                # Flag before stop: partial stop failure also needs restore proof.
                chat_stopped = True
                self.backend.stop_chat(lease)
                if self.backend.cuda_empty() is not True:
                    raise Rejected('chat CUDA contexts remain or unknown')
                if self.backend.chat_pending():
                    result['status'] = 'preempted'
                else:
                    self._transition(State.IMAGE_LOADING)
                    media_attempted = True
                    self.backend.load_media(model, self.snapshots[model.key], lease)
                    lease.progress()
                    self._transition(State.IMAGE_READY)
                    for index in range(job.n):
                        lease.progress()
                        if self.backend.chat_pending() or self.backend.cancelled(job.job_id):
                            result['status'] = 'partial' if result['outputs'] else 'preempted'
                            break
                        self._transition(State.RUNNING)
                        # Exactly one image per call; no batch n=4, no automatic model fallback.
                        output = self.backend.generate_one(job, index, model, lease)
                        lease.progress()
                        result['outputs'].append(output)
                        result['seeds'].append(job.seed + index)
                        self._transition(State.IMAGE_READY)
                    else:
                        result['status'] = 'succeeded'
        except Exception as exc:
            # No prompt, exception text or credentials in journal/logs.
            result['error'] = type(exc).__name__
            result['status'] = 'partial' if result['outputs'] else 'failed'
        finally:
            if paused:
                try:
                    self._restore(deadline - estimate.stop, chat_stopped=chat_stopped,
                                  media_attempted=media_attempted)
                except Exception:
                    self._fail_closed('restore_failed_operator_required')
                    result['status'] = 'failed_closed'
                    result['error'] = 'restore_failed'
            if lease:
                lease.closed = True
            if reserved and self.state != State.FAILED_CLOSED:
                try:
                    self.owner.assert_current(token)
                    self.backend.release(job.job_id, identity)
                except Exception:
                    self._fail_closed('lease_release_failed')
                    result['status'] = 'failed_closed'
        result = self.ledger.finish(job.job_id, result)
        self.results[job.job_id] = result
        return dict(result)

    def close(self):
        if self.state != State.CHAT_READY:
            raise Rejected('cannot release authority before safe restore; operator required')
        self.owner.release()
        self.started = False
