#!/usr/bin/env python3
"""Offline validation of site/release_us (no network)."""
import hashlib, json, os, re, subprocess, sys
H = os.path.dirname(os.path.abspath(__file__)); R = os.path.join(H, '..', sys.argv[1] if len(sys.argv) > 1 else 'release_us'); F = 0
def ck(n, ok, info=''):
    global F; print(('PASS ' if ok else 'FAIL ') + n, info); F += 0 if ok else 1
pub = os.path.join(R, 'public_html')
assets = os.listdir(os.path.join(pub, 'assets')); ck('one new asset', len(assets) == 1, assets)
a = assets[0]; data = open(os.path.join(pub, 'assets', a), 'rb').read(); js = data.decode()
ck('asset name = content hash', re.fullmatch(r'index-uscommerce-i18n(fx2?)?-%s\.js' % hashlib.sha256(data).hexdigest()[:8], a) is not None)
html = open(os.path.join(pub, 'index.html')).read(); live = open(os.path.join(H, 'live_index.html')).read()
ck('index.html references asset', html.count('/assets/' + a) == 1)
ck('index.html differs from live only in script src', html.replace('/assets/' + a, '/assets/index-uscommerce.js?v=20261104-usticari') == live)
ck('rollback = live index', open(os.path.join(R, 'rollback', 'index.live_original.html' if os.path.basename(os.path.normpath(R)).startswith('release_us_fx') else 'index.html')).read() == live)
if os.path.basename(os.path.normpath(R)).startswith('release_us_fx'):
    ck('fx rollback = release_us index', open(os.path.join(R, 'rollback', 'index.html'), 'rb').read() == open(os.path.join(H, '..', 'release_us_fx' if R.rstrip('/').endswith('fx2') else 'release_us', 'public_html', 'index.html'), 'rb').read())
    fx = json.load(open(os.path.join(R, 'public_html', 'fx.json'))); ck('fx.json schema', fx.get('schema') == 'kanus.fx/1' and set(fx['rates']) == {'EUR', 'USD'}, fx.get('tcmb_tarih'))
    ck('no tcmb in bundle', b'tcmb.gov.tr' not in data)
tmp = '/tmp/_vus.mjs'; open(tmp, 'wb').write(data)
ck('node --check', subprocess.run(['node', '--check', tmp], capture_output=True).returncode == 0)
ck('helpers', all(x in js for x in ('function __kDate', 'function __kErr', 'Intl.DateTimeFormat')))
ck('date call sites', js.count('__kDate(') == 4, js.count('__kDate('))
ck('error call sites', len(re.findall(r'catch\(x\)\{[a-z]\(__kErr\(x\)\);return\}', js)) == 3)
for l in ('tr', 'en', 'de', 'ru'): ck('error dict ' + l, re.search(l + r':\{net:"[^"]+",missing:"[^"]+",big:"[^"]+",gen:"[^"]+"\}', js) is not None)
lo = js.lower()
for s in ('yerli teknoloji', 'küresel vizyon', 'built in türkiye', 'designed for the world', 'yerli ve milli', 'homegrown'): ck('no slogan ' + s, s not in lo)
d = json.load(open(os.path.join(H, '..', '..', 'tur8', 'live', 'dicts.json')))
ck('live dict parity 4 langs', all(set(d[l]) == set(d['en']) for l in d) and len(d['en']) == 126, len(d['en']))
r = subprocess.run(['sha256sum', '-c', 'SHA256SUMS'], cwd=R, capture_output=True, text=True)
ck('SHA256SUMS', r.returncode == 0, r.stdout.count(': OK'))
print('ALL PASS' if F == 0 else f'{F} FAIL'); sys.exit(1 if F else 0)
