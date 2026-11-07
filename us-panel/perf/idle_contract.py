"""Policy v5 logical-clock mock, not live GPU verification."""
import json
from pathlib import Path
IDLE_SECONDS=600

def may_stop(now,last_activity,active=0,queued=0):return active==0 and queued==0 and now-last_activity>=IDLE_SECONDS
checks=[('599_not_idle',not may_stop(599,0)),('600_idle',may_stop(600,0)),('active_prevents_stop',not may_stop(601,0,active=1)),('queued_prevents_stop',not may_stop(601,0,queued=1)),('release_resets_timer',not may_stop(601,601)),('new_request_cancels_stop',not may_stop(700,700)),('last_work_plus600_idle',may_stop(1300,700))]
r={'source':'MOCK_LOGICAL_CLOCK','idle_seconds':600,'policy':'v5','decision':'Demir 2026-11-08 00:20 TSİ','live_gpu_tested':False,'checks':[{'name':n,'pass':p} for n,p in checks]}
root=Path(__file__).parent
for p in [root/'results/idle_contract.json',Path('/workspace/us-panel/reports/r20261108/results/idle_policy_v5.json')]:
 p.write_text(json.dumps(r,ensure_ascii=False,indent=2));p.with_suffix('.md').write_text('# Policy v5 —600sn mock\n\n'+json.dumps(r,ensure_ascii=False,indent=2)+'\n')
print(json.dumps({'policy':'v5','idle_seconds':600,'passed':sum(p for _,p in checks),'total':len(checks)}))
