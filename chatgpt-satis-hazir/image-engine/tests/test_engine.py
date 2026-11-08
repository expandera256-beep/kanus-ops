import dataclasses
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import types
import unittest
from unittest.mock import patch

from image_engine import Budget, Job, KLEIN, SCHNELL, Ownership, OwnershipError, Rejected, Snapshot, State, Supervisor
from image_engine.ledger import Ledger
from image_engine.models import DAILY_RUNTIME, MAX_UPTIME, IDLE, REQUIRED_FILES, select_model, sha256
from image_engine.ownership import Fence, atomic_json, boot_id, process_start, validate_fence
from image_engine.supervisor import Estimate, WorkLease
from image_engine.worker import OfflineWorker


class Clock:
    def __init__(self): self.now = 100.0
    def __call__(self): return self.now
    def advance(self, seconds): self.now += seconds


def snapshot_fixture(base, model=KLEIN):
    root = base / 'snapshot'
    root.mkdir()
    entries = {}
    for rel in sorted(REQUIRED_FILES[model.key]):
        path = root / rel
        path.parent.mkdir(exist_ok=True)
        path.write_bytes(b'{}')
        entries[rel] = {'bytes': 2, 'sha256': sha256(path)}
    manifest = base / 'trusted.json'
    manifest.write_text(json.dumps({'model': model.repo, 'revision': model.revision, 'files': entries}))
    return Snapshot(root, manifest, sha256(manifest))


class FakeBackend:
    """CPU only; synthetic state and output refs, no model/GPU/cloud operations."""
    def __init__(self, clock):
        self.clock = clock
        self.events = []
        self.active = 0
        self.pending = False
        self.heartbeat_ok = True
        self.restore_ok = True
        self.smoke_ok = True
        self.context_ok = True
        self.reserve_ok = True
        self.cancel = False
        self.fail_generate = None
        self.fail_load = False
        self.fail_stop = False
        self.after_generate = None
        self.loaded = []
        self.generated = []
        self.identities = []
    def prompt_valid(self, prompt, model): return len(prompt) <= 1000
    def reserve(self, job, identity, seconds):
        self.events.append('reserve'); self.identities.append(identity); return self.reserve_ok
    def heartbeat(self, identity):
        self.events.append('heartbeat'); self.identities.append(identity); return self.heartbeat_ok
    def release(self, job_id, identity): self.events.append('release')
    def chat_pending(self): return self.pending
    def chat_active(self): return self.active > 0
    def pause_chat_admission(self, lease): self.events.append('pause')
    def resume_chat_admission(self, deadline): self.events.append('resume')
    def wait_chat_progress(self, lease):
        self.events.append('drain'); self.active -= 1; self.clock.advance(1)
    def stop_chat(self, lease): self.events.append('stop_chat')
    def cuda_empty(self): self.events.append('cuda_empty'); return self.context_ok
    def load_media(self, model, snapshot, lease):
        self.events.append('load'); self.loaded.append(model.key)
        if self.fail_load: raise RuntimeError('synthetic load error')
    def generate_one(self, job, index, model, lease):
        self.events.append('generate'); self.generated.append((index, model.key))
        if self.fail_generate: raise self.fail_generate('synthetic failure')
        if self.after_generate: self.after_generate()
        return f'synthetic-output-{index}'
    def stop_media(self, deadline):
        self.events.append('stop_media')
        if self.fail_stop: raise RuntimeError('synthetic stop failure')
    def reload_chat(self, deadline):
        self.events.append('reload')
        if not self.restore_ok: raise RuntimeError('synthetic restore failure')
    def chat_ready(self): return True
    def synthetic_smoke(self, deadline): self.events.append('smoke'); return self.smoke_ok
    def cancelled(self, job_id): return self.cancel
    def alert(self, code): self.events.append('alert:' + code)


