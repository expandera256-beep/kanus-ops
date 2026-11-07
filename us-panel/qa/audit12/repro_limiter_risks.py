#!/usr/bin/env python3
"""Loopback/in-process synthetic security regressions. Account-lock DoS intentionally FAILs."""
import argparse,json,subprocess,sys,tempfile
from datetime import datetime
from pathlib import Path
from zoneinfo import ZoneInfo
ROOT=Path(__file__).resolve().parents[2];sys.path.insert(0,str(ROOT/'qa'));from common import cleanenv
p=argparse.ArgumentParser(description=__doc__);p.add_argument('--tree',required=True,type=Path);a=p.parse_args();tree=a.tree.resolve()
assert (tree/'go.mod').is_file()
parent=ROOT/'reports/r20261108/qa-raw/runs';parent.mkdir(exist_ok=True)
out=Path(tempfile.mkdtemp(prefix=datetime.now(ZoneInfo('Europe/Istanbul')).strftime('%Y%m%dT%H%M%S%z')+'-limiter-risks-',dir=parent));print('RUN_DIR '+str(out),flush=True)
overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(tree/'licensesrv/zz_qa_limiter_risks_test.go'):str(ROOT/'qa/audit12/limiter_risks_test.go')}}))
r=subprocess.run(['go','test','-p=1','-overlay',str(overlay),'./licensesrv','-run','^TestQAAuditLimiter','-v','-count=1','-timeout=90s'],cwd=tree,env=cleanenv(),capture_output=True,text=True,timeout=120)
(out/'limiter-risks.log').write_text(r.stdout+r.stderr+'\nexit_code='+str(r.returncode)+'\n');print(r.stdout+r.stderr);print('exit_code='+str(r.returncode));sys.exit(r.returncode)
