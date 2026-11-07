#!/usr/bin/env python3
"""Local-only §12 R13 regression. FAIL is never xfail; unsupported kernel is explicit SKIP."""
import argparse,ctypes,errno,json,os,platform,re,subprocess,sys,tempfile,shutil
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'qa'));from common import cleanenv

def landlock_probe():
 info={'platform':platform.system(),'arch':platform.machine(),'kernel':platform.release(),'supported':False}
 if platform.system()!='Linux':return dict(info,reason='Not a Linux kernel')
 if platform.machine().lower() not in ('x86_64','amd64','aarch64','arm64','riscv64'):return dict(info,reason='Syscall mapping not validated on this architecture')
 libc=ctypes.CDLL(None,use_errno=True);libc.syscall.restype=ctypes.c_long;ctypes.set_errno(0)
 abi=libc.syscall(444,ctypes.c_void_p(None),ctypes.c_size_t(0),ctypes.c_uint(1));err=ctypes.get_errno()
 if abi<1:return dict(info,errno=err,reason='Landlock ABI query unavailable or sandbox-blocked: '+errno.errorcode.get(err,str(err)))
 return dict(info,supported=True,abi=abi,reason='Kernel ABI query succeeded; actual child boundary still must pass')

def main():
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--tree',required=True,type=Path);p.add_argument('--dry-run',action='store_true');a=p.parse_args();tree=a.tree.resolve()
 if not a.dry_run and not (tree/'go.mod').is_file():p.error('--tree must point to source Go module, not an artifact-only directory')
 parent=ROOT/'reports/r20261108/qa-raw/r13-runs';parent.mkdir(parents=True,exist_ok=True)
 out=Path(tempfile.mkdtemp(prefix=datetime.now(ZoneInfo('Europe/Istanbul')).strftime('%Y%m%dT%H%M%S%z')+'-',dir=parent));print('RUN_DIR '+str(out),flush=True)
 probe=landlock_probe();version='unknown'
 if (tree/'release.go').is_file():
  m=re.search(r'const buildVersion = "([^"]+)"',(tree/'release.go').read_text());version=m.group(1) if m else 'unknown'
 rows=[]
 def add(name,status,**kw):
  row=dict(name=name,status=status,**kw);rows.append(row);print(json.dumps(row,ensure_ascii=False),flush=True)
 def case_cli(name,case):
  if a.dry_run:return add(name,'PLANNED',command=['run.py','--tree',str(tree),'--case',case])
  if not shutil.which('go'):return add(name,'NOT TESTED',reason='Native Go toolchain unavailable')
  cmd=[sys.executable,str(HERE/'run.py'),'--tree',str(tree),'--case',case]
  try:r=subprocess.run(cmd,cwd=ROOT,capture_output=True,text=True,timeout=180)
  except subprocess.TimeoutExpired:return add(name,'BLOCKED',reason='Local case timeout, no retry')
  log=out/(name+'.log');log.write_text(r.stdout+r.stderr+'\nexit_code='+str(r.returncode)+'\n')
  run=next((Path(x[8:]) for x in r.stdout.splitlines() if x.startswith('RUN_DIR ')),None)
  result=run/('case-'+case+'-results.json') if run else None
  if not result or not result.is_file():return add(name,'FAIL',reason='Case produced no structured evidence; never assume PASS',log=str(log))
  detail=json.loads(result.read_text());row=detail['results'][0]
  add(name,row['status'],log=str(log),case_evidence=str(result),scope='Local synthetic case; no public endpoint')
 def go_case(name,target,test,fixture,skip_status='SKIP'):
  if a.dry_run:return add(name,'PLANNED',test=test,fixture=fixture)
  if not shutil.which('go'):return add(name,'NOT TESTED',reason='Native Go toolchain unavailable')
  source=tree/('licensesrv' if target=='./licensesrv' else '')/('zz_qa_r13_'+name+'_test.go')
  ov=out/(name+'-overlay.json');ov.write_text(json.dumps({'Replace':{str(source):str(HERE/fixture)}}))
  env=cleanenv();env['US_GPU_WAKE_DISABLE']='1'
  if platform.system()=='Windows':
   for key in ('SYSTEMROOT','SystemRoot','COMSPEC','TEMP','TMP','USERPROFILE'):
    if key in os.environ:env[key]=os.environ[key]
  try:r=subprocess.run(['go','test','-p=1','-overlay',str(ov),target,'-run','^'+test+'$','-json','-count=1','-timeout=90s'],cwd=tree,env=env,capture_output=True,text=True,timeout=150)
  except subprocess.TimeoutExpired:return add(name,'BLOCKED',reason='Local Go test timeout, no retry')
  log=out/(name+'.jsonl');log.write_text(r.stdout+r.stderr+'\nexit_code='+str(r.returncode)+'\n');actions=[];messages=[]
  for line in r.stdout.splitlines():
   try:event=json.loads(line)
   except ValueError:continue
   if event.get('Test')==test:
    if event.get('Action') in ('pass','fail','skip'):actions.append(event['Action'])
    if event.get('Output'):messages.append(event['Output'].strip())
  if 'fail' in actions or r.returncode:return add(name,'FAIL',log=str(log),test=test,reason='Safety regression failed or build/harness error; inspect raw evidence')
  if 'skip' in actions:return add(name,skip_status,reason='; '.join(messages),log=str(log))
  if 'pass' not in actions:return add(name,'NOT TESTED',reason='Expected test did not run; zero-test exit0 is not PASS',log=str(log))
  add(name,'PASS',log=str(log),test=test)
 case_cli('F02_ssrf','ssrf')
 case_cli('F04_account_limiter','rate_limit')
 go_case('F07_account_lock_dos','./licensesrv','TestQAAuditLimiterAccountLockDoS','limiter_risks_test.go')
 if not probe['supported']:add('F01_linux_landlock','SKIP',reason=probe['reason'],kernel_probe=probe)
 else:go_case('F01_linux_landlock','.','TestQAR13LinuxLandlockSiblingRead','r13_linux_boundary_test.go')
 if platform.system()!='Windows':add('Windows_low_il_write_block','NOT TESTED',reason='Native Windows runtime unavailable; cross-compilation is not runtime proof')
 else:go_case('Windows_low_il_write_block','.','TestQAR13WindowsLowILWriteBlock','r13_windows_boundary_test.go',skip_status='NOT TESTED')
 counts={s:sum(x['status']==s for x in rows) for s in ('PASS','FAIL','SKIP','NOT TESTED','BLOCKED','PLANNED')}
 status='DRY RUN' if a.dry_run else 'FAIL' if counts['FAIL'] else 'PARTIAL' if any(counts[s] for s in ('SKIP','NOT TESTED','BLOCKED')) else 'PASS'
 result={'suite':'r13','created_at':datetime.now(ZoneInfo('Europe/Istanbul')).isoformat(),'tree':str(tree),'source_version':version,'status':status,'counts':counts,'landlock_probe':probe,'cases':rows,'remote_requests':0,'authenticated_public_model_calls':0,'max_case_parallelism':1,'note':'F04 preserves current lock contract; F07 intentionally exposes its DoS consequence. API policy decision required; no known failure is hidden.'}
 (out/'summary.json').write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print('SUMMARY '+json.dumps(counts));print('SUITE_STATUS '+status)
 return 0 if a.dry_run or status=='PASS' else 1 if status=='FAIL' else 2
if __name__=='__main__':sys.exit(main())
