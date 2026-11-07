import io
import json
import os
import unittest
from unittest.mock import patch
import runtime

class GoIssuerPolicyTest(unittest.TestCase):
    def configured(self):
        values={'US_WORKERS_CONFIG':'/unused/policy','US_LICENSE_SECRET':'fixture-license-secret-32-characters',
                'US_CREDIT_PRODUCTS':json.dumps({'42':{'prompts':1,'price_minor':14900,'valid_days':90,'currency':'TRY'}}),
                'SHOPIFY_SHOP':'kanus-test.myshopify.com'}
        return patch.object(runtime,'require',side_effect=lambda key:values[key])
    def test_python_cannot_reenable_shopify_issuer(self):
        with patch.dict(os.environ,{'US_SHOPIFY_ENABLED':'true'},clear=True),self.configured(),patch.object(runtime,'configured_store',return_value=object()),patch.object(runtime,'configure_managed',return_value=(object(),None)):
            with self.assertRaisesRegex(RuntimeError,'sole license issuer'):
                runtime.gateway_app()
    def test_python_default_webhook_is_disabled(self):
        with patch.dict(os.environ,{},clear=True),self.configured(),patch.object(runtime,'configured_store',return_value=object()),patch.object(runtime,'configure_managed',return_value=(object(),None)):
            app=runtime.gateway_app();status=[]
            body=b''.join(app({'REQUEST_METHOD':'POST','PATH_INFO':'/webhooks/shopify','CONTENT_LENGTH':'2','wsgi.input':io.BytesIO(b'{}')},lambda s,h:status.append(s)))
            self.assertEqual(status,['503 Service Unavailable'])
            self.assertIn(b'not configured',body)
