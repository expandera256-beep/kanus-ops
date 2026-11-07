"""Opt-in contracts on a real PostgreSQL server, with isolated schemas."""
import os
import unittest
import uuid
from urllib.parse import urlencode
from storage import PostgresConnection
from account_store import AccountStore
from test_account_store import AccountsTest
from test_credit_packs import CreditPacksTest
from test_metered_reservations import ReservationTest
from test_inference_lifecycle import LifecycleTest,Clock,Provider
from inference_lifecycle import LifecycleStore,LifecycleController,WorkerPolicy


class PostgreSQLMixin:
    def setUp(self):
        dsn=os.environ.get('US_POSTGRES_TEST_DSN')
        if not dsn:self.skipTest('opt-in real PostgreSQL contract')
        import psycopg
        schema='contract_'+uuid.uuid4().hex
        admin=psycopg.connect(dsn,autocommit=True)
        admin.execute('CREATE SCHEMA '+schema)
        self.addCleanup(admin.close)
        self.addCleanup(lambda:admin.execute('DROP SCHEMA '+schema+' CASCADE'))
        self.postgres_dsn=dsn+('&' if '?' in dsn else '?')+urlencode({'options':'-csearch_path='+schema})
        super().setUp()
        # Preserve each contract's already initialized fixture by recreating
        # setUp against PostgreSQL rather than substituting a fake driver.
        original=self.store
        self.store=AccountStore(self.postgres_dsn)
        with original._db() as src, self.store._db() as dest:
            for table in ('accounts','sessions','usage','issued_licenses','webhook_orders','delivery_jobs','auth_attempts','credit_packs','metered_requests'):
                rows=src.execute('SELECT * FROM '+table).fetchall()
                for row in rows:
                    dest.execute('INSERT INTO '+table+' VALUES('+','.join('?' for _ in row)+')',row)
            row=dest.execute('SELECT MAX(id) FROM accounts').fetchone()
            if row[0] is not None:dest.execute("SELECT setval('accounts_id_seq', ?)",(row[0],))


class PostgresAccounts(PostgreSQLMixin,AccountsTest):pass
class PostgresCredits(PostgreSQLMixin,CreditPacksTest):pass
class PostgresReservations(PostgreSQLMixin,ReservationTest):pass


class PostgresLifecycle(LifecycleTest):
    def setUp(self):
        dsn=os.environ.get('US_POSTGRES_TEST_DSN')
        if not dsn:self.skipTest('opt-in real PostgreSQL contract')
        import psycopg
        schema='lifecycle_'+uuid.uuid4().hex
        admin=psycopg.connect(dsn,autocommit=True);admin.execute('CREATE SCHEMA '+schema)
        self.addCleanup(admin.close);self.addCleanup(lambda:admin.execute('DROP SCHEMA '+schema+' CASCADE'))
        database=dsn+('&' if '?' in dsn else '?')+urlencode({'options':'-csearch_path='+schema})
        self.store=LifecycleStore(database);self.clock=Clock();self.provider=Provider()
        self.policy=WorkerPolicy('gpu','actual-model','project-one','europe-west4-b','worker','123','http://127.0.0.1:8080',require_watchdog=False)
        self.allowed=set()
        self.controller=LifecycleController(self.store,self.provider,[self.policy],lambda r:r in self.allowed,self.clock,lambda _:None)


def load_tests(loader,tests,pattern):
    suite=unittest.TestSuite()
    for cls in (PostgresAccounts,PostgresCredits,PostgresReservations,PostgresLifecycle):suite.addTests(loader.loadTestsFromTestCase(cls))
    return suite


if __name__=='__main__':unittest.main()
