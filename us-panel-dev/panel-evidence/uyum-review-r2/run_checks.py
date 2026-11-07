import subprocess,os,pathlib,json,time
r=pathlib.Path('/tmp/uyum-r2-sqm3eaol');t=r/'a/b/term';ev=pathlib.Path('/workspace/us-panel-dev/panel-evidence/uyum-review-r2');env=os.environ|dict(GOPROXY='off',GOSUMDB='off',US_REAL_MODEL_TEST='0',US_REAL_MODEL_E2E='0',US_DOOR_PROBE='0',US_REA_NETWORK_TEST='0',US_TASK_STATE_KEY_FILE=str(r/'state.key'))
s=[]
for name,cmd,extra in [('vet',['go','vet','./...'],{}),('race',['go','test','-race','-json','./...'],{}),('win-vet',['go','vet','./...'],{'GOOS':'windows'}),('win-build',['go','build','./...'],{'GOOS':'windows'}),('new-count3',['go','test','-race','-json','./licensesrv','./internal/privacy','-run','Privacy|Erase|Retention|Pref|Handler','-count=3'],{}),('gofmt',['gofmt','-l','internal/privacy','licensesrv/privacy_ops.go','licensesrv/privacy_ops_test.go','cmd/us-license-hook/main.go'],{})]:
 with (ev/(name+'.stdout')).open('w') as o,(ev/(name+'.stderr')).open('w') as e:
  st=time.monotonic();a=subprocess.run(cmd,cwd=t,env=env|extra,stdout=o,stderr=e,timeout=900)
 s.append({'name':name,'command':cmd,'exit':a.returncode,'seconds':round(time.monotonic()-st,2)});(ev/'checks.json').write_text(json.dumps(s,indent=2));print(s[-1],flush=True)
