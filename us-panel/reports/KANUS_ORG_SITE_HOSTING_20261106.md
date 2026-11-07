# kanus.org Hostinger site — salt okuma envanteri

**Tarih:** 2026-11-06 (Europe/Istanbul)  
**İstek:** GLOBAL → HOST (deploy yöntemi, `.htaccess` SPA fallback, kutu kopyası)  
**Kapsam:** Salt okuma. Yazma / deploy / upload yok.  
**Kaynaklar:** Hostinger MCP (`user-Hostinger-hosting`), Hostinger REST (`developers.hostinger.com`), dışarıdan `curl` (kanus.org). Token yazdırılmadı / kaydedilmedi.

---

## Özet

| Soru | Sonuç | Kanıt durumu |
|------|--------|--------------|
| 1) Nasıl yayınlanıyor? | Statik dosya (Vite SPA build + PHP `api/`) — **Git bağlı değil**, **Node.js build yok**, **Builder/Horizons değil** | **Doğrulandı** |
| 1) Kim yayınlıyor? | API’de deploy geçmişi / yükleyen kimlik yok | **Bilinmiyor** |
| 2) SPA fallback `.htaccess`? | Var; `!-f` / `!-d` → `index.html` | **Doğrulandı** |
| 2) `/robots.txt` ve `/sitemap.xml` HTML mi? | Evet, HTTP 200, `content-type: text/html`, gövde = `index.html` | **Doğrulandı** |
| 3) Kutuda site kaynağı / build? | Vite SPA kaynağı veya `assets/` build kopyası **yok** (GLOBAL iskeleti hariç tutuldu) | **Doğrulandı (yok)** |

---

## 1) Yayın yöntemi

### Hosting kaydı (doğrulandı)

`hosting_websites_list?domain=kanus.org` / `GET /api/hosting/v1/websites?domain=kanus.org`:

| Alan | Değer |
|------|--------|
| domain | `kanus.org` |
| username | `u346856246` |
| order_id | `1010284397` |
| website_type | `other` (wordpress / builder / horizons / nodejs **değil**) |
| vhost_type | `main` |
| is_enabled | `true` |
| root_directory | `/home/u346856246/domains/kanus.org/public_html` |
| created_at | `2026-11-02T17:38:51Z` → **2026-11-02 20:38:51 Europe/Istanbul** |

Dışarıdan yanıt başlıkları: `platform: hostinger`, `panel: hpanel`, `server: hcdn` — Shared/Cloud web hosting + hCDN.

### Git / Node deploy (doğrulandı — yok)

| Kontrol | Sonuç |
|---------|--------|
| `hosting_git_list-website-repositories` | `[]` (bağlı repo yok) |
| `hosting_get-deployment` | `build: null`, `buildSettings: null`, `gitAutoDeployment: null` |
| Not | “No Node.js builds found… Static/PHP deploys finish before hosting_deploy returns” |

**Sonuç:** Aktif **Git auto-deploy** ve **Hostinger Node.js build** yok. Site, document root’a konmuş **statik Vite çıktısı** + `api/` PHP dosyaları gibi görünüyor (hPanel File Manager, FTP/SFTP veya tek seferlik `hosting_deploy` / arşiv yükleme ile). Hangisinin kullanıldığı API’den görülmüyor.

### Document root içeriği (doğrulandı, özet)

`hosting_files_list-website-and-directories` (max_depth=2):

- `index.html` (1600 B) — SPA kabuğu; script: `/assets/index-uscommerce.js?v=20261104-usticari`
- `assets/` — birçok `index-us*.js` + `index-BaRMt_G_.css` (Vite tarzı hash/sürüm isimleri)
- `api/` — PHP (`license_mint.php`, `bootstrap.php`, `shopify.php`, …) + `api/.htaccess`
- `downloads/` — `USpanel*.exe`, `US-kaynak.zip`, …
- `storage/` — JSON + Deny `.htaccess`
- `.htaccess` (172 B) — kök SPA kuralı
- `deneme.html`, `us.html`

**Bilinmiyor:** Kim / ne zaman yükledi (statik upload için deploy history endpoint’i yok veya bu sitede boş). `Get Git auto-deployment settings` ayrı okuma op’u search’te gelmedi; `hosting_get-deployment.gitAutoDeployment=null` yeterli sayıldı.

---

