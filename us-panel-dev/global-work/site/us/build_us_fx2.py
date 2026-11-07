#!/usr/bin/env python3
"""Build release_us_fx on top of release_us (local only; never uploads).
Adds fx component to the release_us bundle (new hashed asset) + /fx.json snapshot."""
import hashlib, os, shutil, re, json, subprocess, sys
H = os.path.dirname(os.path.abspath(__file__)); BASE = os.path.join(H, '..', 'release_us'); OUT = os.path.join(H, '..', 'release_us_fx2'); PREVPKG = os.path.join(H, '..', 'release_us_fx')
sha = lambda b: hashlib.sha256(b).hexdigest()
def main(fxjson):
    pub0 = os.path.join(BASE, 'public_html')
    html0 = open(os.path.join(pub0, 'index.html'), encoding='utf-8').read()
    old = re.search(r'/assets/(index-uscommerce-i18n-[0-9a-f]{8}\.js)', html0)[1]
    js = open(os.path.join(pub0, 'assets', old), 'rb').read()
    comp = open(os.path.join(H, 'fx_component.js'), 'rb').read()
    data = comp + b'\n' + js
    name = 'index-uscommerce-i18nfx2-%s.js' % sha(data)[:8]
    assert html0.count('/assets/' + old) == 1
    html = html0.replace('/assets/' + old, '/assets/' + name)
    fx = json.load(open(fxjson)); assert fx['schema'] == 'kanus.fx/1'
    if os.path.isdir(OUT): shutil.rmtree(OUT)
    pub = os.path.join(OUT, 'public_html'); os.makedirs(os.path.join(pub, 'assets')); os.makedirs(os.path.join(OUT, 'rollback')); os.makedirs(os.path.join(OUT, 'cron'))
    open(os.path.join(pub, 'assets', name), 'wb').write(data)
    open(os.path.join(pub, 'index.html'), 'w', encoding='utf-8').write(html)
    shutil.copy(fxjson, os.path.join(pub, 'fx.json'))
    shutil.copy(os.path.join(PREVPKG, 'public_html', 'index.html'), os.path.join(OUT, 'rollback', 'index.html'))
    shutil.copy(os.path.join(pub0, 'index.html'), os.path.join(OUT, 'rollback', 'index.release_us.html'))
    shutil.copy(os.path.join(BASE, 'rollback', 'index.html'), os.path.join(OUT, 'rollback', 'index.live_original.html'))
    shutil.copy(os.path.join(H, 'fx_snapshot.py'), os.path.join(OUT, 'cron', 'fx_snapshot.py'))
    up = os.path.join(OUT, 'cron', 'upload_public_html', '_fx'); os.makedirs(up)
    shutil.copy(os.path.join(H, 'fx_cron.php'), os.path.join(up, 'fx_cron.php'))
    shutil.copy(os.path.join(H, 'fx_htaccess'), os.path.join(up, '.htaccess'))
    shutil.copy(os.path.join(H, 'install_cron.md'), os.path.join(OUT, 'cron', 'INSTALL_CRON.md'))
    ah = sha(data); bh = sha(open(os.path.join(PREVPKG, 'public_html', 'index.html'), 'rb').read())
    old = re.search(r'/assets/(index-uscommerce-i18nfx-[0-9a-f]{8}\.js)', open(os.path.join(PREVPKG, 'public_html', 'index.html')).read())[1]
    v = open(os.path.join(H, 'verify_fx.sh.in')).read().replace('__ASSET__', name).replace('__SHA__', ah).replace('__PREV__', old)
    open(os.path.join(OUT, 'VERIFY_AFTER_UPLOAD.sh'), 'w').write(v); os.chmod(os.path.join(OUT, 'VERIFY_AFTER_UPLOAD.sh'), 0o755)
    cc = open(os.path.join(H, 'conflict_fx2.md.in'), encoding='utf-8').read().replace('__ASSET__', name).replace('__PREV__', old).replace('__BASEIDX__', bh).replace('__LIVEIDX__', sha(open(os.path.join(BASE, 'rollback', 'index.html'), 'rb').read())).replace('__FXDATE__', fx['tcmb_tarih'])
    open(os.path.join(OUT, 'CONFLICT_CHECK.md'), 'w', encoding='utf-8').write(cc)
    shutil.copy(os.path.join(H, 'readme_fx2.md'), os.path.join(OUT, 'README_RELEASE.md'))
    files = []
    for root, _, fs in os.walk(OUT):
        for f in fs:
            if f != 'SHA256SUMS': p = os.path.join(root, f); files.append((os.path.relpath(p, OUT), sha(open(p, 'rb').read())))
    open(os.path.join(OUT, 'SHA256SUMS'), 'w').write(''.join(f'{h}  {p}\n' for p, h in sorted(files)))
    print(name)
if __name__ == '__main__': main(sys.argv[1] if len(sys.argv) > 1 else '/tmp/fx.json')
