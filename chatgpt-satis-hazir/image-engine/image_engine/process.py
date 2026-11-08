"""Bounded JSON-line transport to an owned, separate-venv media process.

This module only launches an explicitly configured LOCAL Python worker; it has
no VM/start/SSH/systemd API. Not invoked by importing the package or by probes.
"""
from __future__ import annotations
import json
import os
import selectors
import signal
import subprocess
import time
from pathlib import Path
from .models import Rejected
from .ownership import Fence, boot_id, process_start


class MediaProcess:
    def __init__(self, interpreter: Path, *, physical_uuid: str = '', clock=time.monotonic):
        self.interpreter = Path(interpreter)
        if physical_uuid and not physical_uuid.startswith('GPU-'):
            raise Rejected('explicit physical GPU UUID required')
        self.physical_uuid = physical_uuid
        self.clock = clock
        self.child = None
        self.identity = None
        self.buffer = bytearray()
        self.sequence = 0

    def start(self):
        if self.child is not None:
            raise Rejected('media process already exists')
        # Explicit separate venv required. No ambient credentials are inherited.
        if not self.interpreter.is_absolute() or not (self.interpreter.parent.parent / 'pyvenv.cfg').is_file():
            raise Rejected('explicit separate media venv required')
        env = {'PATH': '/usr/bin:/bin', 'PYTHONUNBUFFERED': '1',
               'PYTHONPATH': str(Path(__file__).resolve().parent.parent),
               'HF_HUB_OFFLINE': '1', 'TRANSFORMERS_OFFLINE': '1',
               'HF_HUB_DISABLE_TELEMETRY': '1', 'DO_NOT_TRACK': '1',
               'CUDA_VISIBLE_DEVICES': self.physical_uuid}
        self.child = subprocess.Popen([str(self.interpreter), '-m', 'image_engine.worker'],
                                      stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                                      stderr=subprocess.DEVNULL, start_new_session=True,
                                      close_fds=True, env=env)
        try:
            self.identity = Fence(0, boot_id(), self.child.pid, process_start(self.child.pid))
            os.set_blocking(self.child.stdout.fileno(), False)
        except BaseException:
            # Popen handle is still the only owned child; never target an arbitrary PID.
            self.child.terminate()
            self.child.wait(timeout=5)
            raise

    def request(self, payload, lease):
        if self.child is None or self.child.poll() is not None or not self.identity.alive():
            raise Rejected('owned media worker not alive')
        lease.progress()
        self.sequence += 1
        wire = dict(payload, sequence=self.sequence, owner=lease.token.__dict__,
                    owner_file=str(lease.owner.path), job_id=lease.job_id)
        self.child.stdin.write((json.dumps(wire) + '\n').encode())
        self.child.stdin.flush()
        with selectors.DefaultSelector() as selector:
            selector.register(self.child.stdout, selectors.EVENT_READ)
            while True:
                lease.check()
                for _, _ in selector.select(timeout=min(1, max(0, lease.deadline - self.clock()))):
                    block = os.read(self.child.stdout.fileno(), 65536)
                    if not block:
                        raise Rejected('media worker exited')
                    self.buffer.extend(block)
                    if len(self.buffer) > 1024 * 1024:
                        raise Rejected('oversized worker response')
                    if b'\n' in self.buffer:
                        line, _, remaining = self.buffer.partition(b'\n')
                        self.buffer = bytearray(remaining)
                        reply = json.loads(line)
                        if reply.get('sequence') != self.sequence:
                            raise Rejected('stale worker response')
                        if reply.get('error'):
                            if reply['error'] == 'OutOfMemoryError':
                                raise MemoryError('media OOM; no automatic fallback')
                            raise Rejected('media worker failed')
                        lease.progress()
                        return reply
                # Only an admitted in-flight operation can renew; absolute cap is fixed.
                lease.progress()

    def stop(self, deadline):
        """Graceful TERM then bounded KILL, only the original owned process group.

        NVML context verification is separately mandatory before chat reload.
        """
        if self.child is None:
            return
        if self.child.poll() is None:
            if (self.identity is None or not self.identity.alive()
                    or os.getpgid(self.child.pid) != self.child.pid):
                raise Rejected('owned worker identity changed; refusing process signal')
            os.killpg(self.child.pid, signal.SIGTERM)
            try:
                self.child.wait(timeout=max(0, min(5, deadline - self.clock())))
            except subprocess.TimeoutExpired:
                if not self.identity.alive() or os.getpgid(self.child.pid) != self.child.pid:
                    raise Rejected('worker identity changed before kill')
                os.killpg(self.child.pid, signal.SIGKILL)
                self.child.wait(timeout=max(0.01, deadline - self.clock()))
        for stream in (self.child.stdin, self.child.stdout):
            if stream:
                stream.close()
        self.child = None
        self.identity = None
