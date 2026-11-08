"""CPU sandbox coverage; synthetic media, no real NVML/model/services."""
import importlib.util
import json
import os
from pathlib import Path
import tempfile
import threading
import types
import unittest
from unittest.mock import patch

from image_engine.adapters import SandboxChatAdapter, SandboxLifecycleAdapter, LocalStagingAdapter, APPROVED_KLEIN_PREFIX
from image_engine.integration import LocalBackend, NvmlContextProbe
from image_engine.ledger import Ledger
from image_engine.models import Job, KLEIN, SCHNELL, Rejected, sha256
from image_engine.ownership import Ownership
from image_engine.prompt import PinnedTokenizerPromptAdapter
from image_engine.supervisor import Supervisor, State
from test_engine import Clock, snapshot_fixture
from test_integration import FakeNvml

PACKAGE = Path(__file__).resolve().parents[1]
LOCAL_ROOT = Path('/workspace/tmp/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294')


class Lease:
    def __init__(self): self.calls = 0; self.valid = True
    def progress(self):
        if not self.valid: raise Rejected('synthetic expired lease')
        self.calls += 1


class ChatTests(unittest.TestCase):
    def setUp(self):
        self.clock = Clock(); self.chat = SandboxChatAdapter(clock=self.clock); self.lease = Lease()

    def test_pause_drain_stop_reload_resume(self):
        self.chat.set_work(active=1)
        self.chat.drain_step = lambda c, lease: c.set_work(active=0)
        self.chat.pause_chat_admission(self.lease)
        self.assertTrue(self.chat.chat_active())
        self.chat.wait_chat_progress(self.lease)
        self.chat.stop_chat(self.lease)
        self.assertFalse(self.chat.chat_ready())
        self.chat.reload_chat(150)
        self.assertTrue(self.chat.synthetic_smoke(150))
        self.chat.resume_chat_admission(150)
        self.chat.set_work(active=1)
        self.assertTrue(self.chat.chat_active())

    def test_no_forced_active_stop(self):
        self.chat.set_work(active=1); self.chat.pause_chat_admission(self.lease)
        with self.assertRaises(Rejected): self.chat.stop_chat(self.lease)
        with self.assertRaises(Rejected): self.chat.wait_chat_progress(self.lease)
        self.assertTrue(self.chat.chat_active())

    def test_pending_priority_and_pause_new_admission(self):
        self.chat.pause_chat_admission(self.lease)
        self.chat.set_work(pending=1)
        self.assertTrue(self.chat.chat_pending())
        with self.assertRaises(Rejected): self.chat.set_work(active=1)
        with self.assertRaises(Rejected): self.chat.stop_chat(self.lease)

    def test_stale_lease_no_pause(self):
        self.lease.valid = False
        with self.assertRaises(Rejected): self.chat.pause_chat_admission(self.lease)
        self.chat.set_work(active=1)

    def test_deadlines_fail_closed(self):
        for method in [self.chat.reload_chat, self.chat.resume_chat_admission, self.chat.synthetic_smoke]:
            for deadline in [100, float('nan'), float('inf'), None]:
                with self.assertRaises(Rejected): method(deadline)

    def test_smoke_failure_is_not_pass(self):
        self.chat.smoke_ok = False
        self.assertFalse(self.chat.synthetic_smoke(150))

    def test_stop_requires_pause(self):
        with self.assertRaises(Rejected): self.chat.stop_chat(self.lease)

    def test_invalid_counts_and_unbound(self):
        self.assertFalse(self.chat.live_bound)
        for n in [-1, True, 1.5]:
            with self.assertRaises(Rejected): self.chat.set_work(active=n)


