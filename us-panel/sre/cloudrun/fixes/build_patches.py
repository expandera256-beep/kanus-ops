from pathlib import Path
import difflib,shutil,json,subprocess
F=Path(__file__).parent
base={p.name:p.read_text() for p in (F/'base-00007').glob('*.py') if not p.name.startswith('test_')}
final={p.name:p.read_text() for p in (F/'app-00007').glob('*.py')}

def diff(a,b,path):
 text=''.join(''.join(difflib.unified_diff(a.get(k,'').splitlines(keepends=True),b.get(k,'').splitlines(keepends=True),fromfile='a/'+k,tofile='b/'+k)) for k in sorted(set(a)|set(b)) if a.get(k)!=b.get(k))
 (F/path).parent.mkdir(parents=True,exist_ok=True);(F/path).write_text(text)

s1=dict(base);s1['commerce_http.py']=final['commerce_http.py'];s1['readiness_checks.py']=final['readiness_checks.py']
s=base['runtime.py'].replace('from lifecycle_config import configure_managed','from lifecycle_config import configure_managed\nfrom readiness_checks import readiness')
start=final['runtime.py'].split("        elif env.get('REQUEST_METHOD')=='GET' and env.get('PATH_INFO')=='/ready':",1)[1].split("        elif env.get('REQUEST_METHOD')=='POST'",1)[0]
s=s.replace("        elif env.get('REQUEST_METHOD')=='POST' and env.get('PATH_INFO')=='/internal/reconcile':", "        elif env.get('REQUEST_METHOD')=='GET' and env.get('PATH_INFO')=='/ready':"+start+"        elif env.get('REQUEST_METHOD')=='POST' and env.get('PATH_INFO')=='/internal/reconcile':")
s1['runtime.py']=s
s2=dict(s1);s2['rate_limit.py']=final['rate_limit.py'];s=s1['runtime.py'].replace('from readiness_checks import readiness','from readiness_checks import readiness\nfrom rate_limit import rate_limit_app')
old="    return make_app(store,webhook.encode(),license_secret.encode(),require('SHOPIFY_SHOP'),upstream,products,webhooks_enabled=enabled=='true')"
s=s.replace(old,"    app=make_app(store,webhook.encode(),license_secret.encode(),require('SHOPIFY_SHOP'),upstream,products,webhooks_enabled=enabled=='true')\n    return rate_limit_app(app, service='commerce')")
s=s.replace('    return app\n\n\ndef watchdog_app():',"    return rate_limit_app(app, service='watchdog')\n\n\ndef watchdog_app():")
s2['runtime.py']=s
diff(base,s1,'readiness.patch');diff(s1,s2,'rate-limit.patch');diff(s2,final,'structured-logging.patch');diff(base,final,'combined.patch')
# Independent per-service patches, each applied to a clean baseline, never stacked.
commerce=dict(final)
# Keep watchdog factory exactly unchanged while instrumenting only gateway.
commerce['runtime.py']=final['runtime.py'].split('def make_watchdog_app(',1)[0]+'def make_watchdog_app('+base['runtime.py'].split('def make_watchdog_app(',1)[1]
watchdog=dict(final);watchdog['commerce_http.py']=base['commerce_http.py']
# Keep gateway factory exactly unchanged while instrumenting only watchdog.
# Replace only the gateway function body, retain shared instrument helper.
w=final['runtime.py'];head=w.split('def gateway_app():',1)[0];tail='def instrument('+w.split('def instrument(',1)[1]
gateway='def gateway_app():'+base['runtime.py'].split('def gateway_app():',1)[1].split('def make_watchdog_app(',1)[0]
watchdog['runtime.py']=head+gateway+tail
service_trees={'us-commerce':commerce,'us-watchdog':watchdog}
for service,tree in service_trees.items():diff(base,tree,service+'/local-acceptance.patch')
checks=[]
def apply_check(name,patches,target):
 d=F/name
 if d.exists():shutil.rmtree(d)
 d.mkdir()
 for p in (F/'base-00007').glob('*.py'):
  if not p.name.startswith('test_'):shutil.copyfile(p,d/p.name)
 steps=[]
 for patch in patches:
  r=subprocess.run(['patch','--batch','-p1','-i',str(F/patch)],cwd=d,capture_output=True,text=True)
  steps.append({'patch':patch,'returncode':r.returncode,'stdout':r.stdout,'stderr':r.stderr})
 assert all(x['returncode']==0 for x in steps),steps
 matches={k:(d/k).read_text()==v for k,v in target.items()}
 assert all(matches.values()),matches
 checks.append({'tree':name,'steps':steps,'matches':matches})
apply_check('patch_apply_check',['readiness.patch','rate-limit.patch','structured-logging.patch'],final)
for service,tree in service_trees.items():apply_check(service+'-apply-check',[service+'/local-acceptance.patch'],tree)
(F/'patch_apply_check.json').write_text(json.dumps(checks,indent=2))
print('PASS: combined ordered patches and both independent service patches')
