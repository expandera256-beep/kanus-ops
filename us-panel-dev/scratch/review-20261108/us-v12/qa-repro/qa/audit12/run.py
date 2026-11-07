#!/usr/bin/env python3
"""One reproducible case per §12 item. Local default; remote requires --remote.
No product source changes. Expected failing regressions stay FAIL, never xfail.
"""
import argparse,json,os,re,subprocess,sys,tempfile
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo
ROOT=Path(__file__).resolve().parents[2];HERE=Path(__file__).resolve().parent
sys.path.insert(0,str(ROOT/'qa'))
from common import cleanenv,request
SOURCE=ROOT/'work2/kiyas/term';LOGS=ROOT/'reports/r20261108/qa-raw'
CASES={
 'authentication':('./licensesrv','^TestAccountSignedSessionExpiryResetAndVerification$'),
 'api_tokens':('./licensesrv','^Test(ModelProxyJSONAuthBoundsAndMetadata|AccountNoCreditAndInvalidSessionNeverCallsModel)$'),
 'hmac':('./licensesrv','^Test(RejectsBadHMAC|ReversalHMACShopTopicAndInvalidPayload)$'),
 'rate_limit':('./licensesrv','^Test(AccountLimiter(RotatingIPs|SuccessResetsAndWindow|MemoryBound)|QAAuditIPOnlyRateDesign)$'),
 'privilege_escalation':('./licensesrv','^TestCreditHTTPAuthorizationAndIdempotentSettlement$'),
 'sensitive_logs':('./licensesrv','^Test(PrivacyOutboxNoRawKeyLogsNoEmail|PrivacyMaskHelpers|GPUWakeAuthErrorAndTimeoutAndRedaction)$'),
 'command_injection':('.','^TestQAAuditCommandInjection$'),
 'path_traversal':('.','^TestQAAuditPathTraversal$'),
 'file_isolation':('.','^TestQAAuditExecSiblingIsolation$'),
 'ssrf':('.','^TestQAAuditRepairRedirectIsolation$'),
 'sqli':('python','DataSecurity.test_sql_injection'),
 'backup':('v2','HOST-v2-all-synthetic-cases'),
 'restore':('python','DataSecurity.test_backup_restore'),
 'secret_manager':('static','source-binding-and-path-only-scan'),
}

def run_case(name):
 LOGS.mkdir(parents=True,exist_ok=True)
 target,pattern=CASES[name];cwd=SOURCE
 if target=='v2':
  cmd=['python3',str(HERE/'test_backup_v2.py'),'-v'];cwd=ROOT
 elif target=='python':
  cmd=['python3',str(HERE/'test_data_security.py'),pattern,'-v'];cwd=ROOT
 elif target=='static':
  cmd=['python3',str(HERE/'scan_secrets.py'),'--output',str(LOGS/'secret-scan.json')];cwd=ROOT
 else:
  cmd=['go','test']
  if name=='rate_limit':
   overlay=LOGS/'store-overlay.json';overlay.write_text(json.dumps({'Replace':{str(SOURCE/'licensesrv/zz_qa_audit_store_test.go'):str(HERE/'store_backup_test.go')}}))
   cmd+=['-overlay',str(overlay)]
  if target=='.':
   overlay=LOGS/'panel-overlay.json';overlay.write_text(json.dumps({'Replace':{str(SOURCE/'zz_qa_audit12_test.go'):str(HERE/'panel_audit_test.go')}}))
   cmd+=['-overlay',str(overlay)]
  cmd+=[target,'-run',pattern,'-count=1','-v','-timeout=90s']
 try:
  env=cleanenv()
  if name=='sensitive_logs':
   # Existing privacy test invokes only DryRunSender/fakeSender; allow its synthetic fixture.
   env['US_SMTP_ALLOWLIST']='a@example.com'
  p=subprocess.run(cmd,cwd=cwd,env=env,capture_output=True,text=True,timeout=130)
  log=LOGS/(name+'.log');log.write_text(p.stdout+p.stderr+'\nexit_code='+str(p.returncode)+'\n')
  status='PASS' if p.returncode==0 else 'FAIL'
  if name=='restore' and p.returncode==0:
   store_overlay=LOGS/'store-overlay.json'
   store_overlay.write_text(json.dumps({'Replace':{str(SOURCE/'licensesrv/zz_qa_audit_store_test.go'):str(HERE/'store_backup_test.go')}}))
   store=subprocess.run(['go','test','-overlay',str(store_overlay),'./licensesrv','-run','^TestQAAuditGoStoreBackupRestore$','-v','-count=1','-timeout=30s'],cwd=SOURCE,env=env,capture_output=True,text=True,timeout=60)
   (LOGS/'go-store-backup-restore.log').write_text(store.stdout+store.stderr+'\nexit_code='+str(store.returncode)+'\n')
   if store.returncode:status='FAIL'
  if name=='secret_manager':
   # Source checks never import/execute the GCP deploy module or access secrets.
   main=(SOURCE/'cmd/us-license-hook/main.go').read_text()
   config=(ROOT/'work2/kanus/deploy/gateway.service.yaml').read_text()
   checks={'hook_env_loading':'os.Getenv("US_LICENSE_SECRET")' in main,
           'secret_manager_source_binding':'secretKeyRef' in config,
           'scan_executed':p.returncode==0}
   (LOGS/'secret-manager-source-check.json').write_text(json.dumps(checks,indent=2)+'\n')
   status='BLOCKED' if all(checks.values()) else 'FAIL'
  if target not in ('python','v2','static') and '[no tests to run]' in (p.stdout+p.stderr):status='NOT TESTED'
  row={'item':name,'status':status,'log':str(log.relative_to(ROOT)),'exit_code':p.returncode,'scope':'Local synthetic tests only; Secret Manager production IAM/rotation evidence remains BLOCKED'}
 except subprocess.TimeoutExpired:
  row={'item':name,'status':'BLOCKED','reason':'Local test timeout; no retry or fabricated result'}
 print(json.dumps(row,ensure_ascii=False),flush=True);return row

