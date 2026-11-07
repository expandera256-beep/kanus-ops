# kanus.org /tr–/en yayın paketi TASLAĞI — YÜKLENMEDİ (yüklemeyi KAN.US yapar)

Hedef: Hostinger `public_html/` (File Manager/FTP). `.htaccess` DEĞİŞMEZ (`!-f`/`!-d` → gerçek dosya/dizin doğrudan sunulur, gerisi SPA).

## İçerik — 16 dosya (`FILES.txt`, bütünlük `SHA256SUMS.txt`)
- `tr/` ve `en/` altında 7'şer sayfa: ana sayfa, product, pricing, download, docs, api, support (`index.html`).
- `robots.txt`, `sitemap.xml` (14 URL, tr↔en + x-default=/tr/).
- **Hukuki sayfalar (privacy, terms, refund; tr+en) pakette YOK** (Demir kararı 06.11.2026 09:42 TSİ). Menü/footer'da bu sayfalara link yok; hreflang ve sitemap'te yoklar. Taslak metinler yalnız kaynakta (`site/content.json`, `"draft": true`, "TASLAK — YAYINLANMAZ") durur ve build'e girmez.
- `public_html/index.html`, `assets/`, `api/`, `.htaccess` YÜKLENMEZ / DEĞİŞMEZ.

## Tek komutla yeniden üretim (slogan dahil)
```bash
cd /workspace/us-panel-dev/global-work/site
python3 make_release.py --slogan-tr 'Demir’in TR sloganı.' --slogan-en 'Demir’s EN slogan.'
# slogan değişmeyecekse argümansız: python3 make_release.py
```
Sırasıyla: `build.py` (slogan argümanı varsa `content.json` → `"slogan": {"tr","en"}` tek anahtarını günceller) → `validate.py` → `release/public_html` kopyası → `FILES.txt` → `SHA256SUMS.txt` → `sha256sum -c`. Herhangi bir adım başarısızsa durur. Hiçbir şey yüklemez. Ardından `CONFLICT_CHECK.md`'yi tazelemek için o dosyadaki curl döngüsü yeniden çalıştırılır.

### Sloganın yansıdığı alanlar (yalnız `/tr/` ve `/en/` ana sayfaları)
1. `<title>` → `Kanus — <slogan, sondaki nokta olmadan>`
2. `og:title` (aynı metin)
3. JSON-LD `WebPage.name` (aynı metin) ve `WebPage.slogan`
4. JSON-LD `Organization.slogan`
5. Sayfa başlığı `<h1>`
Meta description, `og:description`, alt sayfalar ve sitemap sloganı içermez. Doğrulama: test sloganıyla geçici kopyada üretildi → TR/EN için 6'şar yansıma, eski slogan 0.

## Yükleme öncesi
1. `CONFLICT_CHECK.md`: paketteki 16 yolun tamamı bugün SPA fallback → üzerine yazma yok (yüklemeden hemen önce tekrar çalıştır).
2. `cd public_html && sha256sum -c ../SHA256SUMS.txt`.
3. Onaylar: slogan (Demir), yayın zamanı.

## Yükleme sonrası
`./VERIFY_AFTER_UPLOAD.sh` → hepsi PASS olmalı: 16 dosyanın SHA256'sı sunucuyla aynı, dizin URL'leri canonical'lı, robots/sitemap gerçek dosya, 6 hukuki yol SPA fallback (paketle gelmedi) ve sitemap'te yok, `/` ve SPA fallback çalışıyor.

## Geri alma
`public_html/tr/`, `public_html/en/`, `public_html/robots.txt`, `public_html/sitemap.xml` silinir → yollar bugünkü SPA fallback'e döner. Gerekirse hPanel'den hcdn önbelleği temizlenir.

## Bilinen etki / sınır
- SPA kök `/`: canonical `https://kanus.org/`, `lang="en"`, hreflang yok — değişmez; hreflang kümesinin dışında (x-default=/tr/).
- Statik sayfalar `localStorage['kanus.lang']`'ı URL diline yazar (URL dili öncelikli).
- de/ru yalnız SPA'da istemci tarafı; URL'leri olmadığı için hreflang'e eklenmedi.
