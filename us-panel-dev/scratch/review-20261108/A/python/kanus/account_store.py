"""Local account/quota core, used by the commerce HTTP and Shopify adapters.

Do not expose publicly until authentication, throttling, deployment and request
metering are integrated and independently reviewed. SQLite file must be private.
"""
import base64
import binascii
import contextlib
import hashlib
import hmac
import json
import re
import secrets
import sqlite3
import time
from storage import connect


def _decode(s):
    raw = base64.b64decode(s + '=' * (-len(s) % 4), altchars=b'-_', validate=True)
    if base64.urlsafe_b64encode(raw).decode('ascii').rstrip('=') != s:
        raise ValueError('non-canonical encoding')
    return raw


def _hash(text):
    return hashlib.sha256(text.encode('utf-8')).hexdigest()


def verify_license(token, secret, now=None):
    if not secret or len(token) > 4096:
        raise ValueError('invalid license')
    try:
        prefix, body, signature = token.split('.')
        expected = hmac.new(secret, body.encode('ascii'), hashlib.sha256).digest()
        if prefix != 'us1' or not hmac.compare_digest(_decode(signature), expected):
            raise ValueError('invalid license')
        doc = json.loads(_decode(body))
        moment = int(time.time() if now is None else now)
        if (doc.get('kind') != 'sub' or type(doc.get('exp')) is not int
                or doc['exp'] <= moment or type(doc.get('prompts')) is not int
                or not 1 <= doc['prompts'] <= 100000):
            raise ValueError('unsupported or expired license')
        return doc
    except (ValueError, TypeError, KeyError, UnicodeError, binascii.Error, AttributeError) as exc:
        raise ValueError('invalid license') from exc


