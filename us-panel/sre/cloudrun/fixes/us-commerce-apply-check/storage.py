"""Persistent SQL connection shared by accounts, metering and lifecycle state.

SQLite remains available for a single local host. Production may use PostgreSQL;
no database URL or driver exception containing credentials is logged.
"""
import re
import sqlite3


class PostgresConnection:
    def __init__(self, url):
        import psycopg
        self.driver = psycopg
        try:
            self.db = psycopg.connect(url, autocommit=True, connect_timeout=10)
            self.db.execute("SET statement_timeout='15s'")
            self.db.execute("SET lock_timeout='10s'")
        except psycopg.Error as exc:
            raise RuntimeError('PostgreSQL connection failed') from None

    def execute(self, sql, params=()):
        query = sql.strip()
        if query.upper() == 'BEGIN IMMEDIATE':
            self.db.execute('BEGIN')
            # Match SQLite's write-transaction serialization across processes.
            return self.db.execute('SELECT pg_advisory_xact_lock(718904224)')
        if query.startswith('INSERT OR REPLACE INTO auth_attempts'):
            query = query.replace('INSERT OR REPLACE', 'INSERT', 1)
            query += ' ON CONFLICT(scope) DO UPDATE SET window_start=EXCLUDED.window_start,attempts=EXCLUDED.attempts'
        try:
            return self.db.execute(query.replace('?', '%s'), params)
        except self.driver.IntegrityError:
            raise sqlite3.IntegrityError('database integrity constraint') from None
        except self.driver.Error:
            raise RuntimeError('database operation failed') from None

    def executescript(self, script):
        script = script.replace('id INTEGER PRIMARY KEY', 'id BIGSERIAL PRIMARY KEY')
        script = re.sub(r'\bINTEGER\b', 'BIGINT', script)
        script = re.sub(r'\bBLOB\b', 'BYTEA', script)
        self.execute('BEGIN IMMEDIATE')
        try:
            for statement in script.split(';'):
                if statement.strip():self.execute(statement)
            self.execute('COMMIT')
        except BaseException:
            self.execute('ROLLBACK');raise

    def close(self):
        self.db.close()


def connect(path):
    if str(path).startswith(('postgresql://', 'postgres://')):
        return PostgresConnection(str(path))
    db = sqlite3.connect(str(path), timeout=10, isolation_level=None)
    db.execute('PRAGMA foreign_keys=ON')
    return db