class LifecycleTests(unittest.TestCase):
    def setUp(self):
        self.clock = Clock(); self.life = SandboxLifecycleAdapter(clock=self.clock)
        self.job = Job('cpu-job', 'synthetic')
        self.identity = dict(generation=1, boot_id='sandbox-boot', pid=1, process_start='123', job_id=self.job.job_id, absolute_deadline=200)
    def reserve(self): return self.life.reserve(self.job, self.identity, 100)
    def heartbeat(self, **changes):
        return self.life.heartbeat(dict(self.identity, absolute_deadline=170, ttl=60, **changes))

    def test_reserve_heartbeat_release_idempotent(self):
        self.assertTrue(self.reserve()); self.assertTrue(self.heartbeat())
        self.assertEqual(self.life.budget().available(), 1700)
        self.clock.advance(10)
        self.life.release(self.job.job_id, self.identity)
        self.life.release(self.job.job_id, self.identity)
        self.assertEqual(self.life.budget().available(), 1790)
        self.assertFalse(self.reserve())
        self.assertFalse(self.life.live_bound)

    def test_no_double_reservation(self):
        self.assertTrue(self.reserve()); self.assertFalse(self.reserve())

    def test_expired_heartbeat_cannot_resurrect(self):
        self.assertTrue(self.reserve()); self.clock.advance(60)
        self.assertFalse(self.heartbeat())
        self.assertEqual(self.life.budget().available(), 1700)

    def test_wrong_fence_heartbeat_release_rejected(self):
        self.assertTrue(self.reserve())
        for changes in [dict(generation=2), dict(pid=99), dict(boot_id='other'), dict(process_start='999'), dict(job_id='other')]:
            bad = dict(self.identity, ttl=60); bad.update(changes)
            self.assertFalse(self.life.heartbeat(bad))
            with self.assertRaises(Rejected): self.life.release(self.job.job_id, bad)

    def test_work_deadline_may_shrink_once_never_extend(self):
        self.assertTrue(self.reserve()); self.assertTrue(self.heartbeat())
        for deadline in [180, 201, 169, float('nan')]:
            self.assertFalse(self.life.heartbeat(dict(self.identity, absolute_deadline=deadline, ttl=60)))

    def test_ttl_invalid_or_unbounded(self):
        self.assertTrue(self.reserve())
        for ttl in [0, -1, 61, True, float('nan'), float('inf')]:
            self.assertFalse(self.life.heartbeat(dict(self.identity, ttl=ttl)))

    def test_cancel_stops_progress_and_keeps_cleanup_held(self):
        self.assertTrue(self.reserve()); self.life.cancel(self.job.job_id)
        self.assertTrue(self.life.cancelled(self.job.job_id)); self.assertFalse(self.heartbeat())
        self.life.release(self.job.job_id, self.identity)
        self.assertFalse(self.reserve())

    def test_budget_unknown_or_insufficient_rejected(self):
        for value in [float('nan'), float('inf'), -1, 3601]:
            with self.assertRaises(Rejected): SandboxLifecycleAdapter(daily_remaining=value)
        self.assertFalse(self.life.reserve(self.job, self.identity, 1801))
        for value in [True, 0, float('nan')]:
            self.assertFalse(self.life.reserve(self.job, self.identity, value))

    def test_alert_codes_only(self):
        self.life.alert('restore_failed_operator_required')
        self.assertEqual(self.life.alerts, ['restore_failed_operator_required'])
        with self.assertRaises(Rejected): self.life.alert('prompt content not allowed')

    def test_stale_generation_denied_after_cleanup(self):
        self.identity['generation'] = 2
        self.assertTrue(self.reserve()); self.life.release(self.job.job_id, self.identity)
        job = Job('next-cpu-job', 'synthetic')
        stale = dict(self.identity, generation=1, job_id=job.job_id)
        self.assertFalse(self.life.reserve(job, stale, 100))

    def test_overrun_release_fails_closed(self):
        self.assertTrue(self.reserve()); self.clock.advance(101)
        with self.assertRaises(Rejected): self.life.release(self.job.job_id, self.identity)
        self.assertEqual(self.life.budget().available(), 1700)

    def test_reservations_atomic_under_threads(self):
        results = []
        threads = [threading.Thread(target=lambda: results.append(self.reserve())) for _ in range(12)]
        for t in threads: t.start()
        for t in threads: t.join()
        self.assertEqual(results.count(True), 1)


class StagingTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(); self.addCleanup(self.temp.cleanup)
        self.snap = snapshot_fixture(Path(self.temp.name))
        self.adapter = LocalStagingAdapter(self.snap.root, self.snap.manifest, self.snap.manifest_sha256)
    def test_verified_local_tree(self):
        self.assertEqual(self.adapter.verified_snapshot().root, self.snap.root)
        self.assertEqual(self.adapter.staged_location(), APPROVED_KLEIN_PREFIX)
    def test_schnell_blocked_all_permission_combinations(self):
        for a in [False, True]:
            for b in [False, True]:
                with self.assertRaises(Rejected): self.adapter.verified_snapshot(SCHNELL.key, gated_access=a, fallback_permission=b)
        with self.assertRaises(Rejected): self.adapter.staged_location(SCHNELL.key)
    def test_tamper_detected(self):
        (self.snap.root/'model_index.json').write_text('tamper')
        with self.assertRaises(Rejected): self.adapter.verified_snapshot()
    def test_external_anchor_required(self):
        with self.assertRaises(Rejected): LocalStagingAdapter(self.snap.root, self.snap.root/'inside.json', '0'*64)
        with self.assertRaises(Rejected): LocalStagingAdapter('relative', self.snap.manifest, '0'*64)
        bad = LocalStagingAdapter(self.snap.root, self.snap.manifest, '0'*64)
        with self.assertRaises(Rejected): bad.verified_snapshot()


class PromptTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(); self.addCleanup(self.temp.cleanup)
        self.snap = snapshot_fixture(Path(self.temp.name))
    def adapter(self, **kwargs):
        return PinnedTokenizerPromptAdapter(self.snap.root, self.snap.manifest, self.snap.manifest_sha256, **kwargs)
    def test_unknown_limit_fails_closed(self):
        with self.assertRaises(Rejected): self.adapter()('synthetic', KLEIN.key)
    def test_raw_count_never_silently_used_for_admission(self):
        adapter = self.adapter(max_sequence_length=100, limit_source='synthetic test only')
        with self.assertRaises(Rejected): adapter('synthetic', KLEIN.key)
    def test_sourced_limit_formatter_required(self):
        for limit in [True, 0, -1, 100]:
            with self.assertRaises(Rejected): self.adapter(max_sequence_length=limit)
        with self.assertRaises(Rejected): self.adapter(formatter=lambda s:s)
    def test_no_truncation_and_formatter_accounted(self):
        # A counting stub proves control flow, not real tokenizer semantics.
        adapter = self.adapter(max_sequence_length=3, limit_source='synthetic', formatter=lambda s:'AB'+s, formatter_source='synthetic')
        fake = types.SimpleNamespace(encode=lambda text, **kw: types.SimpleNamespace(ids=list(text)))
        adapter._tokenizer = fake
        self.assertEqual(adapter.count('x'), 1)
        self.assertTrue(adapter('x', KLEIN.key))
        with self.assertRaises(Rejected): adapter('xx', KLEIN.key)
    def test_other_model_and_empty_prompt_rejected(self):
        adapter = self.adapter()
        for text, model in [('synthetic', SCHNELL.key), (' ', KLEIN.key), (None, KLEIN.key)]:
            with self.assertRaises(Rejected): adapter.count(text, model)
    def test_missing_library_fails_closed(self):
        with patch.dict('sys.modules', {'tokenizers':None}):
            with self.assertRaisesRegex(Rejected, 'runtime library unavailable'): self.adapter().count('synthetic')
    def test_tokenizer_integrity_before_import(self):
        (self.snap.root/'tokenizer/tokenizer.json').write_text('tamper')
        with self.assertRaisesRegex(Rejected, 'integrity'): self.adapter().count('synthetic')
    def test_real_pinned_local_tokenizer_cpu(self):
        if importlib.util.find_spec('tokenizers') is None:
            self.skipTest('tokenizers absent in this interpreter; use isolated CPU evidence environment')
        if not LOCAL_ROOT.is_dir(): self.skipTest('pinned local snapshot not on this machine')
        manifest = PACKAGE/'klein.snapshot.manifest.json'
        adapter = PinnedTokenizerPromptAdapter(LOCAL_ROOT, manifest, sha256(manifest))
        for prompt in ['A red cat on a blue chair.', 'İstanbul’da yağmurlu bir akşam, sulu boya.']:
            n = adapter.count(prompt)
            self.assertGreater(n, 0)
            self.assertEqual(n, adapter.count(prompt))
        self.assertIsNone(adapter._tokenizer.truncation)
        self.assertIsNone(adapter._tokenizer.padding)
        with self.assertRaises(Rejected): adapter('synthetic', KLEIN.key)


