import subprocess,json,re,concurrent.futures,datetime,pathlib
ROOT=pathlib.Path(__file__).parent
P="project-27082a40-6083-4952-9db"
commands={
"commerce_describe":["run","services","describe","us-commerce","--region=europe-west4"],
"watchdog_describe":["run","services","describe","us-watchdog","--region=europe-west4"],
"commerce_iam":["run","services","get-iam-policy","us-commerce","--region=europe-west4"],
"watchdog_iam":["run","services","get-iam-policy","us-watchdog","--region=europe-west4"],
"scheduler":["scheduler","jobs","describe","us-idle-watchdog","--location=europe-west4"],
"sql_describe":["sql","instances","describe","us-commerce-pg"],
"secrets":["secrets","list"],
"logs_run_24h":["logging","read",'resource.type="cloud_run_revision" AND (resource.labels.service_name="us-commerce" OR resource.labels.service_name="us-watchdog")',"--freshness=24h","--limit=15000","--order=asc"],
"logs_watchdog_post_deploy":["logging","read",'resource.type="cloud_run_revision" AND resource.labels.service_name="us-watchdog" AND timestamp>="2026-11-07T21:10:00Z"',"--limit=10000","--order=asc"],
"logs_scheduler_24h":["logging","read",'resource.type="cloud_scheduler_job" AND resource.labels.job_id="us-idle-watchdog"',"--freshness=24h","--limit=10000","--order=asc"],
"logs_sql_24h":["logging","read",'resource.type="cloudsql_database" AND resource.labels.database_id="project-27082a40-6083-4952-9db:us-commerce-pg" AND severity>=WARNING',"--freshness=24h","--limit=1000","--order=asc"],
"error_reporting_logs":["logging","read",'(resource.type="cloud_run_revision" AND (resource.labels.service_name="us-commerce" OR resource.labels.service_name="us-watchdog")) AND (severity>=ERROR OR jsonPayload.@type="type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent")',"--freshness=24h","--limit=2000","--order=asc"],
}
def redact(x):
 if isinstance(x,dict):
  if x.get('name','').startswith(('US_','SHOPIFY_')) and 'value' in x and any(t in x['name'].lower() for t in ['token','secret','password','dsn','key']):x=dict(x,value='[REDACTED]')
  return {k:('[REDACTED]' if (k.lower() in ('token','password','authorization','dsn','access_token','id_token') or (k.lower()=='key' and v != 'latest' and not str(v).isdigit())) else redact(v)) for k,v in x.items()}
 if isinstance(x,list):return [redact(v) for v in x]
 if isinstance(x,str):
  x=re.sub(r'postgres(?:ql)?://[^\s"\']+', '[REDACTED_DSN]', x)
  x=re.sub(r'Bearer [A-Za-z0-9._~+/-]+','Bearer [REDACTED]',x)
  x=re.sub(r'eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+','[REDACTED_JWT]',x)
 return x
now=datetime.datetime.now(datetime.timezone.utc)
lo=(now-datetime.timedelta(hours=24)).isoformat().replace('+00:00','Z');hi=now.isoformat().replace('+00:00','Z')
(ROOT/'log_window.json').write_text(json.dumps({'start_utc':lo,'end_utc':hi,'start_istanbul':(now-datetime.timedelta(hours=24)).astimezone().isoformat(),'end_istanbul':now.astimezone().isoformat()},indent=2))
for name,args in commands.items():
 if args[:2]==['logging','read']:
  args[2]+=' AND timestamp<="'+hi+'"'
  if name!='logs_watchdog_post_deploy':args[2]+=' AND timestamp>="'+lo+'"'
def run(item):
 name,args=item;cmd=['gcloud']+args+['--project='+P,'--format=json','--quiet'];t=datetime.datetime.now().astimezone().isoformat()
 try:
  r=subprocess.run(cmd,capture_output=True,text=True,timeout=100)
  try: out=redact(json.loads(r.stdout))
  except ValueError:out=redact(r.stdout)
  (ROOT/(name+'.json')).write_text(json.dumps(out,ensure_ascii=False,indent=2))
  (ROOT/(name+'.meta.json')).write_text(json.dumps({'command':cmd,'started_istanbul':t,'returncode':r.returncode,'stderr':redact(r.stderr)},ensure_ascii=False,indent=2))
  return name,r.returncode,len(r.stdout)
 except Exception as e:return name,type(e).__name__
with concurrent.futures.ThreadPoolExecutor(max_workers=6) as pool:
 for r in pool.map(run,[(k,v) for k,v in commands.items() if v[:2]==["logging","read"]]): print(r,flush=True)