class AccountStore:
    def __init__(self, path):
        self.path = str(path)
        with self._db() as db:
            db.executescript('''
                CREATE TABLE IF NOT EXISTS accounts (
                    id INTEGER PRIMARY KEY, email TEXT NOT NULL UNIQUE,
                    salt BLOB NOT NULL, password BLOB NOT NULL,
                    license_hash TEXT NOT NULL UNIQUE, expires INTEGER NOT NULL,
                    quota INTEGER NOT NULL, used INTEGER NOT NULL DEFAULT 0);
                CREATE TABLE IF NOT EXISTS sessions (
                    token_hash TEXT PRIMARY KEY, account_id INTEGER NOT NULL,
                    expires INTEGER NOT NULL, FOREIGN KEY(account_id) REFERENCES accounts(id));
                CREATE TABLE IF NOT EXISTS usage (
                    account_id INTEGER NOT NULL, request_id TEXT NOT NULL,
                    PRIMARY KEY(account_id, request_id),
                    FOREIGN KEY(account_id) REFERENCES accounts(id));
                CREATE TABLE IF NOT EXISTS issued_licenses (
                    license_hash TEXT PRIMARY KEY, order_id TEXT NOT NULL,
                    token TEXT NOT NULL, email TEXT NOT NULL, status TEXT NOT NULL,
                    UNIQUE(order_id, license_hash));
                CREATE TABLE IF NOT EXISTS webhook_orders (
                    order_id TEXT PRIMARY KEY, body_hash TEXT NOT NULL);
                CREATE TABLE IF NOT EXISTS delivery_jobs (
                    order_id TEXT PRIMARY KEY, email TEXT NOT NULL,
                    state TEXT NOT NULL DEFAULT 'pending', lease_until INTEGER NOT NULL DEFAULT 0);
                CREATE TABLE IF NOT EXISTS auth_attempts (
                    scope TEXT PRIMARY KEY, window_start INTEGER NOT NULL, attempts INTEGER NOT NULL);
                CREATE TABLE IF NOT EXISTS credit_packs (
                    license_hash TEXT PRIMARY KEY, account_id INTEGER NOT NULL,
                    expires INTEGER NOT NULL, quota INTEGER NOT NULL, used INTEGER NOT NULL DEFAULT 0,
                    FOREIGN KEY(license_hash) REFERENCES issued_licenses(license_hash),
                    FOREIGN KEY(account_id) REFERENCES accounts(id));
                CREATE TABLE IF NOT EXISTS metered_requests (
                    reservation TEXT PRIMARY KEY, account_id INTEGER NOT NULL,
                    request_id TEXT NOT NULL, request_hash TEXT NOT NULL,
                    source TEXT NOT NULL, source_id TEXT NOT NULL,
                    state TEXT NOT NULL, created INTEGER NOT NULL, updated INTEGER NOT NULL,
                    UNIQUE(account_id, request_id), FOREIGN KEY(account_id) REFERENCES accounts(id));
            ''')

    @contextlib.contextmanager
    def _db(self):
        db = connect(self.path)
        try:
            yield db
        finally:
            db.close()

    def register(self, email, password, token, secret, now=None):
        email = email.strip().lower()
        if not 3 <= len(email) <= 254 or '@' not in email or not 12 <= len(password) <= 256:
            raise ValueError('invalid account details')
        doc = verify_license(token, secret, now)
        salt = secrets.token_bytes(16)
        digest = hashlib.scrypt(password.encode(), salt=salt, n=16384, r=8, p=1, dklen=32)
        with self._db() as db:
            db.execute('BEGIN IMMEDIATE')
            try:
                issued = db.execute('SELECT status FROM issued_licenses WHERE license_hash=?', (_hash(token),)).fetchone()
                if not issued or issued[0] != 'active':
                    raise ValueError('license not issued or inactive')
                if db.execute('SELECT 1 FROM credit_packs WHERE license_hash=?',(_hash(token),)).fetchone():
                    raise ValueError('license already used')
                db.execute('INSERT INTO accounts(email,salt,password,license_hash,expires,quota) VALUES(?,?,?,?,?,?)',
                           (email, salt, digest, _hash(token), doc['exp'], doc['prompts']))
                db.execute('COMMIT')
            except sqlite3.IntegrityError as exc:
                db.execute('ROLLBACK')
                raise ValueError('account or license already used') from exc
            except Exception:
                db.execute('ROLLBACK')
                raise

    def login(self, email, password, now=None):
        moment = int(time.time() if now is None else now)
        with self._db() as db:
            row = db.execute('SELECT id,salt,password,expires FROM accounts WHERE email=?',
                             (email.strip().lower(),)).fetchone()
            if not row or len(password)>256:
                raise ValueError('invalid credentials')
            digest = hashlib.scrypt(password.encode(), salt=row[1], n=16384, r=8, p=1, dklen=32)
            if not hmac.compare_digest(digest, row[2]):
                raise ValueError('invalid credentials')
            session = secrets.token_urlsafe(32)
            # An expired base entitlement must not lock a customer out of
            # redeeming a newly purchased pack. Access still requires credits.
            db.execute('INSERT INTO sessions VALUES(?,?,?)', (_hash(session), row[0], moment+86400))
            return session

    def redeem_pack(self, session, token, secret, now=None):
        moment = int(time.time() if now is None else now)
        if len(session)>256: raise ValueError('invalid session')
        doc = verify_license(token, secret, moment)
        token_hash = _hash(token)
        with self._db() as db:
            db.execute('BEGIN IMMEDIATE')
            try:
                row = db.execute('SELECT account_id FROM sessions WHERE token_hash=? AND expires>?', (_hash(session),moment)).fetchone()
                issued = db.execute('SELECT status FROM issued_licenses WHERE license_hash=?', (token_hash,)).fetchone()
                if not row or not issued or issued[0] != 'active':
                    raise ValueError('session or license invalid')
                if db.execute('SELECT 1 FROM accounts WHERE license_hash=?',(token_hash,)).fetchone():
                    raise ValueError('license already used')
                db.execute('INSERT INTO credit_packs(license_hash,account_id,expires,quota) VALUES(?,?,?,?)',
                           (token_hash,row[0],doc['exp'],doc['prompts']))
                db.execute('COMMIT')
                return doc['prompts']
            except sqlite3.IntegrityError as exc:
                db.execute('ROLLBACK');raise ValueError('license already used') from exc
            except Exception:
                db.execute('ROLLBACK');raise

    def balance(self, session, now=None):
        moment = int(time.time() if now is None else now)
        if len(session)>256: raise ValueError('invalid session')
        with self._db() as db:
            row=db.execute('SELECT a.id,a.quota,a.used,a.expires,l.status FROM accounts a JOIN sessions s ON s.account_id=a.id JOIN issued_licenses l ON l.license_hash=a.license_hash WHERE s.token_hash=? AND s.expires>?',(_hash(session),moment)).fetchone()
            if not row: raise ValueError('session expired')
            base=max(0,row[1]-row[2]) if row[3]>moment and row[4]=='active' else 0
            extra=db.execute("SELECT COALESCE(SUM(p.quota-p.used),0) FROM credit_packs p JOIN issued_licenses l ON l.license_hash=p.license_hash WHERE p.account_id=? AND p.expires>? AND l.status='active'",(row[0],moment)).fetchone()[0]
            return int(base+extra)

    def check_auth_rate(self, scope, now=None):
        moment=int(time.time() if now is None else now)
        with self._db() as db:
            db.execute('BEGIN IMMEDIATE')
            try:
                db.execute('DELETE FROM auth_attempts WHERE window_start<?',(moment-600,))
                row=db.execute('SELECT window_start,attempts FROM auth_attempts WHERE scope=?',(_hash(scope),)).fetchone()
                if row and moment-row[0]<300 and row[1]>=10:
                    raise ValueError('rate limited')
                if not row or moment-row[0]>=300:
                    db.execute('INSERT OR REPLACE INTO auth_attempts VALUES(?,?,?)',(_hash(scope),moment,1))
                else:
                    db.execute('UPDATE auth_attempts SET attempts=attempts+1 WHERE scope=?',(_hash(scope),))
                db.execute('COMMIT')
            except Exception:
                db.execute('ROLLBACK');raise

    def charge_prompt(self, session, request_id, now=None):
        return self._charge_prompt(session,request_id,now,None,'')

    def reserve_prompt(self, session, request_id, request_hash, now=None):
        if not isinstance(request_hash,str) or len(request_hash)!=64:
            raise ValueError('invalid request fingerprint')
        try: bytes.fromhex(request_hash)
        except ValueError: raise ValueError('invalid request fingerprint')
        reservation=secrets.token_hex(16)
        remaining=self._charge_prompt(session,request_id,now,reservation,request_hash)
        return reservation,remaining

    def reservation_active(self, reservation):
        with self._db() as db:
            return db.execute("SELECT 1 FROM metered_requests WHERE reservation=? AND state='reserved'",(reservation,)).fetchone() is not None

    def finish_reservation(self, reservation, success, now=None):
        if not isinstance(reservation,str) or len(reservation)!=32 or type(success) is not bool:
            raise ValueError('invalid reservation')
        moment=int(time.time() if now is None else now)
        with self._db() as db:
            db.execute('BEGIN IMMEDIATE')
            try:
                row=db.execute("SELECT account_id,source,source_id,state FROM metered_requests WHERE reservation=?",(reservation,)).fetchone()
                if not row: raise ValueError('unknown reservation')
                if row[3]!='reserved':db.execute('COMMIT');return False
                state='completed' if success else 'refunded'
                if not success:
                    if row[1]=='base': changed=db.execute('UPDATE accounts SET used=used-1 WHERE id=? AND used>0',(row[0],)).rowcount
                    else: changed=db.execute('UPDATE credit_packs SET used=used-1 WHERE license_hash=? AND account_id=? AND used>0',(row[2],row[0])).rowcount
                    if changed!=1:raise ValueError('credit reconciliation failed')
                db.execute('UPDATE metered_requests SET state=?,updated=? WHERE reservation=?',(state,moment,reservation))
                db.execute('COMMIT');return True
            except Exception:
                db.execute('ROLLBACK');raise

    def request_history(self, session, now=None):
        moment=int(time.time() if now is None else now)
        with self._db() as db:
            row=db.execute('SELECT account_id FROM sessions WHERE token_hash=? AND expires>?',(_hash(session),moment)).fetchone()
            if not row:raise ValueError('session expired')
            return [dict(zip(('request_id','state','created','updated'),r)) for r in db.execute('SELECT request_id,state,created,updated FROM metered_requests WHERE account_id=? ORDER BY created DESC,reservation DESC LIMIT 20',(row[0],))]

    def unsettled_requests(self):
        # Local operator metadata only; never served over an unauthenticated API.
        with self._db() as db:
            return [dict(zip(('reservation','request_id','source','created','updated'),r)) for r in db.execute("SELECT reservation,request_id,source,created,updated FROM metered_requests WHERE state='reserved' ORDER BY created,reservation")]

    def _charge_prompt(self, session, request_id, now, reservation, request_hash):
        if not isinstance(request_id,str) or not re.fullmatch(r'[A-Za-z0-9._-]{8,128}',request_id) or not isinstance(session,str) or len(session)>256:
            raise ValueError('invalid request')
        moment = int(time.time() if now is None else now)
        with self._db() as db:
            db.execute('BEGIN IMMEDIATE')
            try:
                row = db.execute("SELECT a.id,a.quota,a.used,a.expires,s.expires,l.status FROM accounts a JOIN sessions s ON s.account_id=a.id JOIN issued_licenses l ON l.license_hash=a.license_hash WHERE s.token_hash=?", (_hash(session),)).fetchone()
                if not row or row[4] <= moment:
                    raise ValueError('session expired')
                if db.execute('SELECT 1 FROM usage WHERE account_id=? AND request_id=?', (row[0],request_id)).fetchone():
                    raise ValueError('duplicate request id')
                if reservation and db.execute("SELECT count(*) FROM metered_requests WHERE state='reserved'").fetchone()[0]>=10:
                    raise ValueError('model queue is full')
                base_available = row[3] > moment and row[5] == 'active' and row[2] < row[1]
                pack = None
                if not base_available:
                    pack=db.execute("SELECT p.license_hash FROM credit_packs p JOIN issued_licenses l ON l.license_hash=p.license_hash WHERE p.account_id=? AND p.expires>? AND p.used<p.quota AND l.status='active' ORDER BY p.expires,p.license_hash LIMIT 1",(row[0],moment)).fetchone()
                    if not pack: raise ValueError('quota exhausted')
                db.execute('INSERT INTO usage VALUES(?,?)', (row[0],request_id))
                if base_available: db.execute('UPDATE accounts SET used=used+1 WHERE id=?', (row[0],))
                else: db.execute('UPDATE credit_packs SET used=used+1 WHERE license_hash=?',(pack[0],))
                db.execute('INSERT INTO metered_requests VALUES(?,?,?,?,?,?,?,?,?)',(reservation or secrets.token_hex(16),row[0],request_id,request_hash,'base' if base_available else 'pack','' if base_available else pack[0],'reserved' if reservation else 'completed',moment,moment))
                base=max(0,row[1]-row[2]-(1 if base_available else 0)) if row[3]>moment and row[5]=='active' else 0
                extra=db.execute("SELECT COALESCE(SUM(p.quota-p.used),0) FROM credit_packs p JOIN issued_licenses l ON l.license_hash=p.license_hash WHERE p.account_id=? AND p.expires>? AND l.status='active'",(row[0],moment)).fetchone()[0]
                db.execute('COMMIT'); return int(base+extra)
            except Exception:
                db.execute('ROLLBACK'); raise
