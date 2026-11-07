#!/usr/bin/env python3
"""Build the kanus.org/us i18n release from the LIVE bundle (no SPA source on box).
Local only: never uploads. Inputs: live_app.js, live_index.html (curl, 08.11.2026)."""
import hashlib, os, shutil, json
H = os.path.dirname(os.path.abspath(__file__))
OUT = os.path.join(H, '..', 'release_us')
LIVE_SHA = {'live_app.js': None, 'live_index.html': None}

HELPERS = r'''const __kL=()=>{try{const l=localStorage.getItem("kanus.lang");return({tr:1,de:1,ru:1,en:1})[l]?l:"en"}catch{return"en"}};const __kB={tr:"tr-TR",en:"en-US",de:"de-DE",ru:"ru-RU"};function __kDate(v){const d=new Date(v);if(!v||isNaN(d.getTime()))return v;try{return new Intl.DateTimeFormat(__kB[__kL()],{dateStyle:"medium",timeStyle:"short"}).format(d)}catch{return v}}const __kE={tr:{net:"Sunucuya ulaşılamadı. Bağlantını kontrol edip tekrar dene.",missing:"Zorunlu bir alan boş.",big:"Dosya çok büyük.",gen:"Kaydedilemedi. Lütfen tekrar dene."},en:{net:"Could not reach the server. Check your connection and try again.",missing:"A required field is empty.",big:"The file is too large.",gen:"Could not save. Please try again."},de:{net:"Der Server ist nicht erreichbar. Prüfe deine Verbindung und versuche es erneut.",missing:"Ein Pflichtfeld ist leer.",big:"Die Datei ist zu groß.",gen:"Speichern fehlgeschlagen. Bitte versuche es erneut."},ru:{net:"Не удалось связаться с сервером. Проверьте подключение и повторите попытку.",missing:"Обязательное поле не заполнено.",big:"Файл слишком большой.",gen:"Не удалось сохранить. Повторите попытку."}};function __kErr(x){const m=__kE[__kL()];const s=x&&x.status;if(!s||(x&&x.message==="api-html"))return m.net;if(s===400)return m.missing;if(s===413)return m.big;return m.gen}
'''
REPL = [
    ('children:n.createdAt', 'children:__kDate(n.createdAt)'),
    ('children:[u.author," · ",u.createdAt]', 'children:[u.author," · ",__kDate(u.createdAt)]'),
    ('children:[s.author," · ",s.createdAt]', 'children:[s.author," · ",__kDate(s.createdAt)]'),
    ('const a=await gr("contact",t);', 'let a;try{a=await gr("contact",t)}catch(x){l(__kErr(x));return}'),
    ('await gr("forum",{kind:"topic",...n}),', 'try{await gr("forum",{kind:"topic",...n})}catch(x){i(__kErr(x));return}'),
    ('await gr(e,S,k&&k.size?k:null),', 'try{await gr(e,S,k&&k.size?k:null)}catch(x){s(__kErr(x));return}'),
]
OLD_SRC = '/assets/index-uscommerce.js?v=20261104-usticari'

def sha(b): return hashlib.sha256(b).hexdigest()

def build():
    js = open(os.path.join(H, 'live_app.js'), encoding='utf-8').read()
    for a, b in REPL:
        n = js.count(a)
        assert n == 1, (a, n)
        js = js.replace(a, b)
    js = HELPERS + js
    data = js.encode('utf-8')
    name = 'index-uscommerce-i18n-%s.js' % sha(data)[:8]
    html = open(os.path.join(H, 'live_index.html'), encoding='utf-8').read()
    assert html.count(OLD_SRC) == 1
    html = html.replace(OLD_SRC, '/assets/' + name)
    if os.path.isdir(OUT): shutil.rmtree(OUT)
    pub = os.path.join(OUT, 'public_html'); os.makedirs(os.path.join(pub, 'assets'))
    open(os.path.join(pub, 'assets', name), 'wb').write(data)
    open(os.path.join(pub, 'index.html'), 'w', encoding='utf-8').write(html)
    rb = os.path.join(OUT, 'rollback'); os.makedirs(rb)
    shutil.copy(os.path.join(H, 'live_index.html'), os.path.join(rb, 'index.html'))
    return name


