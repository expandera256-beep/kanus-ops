#!/usr/bin/env python3
"""Read only already-delivered redacted metadata; never access cloud, secret values or token contents."""
import argparse,json,stat
from pathlib import Path
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--evidence',type=Path,required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--assert-safe',action='store_true',help='Negative safety regression: exit1 while policy findings remain');a=p.parse_args()
if a.output.exists():p.error('Evidence output already exists; choose a new path')
def load(n):return json.loads((a.evidence/n).read_text())
secrets=load('secrets_list.json')
if not a.assert_safe:assert len(secrets)==6
rotation=[x['name'].rsplit('/',1)[-1] for x in secrets if not x.get('rotation')]
versions=load('ver_us-worker-policy.json');enabled=sorted(x['name'].rsplit('/',1)[-1] for x in versions if x['state']=='ENABLED')
if not a.assert_safe:assert enabled==['1','3','4','5']
bindings=[]
for s in secrets:
 name=s['name'].rsplit('/',1)[-1];iam=load('iam_'+name+'.json')
 for b in iam.get('bindings',[]):
  assert b['role']=='roles/secretmanager.secretAccessor'
  assert all(m.startswith('serviceAccount:us-gateway@') or m.startswith('serviceAccount:us-watchdog@') for m in b['members'])
  bindings.append({'secret':name,'sa_short_names':[m.split(':',1)[1].split('@')[0] for m in b['members']]})
refs=[]
for svc in load('run_services.json'):
 spec=svc['spec']['template']['spec']
 for c in spec.get('containers',[]):
  for e in c.get('env',[]):
   ref=e.get('valueFrom',{}).get('secretKeyRef')
   if ref:refs.append({'service':svc['metadata']['name'],'secret':ref['name'],'version':ref['key'],'binding':'env'})
 for v in spec.get('volumes',[]):
  for it in v.get('secret',{}).get('items',[]):refs.append({'service':svc['metadata']['name'],'secret':v['secret']['secretName'],'version':it['key'],'binding':'volume'})
if not a.assert_safe:assert refs and any(x['version']=='latest' for x in refs)
vm=[]
for x in load('instances.json'):
 for sa in x.get('serviceAccounts',[]):
  vm.append({'instance':x['name'],'compute_default_sa':sa['email'].endswith('-compute@developer.gserviceaccount.com'),'cloud_platform_scope':'https://www.googleapis.com/auth/cloud-platform' in sa.get('scopes',[])})
if not a.assert_safe:assert all(any(x['instance']==name and x['compute_default_sa'] and x['cloud_platform_scope'] for x in vm) for name in ['ds-a100','usta-wake'])
role=load('role_ustaWake.json');permissions=sorted(role['includedPermissions']);assert all(x.startswith('compute.instances.') for x in permissions)
unused=next((x for x in load('sas.json') if x['email'].startswith('sre-oidc-negtest@')),None)
if not a.assert_safe:assert unused and not unused.get('disabled',False)
project=load('project_iam.json');assert not any(any(m.startswith('serviceAccount:sre-oidc-negtest@') for m in b.get('members',[])) for b in project.get('bindings',[]))
old=Path('/workspace/secrets/usta-wake/a100.token.old');s=old.lstat() if old.exists() else None
if not a.assert_safe:assert s and stat.S_ISREG(s.st_mode)
result={'provenance':'SRE-delivered redacted report/raw metadata, locally cross-checked; not independent live GCP measurement','secret_count':len(secrets),'no_rotation_metadata':rotation,'worker_policy_enabled_versions':enabled,'secret_access_bindings':bindings,'run_bindings':refs,'vm_scope_evidence':vm,'compute_custom_role_permissions':permissions,'sre_oidc_negtest_enabled_no_project_role':True,'us_builder_unused':'NOT VERIFIED','old_token':{'path':str(old),'exists':s is not None,'mode':oct(stat.S_IMODE(s.st_mode)) if s else None,'uid':s.st_uid if s else None,'gid':s.st_gid if s else None,'contents_read':False,'validity':'NOT VERIFIED'},'hook_live_env_permissions':'NOT VERIFIED: HOST evidence needed; source docs600/640 contradictory','secret_manager_assessment':'FAIL: reported configuration/policy gaps; no credential compromise demonstrated'}
issues=[]
if any(name in rotation for name in ['us-license-signing','us-model-token','us-postgres-dsn']):issues.append({'id':'rotation_metadata','severity':'MEDIUM'})
if any(x['version']=='latest' for x in refs) or any(x in enabled for x in ['1','3','4']):issues.append({'id':'version_pinning_policy_retention','severity':'LOW'})
if any(x['compute_default_sa'] and x['cloud_platform_scope'] for x in vm):issues.append({'id':'vm_identity_scope','severity':'MEDIUM'})
if unused and not unused.get('disabled',False):issues.append({'id':'enabled_designated_negative_test_sa','severity':'LOW'})
if s is not None:issues.append({'id':'old_token_residue','severity':'MEDIUM','validity':'NOT VERIFIED'})
result['security_findings']=issues;result['secret_manager_assessment']='FAIL' if issues else 'PASS within metadata-only scope'
a.output.write_text(json.dumps(result,ensure_ascii=False,indent=2)+'\n');print(json.dumps({'status':'FAIL' if a.assert_safe and issues else 'PASS','scope':'negative security policy regression' if a.assert_safe else 'raw-metadata consistency assertions, not security acceptance','output':str(a.output),'secret_count':len(secrets),'old_token_contents_read':False}))

if a.assert_safe:raise SystemExit(1 if issues else 0)
