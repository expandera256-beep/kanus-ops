"""Pinned models, input admission, and offline snapshot integrity validation."""
from __future__ import annotations
import hashlib
import json
from dataclasses import dataclass
from pathlib import Path, PurePosixPath


class Rejected(RuntimeError):
    pass


@dataclass(frozen=True)
class Model:
    repo: str
    revision: str
    guidance: float
    @property
    def key(self):
        return f'{self.repo}@{self.revision}'


KLEIN = Model('black-forest-labs/FLUX.2-klein-4B', 'e7b7dc27f91deacad38e78976d1f2b499d76a294', 1.0)
SCHNELL = Model('black-forest-labs/FLUX.1-schnell', '741f7c3ce8b383c54771c7003378a50191e9efe9', 0.0)
REQUIRED_FILES = {
    KLEIN.key: frozenset({
        'model_index.json', 'scheduler/scheduler_config.json',
        'transformer/config.json', 'vae/config.json', 'text_encoder/config.json',
        'text_encoder/model.safetensors.index.json',
        'text_encoder/model-00001-of-00002.safetensors',
        'text_encoder/model-00002-of-00002.safetensors',
        'transformer/diffusion_pytorch_model.safetensors',
        'vae/diffusion_pytorch_model.safetensors',
        'tokenizer/tokenizer_config.json', 'tokenizer/tokenizer.json',
        'tokenizer/merges.txt', 'tokenizer/vocab.json',
        'tokenizer/added_tokens.json', 'tokenizer/special_tokens_map.json',
        'tokenizer/chat_template.jinja', 'LICENSE.md',
    }),
    SCHNELL.key: frozenset({
        'model_index.json', 'scheduler/scheduler_config.json',
        'transformer/config.json', 'vae/config.json',
        'text_encoder/config.json', 'text_encoder_2/config.json',
        'text_encoder/model.safetensors',
        'text_encoder_2/model-00001-of-00002.safetensors',
        'text_encoder_2/model-00002-of-00002.safetensors',
        'text_encoder_2/model.safetensors.index.json',
        'transformer/diffusion_pytorch_model-00001-of-00003.safetensors',
        'transformer/diffusion_pytorch_model-00002-of-00003.safetensors',
        'transformer/diffusion_pytorch_model-00003-of-00003.safetensors',
        'transformer/diffusion_pytorch_model.safetensors.index.json',
        'vae/diffusion_pytorch_model.safetensors',
        'tokenizer/tokenizer_config.json', 'tokenizer/merges.txt',
        'tokenizer/vocab.json', 'tokenizer/special_tokens_map.json',
        'tokenizer_2/tokenizer_config.json', 'tokenizer_2/tokenizer.json',
        'tokenizer_2/spiece.model', 'tokenizer_2/special_tokens_map.json',
    }),
}
SIZES = frozenset({(1024, 1024), (1344, 768), (768, 1344)})
DAILY_RUNTIME = 3600
MAX_UPTIME = 1800
IDLE = 600


def select_model(key: str, *, fallback_permission=False, gated_access=False,
                 reason: str | None = None) -> Model:
    if key == KLEIN.key:
        return KLEIN
    if key == SCHNELL.key:
        if reason is not None and (not isinstance(reason, str)
                                   or any(marker in reason.lower() for marker in ('oom', 'memory'))):
            raise Rejected('schnell is not an OOM/memory fallback')
        if fallback_permission is not True or gated_access is not True:
            raise Rejected('schnell requires explicit fallback permission and authorized gated access')
        return SCHNELL
    raise Rejected('model/revision not allowlisted')


@dataclass(frozen=True)
class Job:
    job_id: str
    prompt: str
    size: tuple[int, int] = (1024, 1024)
    n: int = 1
    model: str = KLEIN.key
    seed: int = 0

    def validate(self, prompt_validator) -> None:
        if not isinstance(self.job_id, str) or not 1 <= len(self.job_id) <= 128:
            raise Rejected('invalid idempotency key')
        if (not isinstance(self.size, tuple) or len(self.size) != 2
                or any(type(i) is not int for i in self.size) or self.size not in SIZES):
            raise Rejected('size not allowlisted')
        if type(self.n) is not int or not 1 <= self.n <= 4:
            raise Rejected('n must be integer 1..4')
        if type(self.seed) is not int or not 0 <= self.seed <= 2**63 - 5:
            raise Rejected('invalid seed')
        if not isinstance(self.prompt, str) or not self.prompt.strip() or len(self.prompt) > 16384:
            raise Rejected('invalid prompt')
        # Production adapter must count with the pinned tokenizer, not characters.
        if prompt_validator(self.prompt, self.model) is not True:
            raise Rejected('pinned-tokenizer prompt validation failed')

    def fingerprint(self):
        return hashlib.sha256(json.dumps(self.__dict__, sort_keys=True).encode()).hexdigest()


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open('rb') as src:
        for block in iter(lambda: src.read(4 * 1024 * 1024), b''):
            h.update(block)
    return h.hexdigest()


@dataclass(frozen=True)
class Snapshot:
    root: Path
    manifest: Path
    manifest_sha256: str  # Trust anchor must come from reviewed configuration, not the download.

    def verify(self, model: Model) -> Path:
        if self.root.is_symlink() or not self.root.is_dir():
            raise Rejected('missing/unsafe local snapshot')
        if sha256(self.manifest) != self.manifest_sha256:
            raise Rejected('untrusted manifest')
        data = json.loads(self.manifest.read_text())
        if data.get('model') != model.repo or data.get('revision') != model.revision:
            raise Rejected('snapshot identity mismatch')
        entries = data.get('files', {})
        required = REQUIRED_FILES[model.key]
        if not required.issubset(entries):
            raise Rejected('incomplete manifest')
        root = self.root.resolve(strict=True)
        for rel, expected in entries.items():
            relative = PurePosixPath(rel)
            if relative.is_absolute() or '..' in relative.parts or str(relative) != rel:
                raise Rejected('unsafe manifest path')
            candidate = root / rel
            if any(part.is_symlink() for part in [candidate, *candidate.parents] if part != root.parent):
                raise Rejected('snapshot symlinks prohibited')
            if not candidate.is_file() or not candidate.resolve().is_relative_to(root):
                raise Rejected('missing snapshot file')
            if candidate.stat().st_size != expected['bytes'] or sha256(candidate) != expected['sha256']:
                raise Rejected(f'snapshot integrity failure: {rel}')
        # Prevent implicit loading of unverified files and unfinished download markers.
        actual = {str(p.relative_to(root)) for p in root.rglob('*') if p.is_file()}
        if actual != set(entries):
            raise Rejected('unverified extra files or incomplete snapshot')
        for index in root.rglob('*.index.json'):
            for shard in json.loads(index.read_text()).get('weight_map', {}).values():
                if str(index.parent.relative_to(root) / shard) not in entries:
                    raise Rejected('unverified indexed shard')
        return root
