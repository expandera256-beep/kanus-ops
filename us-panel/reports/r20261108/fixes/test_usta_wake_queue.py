import json,tempfile,threading,time,unittest,urllib.request,urllib.error,os
from importlib.machinery import SourceFileLoader
from http.server import ThreadingHTTPServer
SRC='/workspace/us-panel/kan/usta-wake/usta-wake.py.queue'
class Q(unittest.TestCase):
    def setUp(self):
        self.m=m=SourceFileLoader('uw%d'%id(self),SRC).load_module()
        d=tempfile.mkdtemp();open(d+'/tok','w').write('newtok');open(d+'/old','w').write('oldtok');open(d+'/grace','w').close()
        m.A100_TOKEN_FILE=d+'/tok';m.A100_TOKEN_OLD_FILE=d+'/old';m.A100_GRACE_FLAG=d+'/grace'
        m.A100_ALLOWED_IPS={'127.0.0.1'}
        self.status=['STOPPING'];self.calls=[]
        m.instance_status=lambda w:self.status[0] if w['id']=='a100' else 'TERMINATED'
        def power(w,v):self.calls.append((w['id'],v));self.status[0]='STAGING' if v=='start' else self.status[0]
        m.power=power;m.stop_other_gpus=lambda w:None;m.worker_awake=lambda w:False
        m.PENDING_POLL=0.02;m.PENDING_LIMIT=1.0
        self.srv=ThreadingHTTPServer(('127.0.0.1',0),m.A100TlsDoor);threading.Thread(target=self.srv.serve_forever,daemon=True).start()
        self.addCleanup(self.srv.shutdown)
    def req(self,path,method='GET',tok='newtok'):
        r=urllib.request.Request('http://127.0.0.1:%d%s'%(self.srv.server_port,path),method=method,data=b'' if method=='POST' else None,headers={'Authorization':'Bearer '+tok} if tok else {})
        try:
            with urllib.request.urlopen(r,timeout=5) as x:return x.status,json.loads(x.read())
        except urllib.error.HTTPError as e:return e.code,json.loads(e.read() or b'{}')
    def wait(self,cond,t=3):
        end=time.time()+t
        while time.time()<end:
            if cond():return True
            time.sleep(.01)
        return False
    def test_stopping_start_is_queued_then_fired_on_terminated(self):
        c,d=self.req('/a100/start','POST');self.assertEqual((c,d['action'],d['pending_start']),(202,'start_queued',True))
        c,d=self.req('/a100/status');self.assertTrue(d['pending_start']);self.assertIsNotNone(d['pending_since'])
        time.sleep(.1);self.assertEqual(self.calls,[])
        self.status[0]='TERMINATED'
        self.assertTrue(self.wait(lambda:self.calls==[('a100','start')]))
        self.assertTrue(self.wait(lambda:not self.m.a100_pending()['pending_start']))
        c,d=self.req('/a100/status');self.assertFalse(d['pending_start'])
    def test_at_most_one_pending(self):
        a=self.req('/a100/start','POST')[1];b=self.req('/a100/start','POST')[1];c=self.req('/a100/start','POST')[1]
        self.assertEqual([a['action'],b['action'],c['action']],['start_queued','start_already_queued','start_already_queued'])
        self.status[0]='TERMINATED';self.assertTrue(self.wait(lambda:len(self.calls)>=1));time.sleep(.2)
        self.assertEqual(self.calls,[('a100','start')])
    def test_pending_is_bounded(self):
        self.req('/a100/start','POST');time.sleep(1.3)
        self.assertFalse(self.m.a100_pending()['pending_start']);self.assertEqual(self.calls,[])
    def test_pending_clears_if_someone_else_started(self):
        self.req('/a100/start','POST');self.status[0]='RUNNING'
        self.assertTrue(self.wait(lambda:not self.m.a100_pending()['pending_start']));self.assertEqual(self.calls,[])
    def test_terminated_start_unchanged(self):
        self.status[0]='TERMINATED';c,d=self.req('/a100/start','POST');self.assertEqual((c,d['action']),(202,'start_requested'))
        self.assertTrue(self.wait(lambda:self.calls==[('a100','start')]))
    def test_running_start_noop(self):
        self.status[0]='RUNNING';c,d=self.req('/a100/start','POST');self.assertEqual((c,d['action']),(200,'none'))
    def test_auth_unchanged(self):
        self.assertEqual(self.req('/a100/start','POST',tok='bad')[0],403)
        self.assertEqual(self.req('/a100/start','POST',tok=None)[0],403)
        self.assertFalse(self.m.a100_pending()['pending_start'])
        self.assertEqual(self.req('/a100/start','POST',tok='oldtok')[0],202)  # grace still on
        os.remove(self.m.A100_GRACE_FLAG);self.m.PENDING['since']=0.0
        self.assertEqual(self.req('/a100/start','POST',tok='oldtok')[0],403)
    def test_ip_allowlist_unchanged(self):
        self.m.A100_ALLOWED_IPS={'31.220.63.191','82.29.184.225'}
        self.assertEqual(self.req('/a100/start','POST')[0],403);self.assertFalse(self.m.a100_pending()['pending_start'])
unittest.main()
