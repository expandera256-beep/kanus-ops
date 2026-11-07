import tempfile,threading,unittest
from pathlib import Path
from inference_lifecycle import WorkerPolicy,LifecycleStore,LifecycleController,LifecycleError

class Clock:
    def __init__(self):self.value=100000
    def __call__(self):return self.value
class Prov:
    """Cold start: RUNNING after boot_delay seconds; ready when RUNNING."""
    def __init__(self,clock,boot_delay=420):
        self.clock=clock;self.status='TERMINATED';self.t0=None;self.boot=boot_delay;self.starts=[];self.stops=[]
    def describe(self,p):
        if self.status=='STARTING' and self.clock()-self.t0>=self.boot:self.status='RUNNING'
        return {'id':p.instance_id,'status':'STAGING' if self.status=='STARTING' else self.status,'started_at':None}
    def model_ready(self,p):return self.status=='RUNNING'
    def start(self,p,op):self.starts.append(op);self.status='STARTING';self.t0=self.clock()
    def stop(self,p,op):self.stops.append(op);self.status='TERMINATED'

class ReadyWindow(unittest.TestCase):
    def mk(self,boot=420,**kw):
        self.tmp=tempfile.TemporaryDirectory();self.addCleanup(self.tmp.cleanup)
        self.clock=Clock();self.prov=Prov(self.clock,boot)
        args=dict(idle_seconds=600,boot_seconds=600,request_seconds=600,lease_seconds=60,max_uptime_seconds=1800,daily_runtime_seconds=3600,daily_starts=6,require_watchdog=False);args.update(kw)
        self.p=WorkerPolicy('gpu','m','pr','z','ds-a100','42','http://x',**args)
        self.store=LifecycleStore(Path(self.tmp.name)/'s.db')
        def tick(n):self.clock.value+=max(n,5)
        self.c=LifecycleController(self.store,self.prov,[self.p],lambda r:True,self.clock,tick)
    def test_cold_start_gets_full_window_from_ready(self):
        self.mk(420);_,lease=self.c.acquire('r1','m');taken=self.clock()
        self.assertEqual(lease['deadline'],taken+600)
        out=self.c.ensure_ready(self.p,'r1');ready=self.clock()
        self.assertGreaterEqual(ready-taken,420)
        self.assertEqual(out['deadline'],ready+600);self.assertEqual(out['ready_at'],ready)
        self.assertEqual(self.store.view('gpu')['leases']['r1']['deadline'],ready+600)
    def test_warm_request_window_unchanged_length(self):
        self.mk(0);self.prov.status='RUNNING';s=self.store
        self.c.acquire('w','m');self.c._observe(self.p,self.prov.describe(self.p))
        self.store.change('gpu',lambda st:st.update(managed=True))
        t=self.clock();out=self.c.ensure_ready(self.p,'w');self.assertEqual(out['deadline'],self.clock()+600);self.assertLessEqual(out['deadline']-t,600)
    def test_boot_still_capped_at_boot_seconds(self):
        self.mk(10**6);self.c.acquire('r','m');t=self.clock()
        with self.assertRaises(LifecycleError):self.c.ensure_ready(self.p,'r')
        self.assertLessEqual(self.clock()-t,600+10)
    def test_window_never_passes_max_uptime(self):
        self.mk(420,max_uptime_seconds=1000);self.c.acquire('r','m');out=self.c.ensure_ready(self.p,'r')
        rs=self.store.view('gpu')['run_since'];self.assertLessEqual(out['deadline'],rs+1000-30)
    def test_window_never_passes_daily_budget(self):
        self.mk(420,daily_runtime_seconds=900);self.c.acquire('r','m');out=self.c.ensure_ready(self.p,'r')
        s=self.store.view('gpu');self.assertLessEqual(out['deadline']-out['ready_at'],900-s['day_seconds']+1)
        self.assertGreaterEqual(out['deadline'],out['created']+0)
    def test_rebase_only_once(self):
        self.mk(420);self.c.acquire('r','m');a=self.c.ensure_ready(self.p,'r');self.clock.value+=30
        b=self.c.ensure_ready(self.p,'r');self.assertEqual(a['deadline'],b['deadline'])
    def test_waiter_through_boot_also_rebased(self):
        self.mk(100,lease_seconds=120);self.c.acquire('r1','m');self.clock.value+=10;self.c.acquire('r2','m')
        a=self.c.ensure_ready(self.p,'r1');b=self.c.ensure_ready(self.p,'r2')
        self.assertGreaterEqual(b['deadline'],a['ready_at']+600-0);self.assertEqual(len(self.prov.starts),1)
    def test_lease_still_expires_without_heartbeat(self):
        self.mk(420);self.c.acquire('r','m');self.c.ensure_ready(self.p,'r');self.clock.value+=61
        with self.assertRaises(LifecycleError):self.c.heartbeat(self.p,'r')
    def test_idle_stop_after_release_unchanged(self):
        self.mk(420);self.c.acquire('r','m');self.c.ensure_ready(self.p,'r');self.c.release(self.p,'r')
        self.clock.value+=599;self.c.reconcile();self.assertEqual(self.prov.stops,[])
        self.clock.value+=2;self.c.reconcile();self.assertEqual(len(self.prov.stops),1)

class CommerceCap(unittest.TestCase):
    def test_managed_response_carries_ready_marker(self):
        import managed_inference,inspect
        src=inspect.getsource(managed_inference.ManagedResponse)
        self.assertIn('ready_monotonic',src)
if __name__=='__main__':unittest.main()
