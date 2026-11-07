#!/usr/bin/env python3
"""Scratch overlay executes actual panel agent paths without modifying product files."""
import argparse,hashlib,json,os,subprocess,sys,tempfile
from contextlib import ExitStack
from pathlib import Path
sys.path.insert(0,str(Path(__file__).resolve().parents[1]))
from common import QA,SOURCE,Results,cleanenv,local_hook,mock_model
from fixtures import FILES,SERVER
HERE=Path(__file__).resolve().parent
class DeterministicModel:
 def __init__(self):self.phase=0;self.events=[]
 def __call__(self,payload):
  messages=json.dumps(payload.get('messages',[]),ensure_ascii=False)
  if self.phase==0:
   ops=[{'op':'create_file','path':p,'content':c} for p,c in FILES.items()]
   self.phase=1;self.events.append('generate-three-files')
  elif self.phase==1 and 'QA_REPAIR' in messages:
   ops=[{'op':'read_file','path':'server.py'}];self.phase=2;self.events.append('read-server-before-patch')
  elif self.phase==2:
   ops=[{'op':'patch_file','path':'server.py','sha256':hashlib.sha256(SERVER.encode()).hexdigest(),'before':"{'ok': False}",'after':"{'ok': True}"}]
   self.phase=3;self.events.append('patch-failing-health')
  else:
   raise RuntimeError('unexpected mock protocol stage')
  return json.dumps({'version':1,'message':'QA deterministic mock fixture','operations':ops})
def main():
 p=argparse.ArgumentParser();p.add_argument('--real',action='store_true');p.add_argument('--logs',default=str(QA.parent/'reports/qa-20261106'));a=p.parse_args()
 logs=Path(a.logs);logs.mkdir(parents=True,exist_ok=True);r=Results()
 if a.real and (os.environ.get('US_REAL_MODEL_E2E')!='1' or os.environ.get('US_PROXY_URL')!='https://api-test.kanus.org'):
  r.add('real_gate','BLOCKED','--real requires US_REAL_MODEL_E2E=1 and US_PROXY_URL=https://api-test.kanus.org');return r.save(logs/'e2e9-results.json')
 try:
  with ExitStack() as stack:
   model=DeterministicModel()
   if a.real:
    base=os.environ['US_PROXY_URL'];keyfile=Path(os.environ.get('US_TEST_LICENSE','/workspace/secrets/qa-license-a.txt'))
    if not keyfile.is_file():raise FileNotFoundError('private license missing')
    modelname=os.environ.get('US_REAL_MODEL_NAME','deepseek-v4.1-flash')
   else:
    upstream,handler=stack.enter_context(mock_model(model))
    base,mint,work=stack.enter_context(local_hook(upstream,logs/'e2e9-local-hook'))
    keyfile=work/'panel-license.txt';keyfile.write_text(mint(99001));keyfile.chmod(0o600);modelname='qa-mock'
   artifact=Path(tempfile.mkdtemp(prefix='e2e9-real-' if a.real else 'e2e9-mock-',dir=QA/'build'))
   overlay=artifact/'overlay.json'
   overlay.write_text(json.dumps({'Replace':{str(SOURCE/'zz_qa_scratch_panel_test.go'):str(HERE/'panel_driver_test.go')}}))
   env=cleanenv();env.update(QA_PANEL_PROXY=base,QA_PANEL_LICENSE_FILE=str(keyfile),QA_MODEL_NAME=modelname,QA_PANEL_ARTIFACT_DIR=str(artifact/'generated'),US_REAL_MODEL_E2E='1' if a.real else '0')
   cmd=['go','test','-overlay',str(overlay),'.','-run','^TestQAScratchPanel','-count=1','-v','-timeout=30m' if a.real else '-timeout=210s']
   print('CMD: go test -overlay <qa/build overlay.json> . -run ^TestQAScratchPanel -count=1 -v; mode='+('REAL' if a.real else 'MOCK'),flush=True)
   result=subprocess.run(cmd,cwd=SOURCE,env=env,capture_output=True,text=True,timeout=1900 if a.real else 240)
   (logs/'e2e9-panel-driver.log').write_text(result.stdout+result.stderr+'\nexit_code='+str(result.returncode)+'\n')
   r.check('actual_panel_agent_generate_build_repair_tests',result.returncode==0,{'exit_code':result.returncode,'log':'e2e9-panel-driver.log','artifact_dir':str(artifact/'generated'),'mode':'REAL' if a.real else 'MOCK'})
   if not a.real:
    r.check('mock_openai_protocol_and_feedback',model.events==['generate-three-files','read-server-before-patch','patch-failing-health'],{'events':model.events,'upstream_calls':handler.calls})
   if result.returncode==0:
    result=subprocess.run(['python3',str(HERE/'acceptance.py'),str(artifact/'generated')],capture_output=True,text=True,timeout=20)
    (logs/'e2e9-independent-acceptance.log').write_text(result.stdout+result.stderr+'\nexit_code='+str(result.returncode)+'\n')
    r.check('independent_generated_service_http_acceptance',result.returncode==0,{'exit_code':result.returncode,'log':'e2e9-independent-acceptance.log'})
   r.add('full_section9_platform','NOT TESTED','Small stdlib REST fixture only; no frontend/PostgreSQL/Docker/CI production acceptance.')
   if not a.real:r.add('real_model_generation','NOT TESTED','Opt-in not run; requires model tunnel/readiness and explicit cost budget. Mock success is not real model success.')
 except Exception as e:r.add('e2e9_harness_exception','FAIL',{'type':type(e).__name__,'secret_values':'suppressed'})
 return r.save(logs/'e2e9-results.json')
if __name__=='__main__':sys.exit(main())
