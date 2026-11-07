from concurrent.futures import ThreadPoolExecutor
from rate_limit import TokenBuckets,rate_limit_app

class Clock:
    now=0.0
    def __call__(self):return self.now

def app(env,start):start('200 OK',[]);return [b'ok']
def request(app,path='/health',method='GET',**extra):
    r={};env={'REQUEST_METHOD':method,'PATH_INFO':path};env.update(extra)
    def start(status,headers):r.update(status=status,headers=dict(headers))
    r['body']=b''.join(app(env,start));return r

def limited(service='commerce',clock=None,capacity=1):
    buckets=TokenBuckets({k:(1.0,capacity) for k in ('health','read','api')},clock or Clock())
    return rate_limit_app(app,service=service,buckets=buckets),buckets

def test_health_429_retry_after_and_refill():
    c=Clock();a,_=limited(clock=c)
    assert request(a)['status']=='200 OK'
    r=request(a);assert r['status'].startswith('429') and r['headers']['Retry-After']=='1'
    c.now=1;assert request(a)['status']=='200 OK'

def test_read_and_api_separate_buckets():
    a,_=limited()
    assert request(a,'/account/balance','POST')['status']=='200 OK'
    assert request(a,'/account/requests','POST')['status'].startswith('429')
    assert request(a,'/account/login','POST')['status']=='200 OK'
    assert request(a,'/account/register','POST')['status'].startswith('429')
    assert request(a)['status']=='200 OK'

def test_bucket_bounded_under_threads():
    c=Clock();b=TokenBuckets({k:(1.0,7) for k in ('health','read','api')},c)
    with ThreadPoolExecutor(max_workers=16) as pool:results=list(pool.map(lambda _:b.take('api')[0],range(200)))
    assert sum(results)==7 and len(b.state)==3

def test_clock_rollback_never_refills():
    c=Clock();a,_=limited(clock=c);assert request(a)['status']=='200 OK'
    c.now=-10;assert request(a)['status'].startswith('429')

def test_scheduler_oidc_post_bypasses_only_watchdog():
    a,b=limited('watchdog')
    for _ in range(100):assert request(a,'/internal/reconcile','POST',HTTP_AUTHORIZATION='Bearer scheduler-oidc-test')['status']=='200 OK'
    assert all(v[0]==1 for v in b.state.values())
    other,_=limited('commerce');request(other,'/internal/reconcile','POST')
    assert request(other,'/internal/reconcile','POST')['status'].startswith('429')

def test_spoofed_headers_do_not_bypass_general_limit():
    a,_=limited();request(a,'/account/login','POST')
    assert request(a,'/account/login','POST',HTTP_USER_AGENT='Google-Cloud-Scheduler',HTTP_AUTHORIZATION='Bearer anything')['status'].startswith('429')

def test_health_and_ready_share_bucket_and_methods_cannot_bypass():
    a,_=limited();request(a,'/health')
    assert request(a,'/ready','HEAD')['status'].startswith('429')

def test_invalid_limit_configuration_rejected():
    import pytest
    for bad in ({'api':(1,1)}, {k:(0,1) for k in ('health','read','api')}, {k:(float('nan'),1) for k in ('health','read','api')}):
        with pytest.raises(ValueError):TokenBuckets(bad)