class EngineCase(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.base = Path(self.tmp.name)
        self.clock = Clock()
        self.snap = snapshot_fixture(self.base)
        self.owner = Ownership(self.base / 'authority')
        self.ledger = Ledger(self.base / 'journal.sqlite')
        self.backend = FakeBackend(self.clock)
        self.engine = Supervisor(self.owner, self.ledger, self.backend,
                                 {KLEIN.key: self.snap}, clock=self.clock)
        self.engine.start()
    def tearDown(self):
        self.owner.release()
        self.ledger.close()
        self.tmp.cleanup()
    def run_job(self, **kwargs):
        return self.engine.run(Job('synthetic-1', 'A synthetic geometry test', **kwargs), Budget(3600, 1800))

    def test_transition_order_and_restore(self):
        result = self.run_job()
        self.assertEqual(result['status'], 'succeeded')
        self.assertEqual(self.engine.history, [State.CHAT_READY, State.CHAT_DRAINING, State.IMAGE_LOADING,
            State.IMAGE_READY, State.RUNNING, State.IMAGE_READY, State.IMAGE_DRAINING,
            State.CHAT_RELOADING, State.CHAT_READY])
        events = self.backend.events
        self.assertLess(events.index('stop_chat'), events.index('load'))
        self.assertLess(events.index('stop_media'), events.index('reload'))
        self.assertFalse(self.ledger.blocked)

    def test_illegal_transition_rejected(self):
        with self.assertRaises(Rejected): self.engine._transition(State.RUNNING)

    def test_serial_n4(self):
        result = self.run_job(n=4, seed=123)
        self.assertEqual(result['status'], 'succeeded')
        self.assertEqual(result['seeds'], [123, 124, 125, 126])
        self.assertEqual([i for i, _ in self.backend.generated], [0, 1, 2, 3])
        self.assertEqual(self.engine.history.count(State.RUNNING), 4)

    def test_active_chat_drains_without_interruption(self):
        self.backend.active = 3
        self.run_job()
        self.assertEqual(self.backend.events.count('drain'), 3)
        self.assertLess(max(i for i, e in enumerate(self.backend.events) if e == 'drain'),
                        self.backend.events.index('stop_chat'))

    def test_pending_chat_preempts_before_load(self):
        self.backend.pending = True
        result = self.run_job()
        self.assertEqual(result['status'], 'preempted')
        self.assertNotIn('stop_chat', self.backend.events)
        self.assertNotIn('load', self.backend.events)
        self.assertEqual(self.engine.state, State.CHAT_READY)

    def test_chat_arrival_preempts_at_serial_boundary(self):
        self.backend.after_generate = lambda: setattr(self.backend, 'pending', True)
        result = self.run_job(n=4)
        self.assertEqual(result['status'], 'partial')
        self.assertEqual(len(result['outputs']), 1)
        self.assertEqual(self.engine.state, State.CHAT_READY)

    def test_cancel_is_inert_until_admitted_and_preempts(self):
        self.backend.cancel = True
        self.assertEqual(self.run_job()['status'], 'preempted')
        self.assertNotIn('load', self.backend.events)

    def test_drain_timeout_never_stops_chat(self):
        self.backend.active = 1000
        result = self.run_job()
        self.assertEqual(result['status'], 'failed')
        self.assertNotIn('stop_chat', self.backend.events)
        self.assertEqual(self.engine.state, State.CHAT_READY)

    def test_restore_failure_is_durable_fail_closed(self):
        self.backend.restore_ok = False
        result = self.run_job()
        self.assertEqual(result['status'], 'failed_closed')
        with self.assertRaises(Rejected): self.run_job()
        other = Supervisor(self.owner, self.ledger, self.backend, {})
        self.assertEqual(other.state, State.FAILED_CLOSED)
        self.assertNotIn('release', self.backend.events)
        with self.assertRaises(Rejected): self.engine.close()

    def test_smoke_failure_fail_closed(self):
        self.backend.smoke_ok = False
        self.assertEqual(self.run_job()['status'], 'failed_closed')
        self.assertNotIn('resume', self.backend.events)

    def test_chat_context_release_failure_blocks_load_and_reload(self):
        self.backend.context_ok = False
        self.assertEqual(self.run_job()['status'], 'failed_closed')
        self.assertNotIn('load', self.backend.events)
        self.assertNotIn('reload', self.backend.events)

    def test_media_context_release_failure_blocks_chat_reload(self):
        self.backend.after_generate = lambda: setattr(self.backend, 'context_ok', False)
        self.assertEqual(self.run_job()['status'], 'failed_closed')
        self.assertIn('stop_media', self.backend.events)
        self.assertNotIn('reload', self.backend.events)

    def test_lost_fence_during_inference_prevents_restore_mutation(self):
        def lose_fence():
            token = self.owner.fence
            atomic_json(self.owner.path, dict(dataclasses.asdict(token), generation=token.generation+1))
        self.backend.after_generate = lose_fence
        self.assertEqual(self.run_job()['status'], 'failed_closed')
        self.assertNotIn('reload', self.backend.events)
        self.assertNotIn('stop_media', self.backend.events)

    def test_media_stop_failure_blocks_reload(self):
        self.backend.fail_stop = True
        self.assertEqual(self.run_job()['status'], 'failed_closed')
        self.assertNotIn('reload', self.backend.events)

    def test_partial_load_failure_still_stops_media(self):
        self.backend.fail_load = True
        self.assertEqual(self.run_job()['status'], 'failed')
        self.assertIn('stop_media', self.backend.events)
        self.assertEqual(self.engine.state, State.CHAT_READY)

    def test_oom_never_loads_schnell(self):
        self.backend.fail_generate = MemoryError
        self.assertEqual(self.run_job()['status'], 'failed')
        self.assertEqual(self.backend.loaded, [KLEIN.key])
        self.assertIn('stop_media', self.backend.events)

    def test_fallback_requires_permission_and_verified_snapshot(self):
        with self.assertRaises(Rejected): self.run_job(model=SCHNELL.key)
        with self.assertRaises(Rejected):
            self.engine.run(Job('s', 'synthetic', model=SCHNELL.key), Budget(3600, 1800),
                            fallback_permission=True, gated_access=True)
        self.assertNotIn('reserve', self.backend.events)

    def test_health_ready_list_poll_do_not_refresh_or_lease(self):
        before = list(self.backend.events)
        for _ in range(10):
            self.engine.health(); self.engine.ready(); self.engine.list_models(); self.engine.poll('unknown')
        self.assertEqual(before, self.backend.events)
        self.assertEqual(self.engine.health()['policy'], {'daily_runtime': 3600, 'max_uptime': 1800, 'idle': 600})

    def test_health_before_start_does_not_acquire_authority(self):
        owner = Ownership(self.base / 'unstarted')
        engine = Supervisor(owner, self.ledger, self.backend, {})
        engine.health(); engine.poll('x'); engine.list_models()
        self.assertIsNone(owner.fd)
        self.assertFalse(owner.path.exists())

    def test_poll_after_result_is_inert_and_copy_isolated(self):
        self.run_job()
        before = list(self.backend.events)
        self.engine.poll('synthetic-1')['outputs'].append('tamper')
        self.assertEqual(len(self.engine.poll('synthetic-1')['outputs']), 1)
        self.assertEqual(before, self.backend.events)

    def test_budget_reserves_reload_encode_stop(self):
        with self.assertRaises(Rejected):
            self.engine.run(Job('b', 'synthetic'), Budget(100, 100))
        self.assertNotIn('reserve', self.backend.events)

    def test_unknown_or_increased_runtime_refused(self):
        for budget in (Budget(float('nan'), 1800), Budget(3601, 1800), Budget(3600, 1801), Budget(-1, 1800)):
            with self.assertRaises(Rejected): self.engine.run(Job('b', 'synthetic'), budget)

    def test_reservation_denial_no_swap(self):
        self.backend.reserve_ok = False
        self.assertEqual(self.run_job()['status'], 'failed')
        self.assertNotIn('pause', self.backend.events)
        self.assertFalse(self.ledger.blocked)

    def test_heartbeat_refusal_prevents_load(self):
        self.backend.heartbeat_ok = False
        self.assertEqual(self.run_job()['status'], 'failed')
        self.assertNotIn('load', self.backend.events)

    def test_expired_lease_cannot_resurrect(self):
        lease = WorkLease(self.owner, 's', 1000, self.clock, self.backend.heartbeat)
        self.clock.advance(61)
        with self.assertRaises(Rejected): lease.progress()

    def test_absolute_deadline_not_extended_by_progress(self):
        lease = WorkLease(self.owner, 's', 150, self.clock, self.backend.heartbeat)
        self.clock.advance(20); lease.progress()
        self.clock.advance(20); lease.progress()
        self.clock.advance(10)
        with self.assertRaises(Rejected): lease.progress()
        self.assertEqual(lease.deadline, 150)

    def test_missing_heartbeat_during_inference_restores(self):
        self.backend.after_generate = lambda: self.clock.advance(61)
        self.assertEqual(self.run_job()['status'], 'failed')
        self.assertIn('reload', self.backend.events)

    def test_duplicate_job_settles_once_and_never_reexecutes(self):
        first = self.run_job()
        count = len(self.backend.generated)
        self.assertEqual(first, self.run_job())
        self.assertEqual(count, len(self.backend.generated))
        row = self.ledger.db.execute('SELECT settlement_count FROM jobs').fetchone()
        self.assertEqual(row[0], 1)

    def test_changed_payload_same_key_rejected(self):
        self.run_job()
        with self.assertRaises(Rejected): self.run_job(n=2)

    def test_unfinished_crash_reservation_never_retried(self):
        job = Job('synthetic-1', 'A synthetic geometry test')
        self.ledger.begin(job)
        with self.assertRaises(Rejected): self.run_job()
        self.assertNotIn('reserve', self.backend.events)

    def test_concurrent_image_admission_rejected(self):
        def contender():
            with self.assertRaises(Rejected): self.run_job()
        self.backend.after_generate = contender
        self.assertEqual(self.run_job()['status'], 'succeeded')

    def test_stale_fence_blocks_new_admission(self):
        token = self.owner.fence
        atomic_json(self.owner.path, dict(dataclasses.asdict(token), generation=token.generation+1))
        with self.assertRaises(OwnershipError): self.run_job()
        self.assertNotIn('reserve', self.backend.events)

    def test_poll_reads_durable_outcome_without_reexecution(self):
        expected = self.run_job()
        before = list(self.backend.events)
        other = Supervisor(self.owner, self.ledger, self.backend, {})
        self.assertEqual(other.poll('synthetic-1'), expected)
        self.assertEqual(self.backend.events, before)

    def test_initial_unknown_chat_readiness_fails_closed(self):
        self.engine.close()
        self.backend.chat_ready = lambda: None
        self.engine.start()
        self.assertEqual(self.engine.state, State.FAILED_CLOSED)
        with self.assertRaises(Rejected): self.run_job()

    def test_lease_identity_includes_all_fencing_fields(self):
        self.run_job()
        identity = self.backend.identities[-1]
        for key in ('generation', 'boot_id', 'pid', 'process_start', 'job_id', 'absolute_deadline', 'ttl'):
            self.assertIn(key, identity)


class OwnershipTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(); self.base = Path(self.tmp.name)
    def tearDown(self): self.tmp.cleanup()

    def test_os_contention_same_process(self):
        a = Ownership(self.base / 'authority'); b = Ownership(self.base / 'authority')
        try:
            a.acquire()
            with self.assertRaises(OwnershipError): b.acquire()
        finally: a.release(); b.release()

    def test_os_contention_independent_process(self):
        owner = Ownership(self.base / 'authority'); owner.acquire()
        code = '''from pathlib import Path
from image_engine.ownership import Ownership, OwnershipError
import sys
try:
    Ownership(Path(sys.argv[1])).acquire()
except OwnershipError:
    sys.exit(23)
sys.exit(0)
'''
        try:
            result = subprocess.run([sys.executable, '-c', code, str(owner.directory)], timeout=10)
            self.assertEqual(result.returncode, 23)
        finally: owner.release()

    def test_crashed_owner_new_generation_and_old_fence_rejected(self):
        directory = self.base / 'authority'
        code = '''from pathlib import Path
from image_engine.ownership import Ownership
import os,sys
Ownership(Path(sys.argv[1])).acquire()
os._exit(0)
'''
        subprocess.run([sys.executable, '-c', code, str(directory)], check=True, timeout=10)
        old = Fence(**json.loads((directory / 'owner.json').read_text()))
        owner = Ownership(directory)
        try:
            new = owner.acquire()
            self.assertEqual(new.generation, old.generation + 1)
            self.assertFalse(old.alive())
            with self.assertRaises(OwnershipError): validate_fence(owner.path, old)
        finally: owner.release()

    def test_stale_boot_pid_start_and_generation_rejected(self):
        owner = Ownership(self.base / 'authority'); token = owner.acquire()
        try:
            for field, value in [('boot_id', 'old-boot'), ('pid', 999999999),
                                 ('process_start', '0'), ('generation', token.generation-1)]:
                with self.assertRaises(OwnershipError):
                    owner.assert_current(dataclasses.replace(token, **{field: value}))
            fake = dataclasses.replace(token, process_start='0')
            atomic_json(owner.path, dataclasses.asdict(fake))
            with self.assertRaises(OwnershipError): validate_fence(owner.path, fake)
        finally: owner.release()

    def test_release_revokes_worker_fence(self):
        owner = Ownership(self.base / 'authority'); token = owner.acquire(); owner.release()
        with self.assertRaises(OwnershipError): validate_fence(owner.path, token)

    def test_corrupt_generation_cannot_reset(self):
        owner = Ownership(self.base / 'authority'); owner.path.write_text('{broken')
        with self.assertRaises(OwnershipError): owner.acquire()

    def test_replaced_lock_inode_detected(self):
        owner = Ownership(self.base / 'authority'); owner.acquire()
        try:
            owner.lock_path.rename(owner.directory / 'old.lock')
            owner.lock_path.touch()
            with self.assertRaises(OwnershipError): owner.assert_current()
        finally: owner.release()

    def test_process_start_identity_matches_kernel(self):
        self.assertTrue(Fence(1, boot_id(), os.getpid(), process_start(os.getpid())).alive())


class ModelTests(unittest.TestCase):
    def test_allowlist_exact_pin_only(self):
        self.assertEqual(select_model(KLEIN.key), KLEIN)
        for value in ('main', KLEIN.repo, KLEIN.repo+'@main', 'FLUX.2-dev', 'FLUX.2-klein-9B'):
            with self.assertRaises(Rejected): select_model(value)

    def test_explicit_fallback_and_gated_access_both_required(self):
        for permission, access in ((False, False), (True, False), (False, True)):
            with self.assertRaises(Rejected):
                select_model(SCHNELL.key, fallback_permission=permission, gated_access=access)
        self.assertEqual(select_model(SCHNELL.key, fallback_permission=True, gated_access=True), SCHNELL)

    def test_oom_fallback_refused_even_with_permission(self):
        with self.assertRaises(Rejected):
            select_model(SCHNELL.key, fallback_permission=True, gated_access=True, reason='Klein OOM')

    def test_size_and_count_validation(self):
        for size in ((512, 512), (1024.0, 1024), [1024, 1024]):
            with self.assertRaises(Rejected): Job('s', 'synthetic', size=size).validate(lambda *_: True)
        for n in (0, 5, True, 1.0, '4'):
            with self.assertRaises(Rejected): Job('s', 'synthetic', n=n).validate(lambda *_: True)
        for size in ((1024, 1024), (1344, 768), (768, 1344)):
            Job('s', 'synthetic', size=size, n=4).validate(lambda *_: True)

    def test_prompt_gate_cannot_be_bypassed(self):
        with self.assertRaises(Rejected): Job('s', 'synthetic').validate(lambda *_: False)

    def test_truthy_nonboolean_fallback_permission_rejected(self):
        with self.assertRaises(Rejected):
            select_model(SCHNELL.key, fallback_permission='yes', gated_access=True)

    def test_memory_error_reason_rejected_as_fallback(self):
        for reason in ('MemoryError', 'out_of_memory', 'CUDA OOM'):
            with self.assertRaises(Rejected):
                select_model(SCHNELL.key, fallback_permission=True, gated_access=True, reason=reason)

    def test_policy_constants_unchanged(self):
        self.assertEqual((DAILY_RUNTIME, MAX_UPTIME, IDLE), (3600, 1800, 600))


class SnapshotTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory(); self.base = Path(self.tmp.name)
        self.snap = snapshot_fixture(self.base)
    def tearDown(self): self.tmp.cleanup()

    def test_verified_snapshot(self): self.assertEqual(self.snap.verify(KLEIN), self.snap.root)
    def test_hash_tamper(self):
        (self.snap.root/'model_index.json').write_text('tamper')
        with self.assertRaises(Rejected): self.snap.verify(KLEIN)
    def test_extra_partial_file_rejected(self):
        (self.snap.root/'incomplete.partial').touch()
        with self.assertRaises(Rejected): self.snap.verify(KLEIN)
    def test_untrusted_manifest_rejected(self):
        with self.assertRaises(Rejected): dataclasses.replace(self.snap, manifest_sha256='0'*64).verify(KLEIN)
    def test_symlink_rejected(self):
        path = self.snap.root/'model_index.json'; path.unlink(); path.symlink_to('/dev/null')
        with self.assertRaises(Rejected): self.snap.verify(KLEIN)
    def test_wrong_revision_rejected(self):
        with self.assertRaises(Rejected): self.snap.verify(SCHNELL)
    def test_missing_shard_rejected(self):
        data = json.loads(self.snap.manifest.read_text())
        path = self.snap.root/'text_encoder/model.index.json'
        path.write_text(json.dumps({'weight_map': {'a': 'missing.safetensors'}}))
        data['files']['text_encoder/model.index.json'] = {'bytes': path.stat().st_size, 'sha256': sha256(path)}
        self.snap.manifest.write_text(json.dumps(data))
        snap = dataclasses.replace(self.snap, manifest_sha256=sha256(self.snap.manifest))
        with self.assertRaises(Rejected): snap.verify(KLEIN)


class WorkerTests(unittest.TestCase):
    def test_local_only_bf16_single_image_pipeline(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory); snap = snapshot_fixture(base)
            owner = Ownership(base/'authority'); token = owner.acquire()
            calls = []
            class Flux2KleinPipeline:
                def to(self, device): calls.append(('device', device)); return self
            class DiffusionPipeline:
                @staticmethod
                def from_pretrained(*args, **kwargs):
                    calls.append(('load', args, kwargs)); return Flux2KleinPipeline()
            try:
                torch = types.SimpleNamespace(bfloat16='BF16')
                diffusers = types.SimpleNamespace(DiffusionPipeline=DiffusionPipeline)
                with patch.dict(sys.modules, torch=torch, diffusers=diffusers):
                    reply = OfflineWorker().dispatch(dict(op='load', owner=dataclasses.asdict(token),
                        owner_file=str(owner.path), model=KLEIN.key, root=str(snap.root),
                        manifest=str(snap.manifest), manifest_sha256=snap.manifest_sha256))
                self.assertEqual(reply['loaded'], KLEIN.key)
                self.assertTrue(calls[0][2]['local_files_only'])
                self.assertFalse(calls[0][2]['trust_remote_code'])
                self.assertEqual(calls[0][2]['torch_dtype'], 'BF16')
                self.assertEqual(calls[0][1], (str(snap.root),))
            finally: owner.release()

    def test_stale_worker_fence_refuses_before_dependency_import(self):
        with tempfile.TemporaryDirectory() as directory:
            owner = Ownership(Path(directory)/'authority'); token = owner.acquire(); owner.release()
            with self.assertRaises(OwnershipError):
                OfflineWorker().dispatch(dict(op='load', owner=dataclasses.asdict(token), owner_file=str(owner.path)))


if __name__ == '__main__': unittest.main()