class LockTests(unittest.TestCase):
    def setUp(self): self.lock = json.loads((PACKAGE/'dependencies.lock.json').read_text())
    def test_sourced_exact_pins(self):
        pins = {p['name']:p for p in self.lock['packages']}
        self.assertEqual(pins['transformers']['version_or_commit'], '4.56.1')
        self.assertEqual(pins['diffusers']['version_or_commit'], '0.37.0')
        for name in ['transformers', 'diffusers', 'tokenizers']:
            self.assertEqual(pins[name]['status'], 'PINNED sourced')
            self.assertTrue(pins[name]['source_of_pin'])
            for wheel in pins[name]['wheels']:
                self.assertRegex(wheel['sha256'], '^[a-f0-9]{64}$')
                self.assertTrue(wheel['filename'].endswith('.whl'))
                self.assertTrue(wheel['url'].startswith('https://files.pythonhosted.org/'))
    def test_no_floating_git_or_unknown_promoted_to_ready(self):
        self.assertFalse(self.lock['production_installable'])
        self.assertFalse(self.lock['floating_git_allowed'])
        self.assertIsNone(self.lock['prompt_limit']['value'])
        pins = {p['name']:p for p in self.lock['packages']}
        for name in ['torch', 'torchvision', 'CUDA', 'pynvml']:
            self.assertEqual(pins[name]['status'], 'UNKNOWN')
            self.assertIsNone(pins[name]['version_or_commit'])
        self.assertEqual(self.lock['schnell']['status'], 'BLOCKED')
    def test_manifest_matches_reviewed_upload_evidence(self):
        manifest = json.loads((PACKAGE/'klein.snapshot.manifest.json').read_text())
        source = Path(manifest['provenance']['source'])
        self.assertEqual(sha256(source), manifest['provenance']['sha256'])
        expected = {e['path']:{'bytes':e['bytes'],'sha256':e['sha256']} for e in json.loads(source.read_text())}
        self.assertEqual(manifest['files'], expected)
        self.assertEqual(len(expected), 24)
        self.assertEqual(sum(e['bytes'] for e in expected.values()), 15980183023)
        self.assertEqual(expected['tokenizer/tokenizer.json']['sha256'], 'aeb13307a71acd8fe81861d94ad54ab689df773318809eed3cbe794b4492dae4')


class SandboxWiringTests(unittest.TestCase):
    def test_new_adapters_through_existing_local_backend_supervisor(self):
        clock = Clock()
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory); snap = snapshot_fixture(base)
            owner = Ownership(base/'authority'); ledger = Ledger(base/'ledger.sqlite')
            chat = SandboxChatAdapter(clock=clock); life = SandboxLifecycleAdapter(clock=clock)
            class Media:
                physical_uuid = 'GPU-synthetic'
                def start(self): pass
                def request(self, request, lease):
                    lease.progress()
                    return {'output':'synthetic-output-only'}
                def stop(self, deadline): pass
            backend = LocalBackend(chat=chat, lifecycle=life, probe=NvmlContextProbe('GPU-synthetic', FakeNvml()), media=Media(), prompt_validator=lambda p,m:True, output_directory=base, max_prompt_tokens=100)
            supervisor = Supervisor(owner, ledger, backend, {KLEIN.key:snap}, clock=clock)
            supervisor.start()
            try:
                result = supervisor.run(Job('sandbox-wiring', 'synthetic'), life.budget())
                self.assertEqual(result['status'], 'succeeded')
                self.assertEqual(supervisor.state, State.CHAT_READY)
                self.assertIsNone(life._record)
                self.assertTrue(chat.chat_ready())
            finally:
                supervisor.close()
                ledger.db.close()
