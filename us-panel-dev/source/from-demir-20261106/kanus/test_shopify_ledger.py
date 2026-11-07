import base64
import hashlib
import hmac
import json
import tempfile
import unittest
from pathlib import Path
from account_store import AccountStore
from shopify_ledger import receive_paid_order, receive_cancelled_order, receive_full_refund
from model_gateway import MeteredGateway


class ShopifyLedgerTest(unittest.TestCase):
    def setUp(self):
        self.tmp=tempfile.TemporaryDirectory(); self.addCleanup(self.tmp.cleanup)
        self.store=AccountStore(Path(self.tmp.name)/'db.sqlite')
        self.webhook=b'webhook-secret'; self.license=b'license-secret'
        self.order={'id':1234,'email':'buyer@example.com','financial_status':'paid','line_items':[{'product_id':15424693010660,'quantity':1}]}

    def sign(self, payload):
        return base64.b64encode(hmac.new(self.webhook,payload,hashlib.sha256).digest()).decode('ascii')

    def test_paid_order_register_use_and_cancel(self):
        raw=json.dumps(self.order).encode(); sig=self.sign(raw)
        keys=receive_paid_order(self.store,raw,sig,self.webhook,self.license,now=100)
        self.assertEqual(keys,receive_paid_order(self.store,raw,sig,self.webhook,self.license,now=101))
        self.store.register('buyer@example.com','long-password-123',keys[0],self.license,now=102)
        session=self.store.login('buyer@example.com','long-password-123',now=103)
        self.assertEqual(self.store.charge_prompt(session,'request-0001',now=104),499)
        cancel=json.dumps({'id':1234,'cancelled_at':'2026-11-04T00:00:00Z'}).encode()
        self.assertEqual(receive_cancelled_order(self.store,cancel,self.sign(cancel),self.webhook),1)
        self.assertEqual(receive_cancelled_order(self.store,cancel,self.sign(cancel),self.webhook),0)
        with self.assertRaises(ValueError): self.store.charge_prompt(session,'request-0002',now=105)

    def test_bad_signature_and_wrong_product_never_issue(self):
        raw=json.dumps(self.order).encode()
        with self.assertRaises(ValueError): receive_paid_order(self.store,raw,'bad',self.webhook,self.license,now=1)
        self.order['line_items']=[{'title':'monthly','price':'10000','quantity':1}]
        raw=json.dumps(self.order).encode()
        with self.assertRaises(ValueError): receive_paid_order(self.store,raw,self.sign(raw),self.webhook,self.license,now=1)
        with self.store._db() as db: self.assertEqual(db.execute('SELECT count(*) FROM issued_licenses').fetchone()[0],0)

    def test_same_order_changed_body_rejected(self):
        raw=json.dumps(self.order).encode(); receive_paid_order(self.store,raw,self.sign(raw),self.webhook,self.license,now=1)
        self.order['line_items'][0]['quantity']=2; changed=json.dumps(self.order).encode()
        with self.assertRaises(ValueError): receive_paid_order(self.store,changed,self.sign(changed),self.webhook,self.license,now=2)

    def test_full_refund_denies_gateway_and_partial_does_not_revoke(self):
        raw=json.dumps(self.order).encode(); key=receive_paid_order(self.store,raw,self.sign(raw),self.webhook,self.license,now=100)[0]
        self.store.register('buyer@example.com','long-password-123',key,self.license,now=101)
        session=self.store.login('buyer@example.com','long-password-123',now=102)
        hits=[]; gate=MeteredGateway(self.store,lambda payload: hits.append(payload) or 'answer')
        self.assertEqual(gate.complete(session,'request-01','first',now=103)['response'],'answer')
        with self.assertRaises(ValueError): gate.complete(session,'request-01','duplicate',now=103)
        self.assertEqual(hits,['first'])
        partial=json.dumps({'id':1234,'financial_status':'partially_refunded'}).encode()
        with self.assertRaises(ValueError): receive_full_refund(self.store,partial,self.sign(partial),self.webhook)
        full=json.dumps({'id':1234,'financial_status':'refunded'}).encode()
        self.assertEqual(receive_full_refund(self.store,full,self.sign(full),self.webhook),1)
        with self.assertRaises(ValueError): gate.complete(session,'request-02','after-refund',now=104)
        self.assertEqual(hits,['first'])

if __name__=='__main__': unittest.main()
