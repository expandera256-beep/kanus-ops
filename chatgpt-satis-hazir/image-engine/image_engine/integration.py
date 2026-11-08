"""Explicit local integration glue, inert until called by an approved host.

No systemd edits or service discovery. Caller supplies reviewed chat/lifecycle
adapters; these must share this supervisor's authority rather than independently
starting CUDA workers. This file is preparation, not a live installation.
"""
from __future__ import annotations
from pathlib import Path
from .models import KLEIN, SCHNELL, Rejected
from .process import MediaProcess


class NvmlContextProbe:
    """Unknown, unavailable, unsupported, permission-denied => NOT empty.

    Uses a pinned physical UUID, never assumes logical cuda:0 is the right GPU.
    Real NVML calls were NOT exercised on the box. Injected provider enables CPU
    testing without importing or initializing NVML.
    """
    def __init__(self, physical_uuid: str, provider=None):
        if not physical_uuid.startswith('GPU-'):
            raise Rejected('explicit physical GPU UUID required')
        self.uuid, self.provider = physical_uuid, provider

    def empty(self):
        api = self.provider
        try:
            if api is None:
                import pynvml
                api = pynvml
            api.nvmlInit()
            handle = api.nvmlDeviceGetHandleByUUID(self.uuid)
            actual = api.nvmlDeviceGetUUID(handle)
            if isinstance(actual, bytes):
                actual = actual.decode('ascii')
            if actual != self.uuid:
                return False
            return (len(api.nvmlDeviceGetComputeRunningProcesses(handle)) == 0
                    and len(api.nvmlDeviceGetGraphicsRunningProcesses(handle)) == 0)
        except Exception:
            return False
        finally:
            if api is not None:
                try:
                    api.nvmlShutdown()
                except Exception:
                    pass


class LocalBackend:
    """Concrete backend binding the media subprocess to reviewed local adapters.

    chat implements chat_pending/chat_active/pause_chat_admission/etc;
    lifecycle implements reserve/heartbeat/release/cancelled/alert. Constructor
    does not invoke any adapter, read credentials, start processes, or load CUDA.
    Production binding and dependency wheel hashes are deployment blockers.
    """
    def __init__(self, *, chat, lifecycle, probe: NvmlContextProbe,
                 media: MediaProcess, prompt_validator, output_directory: Path,
                 max_prompt_tokens: int, fallback_permission=False, gated_access=False):
        if type(max_prompt_tokens) is not int or max_prompt_tokens < 1:
            raise Rejected('pinned-config prompt token limit required')
        output = Path(output_directory)
        if (not output.is_absolute() or not output.is_dir() or output.is_symlink()
                or output.stat().st_mode & 0o077):
            raise Rejected('existing private local output directory required')
        if media.physical_uuid != probe.uuid:
            raise Rejected('worker/probe physical GPU UUID mismatch')
        self.chat, self.lifecycle, self.probe, self.media = chat, lifecycle, probe, media
        self.prompt_validator, self.output_directory = prompt_validator, output
        self.max_prompt_tokens = max_prompt_tokens
        self.fallback_permission, self.gated_access = fallback_permission, gated_access

    def prompt_valid(self, prompt, model): return self.prompt_validator(prompt, model)
    def reserve(self, job, identity, seconds): return self.lifecycle.reserve(job, identity, seconds)
    def heartbeat(self, identity): return self.lifecycle.heartbeat(identity)
    def release(self, job_id, identity): return self.lifecycle.release(job_id, identity)
    def cancelled(self, job_id): return self.lifecycle.cancelled(job_id)
    def alert(self, code): return self.lifecycle.alert(code)
    def chat_pending(self): return self.chat.chat_pending()
    def chat_active(self): return self.chat.chat_active()
    def pause_chat_admission(self, lease): return self.chat.pause_chat_admission(lease)
    def resume_chat_admission(self, deadline): return self.chat.resume_chat_admission(deadline)
    def wait_chat_progress(self, lease): return self.chat.wait_chat_progress(lease)
    def stop_chat(self, lease): return self.chat.stop_chat(lease)
    def cuda_empty(self): return self.probe.empty()
    def reload_chat(self, deadline): return self.chat.reload_chat(deadline)
    def chat_ready(self): return self.chat.chat_ready()
    def synthetic_smoke(self, deadline): return self.chat.synthetic_smoke(deadline)

    def load_media(self, model, snapshot, lease):
        lease.progress()
        snapshot.verify(model)
        self.media.start()
        self.media.request(dict(op='load', model=model.key, root=str(snapshot.root),
                                manifest=str(snapshot.manifest), manifest_sha256=snapshot.manifest_sha256,
                                fallback_permission=self.fallback_permission,
                                gated_access=self.gated_access), lease)

    def generate_one(self, job, index, model, lease):
        # An output path is returned only after local PNG encode+fsync completes.
        response = self.media.request(dict(op='generate', model=model.key, prompt=job.prompt,
            size=list(job.size), seed=job.seed + index, max_prompt_tokens=self.max_prompt_tokens,
            output_directory=str(self.output_directory)), lease)
        return response['output']

    def stop_media(self, deadline): return self.media.stop(deadline)
