"""Pinned local token counting. No model import, network, or silent truncation.

Raw counts are diagnostic, NOT pipeline admission counts. The host must inject a
reviewed formatter matching the pinned pipeline's chat template and a sourced
limit. The snapshot contains no max_sequence_length; defaults stay UNKNOWN and
admission fails closed. max_position_embeddings/model_max_length are NOT caps.
"""
from __future__ import annotations
import json
from pathlib import Path
from .models import KLEIN, Rejected, sha256


class PinnedTokenizerPromptAdapter:
    def __init__(self, root, manifest, manifest_sha256, *, max_sequence_length=None,
                 limit_source=None, formatter=None, formatter_source=None):
        self.root, self.manifest = Path(root), Path(manifest)
        self.manifest_sha256 = manifest_sha256
        if max_sequence_length is not None and (
                type(max_sequence_length) is not int or max_sequence_length < 1 or not limit_source):
            raise Rejected('positive sourced prompt limit required')
        if formatter is not None and (not callable(formatter) or not formatter_source):
            raise Rejected('reviewed pipeline formatter source required')
        self.max_sequence_length, self.limit_source = max_sequence_length, limit_source
        self.formatter, self.formatter_source = formatter, formatter_source
        self._tokenizer = None

    def _load(self):
        if self._tokenizer is not None:
            return self._tokenizer
        if self.root.is_symlink() or not self.root.is_dir() or sha256(self.manifest) != self.manifest_sha256:
            raise Rejected('untrusted pinned tokenizer manifest')
        data = json.loads(self.manifest.read_text())
        if data.get('model') != KLEIN.repo or data.get('revision') != KLEIN.revision:
            raise Rejected('tokenizer model/revision mismatch')
        entries = data.get('files', {})
        required = {'tokenizer/tokenizer.json', 'tokenizer/tokenizer_config.json',
                    'tokenizer/merges.txt', 'tokenizer/vocab.json', 'tokenizer/added_tokens.json',
                    'tokenizer/special_tokens_map.json', 'tokenizer/chat_template.jinja'}
        if not required.issubset(entries):
            raise Rejected('missing tokenizer sibling anchors')
        root = self.root.resolve(strict=True)
        for rel in required:
            path = root / rel
            entry = entries[rel]
            if (path.is_symlink() or path.parent.is_symlink() or not path.is_file()
                    or not path.resolve().is_relative_to(root)
                    or path.stat().st_size != entry['bytes'] or sha256(path) != entry['sha256']):
                raise Rejected('pinned tokenizer integrity failure')
        try:
            from tokenizers import Tokenizer
        except ImportError as exc:
            raise Rejected('tokenizers runtime library unavailable; NOT TESTED') from exc
        tokenizer = Tokenizer.from_file(str(root / 'tokenizer/tokenizer.json'))
        tokenizer.no_truncation()
        tokenizer.no_padding()
        self._tokenizer = tokenizer
        return tokenizer

    def count(self, prompt, model=KLEIN.key, *, pipeline_formatted=False):
        if model != KLEIN.key or not isinstance(prompt, str) or not prompt.strip():
            raise Rejected('unsupported tokenizer model or empty prompt')
        text = prompt
        if pipeline_formatted:
            if self.formatter is None:
                raise Rejected('pipeline formatter UNKNOWN; no admission')
            text = self.formatter(prompt)
            if not isinstance(text, str):
                raise Rejected('invalid pipeline formatter result')
        return len(self._load().encode(text, add_special_tokens=True).ids)

    def __call__(self, prompt, model):
        if self.max_sequence_length is None or self.formatter is None:
            raise Rejected('prompt admission limit/formatter UNKNOWN / NOT TESTED')
        if self.count(prompt, model, pipeline_formatted=True) > self.max_sequence_length:
            raise Rejected('prompt_too_long; silent truncation forbidden')
        return True
