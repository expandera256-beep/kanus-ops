"""Local durable idempotency and fail-closed recovery journal. No billing API."""
import json
import sqlite3
from pathlib import Path
from .models import Rejected


class Ledger:
    def __init__(self, path: Path):
        self.db = sqlite3.connect(path)
        self.db.execute('PRAGMA journal_mode=WAL')
        self.db.execute('PRAGMA synchronous=FULL')
        self.db.executescript('''
            CREATE TABLE IF NOT EXISTS jobs (
                id TEXT PRIMARY KEY, fingerprint TEXT NOT NULL,
                outcome TEXT, settlement_count INTEGER NOT NULL DEFAULT 0);
            CREATE TABLE IF NOT EXISTS control (key TEXT PRIMARY KEY, value TEXT NOT NULL);
        ''')
        self.db.commit()

    @property
    def blocked(self):
        return self.db.execute("SELECT value FROM control WHERE key='blocked'").fetchone()

    def block(self, reason):
        with self.db:
            self.db.execute("INSERT OR REPLACE INTO control VALUES ('blocked', ?)", (reason,))

    def begin(self, job):
        row = self.db.execute('SELECT fingerprint,outcome FROM jobs WHERE id=?', (job.job_id,)).fetchone()
        if row:
            if row[0] != job.fingerprint():
                raise Rejected('idempotency key payload mismatch')
            if row[1] is None:
                raise Rejected('unfinished reservation; explicit reconciliation required, no retry')
            return json.loads(row[1])
        with self.db:
            self.db.execute('INSERT INTO jobs(id,fingerprint) VALUES (?,?)', (job.job_id, job.fingerprint()))
        return None

    def finish(self, job_id, result):
        with self.db:
            self.db.execute('UPDATE jobs SET outcome=?,settlement_count=1 WHERE id=? AND outcome IS NULL',
                            (json.dumps(result, sort_keys=True), job_id))
        return json.loads(self.db.execute('SELECT outcome FROM jobs WHERE id=?', (job_id,)).fetchone()[0])

    def poll(self, job_id):
        row = self.db.execute('SELECT outcome FROM jobs WHERE id=?', (job_id,)).fetchone()
        if row is None:
            return {'job_id': job_id, 'status': 'unknown'}
        if row[0] is None:
            return {'job_id': job_id, 'status': 'pending_reconciliation'}
        return json.loads(row[0])

    def close(self):
        self.db.close()