def remote():
 from precheck import check
 gate=check(LOGS/'new-host-precheck.json')
 if gate['status']!='PASS':
  rows=[{'name':'new_host_precheck','status':'BLOCKED','reason':'DNS/TLS peer/health precondition not met; zero negative probes, no fallback'}]
  (LOGS/'remote-negative.json').write_text(json.dumps({'results':rows,'max_parallel':1,'http_requests':gate['http_requests']},indent=2)+'\n')
  print(json.dumps(rows[0]));return rows
 rows=[]
 probes=[('health','/healthz',None,None,None,{200}),
  ('missing_auth','/v1/credits/status',None,None,None,{401}),
  ('fake_license','/v1/models',None,'us1.qa.invalid',None,{401}),
  ('fake_session','/account/balance',{},'qa-invalid-session',None,{401}),
  ('missing_hmac','/webhooks/shopify/orders-paid',{},None,None,{401}),
  ('wrong_hmac','/webhooks/shopify/orders-paid',{},None,{'X-Shopify-Hmac-Sha256':'aW52YWxpZA=='},{401}),
  ('traversal_gate','/v1/%2e%2e/%2e%2e/qa-synthetic',None,None,None,{400,401,404}),
  ('command_gate','/v1/credits/status?command=qa%3Becho%20synthetic',None,None,None,{401})]
 for name,path,body,token,headers,expected in probes:
  try:
   code,_,_=request('https://api-test.kanus.org',path,body,token,headers,timeout=120)
   row={'name':name,'http':code,'status':'PASS' if code in expected else 'FAIL','time':datetime.now(ZoneInfo('Europe/Istanbul')).isoformat(),'scope':'Negative gate only; not a full filesystem/terminal/SSRF proof'}
   rows.append(row)
   if code>=500:break
  except Exception as e:
   rows.append({'name':name,'status':'BLOCKED','type':type(e).__name__,'message':'No raw exception/URL/body; stop without retry'});break
 (LOGS/'remote-negative.json').write_text(json.dumps({'results':rows,'max_parallel':1,'hard_cap':8,'secrets_read':False},ensure_ascii=False,indent=2)+'\n')
 print(json.dumps({'remote_requests':len(rows),'scope':'Non-destructive negative probes, max1,120s timeout'},ensure_ascii=False))
 return rows

def main():
 global SOURCE,LOGS
 p=argparse.ArgumentParser(description=__doc__);p.add_argument('--case',choices=CASES);p.add_argument('--remote',action='store_true');p.add_argument('--tree',type=Path,default=SOURCE,help='Go module tree to test; Python fixtures and shared scans remain QA-local');a=p.parse_args()
 SOURCE=a.tree.resolve()
 if not (SOURCE/'go.mod').is_file():p.error('--tree must name a Go module directory')
 parent=LOGS/'runs';parent.mkdir(parents=True,exist_ok=True)
 stamp=datetime.now(ZoneInfo('Europe/Istanbul')).strftime('%Y%m%dT%H%M%S_%f%z')
 LOGS=Path(tempfile.mkdtemp(prefix=stamp+'-',dir=parent))
 (LOGS/'run-context.json').write_text(json.dumps({'tree':str(SOURCE),'timezone':'Europe/Istanbul','created_at':datetime.now(ZoneInfo('Europe/Istanbul')).isoformat(),'case':a.case,'remote':a.remote},indent=2)+'\n')
 print('RUN_DIR '+str(LOGS),flush=True)
 LOGS.mkdir(parents=True,exist_ok=True)
 if a.remote:
  rows=remote();return 1 if any(x['status']=='FAIL' for x in rows) else 2 if any(x['status']=='BLOCKED' for x in rows) else 0
 rows=[run_case(n) for n in ([a.case] if a.case else CASES)]
 out={'results':rows,'counts':{s:sum(r['status']==s for r in rows) for s in ('PASS','FAIL','BLOCKED','NOT TESTED')},'scope':'Local automated case outcomes; production backup/restore and cloud IAM require separate read-only evidence'}
 (LOGS/('case-'+a.case+'-results.json' if a.case else 'all-cases-results.json')).write_text(json.dumps(out,ensure_ascii=False,indent=2)+'\n')
 print('SUMMARY '+json.dumps(out['counts']));return int(any(r['status']=='FAIL' for r in rows))
if __name__=='__main__':sys.exit(main())
