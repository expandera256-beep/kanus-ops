#!/usr/bin/env python3
"""One final three-request probe: total task login cap15 including prior12."""
import hashlib,json,sys,time
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import QA,Results,request
LOGS=QA.parent/'reports/qa-20261106'
def main():
 r=Results();rows=[]
 try:
  for j in range(3):
   headers={} if j==0 else {'X-Forwarded-For':'192.0.2.'+str(20+j)}
   start=time.monotonic();code,raw,_=request('https://api-test.kanus.org','/account/login',{'email':'qa-acct-20261106@kanus.org','password':'x'},headers=headers,timeout=120)
   row={'label':'control' if j==0 else 'rotated_XFF_'+str(j),'http':code,'ms':round((time.monotonic()-start)*1000,3),'body_bytes':len(raw),'body_sha256':hashlib.sha256(raw).hexdigest(),'time':datetime.now(ZoneInfo('Europe/Istanbul')).isoformat()};rows.append(row);print(json.dumps(row),flush=True)
   if code>=500:
    r.add('XFF_rotation_limiter_bypass','BLOCKED',{'requests':rows,'reason':'Server error; immediate stop/no retries'});break
  else:
   passed=rows[0]['http']==429 and all(x['http']==429 for x in rows[1:])
   r.add('XFF_rotation_limiter_bypass','PASS' if passed else 'NOT TESTED',{'requests':rows,'reason':'No bypass of observed429' if passed else 'Baseline429 absent or time/burst confounded; cannot conclude XFF bypass. No escalation.','total_task_login_requests':15})
 except Exception as e:r.add('XFF_rotation_limiter_bypass','BLOCKED',{'type':type(e).__name__,'requests':rows,'reason':'Network error; immediate stop/no retries'})
 return r.save(LOGS/'accounts-xff-results.json')
if __name__=='__main__':sys.exit(main())