VERIFY = r'''#!/usr/bin/env bash
# Run AFTER KAN.US uploads public_html/* to Hostinger public_html/. Read-only checks (curl GET).
# Usage: VERIFY_AFTER_UPLOAD.sh [https://kanus.org]   exit 0 = all PASS
set -u
B="${1:-https://kanus.org}"; ASSET="__ASSET__"; WANT="__SHA__"; F=0
ok(){ echo "PASS $1"; }; bad(){ echo "FAIL $1"; F=$((F+1)); }
for p in / /us /contact /news; do
  h=$(curl -fsS "$B$p") || { bad "$p fetch"; continue; }
  echo "$h" | grep -q "/assets/$ASSET" && ok "$p references $ASSET" || bad "$p does not reference $ASSET (cache? old index.html?)"
done
got=$(curl -fsS "$B/assets/$ASSET" | sha256sum | cut -d' ' -f1)
[ "$got" = "$WANT" ] && ok "asset sha256" || bad "asset sha256 $got != $WANT"
ct=$(curl -fsSI "$B/assets/$ASSET" | tr -d '\r' | awk -F': ' 'tolower($1)=="content-type"{print $2}')
echo "$ct" | grep -qi javascript && ok "asset content-type $ct" || bad "asset content-type '$ct' (SPA fallback?)"
curl -fsS "$B/assets/index-uscommerce.js?v=20261104-usticari" >/dev/null && ok "old asset still served (rollback safe)" || bad "old asset missing"
curl -fsS "$B/assets/$ASSET" | grep -q '__kDate' && ok "i18n helpers present" || bad "helpers missing"
for s in 'Yerli Teknoloji' 'Küresel Vizyon' 'Built in Türkiye' 'Designed for the World'; do
  curl -fsS "$B/assets/$ASSET" | grep -qi "$s" && bad "forbidden slogan: $s" || ok "no '$s'"
done
echo "Manual: open $B/us, switch TR/EN/DE/RU; /news dates localized; failed form submit shows a localized notice."
echo "Rollback: upload rollback/index.html over public_html/index.html (old asset untouched)."
[ $F -eq 0 ] && echo "ALL PASS" || echo "$F FAIL"; exit $F
'''

def finalize(name):
    pub = os.path.join(OUT, 'public_html')
    data = open(os.path.join(pub, 'assets', name), 'rb').read()
    v = VERIFY.replace('__ASSET__', name).replace('__SHA__', sha(data))
    open(os.path.join(OUT, 'VERIFY_AFTER_UPLOAD.sh'), 'w').write(v); os.chmod(os.path.join(OUT, 'VERIFY_AFTER_UPLOAD.sh'), 0o755)
    live = {k: sha(open(os.path.join(H, k), 'rb').read()) for k in ('live_app.js', 'live_index.html')}
    cc = f"""# CONFLICT_CHECK — kanus.org/us i18n paketi (YÜKLENMEDİ)

Kaynak: canlı dosyalar (curl, 08.11.2026 ~00:22 TSİ). SPA kaynağı box'ta yok; paket canlı derlenmiş bundle'ın yerel kopyası üzerine 6 noktasal değişiklik + yardımcılar.
- Canlı `index.html` sha256 `{live['live_index.html']}` (1600 B, last-modified 03.11.2026 23:22 GMT)
- Canlı `/assets/index-uscommerce.js` sha256 `{live['live_app.js']}`

| Paket dosyası | Sunucuda | Etki |
|---|---|---|
| `public_html/assets/{name}` | YOK (yeni ad, içerik hash'li) | Yeni dosya, üzerine yazma yok |
| `public_html/index.html` | VAR — **ÜZERİNE YAZAR** | Tek fark script src satırı; SPA fallback (.htaccess) tüm yollar için bu dosyayı sunar → tüm site yeni bundle'ı yükler |

- `.htaccess`, `api/`, `robots.txt`, `sitemap.xml`, eski asset ve CSS'e dokunulmaz. Eski bundle sunucuda kalır.
- Yüklemeden hemen önce canlı index.html sha256'sı yukarıdakiyle eşleşmeli; eşleşmiyorsa biri siteyi değiştirmiştir → DUR, paketi yeniden üret (`python3 site/us/build_us.py` güncel canlı dosyalarla).
- Önceki statik /tr–/en paketi (`site/release/`) ile dosya çakışması yok (ayrı yollar). İkisi birlikte yüklenebilir.
- Geri alma: `rollback/index.html` (canlı orijinal) → public_html/index.html.
"""
    open(os.path.join(OUT, 'CONFLICT_CHECK.md'), 'w').write(cc)
    rd = f"""# release_us — kanus.org/us çok dil düzeltmesi (YEREL, YÜKLENMEDİ)
Yükleme KAN.US'te ve Demir onayına bağlı. DNS/Shopify/GitHub değişikliği yok.
Değişiklikler (canlı bundle üzerinde): tarih gösterimi dile göre Intl.DateTimeFormat (tr-TR/en-US/de-DE/ru-RU); iletişim, forum konusu ve yükleme formlarında başarısız gönderimde 4 dilde hata mesajı (önce sessizce düşüyordu). Metin sözlüklerine dokunulmadı (TR/EN/DE/RU 126 anahtar zaten tam).
Fiyat: canlı sitede fiyat gösterilmiyor (payments rotası /us'a yönleniyor); döviz gösterimi eklenmedi, kur kaynağı yok.
Yükleme: public_html/ içeriğini sunucu public_html/ köküne. Sonra `./VERIFY_AFTER_UPLOAD.sh`.
Yeniden üretim: `python3 site/us/build_us.py`.
"""
    open(os.path.join(OUT, 'README_RELEASE.md'), 'w').write(rd)
    files = []
    for root, _, fs in os.walk(OUT):
        for f in fs:
            if f == 'SHA256SUMS': continue
            p = os.path.join(root, f); files.append((os.path.relpath(p, OUT), sha(open(p, 'rb').read())))
    open(os.path.join(OUT, 'SHA256SUMS'), 'w').write(''.join(f'{h}  {p}\n' for p, h in sorted(files)))

if __name__ == '__main__':
    finalize(build())
