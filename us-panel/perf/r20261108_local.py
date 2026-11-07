import argparse,asyncio,json,sys,shutil
from pathlib import Path
import run_suite,login_bench
OUT=Path('/workspace/us-panel/reports/r20261108')
async def main():
    OUT.mkdir(exist_ok=True);(OUT/'results').mkdir(exist_ok=True);(OUT/'bin').mkdir(exist_ok=True)
    run_suite.ROOT=OUT
    await run_suite.run(argparse.Namespace(source='/workspace/us-panel/work2/kiyas/term'))
    login_bench.ROOT=OUT
    await login_bench.run(argparse.Namespace(source='/workspace/us-panel/work2/kiyas/term',baseline='/workspace/us-panel/perf/bin/hook-login-baseline',cpus=[0,1],samples=50,warmup=4,overload_probe=True))
if __name__=='__main__':asyncio.run(main())
