#!/usr/bin/env python3
"""One baseline+one spoof burst,5 concurrent each; >=60s cooldown. Short invalid password avoids bcrypt load."""
import argparse,concurrent.futures,hashlib,json,sys,threading,time
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import QA,Results,request
LOGS=QA.parent/'reports/qa-20261106';TZ=ZoneInfo('Europe/Istanbul')
def main():
 p=argparse.ArgumentParser();p.add_argument('--mode',choices=('baseline','spoof'),required=True);p.add_argument('--repeat-baseline',action='store_true');a=p.parse_args();r=Results()
 label=a.mode+('-repeat' if a.repeat_baseline else '')
 outfile=LOGS/('accounts-burst-'+label+'-results.json')
 if outfile.exists():print('BLOCKED: named burst already executed; no automatic repeat',flush=True);return 2
 history=[]
 for name in ('baseline','baseline-repeat','spoof'):
  f=LOGS/('accounts-burst-'+name+'-results.json')
  if f.is_file():history.append(json.loads(f.read_text()))
 if a.repeat_baseline and (a.mode!='baseline' or not any(x.get('mode')=='baseline' for x in history)):
  print('BLOCKED: repeat requires prior baseline',flush=True);return 2
 if a.mode=='spoof' and not any(x.get('mode')=='baseline' for x in history):
  print('BLOCKED: baseline required',flush=True);return 2
 if any(x.get('safety_stop') for x in history):print('BLOCKED: prior server/network error; no further bursts',flush=True);return 2
 gap=time.monotonic()-max((x['finished_monotonic'] for x in history),default=0)
 if history and gap<60:print('BLOCKED: cooldown remaining_seconds='+str(round(60-gap,3)),flush=True);return 2
 barrier=threading.Barrier(6);origin=time.monotonic();started=datetime.now(TZ).isoformat()
 def call(index):
  headers={}
  if a.mode=='spoof':
   ip='203.0.113.'+str(20+index);headers={'X-Forwarded-For':ip,'X-Real-IP':ip}
  barrier.wait();start=time.monotonic()
  row={'index':index,'start_offset_ms':round((start-origin)*1000,3)}
  try:
   code,raw,h=request('https://api-test.kanus.org','/account/login',{'email':'qa-acct-20261106@kanus.org','password':'x'},headers=headers,timeout=120)
   row.update(http=code,ms=round((time.monotonic()-start)*1000,3),body_bytes=len(raw),body_sha256=hashlib.sha256(raw).hexdigest(),content_type=h.get('Content-Type',h.get('content-type','')))
  except Exception as e:row.update(http=0,error_type=type(e).__name__,ms=round((time.monotonic()-start)*1000,3))
  print(json.dumps(row),flush=True);return row
 with concurrent.futures.ThreadPoolExecutor(max_workers=5) as pool:
  futures=[pool.submit(call,i) for i in range(5)];barrier.wait();rows=[f.result() for f in futures]
 safety=any(x['http']==0 or x['http']>=500 for x in rows)
 blocked=sum(x['http']==429 for x in rows);allowed=sum(x['http']==400 for x in rows)
 stats={'mode':label,'started_tsi':started,'finished_tsi':datetime.now(TZ).isoformat(),'finished_monotonic':time.monotonic(),'cooldown_before_seconds':round(gap,3) if history else None,'requests':rows,'concurrency':5,'count':5,'password_kind':'invalid-length fixture, not actual credential; allowed400 avoids bcrypt','safety_stop':safety,'http429':blocked,'http400':allowed,'start_spread_ms':round(max(x['start_offset_ms'] for x in rows)-min(x['start_offset_ms'] for x in rows),3)}
 if safety:r.add('burst_safety','BLOCKED',{'reason':'Network/5xx error; no more bursts allowed','requests':rows})
 else:
  r.check(label+'_no_wrong_password_session',not any(x['http']==200 for x in rows),{'http':[x['http'] for x in rows]})
  if a.mode=='baseline':r.add('targeted_burst_limiter','PASS' if blocked>=2 else 'NOT TESTED',{'http429':blocked,'http400':allowed,'expected':'At least2 overflow429 for5 clients with burst2; no conclusion if source-IP buckets vary'})
  else:
   control=next((x for x in reversed(history) if x['mode'].startswith('baseline')),None)
   proven=control and control.get('http429',0)>=2
   if proven and blocked>=2:r.add('XFF_XRealIP_spoof_burst','PASS',{'baseline429':control['http429'],'spoof429':blocked,'spoof400':allowed,'scope':'Both client headers rotated; limiter overflow still enforced; per-IP identity not independently logged'})
   elif proven and blocked==0:r.add('XFF_XRealIP_spoof_burst','NOT TESTED',{'baseline429':control['http429'],'spoof429':blocked,'reason':'Differential possible bypass, but observed source-IP buckets/config unknown; controlled baseline repeat/log review needed, no exploit claim'})
   else:r.add('XFF_XRealIP_spoof_burst','NOT TESTED',{'baseline429':control.get('http429') if control else None,'spoof429':blocked,'reason':'No matched overflow control or insufficient429; no bypass conclusion'})
 stats['results']=r.rows;stats['counts']={s:sum(x['status']==s for x in r.rows) for s in ('PASS','FAIL','BLOCKED','NOT TESTED')}
 outfile.write_text(json.dumps(stats,ensure_ascii=False,indent=2));print('SUMMARY '+json.dumps(stats['counts']),flush=True)
 return int(stats['counts']['FAIL']>0)
if __name__=='__main__':sys.exit(main())
