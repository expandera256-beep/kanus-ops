import contextlib
import io
import json
from types import SimpleNamespace
from commerce_http import make_app
from runtime import make_watchdog_app
from readiness_checks import readiness

class Store:
    def __init__(self, fail=False): self.sql=[]; self.fail=fail
    @contextlib.contextmanager
    def db(self):
        if self.fail: raise RuntimeError('postgresql://private:password@host/db')
        yield self
    _db=db
    def execute(self, sql): self.sql.append(sql)

class Controller:
    def __init__(self, store): self.store=store; self.policies={'gpu':SimpleNamespace(project='test-project')}; self.calls=[]
    def reconcile(self, external=False): self.calls.append('reconcile'); return []
    def __getattr__(self, name): raise AssertionError('Forbidden lifecycle/provider access: '+name)

def get(app,path):
    result={}
    def start(status,headers):result.update(status=status,headers=dict(headers))
    result['body']=json.loads(b''.join(app({'REQUEST_METHOD':'GET','PATH_INFO':path,'wsgi.input':io.BytesIO()},start)))
    return result

def commerce(store,upstream=object()):return make_app(store,b'webhook',b'license','test.invalid',upstream)

def test_commerce_ready_200_and_only_select():
    s=Store();r=get(commerce(s),'/ready')
    assert r['status']=='200 OK' and r['body']=={'ready':True,'checks':{'config':True,'store':True}}
    assert s.sql==['SELECT 1']

def test_commerce_ready_503_sanitizes_driver_failure():
    r=get(commerce(Store(True)),'/ready')
    assert r['status'].startswith('503') and r['body']['checks']['store'] is False
    assert 'password' not in json.dumps(r)

def test_commerce_missing_config_503():
    assert get(commerce(Store(),None),'/ready')['status'].startswith('503')

def test_watchdog_ready_does_not_reconcile_or_touch_provider():
    s=Store();c=Controller(s);app=make_watchdog_app(c)
    for _ in range(3):assert get(app,'/ready')['status']=='200 OK'
    assert s.sql==['SELECT 1']*3 and c.calls==[]

def test_watchdog_failed_store_503():
    c=Controller(Store(True));assert get(make_watchdog_app(c),'/ready')['status'].startswith('503');assert c.calls==[]

def test_watchdog_no_policy_503():
    c=Controller(Store());c.policies={}
    assert get(make_watchdog_app(c),'/ready')['status'].startswith('503')

def test_health_does_not_probe_db_or_heartbeat():
    s=Store();c=Controller(s)
    assert get(commerce(s),'/health')['status']=='200 OK'
    assert get(make_watchdog_app(c),'/health')['status']=='200 OK'
    assert s.sql==[] and c.calls==[]

def test_readyz_not_alias_or_gpu_poll():
    s=Store();c=Controller(s)
    assert get(commerce(s),'/readyz')['status'].startswith('405')
    assert get(make_watchdog_app(c),'/readyz')['status'].startswith('404')
    assert s.sql==[] and c.calls==[]
