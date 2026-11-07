import base64
import hashlib
import hmac
import io
import json
import tempfile
import unittest
from pathlib import Path
from wsgiref.util import setup_testing_defaults
from account_store import AccountStore
from commerce_http import make_app
from delivery import deliver_one


class Upstream:
    status=200
    headers={'Content-Type':'text/event-stream'}
    def __init__(self):self.stream=io.BytesIO(b'data: {"choices":[{"delta":{"content":"answer"},"finish_reason":null}]}\n\ndata: {"choices":[{"delta":{},"finish_reason":"stop"}]}\n\ndata: [DONE]\n\n')
    def read(self,n):return self.stream.read(n)
    def close(self):self.stream.close()


class CommerceHTTPTest(unittest.TestCase):
    def setUp(self):
        self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup)
        self.store=AccountStore(Path(self.tmp.name)/'commerce.db');self.webhook=b'webhook-secret';self.key=b'license-secret';self.calls=[]
        self.app=make_app(self.store,self.webhook,self.key,'test.myshopify.com',lambda raw:self.calls.append(raw) or Upstream())

    def request(self,path,body,headers=None):
        raw=body if isinstance(body,bytes) else json.dumps(body).encode();env={};setup_testing_defaults(env)
        env.update(REQUEST_METHOD='POST',PATH_INFO=path,CONTENT_LENGTH=str(len(raw)),REMOTE_ADDR='127.0.0.1',**{'wsgi.input':io.BytesIO(raw)})
        env.update(headers or {}); result={}
        def start(status,headers):result['status']=status;result['headers']=headers
        result['body']=b''.join(self.app(env,start));return result

    def test_webhook_delivery_account_and_metered_model(self):
        order={'id':10,'email':'buyer@example.com','financial_status':'paid','line_items':[{'product_id':15424693010660,'quantity':1}]}
        raw=json.dumps(order).encode();sig=base64.b64encode(hmac.new(self.webhook,raw,hashlib.sha256).digest()).decode()
        headers={'HTTP_X_SHOPIFY_SHOP_DOMAIN':'test.myshopify.com','HTTP_X_SHOPIFY_TOPIC':'orders/paid','HTTP_X_SHOPIFY_HMAC_SHA256':sig}
        self.assertTrue(self.request('/webhooks/shopify',raw,headers)['status'].startswith('200'))
        self.assertTrue(self.request('/webhooks/shopify',raw,headers)['status'].startswith('200'))
        sent=[];self.assertTrue(deliver_one(self.store,lambda to,keys:sent.append((to,keys))));self.assertFalse(deliver_one(self.store,lambda *_:self.fail('sent twice')))
        self.assertEqual(sent[0][0],'buyer@example.com');license=sent[0][1][0]
        creds={'email':'buyer@example.com','password':'long-password-123','license':license}
        self.assertTrue(self.request('/account/register',creds)['status'].startswith('201'))
        response=self.request('/account/login',creds);self.assertTrue(response['status'].startswith('200'))
        session=json.loads(response['body'])['session'];model={'model':'example','messages':[{'role':'user','content':'hi'}]}
        auth={'HTTP_AUTHORIZATION':'Bearer '+session,'HTTP_X_US_REQUEST_ID':'request-0001'}
        response=self.request('/v1/chat/completions',model,auth);self.assertIn(b'[DONE]',response['body']);self.assertEqual(len(self.calls),1)
        self.assertTrue(self.request('/v1/chat/completions',model,auth)['status'].startswith('400'));self.assertEqual(len(self.calls),1)

    def test_bad_shop_and_signature_never_send(self):
        order={'id':10,'email':'buyer@example.com','financial_status':'paid','line_items':[{'product_id':15424693010660,'quantity':1}]}
        raw=json.dumps(order).encode();headers={'HTTP_X_SHOPIFY_SHOP_DOMAIN':'other.myshopify.com','HTTP_X_SHOPIFY_TOPIC':'orders/paid','HTTP_X_SHOPIFY_HMAC_SHA256':'bad'}
        self.assertTrue(self.request('/webhooks/shopify',raw,headers)['status'].startswith('403'))
        headers['HTTP_X_SHOPIFY_SHOP_DOMAIN']='test.myshopify.com'
        self.assertTrue(self.request('/webhooks/shopify',raw,headers)['status'].startswith('400'));self.assertFalse(deliver_one(self.store,lambda *_:self.fail('sent')))

    def test_unconfigured_shopify_is_explicitly_disabled(self):
        self.app=make_app(self.store,b'',self.key,'test.myshopify.com',lambda _:self.fail('model must stay asleep'),webhooks_enabled=False)
        self.assertTrue(self.request('/webhooks/shopify',{})['status'].startswith('503'))
        env={};setup_testing_defaults(env);env.update(REQUEST_METHOD='GET',PATH_INFO='/healthz')
        result={}
        def start(status,headers):result['status']=status
        self.assertEqual(json.loads(b''.join(self.app(env,start))),{'ok':True,'shopify_webhooks':False})
        self.assertEqual(result['status'],'200 OK')
        with self.store._db() as db:self.assertEqual(db.execute('SELECT COUNT(*) FROM issued_licenses').fetchone()[0],0)
        with self.assertRaises(ValueError):make_app(self.store,b'',self.key,'test.myshopify.com',lambda _:None)

    def test_mail_retry_retains_same_license(self):
        order={'id':11,'email':'buyer@example.com','financial_status':'paid','line_items':[{'product_id':15424693010660,'quantity':1}]}
        raw=json.dumps(order).encode();sig=base64.b64encode(hmac.new(self.webhook,raw,hashlib.sha256).digest()).decode();headers={'HTTP_X_SHOPIFY_SHOP_DOMAIN':'test.myshopify.com','HTTP_X_SHOPIFY_TOPIC':'orders/paid','HTTP_X_SHOPIFY_HMAC_SHA256':sig}
        self.request('/webhooks/shopify',raw,headers);seen=[]
        def fail(to,keys):seen.append(keys[0]);raise OSError('SMTP unavailable')
        with self.assertRaises(OSError):deliver_one(self.store,fail)
        self.assertTrue(deliver_one(self.store,lambda to,keys:seen.append(keys[0])));self.assertEqual(seen[0],seen[1])

    def test_credit_pack_redeem_and_balance_via_http(self):
        catalog={'789':{'prompts':5,'price_minor':69000,'valid_days':90,'currency':'TRY'}}
        self.app=make_app(self.store,self.webhook,self.key,'test.myshopify.com',lambda raw:self.calls.append(raw) or Upstream(),catalog)
        def buy(number):
            order={'id':number,'email':'buyer@example.com','financial_status':'paid','currency':'TRY','line_items':[{'variant_id':789,'quantity':1,'price':'690.00','discount_allocations':[]}]}
            raw=json.dumps(order).encode();sig=base64.b64encode(hmac.new(self.webhook,raw,hashlib.sha256).digest()).decode()
            headers={'HTTP_X_SHOPIFY_SHOP_DOMAIN':'test.myshopify.com','HTTP_X_SHOPIFY_TOPIC':'orders/paid','HTTP_X_SHOPIFY_HMAC_SHA256':sig}
            self.assertTrue(self.request('/webhooks/shopify',raw,headers)['status'].startswith('200'))
            sent=[];self.assertTrue(deliver_one(self.store,lambda to,keys:sent.extend(keys)))
            return sent[0]
        first=buy(20)
        creds={'email':'buyer@example.com','password':'long-password-123','license':first}
        self.assertTrue(self.request('/account/register',creds)['status'].startswith('201'))
        session=json.loads(self.request('/account/login',creds)['body'])['session']
        auth={'HTTP_AUTHORIZATION':'Bearer '+session}
        self.assertEqual(json.loads(self.request('/account/balance',{},auth)['body'])['remaining'],5)
        second=buy(21)
        self.assertEqual(json.loads(self.request('/account/redeem',{'license':second},auth)['body'])['remaining'],10)
        self.assertTrue(self.request('/account/redeem',{'license':second},auth)['status'].startswith('400'))
        for model in [
            {'model':'example','messages':[{'role':'user','content':'x'}],'max_tokens':2049},
            {'model':'example','messages':[{'role':'user','content':'x'}],'n':10},
            {'model':'example','messages':[{'role':'user','content':'x'*32769}]},
        ]:
            self.assertTrue(self.request('/v1/chat/completions',model,auth)['status'].startswith('400'))
        self.assertEqual(self.calls,[])
        self.assertEqual(json.loads(self.request('/account/balance',{},auth)['body'])['remaining'],10)

if __name__=='__main__':unittest.main()
