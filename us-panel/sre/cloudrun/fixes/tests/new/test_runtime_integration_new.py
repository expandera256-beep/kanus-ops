import contextlib
import io
import json
from types import SimpleNamespace
import runtime

class Store:
    @contextlib.contextmanager
    def db(self):yield self
    _db=db
    def execute(self,sql):assert sql=='SELECT 1'

class Controller:
    def __init__(self):self.store=Store();self.policies={'gpu':SimpleNamespace(project='test-project')};self.calls=0
    def reconcile(self,external=False):self.calls+=1;assert external;return []

def call(app,path,method='GET'):
    result={}
    def start(s,h):result.update(status=s,headers=dict(h))
    result['body']=b''.join(app({'REQUEST_METHOD':method,'PATH_INFO':path,'CONTENT_LENGTH':'0','wsgi.input':io.BytesIO()},start));return result

def test_real_watchdog_factory_keeps_scheduler_exempt_and_logs(capsys):
    c=Controller();app=runtime.make_watchdog_app(c)
    for _ in range(100):assert call(app,'/internal/reconcile','POST')['status']=='200 OK'
    assert c.calls==100
    logs=[json.loads(x) for x in capsys.readouterr().out.splitlines()]
    assert len(logs)==100 and all(x['httpRequest']['status']==200 for x in logs)
    assert all(x['logging.googleapis.com/trace'].startswith('projects/test-project/traces/') for x in logs)

def test_gateway_factory_installs_ready_limiter_and_logging(monkeypatch,capsys):
    c=Controller();settings={'US_WORKERS_CONFIG':'not-read','US_LICENSE_SECRET':'x'*32,'US_CREDIT_PRODUCTS':'{"123":{"prompts":1,"price_minor":100,"valid_days":90,"currency":"TRY"}}','SHOPIFY_SHOP':'test.invalid'}
    monkeypatch.setenv('US_SHOPIFY_ENABLED','false');monkeypatch.setattr(runtime,'require',settings.__getitem__)
    monkeypatch.setattr(runtime,'configured_store',lambda:Store())
    monkeypatch.setattr(runtime,'configure_managed',lambda *a:(object(),c))
    app=runtime.gateway_app();assert call(app,'/ready')['status']=='200 OK'
    responses=[call(app,'/health')['status'] for _ in range(61)]
    assert any(s.startswith('429') for s in responses)
    logs=[json.loads(x) for x in capsys.readouterr().out.splitlines()]
    assert logs[0]['logging.googleapis.com/trace'].startswith('projects/test-project/traces/')
    assert logs[-1]['severity']=='WARNING' and logs[-1]['httpRequest']['status']==429
