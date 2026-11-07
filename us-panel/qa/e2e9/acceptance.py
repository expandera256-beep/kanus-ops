#!/usr/bin/env python3
"""Independent HTTP acceptance; never supplied to the model."""
import importlib.util,json,sys,threading
from pathlib import Path
from urllib.request import Request,urlopen
from urllib.error import HTTPError
from http.server import ThreadingHTTPServer
root=Path(sys.argv[1]); spec=importlib.util.spec_from_file_location('generated_service',root/'server.py')
module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
server=ThreadingHTTPServer(('127.0.0.1',0),module.Handler)
thread=threading.Thread(target=server.serve_forever,daemon=True);thread.start()
base='http://127.0.0.1:'+str(server.server_port)
def call(path,body=None,user=None):
 headers={'Content-Type':'application/json'}
 if user: headers['Authorization']='Bearer '+user
 req=Request(base+path,data=json.dumps(body).encode() if body is not None else None,headers=headers)
 try:
  with urlopen(req,timeout=3) as r:return r.status,json.load(r)
 except HTTPError as e:return e.code,json.load(e)
try:
 assert call('/healthz')==(200,{'ok':True})
 assert call('/tasks')[0]==401
 assert call('/tasks',{'title':'independent-A'},'alice')[0]==201
 assert call('/tasks',user='bob')==(200,{'tasks':[]})
 assert call('/tasks',user='alice')[1]['tasks'][0]['title']=='independent-A'
 assert call('/tasks',{'title':''},'bob')[0]==400
 print('INDEPENDENT_HTTP_ACCEPTANCE PASS health/auth/create/list/cross-user/validation')
finally:
 server.shutdown();server.server_close();thread.join(timeout=2)
