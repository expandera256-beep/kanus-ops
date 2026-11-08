"""Optional offline Diffusers worker entry point. NOT EXECUTED during CPU tests.

Each process owns exactly one pipeline. There is no download, fallback, process
restart, environment credential discovery, VM wake, or model-name resolution.
"""
from __future__ import annotations
import contextlib
import json
import os
import sys
from pathlib import Path
from .models import Job, KLEIN, SCHNELL, Snapshot, Rejected, select_model
from .ownership import Fence, validate_fence


class OfflineWorker:
    def __init__(self):
        self.pipeline = None
        self.model = None
        self.owner = None
        self.snapshot = None
        self.job_id = None

    def dispatch(self, request):
        owner = Fence(**request['owner'])
        validate_fence(Path(request['owner_file']), owner)
        if self.owner is not None and owner != self.owner:
            raise Rejected('worker cannot cross supervisor generations')
        self.owner = owner
        if request['op'] == 'load':
            if self.pipeline is not None:
                raise Rejected('one pipeline per process')
            model = select_model(request['model'],
                                 fallback_permission=request.get('fallback_permission') is True,
                                 gated_access=request.get('gated_access') is True,
                                 reason=request.get('fallback_reason'))
            snapshot = Snapshot(Path(request['root']), Path(request['manifest']), request['manifest_sha256'])
            root = snapshot.verify(model)
            # Offline environment must be set before these optional imports.
            import torch
            from diffusers import DiffusionPipeline
            pipeline = DiffusionPipeline.from_pretrained(
                str(root), local_files_only=True, torch_dtype=torch.bfloat16,
                use_safetensors=True, trust_remote_code=False)
            expected = 'Flux2KleinPipeline' if model == KLEIN else 'FluxPipeline'
            if type(pipeline).__name__ != expected:
                raise Rejected('unexpected local pipeline class')
            validate_fence(Path(request['owner_file']), owner)
            self.pipeline = pipeline.to('cuda')
            self.model, self.snapshot = model, root
            return {'loaded': model.key}
        if request['op'] != 'generate' or self.pipeline is None:
            raise Rejected('worker is not loaded')
        if request['model'] != self.model.key:
            raise Rejected('generation model mismatch')
        # Protocol never accepts batch count: exactly one image per request.
        job = Job(request['job_id'], request['prompt'], tuple(request['size']), 1,
                  self.model.key, request['seed'])
        token_limit = request['max_prompt_tokens']
        if type(token_limit) is not int or token_limit < 1:
            raise Rejected('reviewed prompt token limit required')
        def prompt_valid(prompt, model):
            tokenizers = [self.pipeline.tokenizer]
            if getattr(self.pipeline, 'tokenizer_2', None) is not None:
                tokenizers.append(self.pipeline.tokenizer_2)
            return all(len(t(prompt, truncation=False)['input_ids']) <= token_limit for t in tokenizers)
        job.validate(prompt_valid)
        output_root = Path(request['output_directory'])
        if (not output_root.is_absolute() or output_root.is_symlink() or not output_root.is_dir()
                or output_root.stat().st_uid != os.getuid() or output_root.stat().st_mode & 0o077):
            raise Rejected('reviewed private local output directory required')
        import torch
        validate_fence(Path(request['owner_file']), owner)
        image = self.pipeline(prompt=job.prompt, width=job.size[0], height=job.size[1],
                              num_images_per_prompt=1, num_inference_steps=4,
                              guidance_scale=self.model.guidance,
                              generator=torch.Generator(device='cuda').manual_seed(job.seed)).images[0]
        validate_fence(Path(request['owner_file']), owner)
        # Fresh private file; no user-controlled filename, overwrite, remote URL or upload.
        import tempfile
        fd, path = tempfile.mkstemp(prefix='image-', suffix='.png', dir=output_root)
        with os.fdopen(fd, 'wb') as output:
            image.save(output, format='PNG')
            output.flush()
            os.fsync(output.fileno())
        return {'output': path, 'seed': job.seed}


def main():
    for key in ('HF_HUB_OFFLINE', 'TRANSFORMERS_OFFLINE'):
        os.environ[key] = '1'
    worker = OfflineWorker()
    for line in sys.stdin:
        sequence = None
        try:
            if len(line) > 65536:
                raise Rejected('oversized request')
            request = json.loads(line)
            sequence = request['sequence']
            # Keep stdout protocol-only; dependencies can write ordinary chatter.
            with contextlib.redirect_stdout(sys.stderr):
                reply = worker.dispatch(request)
            reply['sequence'] = sequence
        except Exception as exc:
            reply = {'sequence': sequence, 'error': type(exc).__name__}
        print(json.dumps(reply), flush=True)


if __name__ == '__main__':
    main()
