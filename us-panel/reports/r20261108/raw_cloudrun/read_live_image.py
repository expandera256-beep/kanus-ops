import subprocess,json,pathlib,datetime,urllib.request,hashlib,tarfile,io
R=pathlib.Path(__file__).parent;P='project-27082a40-6083-4952-9db'
cmd=['gcloud','run','services','describe','us-watchdog','--project='+P,'--region=europe-west4','--format=json']
r=subprocess.run(cmd,capture_output=True,text=True,check=True,timeout=40);d=json.loads(r.stdout)
for c in d['spec']['template']['spec']['containers']:
 for e in c.get('env',[]):
  if 'value' in e and any(s in e.get('name','').lower() for s in ['secret','token','password','dsn','key']):e['value']='[REDACTED]'
(R/'watchdog_final.json').write_text(json.dumps(d,indent=2));(R/'watchdog_final.meta.json').write_text(json.dumps({'command':cmd,'time_istanbul':datetime.datetime.now().astimezone().isoformat()},indent=2))
image=d['spec']['template']['spec']['containers'][0]['image'];print(d['status']['latestReadyRevisionName'],image,flush=True)
# Official OCI Distribution read-only endpoints; gcloud access token stays in memory only.
t=subprocess.run(['gcloud','auth','print-access-token'],capture_output=True,text=True,check=True,timeout=30).stdout.strip()
host,repo=image.split('/',1)
if '@' in repo:repo,ref=repo.split('@',1)
else:repo,ref=repo.rsplit(':',1)
base='https://'+host+'/v2/'+repo
steps=[]
def get(path,accept=None):
 h={'Authorization':'Bearer '+t}
 if accept:h['Accept']=accept
 with urllib.request.urlopen(urllib.request.Request(base+path,headers=h),timeout=90) as f:data=f.read()
 steps.append({'url':base+path,'http_status':200,'bytes':len(data),'sha256':hashlib.sha256(data).hexdigest()})
 return data
raw=get('/manifests/'+ref,'application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json')
manifest=json.loads(raw);(R/'live_image_manifest.json').write_text(json.dumps(manifest,indent=2))
if ref.startswith('sha256:'):assert 'sha256:'+hashlib.sha256(raw).hexdigest()==ref
config=json.loads(get('/blobs/'+manifest['config']['digest']));(R/'live_image_config.json').write_text(json.dumps(config,indent=2))
# Image source is in the uppermost small app layer. Never extract arbitrary tar paths.
found={};want={'gce_provider.py','runtime.py','commerce_http.py'}
for layer in reversed(manifest['layers']):
 if layer['size']>2_000_000:continue
 blob=get('/blobs/'+layer['digest']);assert 'sha256:'+hashlib.sha256(blob).hexdigest()==layer['digest']
 with tarfile.open(fileobj=io.BytesIO(blob),mode='r:*') as tar:
  for m in tar.getmembers():
   name=pathlib.PurePosixPath(m.name).name
   if name in want and m.isfile() and m.size<500_000 and name not in found:
    content=tar.extractfile(m).read();(R/('live_'+name)).write_bytes(content);found[name]={'tar_path':m.name,'layer_digest':layer['digest'],'sha256':hashlib.sha256(content).hexdigest()}
 if want<=found.keys():break
(R/'live_image_read_evidence.json').write_text(json.dumps({'image':image,'revision':d['status']['latestReadyRevisionName'],'time_istanbul':datetime.datetime.now().astimezone().isoformat(),'manifest_sha256':hashlib.sha256(raw).hexdigest(),'files':found,'requests':steps,'token_persisted':False},indent=2))
print('independently extracted',list(found),flush=True)
