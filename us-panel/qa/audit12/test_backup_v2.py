"""HOST v2 shell logic retest, synthetic temp paths only; no VPS/system secret reads."""
import hashlib,json,os,subprocess,tempfile,unittest
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]
SOURCE=ROOT/'deploy/host-hardening/backup/backup-us-license-hook-hourly.v2.sh'

class BackupV2(unittest.TestCase):
 def setUp(self):
  self.tmp=tempfile.TemporaryDirectory(prefix='qa-backup-v2-');self.addCleanup(self.tmp.cleanup)
  self.root=Path(self.tmp.name);(self.root/'locks').mkdir()
  source=SOURCE.read_text()
  # ONLY sandbox path/owner remapping; no backup, validation, exit or instance logic changed.
  source=source.replace('BASE=/opt/us-license-hook; ENVF=/etc/us-license-hook.env',f'BASE={self.root}/prod; ENVF={self.root}/prod.env')
  source=source.replace('BASE=/opt/us-license-hook-test; ENVF=/etc/us-license-hook-test.env',f'BASE={self.root}/test; ENVF={self.root}/test.env')
  source=source.replace('/var/backups/$NAME',str(self.root/'backups')+'/$NAME')
  source=source.replace('/run/$NAME-backup.lock',str(self.root/'locks')+'/$NAME-backup.lock')
  source=source.replace('-o root -g root',f'-o {os.getuid()} -g {os.getgid()}')
  self.script=self.root/'backup-v2.sh';self.script.write_text(source)
  for instance in ('prod','test'):
   d=self.root/instance;d.mkdir();(d/'licenses.json').write_text(json.dumps({'synthetic_instance':instance,'credits':10}))
   (d/'plans.json').write_text('{}');(self.root/(instance+'.env')).write_text('US_QA_SYNTHETIC_ONLY=1\n')
 def run_backup(self,instance):
  p=subprocess.run(['bash',str(self.script),instance],env={'PATH':os.environ['PATH']},capture_output=True,text=True,timeout=15)
  print(json.dumps({'case':self._testMethodName,'instance':instance,'exit_code':p.returncode}),flush=True)
  return p
 def test_missing_store_exit2(self):
  (self.root/'test/licenses.json').unlink();self.assertEqual(self.run_backup('test').returncode,2)
 def test_missing_plans_exit2(self):
  (self.root/'test/plans.json').unlink();self.assertEqual(self.run_backup('test').returncode,2)
 def test_invalid_store_json_exit3(self):
  (self.root/'test/licenses.json').write_text('{invalid');self.assertEqual(self.run_backup('test').returncode,3)
 def test_symlink_store_exit2(self):
  p=self.root/'test/licenses.json';p.unlink();p.symlink_to(self.root/'prod/licenses.json');self.assertEqual(self.run_backup('test').returncode,2)
 def test_unknown_instance_exit64(self):self.assertEqual(self.run_backup('other').returncode,64)
 def test_both_instances_separate_and_verified(self):
  destinations=[]
  for instance,name in [('prod','us-license-hook'),('test','us-license-hook-test')]:
   self.assertEqual(self.run_backup(instance).returncode,0)
   dest=self.root/'backups'/name;destinations.append(dest)
   backups=list(dest.glob('licenses.json.*'));self.assertEqual(len(backups),1)
   self.assertEqual(backups[0].read_bytes(),(self.root/instance/'licenses.json').read_bytes())
   self.assertEqual(json.loads(backups[0].read_text())['synthetic_instance'],instance)
   manifests=list(dest.glob('manifest.*.sha256'));self.assertEqual(len(manifests),1)
   manifest=manifests[0].read_text().splitlines();self.assertEqual(len(manifest),3)
   for row in manifest:
    digest,name=row.split(None,1);self.assertEqual(hashlib.sha256((dest/name.strip()).read_bytes()).hexdigest(),digest)
   self.assertTrue((dest/'LAST_SUCCESS').is_file())
  self.assertNotEqual(*destinations)
 def test_missing_optional_env_success(self):
  (self.root/'test.env').unlink();self.assertEqual(self.run_backup('test').returncode,0)

if __name__=='__main__':unittest.main()
