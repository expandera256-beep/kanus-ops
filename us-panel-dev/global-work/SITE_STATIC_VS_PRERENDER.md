# kanus.org — statik /tr–/en vs. SPA prerender/SSG (teknik not, 06.11.2026 TSİ)

## Doğrulanan durum (salt okuma curl)
- Hostinger shared/cloud, hcdn CDN; A 84.32.84.209, 2.57.91.238; www → www.kanus.org.cdn.hstgr.net (ORG/HOST raporu). Yayın: dosya yükleme (File Manager/FTP); git/Node build/auto-deploy yok (HOST raporu).
- Kök `index.html` (last-modified 03.11.2026 23:22 GMT = 04.11.2026 02:22 TSİ) → Vite + React Router SPA, `/assets/index-uscommerce.js?v=20261104-usticari` (248 KB). Dil: istemci tarafı, `localStorage['kanus.lang']`, sözlükler `{en,tr,de,ru}`, varsayılan **en** (`qy()` fonksiyonu). URL'de dil yok.
- `/tr/`, `/en/`, `/robots.txt`, `/sitemap.xml` ve tüm alt yollar → aynı 1600 B SPA kabuğu (HTTP 200 text/html). Sebep `.htaccess`: `RewriteCond %{REQUEST_FILENAME} !-f` / `!-d` / `RewriteRule ^ index.html [L]`.
- SPA kaynağı ve build çıktısı kutuda yok (`source/` zip'i ve from-demir-20261106 içinde `index-uscommerce`/`kanus.lang` eşleşmesi yok; yalnız PHP `kanus/api/`).
- §7.2 düzeltmesi: "çok dilli site" = PARTIAL (istemci tarafında 4 dil var), ama URL bazlı /tr–/en, hreflang, gerçek sitemap/robots = MISSING/BROKEN.

## Karşılaştırma
| Ölçüt | A) Statik /tr/, /en/ (önerilen ilk adım) | B) SPA prerender/SSG |
|---|---|---|
| SEO/hreflang güvenilirliği | Yüksek: her URL'de sunucu HTML'inde lang, canonical, hreflang, OG, JSON-LD | Yüksek, ama yalnız prerender her rota×dil için doğru head üretirse; React Router'a URL dil öneki eklenmeli |
| Crawler JS ihtiyacı | Yok | Prerender edilmiş rotalarda yok; edilmeyenlerde (fallback) var |
| Performans | ~2.7 KB/sayfa, JS/harici font yok | SPA bundle (248 KB JS) + hidrasyon; ilk boyama HTML ile hızlanır |
| Bakım | İçerik iki yerde (SPA sözlüğü + `content.json`); senkron disiplini gerekir | Tek kaynak; ama build zinciri gerekir |
| Kaynak gereksinimi | Yok — mevcut SPA'ya dokunmaz | **SPA kaynağı gerekli** (Demir'den) ya da canlıdan indirilen dist üzerinde headless (Chromium) prerender — kırılgan, her SPA sürümünde tekrar |
| Barındırma | Mevcut `.htaccess` DEĞİŞMEDEN çalışır (`!-d` → `/tr/` dizini, DirectoryIndex → `index.html`) | Statik çıktı yine dosya olarak yüklenir; rota başına dizin; yerel/CI'da Node + headless tarayıcı |
| sitemap/robots | Gerçek dosya yüklenince `!-f` sayesinde doğrudan sunulur | Aynı |
| Geri alma | Eklenen 22 dosyayı sil → bugünkü hâl | Tüm SPA dist'i önceki sürümle değiştirmek gerekir (yedek şart) |
| Risk | SPA kök `/` kendi canonical'ında kalır (hreflang kümesi dışında) | SPA'yı bozma riski; tüm site yeniden yüklenir |

## Önerilen Apache kuralı (DEĞİŞİKLİK GEREKMİYOR)
Mevcut kural statik dizinleri zaten ayırır:
```apache
RewriteEngine On
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule ^ index.html [L]
```
İsteğe bağlı (ayrı onay): `/tr` → `/tr/` 301 için `DirectorySlash On` varsayılan zaten yönlendirir; ek kural gerekmez. `/` için dil yönlendirmesi önerilmez (SPA'yı etkiler).

## kanus.lang uyumu
- URL dili önceliklidir: statik sayfa açılınca `localStorage['kanus.lang']` = URL dili (`tr`/`en`) yazılır; kullanıcı SPA'ya geçerse aynı dili görür. SPA'nın kendi seçimi sonra yine geçerlidir.
- de/ru: SPA'da yalnız istemci tarafı; URL'leri olmadığı için hreflang'e EKLENMEDİ (hreflang yalnız gerçekten var olan, dili sabit URL'leri göstermeli). İleride `/de/`, `/ru/` statik sayfaları eklenirse kümeye eklenir; içerik metinleri ORG/Demir onayı ister.

## Varsayımlar
- `DirectoryIndex index.html` açık (Hostinger varsayılanı); yüklemeden sonra `VERIFY_AFTER_UPLOAD.sh` ile doğrulanır.
- hcdn önbelleği yeni dosyaları kısa sürede yansıtır; gerekirse hPanel'den temizlenir.
- `/tr/...` yollarını SPA kendi içinde kullanmıyor (bundle'da `/tr/` rota dizesi yok).

## Geri alma
`public_html/tr/`, `public_html/en/`, `robots.txt`, `sitemap.xml` silinir → yollar yeniden SPA fallback döner. Paket: `global-work/site/release/` (README_RELEASE.md, CONFLICT_CHECK.md, VERIFY_AFTER_UPLOAD.sh).

## Öneri
1) A'yı (statik) onayla ve yükle; 2) SPA kaynağı Demir'den gelirse B'ye (veya SPA'ya `/tr`, `/en` rotaları + head yönetimi) geç; o zaman `content.json` metinleri tek kaynağa taşınır.