## 2) `.htaccess` SPA fallback + robots/sitemap

### Kök `.htaccess` (doğrulandı — birebir)

`hosting_files_website-content` path=`.htaccess`:

```
RewriteEngine On
RewriteRule ^storage/ - [F,L]
RewriteRule ^api/ - [L]
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule ^ index.html [L]
```

- `storage/` → 403  
- `api/` → rewrite yok (PHP’ye bırak)  
- Dosya/dizin yoksa → `index.html` (**SPA fallback**)

### `index.html` (doğrulandı)

Aynı 1600 B gövde; `<div id="root"></div>` + `index-uscommerce.js?v=20261104-usticari`. Title: `KAN.US) — Space for your next idea`.

### curl: robots.txt / sitemap.xml (doğrulandı)

| URL | Status | content-type | Boyut | İlk baytlar |
|-----|--------|--------------|-------|-------------|
| `https://kanus.org/robots.txt` | **200** | **text/html** | 1600 | `<!DOCTYPE html><html lang="en">…` (SPA `index.html`) |
| `https://kanus.org/sitemap.xml` | **200** | **text/html** | 1600 | Aynı SPA kabuğu |

Ortak başlıklar: `platform: hostinger`, `panel: hpanel`, `server: hcdn`.  
`Last-Modified: Tue, 03 Nov 2026 23:22:22 GMT` → **2026-11-04 02:22:22 Europe/Istanbul**.

**Neden HTML:** Gerçek `robots.txt` / `sitemap.xml` dosyası yok → `!-f`/`!-d` kuralı `index.html` döndürüyor.

---

## 3) Kutuda kaynak / build kopyası

**Arama:** `rg` / `find` under `/workspace` ve `/home/box`; hariç: `/workspace/us-panel-dev/global-work/site/` (GLOBAL iskeleti), `node_modules`, `.git` içerikleri.

| Bulgu | Durum |
|-------|--------|
| `vite.config.*` | **Yok** |
| `package.json` içinde `"vite"` | **Yok** |
| Canlı `index-uscommerce.js` / `index-BaRMt_G_.css` kopyası | **Yok** |
| `/workspace/us-panel/dist`, `/workspace/us-panel-dev/dist` | US Panel **Windows/Linux binary**’leri — web SPA değil |
| PHP API parçası | `/workspace/us-panel/source/from-demir-20261106/kanus/api/license_mint.php` (+ test); kopya: `us-panel-dev/source/...`, `us-panel/work2/kanus/api/` — **yalnızca API PHP**, SPA değil |
| `/workspace/us-panel/work2/kanus/` | Python gateway / commerce / deploy (GCP) — **kanus.org Vite sitesi değil** |
| Marka notları | `docs/BRAND_A_COPY_20261106.md` canlı siteden metin çekmiş; kaynak repo değil |
| GLOBAL iskeleti | `/workspace/us-panel-dev/global-work/site/` — **hariç tutuldu** (GLOBAL’in kendi çalışması) |

**Sonuç:** Kutuda canlı kanus.org Vite SPA kaynağı veya derlenmiş `assets/` ağacı **yok**. Canlı site yalnızca Hostinger `public_html` üzerinde.

---

## Bilinmeyenler / sınırlar

1. Statik dosyaları kimin (Demir / ORG / başka bot) ve hangi araçla yüklediği — API history yok.  
2. Ayrı “Get Git auto-deployment settings” MCP op adı search’te çıkmadı; `gitAutoDeployment: null` ile çapraz doğrulandı.  
3. `api/secrets.local.php` listede var; **içerik okunmadı** (secret).  
4. Yazma / deploy yapılmadı.

---

## GLOBAL için kısa cevap (TR)

1. **Yayın:** Git deploy değil; Node.js Hostinger build değil; website builder değil. Statik Vite SPA + PHP `api/` dosya yüklemesi (hPanel/FTP veya eşdeğeri). **Kim yükledi: bilinmiyor.**  
2. **Evet**, kök `.htaccess` SPA fallback var (`RewriteCond %{REQUEST_FILENAME} !-f` / `!-d` → `index.html`). Bu yüzden `/robots.txt` ve `/sitemap.xml` şu an **HTML** dönüyor (200, text/html).  
3. Kutuda SPA kaynağı/build **yok** (GLOBAL’in `global-work/site/` iskeleti hariç; orası senin taslağın).
