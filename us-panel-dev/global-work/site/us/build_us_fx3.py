#!/usr/bin/env python3
"""release_us_fx3: live ee14acd6 (SHOP TR TL strings, Demir-approved 01:10 TSİ) + fixed fx component.
Local only; never uploads. Proves the bundle body (all dictionaries incl. TR price strings) is byte-identical."""
import hashlib, os, re, shutil
H = os.path.dirname(os.path.abspath(__file__)); S = os.path.join(H, '..'); OUT = os.path.join(S, 'release_us_fx3')
LIVE = os.path.join(H, 'live_20261108_0110'); LIVEJS = 'index-uscommerce-i18nfx-trtl-ee14acd6.js'
OLDCOMP_LEN = 2342  # old component + '\n' prefix of 5a375f29 (= ee14acd6 prefix)
sha = lambda b: hashlib.sha256(b).hexdigest()
live = open(os.path.join(LIVE, LIVEJS), 'rb').read(); lidx = open(os.path.join(LIVE, 'live_index.html'), encoding='utf-8').read()
old5a = open(os.path.join(S, 'release_us_fx/public_html/assets/index-uscommerce-i18nfx-5a375f29.js'), 'rb').read()
assert sha(live).startswith('ee14acd6') and live[:OLDCOMP_LEN] == old5a[:OLDCOMP_LEN]
body = live[OLDCOMP_LEN:]
comp = open(os.path.join(H, 'fx_component.js'), 'rb').read()
data = comp + b'\n' + body
assert data[len(comp) + 1:] == body  # body unchanged
for s in ['1 Aylık Erişim / 200 İstem: 18.900 TL. 1 Saatlik Kullanım: 1.990 TL.', '1 saat işlem veya açık panel: 1.990 TL.']:
    assert live.count(s.encode()) == data.count(s.encode()) >= 1, s
name = 'index-uscommerce-i18nfx3-%s.js' % sha(data)[:8]
assert lidx.count('/assets/' + LIVEJS) == 1
html = lidx.replace('/assets/' + LIVEJS, '/assets/' + name)
if os.path.isdir(OUT): shutil.rmtree(OUT)
pub = os.path.join(OUT, 'public_html'); os.makedirs(os.path.join(pub, 'assets')); os.makedirs(os.path.join(OUT, 'rollback'))
open(os.path.join(pub, 'assets', name), 'wb').write(data); open(os.path.join(pub, 'index.html'), 'w', encoding='utf-8').write(html)
shutil.copy(os.path.join(LIVE, 'live_index.html'), os.path.join(OUT, 'rollback', 'index.html'))
shutil.copytree(os.path.join(S, 'release_us_fx2', 'cron'), os.path.join(OUT, 'cron'))
open(os.path.join(OUT, 'BODY_DIFF.txt'), 'w').write('live ee14acd6 body (after old component, %d B) vs fx3 body: %s\nlive sha256 %s\nfx3 sha256 %s\n' % (OLDCOMP_LEN, 'IDENTICAL' if data[len(comp)+1:] == body else 'DIFFERENT', sha(live), sha(data)))
v = open(os.path.join(H, 'verify_fx.sh.in')).read().replace('__ASSET__', name).replace('__SHA__', sha(data)).replace('__PREV__', LIVEJS)
open(os.path.join(OUT, 'VERIFY_AFTER_UPLOAD.sh'), 'w').write(v); os.chmod(os.path.join(OUT, 'VERIFY_AFTER_UPLOAD.sh'), 0o755)
open(os.path.join(OUT, 'CONFLICT_CHECK.md'), 'w', encoding='utf-8').write(f"""# CONFLICT_CHECK — release_us_fx3 (YÜKLENMEDİ)

Taban: canlı durum, curl ile 08.11.2026 ~01:20 TSİ indirildi (salt okuma).
- Canlı `index.html` sha256 `{sha(lidx.encode())}` → `/assets/{LIVEJS}`
- Canlı bundle sha256 `{sha(live)}` (SHOP: 5a375f29 + TR usCard/appsIntro/usIntro TL; Demir onayı 01:10 TSİ)
Yüklemeden hemen önce canlı index.html sha256 yukarıdakiyle eşleşmeli; değilse DUR ve yeniden üret.

| Dosya | Sunucuda | Etki |
|---|---|---|
| `public_html/assets/{name}` | YOK | Yeni |
| `public_html/index.html` | VAR — ÜZERİNE YAZAR | Yalnız script src |
- `fx.json` pakette YOK (canlıdaki/cron'un dosyasına dokunulmaz). `cron/` = fx2 ile aynı (ayrı adım, INSTALL_CRON.md).
- Bundle gövdesi canlıyla bayt bayt aynı (BODY_DIFF.txt); TR TL fiyat dizeleri korunur. Yalnız baştaki fx bileşeni değişir.
- Geri alma: `rollback/index.html` (= canlı {LIVEJS} index'i). Canlı asset sunucuda kalır.
- release_us_fx2 DEPRECATED: eski tabana dayanır, yüklenirse TR fiyat dizelerini geri alır.
""")
open(os.path.join(OUT, 'README_RELEASE.md'), 'w', encoding='utf-8').write("""# release_us_fx3 (YEREL, YÜKLENMEDİ)
Canlı ee14acd6 + düzeltilmiş fx bileşeni (R4B FX01/FX02: TSİ takvimi bugün-5..bugün; kuruş/BigInt yarımda sıfırdan uzağa).
Sitedeki fiyat metinleri (usCard/appsIntro/usIntro; TR'de TL, EN/DE/RU'da TRY) düz sözlük metni; `data-kanus-try` bağlantısı yok, bileşen onlara dokunmaz — uydurma bağlama yapılmadı.
Yeniden üretim: `python3 site/us/build_us_fx3.py`.
""")
files = []
for root, _, fs in os.walk(OUT):
    for f in fs:
        if f != 'SHA256SUMS': p = os.path.join(root, f); files.append((os.path.relpath(p, OUT), sha(open(p, 'rb').read())))
open(os.path.join(OUT, 'SHA256SUMS'), 'w').write(''.join(f'{h}  {p}\n' for p, h in sorted(files)))
print(name)
