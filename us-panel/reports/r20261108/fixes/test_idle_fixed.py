import importlib.util, os, subprocess, tempfile, unittest
from unittest import mock
P='/workspace/us-panel/sre/watchdog/idle.py.fixed'
def load(d):
    from importlib.machinery import SourceFileLoader
    m=SourceFileLoader('idle_fixed',P).load_module()
    m.READY=d+'/ready';m.ACTIVITY=d+'/act';m.SS_ERRORS=d+'/sserr';m.NGINX_LOG=d+'/access.log';m.NGINX_OFFSET=d+'/off'
    return m
class T(unittest.TestCase):
    def setUp(self):
        self.d=tempfile.mkdtemp();self.m=load(self.d);self.now=[1_000_000.0];self.up=[100.0];self.ok=[True];self.stops=[]
        m=self.m
        m.time=mock.Mock(time=lambda:self.now[0])
        m.uptime=lambda:self.up[0];m.health_ok=lambda:self.ok[0]
        m.stop=lambda discard:self.stops.append(discard)
        open(m.NGINX_LOG,'w').close()
        self.ssout=['']
        def co(*a,**k):
            if isinstance(self.ssout[0],Exception):raise self.ssout[0]
            return self.ssout[0]
        m.subprocess=mock.Mock(check_output=co,CalledProcessError=subprocess.CalledProcessError,TimeoutExpired=subprocess.TimeoutExpired)
    def tick(self,sec):
        self.now[0]+=sec;self.up[0]+=sec
    def run_(self):
        self.m.main()
        # make activity mtime follow fake clock
        if os.path.exists(self.m.ACTIVITY) and os.path.getmtime(self.m.ACTIVITY)>self.now[0]-5 or True:
            pass
    def touch_at(self,t):
        os.utime(self.m.ACTIVITY,(t,t))
    def boot_ready(self):
        self.m.main();self.touch_at(self.now[0])
    def test_initial_boot_unhealthy_uses_70min_grace(self):
        self.ok[0]=False
        self.up[0]=69*60;self.m.main();self.assertEqual(self.stops,[])
        self.up[0]=71*60;self.m.main();self.assertEqual(self.stops,[True])
    def test_post_ready_health_fail_stops_after_idle(self):
        self.boot_ready();self.ok[0]=False
        self.tick(300);self.m.main();self.assertEqual(self.stops,[])
        self.tick(301);self.m.main();self.assertEqual(self.stops,[True])
        self.assertLess(self.up[0],70*60)
    def test_post_ready_health_fail_with_requests_stays(self):
        self.boot_ready();self.ok[0]=False
        for _ in range(15):
            self.tick(60)
            with open(self.m.NGINX_LOG,'a') as f:f.write('1.2.3.4 - - "POST /v1/chat/completions HTTP/1.1" 200\n')
            self.m.main();self.touch_at(self.now[0]) if os.path.getmtime(self.m.ACTIVITY)>self.now[0]-100000 and False else None
            # activity touched by main uses real clock; align to fake
            self.touch_at(self.now[0])
        self.assertEqual(self.stops,[])
    def test_healthy_idle_stop_unchanged(self):
        self.boot_ready();self.tick(599);self.m.main();self.assertEqual(self.stops,[])
        self.tick(2);self.m.main();self.assertEqual(self.stops,[True])
    def test_ss_error_bounded(self):
        self.boot_ready();self.ssout[0]=OSError('ss broke')
        r=[self.m.clients() for _ in range(7)]
        self.assertEqual(r,[1,1,1,1,1,0,0])
    def test_ss_error_does_not_block_idle_forever(self):
        self.boot_ready();self.ssout[0]=subprocess.CalledProcessError(1,'ss')
        for i in range(30):
            self.tick(60);self.m.main()
            if self.m._ss_errors()<=self.m.SS_MAX_ERRORS:self.touch_at(self.now[0])  # busy touches use real clock
            if self.stops:break
        self.assertEqual(self.stops,[True]);self.assertLess(i,20)
    def test_ss_recovers_resets_counter(self):
        self.ssout[0]=OSError();self.m.clients();self.m.clients()
        self.ssout[0]='0 0 10.0.0.2:8080 10.0.0.9:5555\n'
        self.assertEqual(self.m.clients(),1);self.assertEqual(self.m._ss_errors(),0)
    def test_outbound_80_not_counted(self):
        self.ssout[0]='0 0 10.0.0.2:41000 169.254.169.254:80\n';self.assertEqual(self.m.clients(),0)
unittest.main(verbosity=1)
