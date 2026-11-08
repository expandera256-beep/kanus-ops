"""Integration boundary CPU tests. No CUDA dependencies or GPU load."""
import dataclasses
import json
from pathlib import Path
import tempfile
import time
import types
import unittest
import venv
from unittest.mock import patch
from image_engine.integration import LocalBackend, NvmlContextProbe
from image_engine.models import Job, KLEIN, Rejected
from image_engine.ownership import Ownership, OwnershipError, atomic_json
from image_engine.process import MediaProcess
from image_engine.supervisor import WorkLease
from image_engine.worker import OfflineWorker
from test_engine import Clock, FakeBackend, snapshot_fixture


class FakeNvml:
    def __init__(self):
        self.compute = []; self.graphics = []; self.uuid = 'GPU-synthetic'; self.fail = False
    def nvmlInit(self):
        if self.fail: raise PermissionError('synthetic permission denial')
    def nvmlShutdown(self): pass
    def nvmlDeviceGetHandleByUUID(self, uuid): return 'synthetic-device'
    def nvmlDeviceGetUUID(self, handle): return self.uuid
    def nvmlDeviceGetComputeRunningProcesses(self, handle): return self.compute
    def nvmlDeviceGetGraphicsRunningProcesses(self, handle): return self.graphics


class ProbeTests(unittest.TestCase):
    def test_empty_compute_and_graphics_required(self):
        api = FakeNvml(); probe = NvmlContextProbe(api.uuid, api)
        self.assertTrue(probe.empty())
        api.compute = [123]
        self.assertFalse(probe.empty())
        api.compute = []; api.graphics = [456]
        self.assertFalse(probe.empty())

    def test_unknown_nvml_is_not_empty(self):
        api = FakeNvml(); api.fail = True
        self.assertFalse(NvmlContextProbe(api.uuid, api).empty())

    def test_wrong_physical_device_is_not_empty(self):
        api = FakeNvml()
        self.assertFalse(NvmlContextProbe('GPU-other', api).empty())

    def test_ordinal_not_accepted_as_uuid(self):
        with self.assertRaises(Rejected): NvmlContextProbe('0', FakeNvml())


class ProcessTests(unittest.TestCase):
    def test_requires_explicit_separate_venv(self):
        process = MediaProcess(Path('/usr/bin/python3'))
        with self.assertRaises(Rejected): process.start()
        self.assertIsNone(process.child)

    def test_local_cpu_worker_transport_and_owned_stop(self):
        # Actual LOCAL subprocess with CUDA_VISIBLE_DEVICES empty, no load op.
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            venv.EnvBuilder(with_pip=False).create(base/'venv')
            owner = Ownership(base/'authority'); owner.acquire()
            process = MediaProcess(base/'venv/bin/python3')
            lease = WorkLease(owner, 'synthetic', time.monotonic()+20,
                              time.monotonic, lambda identity: True)
            try:
                process.start()
                self.assertTrue(process.identity.alive())
                with self.assertRaises(Rejected):
                    process.request({'op': 'unsupported-cpu-test'}, lease)
                process.stop(time.monotonic()+10)
                self.assertIsNone(process.child)
            finally:
                process.stop(time.monotonic()+10)
                owner.release()

    def test_refuses_signal_when_process_identity_stale(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory)
            venv.EnvBuilder(with_pip=False).create(base/'venv')
            process = MediaProcess(base/'venv/bin/python3')
            process.start()
            original = process.identity
            process.identity = dataclasses.replace(original, process_start='0')
            try:
                with self.assertRaises(Rejected): process.stop(time.monotonic()+10)
            finally:
                process.identity = original
                process.stop(time.monotonic()+10)

    def test_releasing_stale_owner_never_decreases_generation(self):
        with tempfile.TemporaryDirectory() as directory:
            owner = Ownership(Path(directory)/'authority'); token = owner.acquire()
            value = dict(dataclasses.asdict(token), generation=token.generation+9)
            atomic_json(owner.path, value)
            owner.release()
            self.assertEqual(json.loads(owner.path.read_text())['generation'], token.generation+9)


class ConcreteBackendTests(unittest.TestCase):
    def test_constructor_is_inert_and_uuid_bound(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            clock = Clock(); fake = FakeBackend(clock); probe = NvmlContextProbe('GPU-synthetic', FakeNvml())
            process = MediaProcess(output/'venv/bin/python3', physical_uuid='GPU-synthetic')
            backend = LocalBackend(chat=fake, lifecycle=fake, probe=probe, media=process,
                prompt_validator=fake.prompt_valid, output_directory=output, max_prompt_tokens=100)
            self.assertEqual(fake.events, [])
            self.assertIsNone(process.child)
            self.assertTrue(backend.prompt_valid('synthetic', KLEIN.key))
            with self.assertRaises(Rejected):
                LocalBackend(chat=fake, lifecycle=fake, probe=probe,
                    media=MediaProcess(output/'python', physical_uuid='GPU-other'),
                    prompt_validator=fake.prompt_valid, output_directory=output, max_prompt_tokens=100)

    def test_worker_calls_single_image_four_steps_and_correct_guidance(self):
        with tempfile.TemporaryDirectory() as directory:
            base = Path(directory); owner = Ownership(base/'authority'); token = owner.acquire()
            worker = OfflineWorker(); worker.owner = token; worker.model = KLEIN
            calls = []
            class Image:
                def save(self, output, format): output.write(b'synthetic-png-not-a-real-image')
            class Pipeline:
                def tokenizer(self, prompt, truncation): return {'input_ids': [1,2,3]}
                def __call__(self, **kwargs):
                    calls.append(kwargs); return types.SimpleNamespace(images=[Image()])
            class Generator:
                def __init__(self, device): pass
                def manual_seed(self, seed): return self
            worker.pipeline = Pipeline()
            request = dict(op='generate', owner=dataclasses.asdict(token), owner_file=str(owner.path),
                job_id='synthetic', prompt='synthetic', size=[1024,1024], seed=11,
                model=KLEIN.key, max_prompt_tokens=10, output_directory=str(base))
            try:
                with patch.dict('sys.modules', torch=types.SimpleNamespace(Generator=Generator)):
                    reply = worker.dispatch(request)
                self.assertTrue(Path(reply['output']).is_file())
                self.assertEqual(calls[0]['num_images_per_prompt'], 1)
                self.assertEqual(calls[0]['num_inference_steps'], 4)
                self.assertEqual(calls[0]['guidance_scale'], 1.0)
            finally: owner.release()
