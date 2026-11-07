import asyncio,json,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parent
async def run():
    p=await asyncio.create_subprocess_exec(sys.executable,str(ROOT/'mock_stack.py'),'--upstream-port','0','--hook-port','0',stdout=asyncio.subprocess.PIPE,stderr=asyncio.subprocess.PIPE)
    try:
        info=json.loads(await asyncio.wait_for(p.stdout.readline(),5)); rows=[]
        for cmd in ([sys.executable,str(ROOT/'load.py'),'--url',info['hook']+'/v1/chat/completions','--users','1,5','--repetitions','1','--output',str(ROOT/'results/cli_load.json')],[sys.executable,str(ROOT/'latency.py'),'--url',info['hook']+'/healthz','--count','1','--output-prefix','local_network_results']):
            q=await asyncio.create_subprocess_exec(*cmd,stdout=asyncio.subprocess.PIPE,stderr=asyncio.subprocess.PIPE)
            out,err=await asyncio.wait_for(q.communicate(),15)
            rows.append({'script':Path(cmd[1]).name,'exit_code':q.returncode,'stdout':out.decode(),'stderr':err.decode()})
        result={'source':'MOCK','steps':rows,'pass':all(x['exit_code']==0 for x in rows)}
        (ROOT/'results/cli_smoke.json').write_text(json.dumps(result,indent=2)); print(json.dumps({'source':'MOCK','pass':result['pass']}))
    finally:
        p.terminate(); await p.wait()
if __name__=='__main__': asyncio.run(run())
