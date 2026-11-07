#!/usr/bin/env python3
"""One command: build -> validate -> release/public_html -> FILES.txt -> SHA256SUMS.txt -> sha256sum -c.
Never uploads. Usage: python3 make_release.py [--slogan-tr '...'] [--slogan-en '...']"""
import os, shutil, subprocess, sys
HERE = os.path.dirname(os.path.abspath(__file__)); REL = os.path.join(HERE, "release"); PUB = os.path.join(REL, "public_html")
def run(cmd, cwd=HERE):
    print("+", " ".join(cmd)); r = subprocess.run(cmd, cwd=cwd)
    if r.returncode: sys.exit(f"FAILED: {' '.join(cmd)}")
run([sys.executable, "build.py", *sys.argv[1:]])
run([sys.executable, "validate.py"])
shutil.rmtree(PUB, ignore_errors=True); os.makedirs(PUB)
for item in ("tr", "en", "robots.txt", "sitemap.xml"):
    s = os.path.join(HERE, "out", item)
    (shutil.copytree if os.path.isdir(s) else shutil.copy2)(s, os.path.join(PUB, item))
files = sorted(os.path.relpath(os.path.join(d, f), PUB) for d, _, fs in os.walk(PUB) for f in fs)
open(os.path.join(REL, "FILES.txt"), "w").write("\n".join(files) + "\n")
with open(os.path.join(REL, "SHA256SUMS.txt"), "w") as fh:
    subprocess.run(["sha256sum", *files], cwd=PUB, stdout=fh, check=True)
run(["sha256sum", "-c", "--quiet", "../SHA256SUMS.txt"], cwd=PUB)
print(f"OK: {len(files)} files in release/public_html; sha256sum -c passed. NOT uploaded.")
