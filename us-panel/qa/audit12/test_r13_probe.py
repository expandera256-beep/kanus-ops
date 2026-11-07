#!/usr/bin/env python3
import errno,unittest,contextlib,io,json,tempfile
from pathlib import Path
from unittest.mock import patch,MagicMock
import regress_r13
class ProbeTests(unittest.TestCase):
 def test_not_linux(self):
  with patch.object(regress_r13.platform,'system',return_value='Windows'):
   r=regress_r13.landlock_probe();self.assertFalse(r['supported']);self.assertIn('Not a Linux',r['reason'])
 def test_unknown_arch_is_not_pass(self):
  with patch.object(regress_r13.platform,'system',return_value='Linux'),patch.object(regress_r13.platform,'machine',return_value='unmapped'):
   self.assertFalse(regress_r13.landlock_probe()['supported'])
 def probe(self,n,err):
  libc=MagicMock();libc.syscall.return_value=n
  with patch.object(regress_r13.platform,'system',return_value='Linux'),patch.object(regress_r13.platform,'machine',return_value='x86_64'),patch.object(regress_r13.ctypes,'CDLL',return_value=libc),patch.object(regress_r13.ctypes,'get_errno',return_value=err):return regress_r13.landlock_probe()
 def test_enosys_is_skip_reason(self):
  r=self.probe(-1,errno.ENOSYS);self.assertFalse(r['supported']);self.assertIn('ENOSYS',r['reason'])
 def test_seccomp_denial_is_not_pass(self):
  r=self.probe(-1,errno.EPERM);self.assertFalse(r['supported']);self.assertIn('EPERM',r['reason'])
 def test_abi_is_not_boundary_acceptance(self):
  r=self.probe(6,0);self.assertTrue(r['supported']);self.assertEqual(r['abi'],6);self.assertIn('boundary still must pass',r['reason'])
 def test_unsupported_kernel_suite_row_is_skip_not_pass(self):
  with tempfile.TemporaryDirectory(prefix='qa-r13-probe-test-') as tmp:
   stdout=io.StringIO()
   with patch.object(regress_r13,'ROOT',Path(tmp)),patch.object(regress_r13,'landlock_probe',return_value={'supported':False,'reason':'simulated ENOSYS'}),patch.object(regress_r13.sys,'argv',['regress_r13.py','--tree',tmp,'--dry-run']),contextlib.redirect_stdout(stdout):
    self.assertEqual(regress_r13.main(),0)
   rows=[]
   for line in stdout.getvalue().splitlines():
    try:rows.append(json.loads(line))
    except ValueError:pass
   row=next(x for x in rows if x.get('name')=='F01_linux_landlock')
   self.assertEqual(row['status'],'SKIP');self.assertIn('ENOSYS',row['reason'])
   self.assertFalse(any(x.get('status')=='PASS' for x in rows))
if __name__=='__main__':unittest.main()
