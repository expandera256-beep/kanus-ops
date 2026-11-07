import subprocess,json,pathlib,datetime,re,hashlib,collections
R=pathlib.Path(__file__).parent;P='project-27082a40-6083-4952-9db';now=datetime.datetime.now(datetime.timezone.utc);hi=now.isoformat().replace('+00:00','Z')
commands={
'revision':['run','revisions','describe','us-watchdog-00007-fnc','--region=europe-west4'],
'logs_after_0034':['logging','read','resource.type="cloud_run_revision" AND resource.labels.service_name="us-watchdog" AND timestamp>="2026-11-07T21:34:00Z" AND timestamp<="'+hi+'"','--order=asc','--limit=10000'],
'scheduler_after_0034':['logging','read','resource.type="cloud_scheduler_job" AND resource.labels.job_id="us-idle-watchdog" AND timestamp>="2026-11-07T21:34:00Z" AND timestamp<="'+hi+'"','--order=asc','--limit=1000']}
def clean(x):
 if isinstance(x,dict):
  if 'value' in x and any(k in x.get('name','').lower() for k in ['token','secret','dsn','password']):x=dict(x,value='[REDACTED]')
  return {k:clean(v) for k,v in x.items()}
 if isinstance(x,list):return [clean(v) for v in x]
 if isinstance(x,str):
  x=re.sub(r'postgres(?:ql)?://[^\s"\']+','[REDACTED_DSN]',x);x=re.sub(r'Bearer [A-Za-z0-9._~+/-]+','Bearer [REDACTED]',x)
 return x
for name,args in commands.items():
 cmd=['gcloud']+args+['--project='+P,'--format=json'];p=subprocess.run(cmd,capture_output=True,text=True,check=True,timeout=50);d=clean(json.loads(p.stdout));(R/(name+'.json')).write_text(json.dumps(d,indent=2));(R/(name+'.meta.json')).write_text(json.dumps({'command':cmd,'end_istanbul':now.astimezone().isoformat(),'returncode':p.returncode},indent=2))
a=json.loads((R/'logs_after_0034.json').read_text());rev=json.loads((R/'revision.json').read_text());image=json.loads((R/'live_image_read_evidence.json').read_text());base=pathlib.Path('/workspace/us-panel/sre/cloudrun/fixes/base-00007')
summary={'end_istanbul':now.astimezone().isoformat(),'revision_image_matches_registry_manifest':rev['status']['imageDigest'].endswith('@sha256:'+image['manifest_sha256']),'source_matches_release':{n:(R/('live_'+n)).read_bytes()==(base/n).read_bytes() for n in image['files']},'records':len(a),'http_5xx':sum(x.get('httpRequest',{}).get('status',0)>=500 for x in a),'LifecycleError_or_refusing':sum(any(t in json.dumps(x).lower() for t in ['lifecycleerror','refusing']) for x in a),'revision_00007_reconcile_http_status':dict(collections.Counter(str(x['httpRequest']['status']) for x in a if x['resource']['labels'].get('revision_name')=='us-watchdog-00007-fnc' and '/internal/reconcile' in x.get('httpRequest',{}).get('requestUrl','')))}
(R/'summary.json').write_text(json.dumps(summary,indent=2));print(json.dumps(summary,indent=2))
