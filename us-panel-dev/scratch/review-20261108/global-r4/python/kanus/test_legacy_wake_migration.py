import unittest
from legacy_wake_migration import guarded_source


class LegacyGuardTest(unittest.TestCase):
    def test_migrated_reads_cannot_start_resume_or_stop_workers(self):
        source='''def ensure(worker):
    if worker_awake(worker):return
    instance_status(worker)
    power(worker, 'start')
'''
        called=[];scope={'worker_awake':lambda _:True,'power':lambda *args:called.append(args),'instance_status':lambda _:called.append('read')}
        exec(guarded_source(source),scope);scope['ensure']({})
        scope['worker_awake']=lambda _:False
        with self.assertRaises(TimeoutError):scope['ensure']({})
        self.assertEqual(called,[])

    def test_unknown_revision_is_not_replaced(self):
        with self.assertRaises(ValueError):guarded_source('def ensure(worker): return None\n')


if __name__=='__main__':unittest.main()
