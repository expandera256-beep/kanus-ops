"""Capture/verify a local SSD canary plus bounded samples of every model file.

These are preservation checks, not a complete cryptographic backup verification.
No model contents, credentials, or file hashes are emitted to stdout.
"""
import ast
import hashlib
import json
import pathlib
import secrets
import sys

root=pathlib.Path('/opt/ds/us-backups-20261105')
manifest=root/'ssd-preservation.json'
canary=pathlib.Path('/mnt/nvme/.us-preservation-probe-20261105')
tree=ast.parse(pathlib.Path('/opt/ds/fetch-weights.py').read_text())
jobs=next(ast.literal_eval(node.value) for node in tree.body if isinstance(node,ast.Assign) and any(isinstance(target,ast.Name) and target.id=='JOBS' for target in node.targets))

def snapshot():
    rows=[]
    for name,dest,expected_size in jobs:
        path=pathlib.Path(dest)
        if path.stat().st_size!=expected_size:raise SystemExit('model size mismatch')
        with path.open('rb') as f:
            first=f.read(4*1024*1024);f.seek(max(0,expected_size-4*1024*1024));last=f.read(4*1024*1024)
        rows.append({'name':name,'size':expected_size,'sample_sha256':hashlib.sha256(first+last).hexdigest()})
    return {'models':rows,'canary_sha256':hashlib.sha256(canary.read_bytes()).hexdigest()}

action=sys.argv[1] if len(sys.argv)==2 else ''
if action=='capture':
    root.mkdir(mode=0o700,exist_ok=True)
    if manifest.exists() or canary.exists():raise SystemExit('baseline already exists; do not overwrite')
    canary.write_bytes(secrets.token_bytes(64));canary.chmod(0o600)
    manifest.write_text(json.dumps(snapshot()));manifest.chmod(0o600)
    print('SSD_BASELINE_CAPTURED; seven model sizes and samples; 64-byte canary')
elif action=='verify':
    if snapshot()!=json.loads(manifest.read_text()):raise SystemExit('SSD preservation mismatch')
    print('SSD_PRESERVATION_VERIFIED; seven model sizes and samples; canary unchanged')
else:raise SystemExit('choose capture or verify')
