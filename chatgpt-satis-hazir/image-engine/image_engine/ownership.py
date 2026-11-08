"""Linux advisory ownership with durable, monotonically increasing fencing.

Lock directory must be local, owner-controlled storage, never NFS. Never unlink
owner.lock: replacing its inode would create a second independent lock domain.
"""
from __future__ import annotations
import fcntl
import json
import os
import stat
from dataclasses import asdict, dataclass
from pathlib import Path


class OwnershipError(RuntimeError):
    pass


def boot_id() -> str:
    return Path('/proc/sys/kernel/random/boot_id').read_text().strip()


def process_start(pid: int) -> str:
    # comm can contain spaces and ')'; field 22 is offset 19 after closing comm.
    return Path(f'/proc/{pid}/stat').read_text().rsplit(')', 1)[1].split()[19]


def atomic_json(path: Path, value: dict) -> None:
    import tempfile
    fd, name = tempfile.mkstemp(prefix='.' + path.name, dir=path.parent)
    try:
        with os.fdopen(fd, 'w') as out:
            json.dump(value, out, sort_keys=True)
            out.flush()
            os.fsync(out.fileno())
        os.replace(name, path)
        dfd = os.open(path.parent, os.O_RDONLY | os.O_DIRECTORY)
        try:
            os.fsync(dfd)
        finally:
            os.close(dfd)
    finally:
        if os.path.exists(name):
            os.unlink(name)


@dataclass(frozen=True)
class Fence:
    generation: int
    boot_id: str
    pid: int
    process_start: str

    def alive(self) -> bool:
        try:
            return self.boot_id == boot_id() and self.process_start == process_start(self.pid)
        except (OSError, IndexError, ValueError):
            return False


class Ownership:
    def __init__(self, directory: Path):
        self.directory = Path(directory)
        self.directory.mkdir(mode=0o700, parents=True, exist_ok=True)
        info = self.directory.lstat()
        if not stat.S_ISDIR(info.st_mode) or info.st_uid != os.getuid() or info.st_mode & 0o022:
            raise OwnershipError('ownership directory must be private and owner controlled')
        self.path = self.directory / 'owner.json'
        self.lock_path = self.directory / 'owner.lock'
        self.fd = None
        self.fence = None

    def acquire(self) -> Fence:
        if self.fd is not None:
            raise OwnershipError('already acquired')
        fd = os.open(self.lock_path, os.O_CREAT | os.O_RDWR | os.O_NOFOLLOW, 0o600)
        try:
            info = os.fstat(fd)
            if info.st_uid != os.getuid() or not stat.S_ISREG(info.st_mode) or info.st_nlink != 1:
                raise OwnershipError('unsafe lock inode')
            try:
                fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError as exc:
                raise OwnershipError('GPU authority already held') from exc
            generation = 0
            if self.path.exists():
                if self.path.is_symlink():
                    raise OwnershipError('symlink fencing state')
                try:
                    old = json.loads(self.path.read_text())
                    generation = old['generation']
                    if type(generation) is not int or generation < 1:
                        raise ValueError('invalid generation')
                    Fence(**{k: old[k] for k in Fence.__dataclass_fields__})
                except (KeyError, ValueError, TypeError) as exc:
                    raise OwnershipError('corrupt fencing state; operator recovery required') from exc
            fence = Fence(generation + 1, boot_id(), os.getpid(), process_start(os.getpid()))
            atomic_json(self.path, asdict(fence))
            self.fd, self.fence = fd, fence
            return fence
        except BaseException:
            os.close(fd)
            raise

    def assert_current(self, token: Fence | None = None) -> Fence:
        token = token or self.fence
        if self.fd is None or token is None or token != self.fence:
            raise OwnershipError('authority not held / stale fencing generation')
        # Protect against using an inherited fd after fork and lock-file replacement.
        if token.pid != os.getpid() or not token.alive():
            raise OwnershipError('stale boot/PID/process-start identity')
        if os.fstat(self.fd).st_ino != self.lock_path.stat().st_ino:
            raise OwnershipError('lock inode replaced')
        validate_fence(self.path, token)
        return token

    def release(self) -> None:
        if self.fd is not None:
            # A forked child must not unlock the parent's shared open description.
            if self.fence is not None and self.fence.pid != os.getpid():
                os.close(self.fd)
            else:
                # Revoke any still-running worker before dropping authority.
                if self.fence is not None:
                    try:
                        current = json.loads(self.path.read_text())
                    except (OSError, ValueError):
                        current = None
                    if current == asdict(self.fence):
                        atomic_json(self.path, dict(asdict(self.fence), released=True))
                fcntl.flock(self.fd, fcntl.LOCK_UN)
                os.close(self.fd)
            self.fd = None
            self.fence = None

    def __enter__(self):
        self.acquire()
        return self

    def __exit__(self, *args):
        self.release()


def validate_fence(path: Path, token: Fence) -> None:
    """Worker-side authorization; never grants ownership merely from this file."""
    try:
        if path.is_symlink() or json.loads(path.read_text()) != asdict(token) or not token.alive():
            raise OwnershipError('stale fence / boot / PID identity')
    except (OSError, ValueError) as exc:
        raise OwnershipError('unreadable fencing state') from exc
