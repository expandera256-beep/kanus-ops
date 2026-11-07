import threading, time, unittest, usta_wake as U
class T(unittest.TestCase):
    def setUp(s):
        s.st={'ds-a100':'TERMINATED'}; s.calls=[]; s.l=threading.Lock()
        U.instance_status=lambda w: s.st[w['instance']]
        def power(w,verb):
            time.sleep(0.05)
            with s.l: s.calls.append((w['instance'],verb)); s.st[w['instance']]='STAGING'
        U.power=power; U.stop_other_gpus=lambda w: None
    def test_U01_power_verbs(s):
        s.assertEqual([U.power_verb(x) for x in ('TERMINATED','RUNNING','STAGING','STOPPING','SUSPENDED')],['start','','','','resume'])
    def test_U02_concurrent_start_single_power(s):
        for _ in range(8): U.a100_start_async()
        time.sleep(1); s.assertEqual(s.calls,[('ds-a100','start')])
    def test_U03_start_while_stopping_is_noop(s):
        s.st['ds-a100']='STOPPING'; U.a100_start_async(); time.sleep(0.3); s.assertEqual(s.calls,[])
    def test_U04_a100_discards_local_ssd_flag(s):
        s.assertTrue(U.WORKERS['a100']['discard_local_ssd'])
if __name__=='__main__': unittest.main(verbosity=2)
