import pathlib,subprocess,json,datetime,concurrent.futures,re
R=pathlib.Path(__file__).parent;P='project-27082a40-6083-4952-9db';now=datetime.datetime.now(datetime.timezone.utc);hi=now.isoformat().replace('+00:00','Z')
commands={
'commerce_post_promote_logs':['logging','read','resource.type="cloud_run_revision" AND resource.labels.service_name="us-commerce" AND resource.labels.revision_name="us-commerce-sre-ready-r1" AND timestamp>="2026-11-07T21:48:30Z" AND timestamp<="'+hi+'"','--order=asc','--limit=10000'],
'global_builds':['builds','list','--region=global','--filter=createTime>="2026-11-07T21:44:00Z"','--limit=20'],
'commerce_revision':['run','revisions','describe','us-commerce-sre-ready-r1','--region=europe-west4']}
def clean(x):
 if isinstance(x,dict):
  if 'value' in x and any(t in x.get('name','').lower() for t in ['secret','password','token','dsn']):x=dict(x,value='[REDACTED]')
  return {k:('[REDACTED]' if k.lower() in ['authorization','password','access_token','id_token','dsn'] else clean(v)) for k,v in x.items()}
 if isinstance(x,list):return [clean(v) for v in x]
 if isinstance(x,str):
  x=re.sub(r'postgres(?:ql)?://[^\s"\']+','[REDACTED_DSN]',x);x=re.sub(r'Bearer [A-Za-z0-9._~+/-]+','Bearer [REDACTED]',x)
 return x
def run(item):
 n,a=item;cmd=['gcloud']+a+['--project='+P,'--format=json'];p=subprocess.run(cmd,capture_output=True,text=True,timeout=70)
 try:d=clean(json.loads(p.stdout))
 except ValueError:d={'unavailable':True}
 (R/(n+'.json')).write_text(json.dumps(d,indent=2));(R/(n+'.meta.json')).write_text(json.dumps({'command':cmd,'end_istanbul':now.astimezone().isoformat(),'returncode':p.returncode,'stderr':clean(p.stderr)},indent=2));return n,p.returncode,len(p.stdout)
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
 for row in pool.map(run,commands.items()):print(row,flush=True)
