"""Safe local mock verifier; real projects default to plan-only."""
import argparse, json, os, subprocess, sys, tempfile
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
FIXTURE={
 'backend.py':'def health():\n    return {"ok": True, "source": "MOCK"}\n',
 'test_backend.py':'import unittest\nfrom backend import health\nclass TestHealth(unittest.TestCase):\n    def test_health(self): self.assertEqual(health(), {"ok":True,"source":"MOCK"})\n',
 'frontend/index.html':'<!doctype html><title>MOCK SaaS fixture</title><h1>MOCK only</h1>\n',
 'README.md':'MOCK fixture; no PostgreSQL/auth/RBAC/Docker/CI acceptance claimed.\n'}
REAL_PLAN=[['docker','compose','config','--quiet'],['docker','compose','build'],['docker','compose','up','-d','--wait'],['docker','compose','exec','-T','backend','python','-m','pytest','-q'],['docker','compose','exec','-T','frontend','npm','test','--','--run'],['docker','compose','exec','-T','frontend','npm','run','build']]
def limits():
    import resource
    resource.setrlimit(resource.RLIMIT_CPU,(15,15)); resource.setrlimit(resource.RLIMIT_AS,(512*1024*1024,512*1024*1024)); resource.setrlimit(resource.RLIMIT_FSIZE,(10*1024*1024,10*1024*1024))
def run_mock():
    result={'source':'MOCK','not_real_saas_acceptance':True,'steps':[]}
    with tempfile.TemporaryDirectory(prefix='perf-saas-fixture-') as d:
        p=Path(d)
        for name,text in FIXTURE.items(): (p/name).parent.mkdir(parents=True,exist_ok=True); (p/name).write_text(text)
        for cmd in ([sys.executable,'-m','compileall','-q','.'],[sys.executable,'-m','unittest','discover','-v']):
            env={'PATH':os.environ.get('PATH',''),'HOME':d,'PYTHONNOUSERSITE':'1','LANG':'C.UTF-8'}
            r=subprocess.run(cmd,cwd=d,env=env,capture_output=True,text=True,timeout=20,preexec_fn=limits)
            # Only fixed trusted fixture output; not arbitrary generated code/logs.
            result['steps'].append({'command':cmd[1:],'exit_code':r.returncode,'stdout':r.stdout,'stderr':r.stderr})
        result['pass']=all(s['exit_code']==0 for s in result['steps'])
    return result
if __name__=='__main__':
    a=argparse.ArgumentParser(); a.add_argument('--mock',action='store_true'); a.add_argument('--project'); args=a.parse_args()
    result=run_mock() if args.mock else {'source':'PLAN_ONLY_NOT_TESTED','project':args.project,'commands':REAL_PLAN,'warning':'Run only in an approved isolated disposable environment with budget/timeout limits; scripts never auto-start Docker/GPU or execute untrusted project code.'}
    out=ROOT/'results/saas_verification.json'; out.write_text(json.dumps(result,indent=2)); out.with_suffix('.md').write_text('# '+result['source']+' — SaaS verifier\n\n```json\n'+json.dumps(result,indent=2)+'\n```\n'); print(json.dumps({'source':result['source'],'pass':result.get('pass'),'report':str(out)}))
