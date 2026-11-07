"""Fail-closed new-host DNS/TLS peer/anonymous-health check; never contacts a model."""
import http.client,json,socket,ssl
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo
HOST='api-test.kanus.org';IP='82.29.184.225'

def check(output):
 d={'time':datetime.now(ZoneInfo('Europe/Istanbul')).isoformat(),'host':HOST,'expected_ip':IP,'authenticated_requests':0,'model_requests':0,'http_cap':1,'http_requests':0}
 conn=None
 try:
  addresses=sorted({r[4][0] for r in socket.getaddrinfo(HOST,443,type=socket.SOCK_STREAM)})
  d['dns_addresses']=addresses;d['dns_status']='PASS' if addresses==[IP] else 'BLOCKED'
  if addresses!=[IP]:
   d.update(status='BLOCKED',reason='DNS precondition not independently met; may be synthetic proxy DNS. No HTTP or alternate route.')
  else:
   conn=http.client.HTTPSConnection(HOST,timeout=20,context=ssl.create_default_context());conn.connect()
   d['tls_peer_ip']=conn.sock.getpeername()[0];d['tls_hostname_verified']=True
   if d['tls_peer_ip']!=IP:
    d.update(status='BLOCKED',reason='Unexpected peer; no HTTP sent')
   else:
    conn.request('GET','/healthz');d['http_requests']=1;r=conn.getresponse();d['health_http']=r.status;r.read(4096)
    d['status']='PASS' if r.status==200 else 'BLOCKED'
 except Exception as e:
  d.update(status='BLOCKED',error_type=type(e).__name__,reason='No fallback/retry/authentication/model request')
 finally:
  if conn:conn.close()
 Path(output).parent.mkdir(parents=True,exist_ok=True);Path(output).write_text(json.dumps(d,ensure_ascii=False,indent=2)+'\n')
 return d

if __name__=='__main__':
 result=check(Path(__file__).resolve().parents[2]/'reports/r20261108/qa-raw/new-host-precheck.json')
 print(json.dumps(result,ensure_ascii=False));raise SystemExit(0 if result['status']=='PASS' else 2)
