"""Guard the audited existing boot script; never initialize or format model disks."""
import hashlib,os,pathlib,subprocess
path=pathlib.Path('/opt/ds/boot.sh')
old=path.read_bytes()
expected='afd8ed75048ae89635491810915f72d25dd481a7aebf96353142f640d7cbab01'
if hashlib.sha256(old).hexdigest()!=expected:raise SystemExit('boot source changed; inspect before applying')
# The audited file may indent by two spaces. Match actual line indentation.
lines=old.splitlines(keepends=True)
create=[i for i,l in enumerate(lines) if b'mdadm --create' in l]
if len(create)!=1 or b'mkfs.ext4' not in lines[create[0]+1]:raise SystemExit('unexpected disk initialization block')
i=create[0];indent=lines[i][:len(lines[i])-len(lines[i].lstrip())]
lines[i:i+2]=[indent+b'echo "RAID assembly failed; refusing to initialize or format existing model disks" >&2\n',indent+b'exit 1\n']
new=b''.join(lines)
if b'mkfs.' in new or b'mdadm --create' in new:raise SystemExit('unsafe boot command remains')
backup=path.parent/'us-backups-20261105'
backup.mkdir(mode=0o700,exist_ok=True)
b=backup/'boot.sh.original'
if b.exists() and b.read_bytes()!=old:raise SystemExit('backup conflict')
b.write_bytes(old);b.chmod(0o600)
tmp=path.with_name('boot.sh.us-new');tmp.write_bytes(new);tmp.chmod(path.stat().st_mode & 0o777)
subprocess.run(['bash','-n',str(tmp)],check=True)
os.replace(tmp,path)
print('BOOT_GUARD_APPLIED',hashlib.sha256(new).hexdigest())
