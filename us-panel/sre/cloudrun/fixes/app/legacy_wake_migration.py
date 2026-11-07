"""Reviewable migration guard for the old gateway; dry-run by default.

Use only after routing paid work through the new managed commerce gateway.
Already-running workers remain reachable. The old gateway can no longer start,
resume or stop VMs in response to health, static, login or metadata requests.
This tool does not restart services or alter a VM's power state.
"""
import argparse
import ast
import hashlib
import os
from pathlib import Path
import stat
import tempfile


def guarded_source(source):
    tree=ast.parse(source)
    functions=[x for x in tree.body if isinstance(x,ast.FunctionDef) and x.name=='ensure']
    if len(functions)!=1:raise ValueError('expected exactly one legacy ensure function')
    function=functions[0]
    calls={ast.unparse(n.func) for n in ast.walk(function) if isinstance(n,ast.Call)}
    if not {'worker_awake','power','instance_status'}<=calls:raise ValueError('unexpected legacy gateway revision')
    replacement='''def ensure(worker: dict) -> None:
    # Power operations belong exclusively to the metered commerce controller.
    if worker_awake(worker):
        return
    raise TimeoutError('İşçi kapalı; yetkili ticari model geçidini kullanın.')
'''
    lines=source.splitlines(keepends=True)
    output=''.join(lines[:function.lineno-1])+replacement+''.join(lines[function.end_lineno:])
    compile(output,'legacy-wake-guard','exec')
    return output


def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('source');parser.add_argument('--apply',action='store_true');parser.add_argument('--expected-sha256')
    args=parser.parse_args();path=Path(args.source)
    if path.is_symlink() or not path.is_file() or path.stat().st_size>1_000_000:raise ValueError('regular bounded source required')
    before=path.read_bytes();digest=hashlib.sha256(before).hexdigest();after=guarded_source(before.decode()).encode()
    print('before_sha256='+digest);print('after_sha256='+hashlib.sha256(after).hexdigest())
    if not args.apply:print('DRY RUN: no source, service or power state changed');return
    if args.expected_sha256!=digest:raise ValueError('explicit matching expected hash required')
    backup=path.with_name(path.name+'.us-before-'+digest[:16])
    with backup.open('xb') as file:
        os.chmod(backup,0o600);file.write(before);file.flush();os.fsync(file.fileno())
    fd,temp=tempfile.mkstemp(prefix=path.name+'.us-',dir=path.parent)
    try:
        with os.fdopen(fd,'wb') as file:
            os.fchmod(file.fileno(),stat.S_IMODE(path.stat().st_mode));file.write(after);file.flush();os.fsync(file.fileno())
        if hashlib.sha256(path.read_bytes()).hexdigest()!=digest:raise ValueError('source changed during migration')
        os.replace(temp,path)
    finally:
        if os.path.exists(temp):os.unlink(temp)
    print('Guard saved with private backup; service restart is a separate explicit deployment step')


if __name__=='__main__':main()
