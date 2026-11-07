import importlib, os, tempfile, time, unittest
def load(name):
    m=importlib.import_module(name); importlib.reload(m); d=tempfile.mkdtemp()
    m.READY=d+'/ready'; m.ACTIVITY=d+'/act'; m.NGINX_LOG=d+'/access.log'; m.NGINX_OFFSET=d+'/off'
    m.stops=[]; m.stop=lambda *,discard: m.stops.append(discard)
    m._health=True; m.health_ok=lambda: m._health; m._up=600.0; m.uptime=lambda: m._up
    m._ss=''; import subprocess
    m.subprocess=type('S',(),{'check_output':staticmethod(lambda *a,**k: m._ss),'CalledProcessError':subprocess.CalledProcessError})
    open(m.NGINX_LOG,'w').close(); return m
def ago(m,sec): t=time.time()-sec; os.utime(m.ACTIVITY,(t,t))
def log(m,line):
    with open(m.NGINX_LOG,'a') as f: f.write(line+'\n')
class T(unittest.TestCase):
    def test_I01_constants(self):
        m=load('idle600'); self.assertEqual((m.IDLE,m.FAIL_STOP),(600,4200))
    def test_I02_ready_marks_activity_no_stop_on_first_tick(self):
        m=load('idle600'); m.main(); self.assertTrue(os.path.exists(m.READY)); self.assertEqual(m.stops,[])
    def test_I03_599s_no_stop(self):
        m=load('idle600'); m.main(); ago(m,599); m.main(); self.assertEqual(m.stops,[])
    def test_I04_601s_stop_discard(self):
        m=load('idle600'); m.main(); ago(m,601); m.main(); self.assertEqual(m.stops,[True])
    def test_I05_model_request_resets(self):
        m=load('idle600'); m.main(); ago(m,601); log(m,'1.2.3.4 - - [x] "POST /v1/chat/completions HTTP/1.1" 200'); m.main(); self.assertEqual(m.stops,[])
    def test_I06_health_poll_does_not_reset(self):
        m=load('idle600'); m.main(); ago(m,601); log(m,'127.0.0.1 - - [x] "GET /health HTTP/1.1" 200'); m.main(); self.assertEqual(m.stops,[True])
    def test_I07_inbound_client_counts(self):
        m=load('idle600'); m.main(); ago(m,601); m._ss='0 0 127.0.0.1:80 127.0.0.1:51234\n'; m.main(); self.assertEqual(m.stops,[])
    def test_I08_outbound_metadata_not_counted(self):
        m=load('idle600'); m.main(); ago(m,601); m._ss='0 0 10.164.0.6:51234 169.254.169.254:80\n'; m.main(); self.assertEqual(m.stops,[True])
    def test_I09_not_ready_no_idle_stop(self):
        m=load('idle600'); m._health=False; m._up=4000; m.main(); self.assertEqual(m.stops,[])
    def test_I10_health_fail_after_70min_stops(self):
        m=load('idle600'); m._health=False; m._up=4201; m.main(); self.assertEqual(m.stops,[True])
    def test_I11_log_rotation(self):
        m=load('idle600'); log(m,'x'*500); m.main(); open(m.NGINX_LOG,'w').close(); log(m,'"GET /v1/models HTTP/1.1"'); ago(m,601); m.main(); self.assertEqual(m.stops,[])
    def test_I12_old_15min_would_not_stop_at_601(self):
        m=load('idle900'); m.main(); ago(m,601); m.main(); self.assertEqual(m.stops,[])
    def test_I13_ss_failure_counts_as_busy(self):
        m=load('idle600'); m.main(); ago(m,601)
        def boom(*a,**k): raise OSError()
        m.subprocess.check_output=staticmethod(boom); m.main(); self.assertEqual(m.stops,[])
    def test_I14_health_drop_while_ready_no_idle_stop_until_70min(self):
        m=load('idle600'); m.main(); ago(m,3000); m._health=False; m._up=3000; m.main(); self.assertEqual(m.stops,[])
if __name__=='__main__': unittest.main(verbosity=2)
