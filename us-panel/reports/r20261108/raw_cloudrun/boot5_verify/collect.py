import subprocess,json,pathlib,datetime,concurrent.futures,re
R=pathlib.Path(__file__).parent;P='project-27082a40-6083-4952-9db';now=datetime.datetime.now(datetime.timezone.utc);hi=now.isoformat().replace('+00:00','Z')
cmds={
'audit_stops':['logging','read','logName="projects/'+P+'/logs/cloudaudit.googleapis.com%2Factivity" AND resource.type="gce_instance" AND resource.labels.instance_id="6979174631840438511" AND protoPayload.methodName:"compute.instances.stop" AND timestamp>="2026-11-07T21:28:00Z" AND timestamp<="'+hi+'"','--order=asc','--limit=1000'],
'watchdog_logs':['logging','read','resource.type="cloud_run_revision" AND resource.labels.service_name="us-watchdog" AND timestamp>="2026-11-07T21:28:00Z" AND timestamp<="'+hi+'"','--order=asc','--limit=10000'],
'scheduler_logs':['logging','read','resource.type="cloud_scheduler_job" AND resource.labels.job_id="us-idle-watchdog" AND timestamp>="2026-11-07T21:28:00Z" AND timestamp<="'+hi+'"','--order=asc','--limit=10000'],
'vm_state':['compute','instances','describe','ds-a100','--zone=europe-west4-b','--format=json(name,id,status,zone,lastStartTimestamp,lastStopTimestamp,serviceAccounts.email,disks.type,disks.savedState,networkInterfaces.accessConfigs.natIP)'],
'commerce_current':['run','services','describe','us-commerce','--region=europe-west4'],
'watchdog_current':['run','services','describe','us-watchdog','--region=europe-west4']}
def clean(x):
 if isinstance(x,dict):
  if 'value' in x and any(t in x.get('name','').lower() for t in ['secret','password','token','dsn']):x=dict(x,value='[REDACTED]')
  return {k:('[REDACTED]' if k.lower() in ('authorization','access_token','id_token','password','dsn') else clean(v)) for k,v in x.items()}
 if isinstance(x,list):return [clean(v) for v in x]
 if isinstance(x,str):
  x=re.sub(r'postgres(?:ql)?://[^\s"\']+','[REDACTED_DSN]',x);x=re.sub(r'Bearer [A-Za-z0-9._~+/-]+','Bearer [REDACTED]',x);x=re.sub(r'eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+','[REDACTED_JWT]',x)
 return x
def run(item):
 n,a=item;cmd=['gcloud']+a+['--project='+P]
 if not any(x.startswith('--format=') for x in a):cmd+=['--format=json']
 r=subprocess.run(cmd,capture_output=True,text=True,timeout=70)
 try:d=clean(json.loads(r.stdout))
 except ValueError:d={'unavailable':True}
 (R/(n+'.json')).write_text(json.dumps(d,indent=2));(R/(n+'.meta.json')).write_text(json.dumps({'command':cmd,'returncode':r.returncode,'stderr':clean(r.stderr),'end_istanbul':now.astimezone().isoformat()},indent=2));return n,r.returncode,len(r.stdout)
with concurrent.futures.ThreadPoolExecutor(max_workers=6) as pool:
 for row in pool.map(run,cmds.items()):print(row,flush=True)
