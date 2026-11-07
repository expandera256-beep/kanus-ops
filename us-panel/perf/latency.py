"""Low-rate network timing. Never a load runner. No redirects/auth/GPU paths."""
import argparse, asyncio, json, time
from datetime import datetime
from pathlib import Path
from urllib.parse import urlsplit
from zoneinfo import ZoneInfo
ROOT=Path(__file__).resolve().parent
async def curl(url,body=False):
    args=['curl','--silent','--show-error','--max-time','15','--connect-timeout','5','--proto','=http,https','--max-redirs','0']
    if body:
        args += [url]
    else: args += ['--output','/dev/null','--write-out','%{json}',url]
    p=await asyncio.create_subprocess_exec(*args,stdout=asyncio.subprocess.PIPE,stderr=asyncio.subprocess.PIPE)
    out,err=await p.communicate()
    if p.returncode: return {'error':'curl_exit_'+str(p.returncode),'detail':err.decode().strip()}
    try: return json.loads(out)
    except ValueError: return {'error':'non_json_result'}
async def run(args):
    if not args.output_prefix.replace('_','').isalnum(): raise SystemExit('invalid output prefix')
    u=urlsplit(args.url)
    if u.username or u.password or u.query or u.fragment: raise SystemExit('credentials/query/fragment forbidden')
    local=u.scheme=='http' and u.hostname=='127.0.0.1' and u.path=='/healthz'
    approved=args.allow_public_health and u.scheme=='https' and u.hostname=='api-test.kanus.org' and u.path=='/healthz' and u.port in (None,443)
    if not (local or approved): raise SystemExit('only loopback health or explicitly enabled api-test HTTPS /healthz')
    if not 1<=args.count<=10 or args.interval<2: raise SystemExit('count1..10 and interval>=2 required')
    result={'source':'REAL_NETWORK_NOT_GPU' if approved else 'MOCK_NETWORK','target':args.url,'timestamp':datetime.now(ZoneInfo('Europe/Istanbul')).isoformat(),'zone':'TSİ / Europe/Istanbul','samples':[],'notes':['TCP connect is connect completion minus DNS, NOT pure RTT.','TLS increment can include network/proxy overhead; no packet-level RTT inferred.','Health HTTP TTFB is NOT LLM TTFT.','Single box vantage does not measure US/Europe end-user latency.']}
    if args.egress:
        geo=await curl('https://ipinfo.io/json',True)
        result['egress']={k:geo[k] for k in ('ip','city','region','country','loc','org','timezone','error','detail') if k in geo}
        result['notes'].append('Egress region is ipinfo geolocation estimate, not verified hosting location.')
        if 'error' in geo:
            result['notes'].append('Egress lookup failed; no alternate network route used.')
    for i in range(args.count):
        start=time.perf_counter(); row=await curl(args.url)
        if 'error' in row:
            result['samples'].append(row)
            # Stop on network/proxy failures: do not reroute/retry blocked traffic.
            break
        sample={k:row.get(k) for k in ('http_code','remote_ip','http_connect','num_connects','size_download','size_header','ssl_verify_result','proxy_used')}
        for name,key in [('dns_ms','time_namelookup'),('ttfb_ms','time_starttransfer'),('total_ms','time_total')]: sample[name]=1000*row.get(key,0)
        sample['tcp_connect_ms']=1000*(row.get('time_connect',0)-row.get('time_namelookup',0))
        sample['tls_ms']=1000*(row.get('time_appconnect',0)-row.get('time_connect',0)) if u.scheme=='https' else None
        sample['post_tls_to_first_byte_ms']=1000*(row.get('time_starttransfer',0)-row.get('time_appconnect',0)) if u.scheme=='https' else None
        sample['ttft_ms']=None; sample['started_monotonic']=start
        result['samples'].append(sample)
        if i+1<args.count: await asyncio.sleep(args.interval)
    for i,s in enumerate(result['samples']):
        if i and 'started_monotonic' in s and 'started_monotonic' in result['samples'][i-1]: s['gap_from_previous_start_seconds']=s['started_monotonic']-result['samples'][i-1]['started_monotonic']
    for s in result['samples']: s.pop('started_monotonic',None)
    (ROOT/'results').mkdir(exist_ok=True)
    (ROOT/('results/'+args.output_prefix+'.json')).write_text(json.dumps(result,indent=2,ensure_ascii=False))
    md=['# '+result['source']+' — düşük hızlı health ölçümü','','Kaynak zamanı: '+result['timestamp']+' (TSİ)','','| HTTP | DNS ms | TCP connect ms | TLS ms | HTTP TTFB ms | Toplam ms |','|---|---:|---:|---:|---:|---:|']
    for s in result['samples']:
        if 'error' in s: md.append('| ERROR: '+s['error']+' | | | | | |'); continue
        md.append('| '+str(s['http_code'])+' | '+' | '.join(f'{s[k]:.3f}' if isinstance(s.get(k),(int,float)) else 'N/A' for k in ('dns_ms','tcp_connect_ms','tls_ms','ttfb_ms','total_ms'))+' |')
    md+=['','Egress: `'+json.dumps(result.get('egress',{}),ensure_ascii=False)+'`','','\n'.join('- '+n for n in result['notes'])]
    (ROOT/('results/'+args.output_prefix+'.md')).write_text('\n'.join(md)+'\n')
    print(json.dumps(result,ensure_ascii=False))
if __name__=='__main__':
    p=argparse.ArgumentParser(); p.add_argument('--output-prefix',default='network_results'); p.add_argument('--url',default='http://127.0.0.1:18101/healthz'); p.add_argument('--count',type=int,default=3); p.add_argument('--interval',type=float,default=2); p.add_argument('--egress',action='store_true'); p.add_argument('--allow-public-health',action='store_true'); asyncio.run(run(p.parse_args()))
