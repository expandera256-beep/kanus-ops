import asyncio,argparse
from pathlib import Path
import run_suite
run_suite.ROOT=Path('/workspace/us-panel/reports/r20261108')
asyncio.run(run_suite.run(argparse.Namespace(source='/workspace/us-panel/work2/kiyas/term')))
