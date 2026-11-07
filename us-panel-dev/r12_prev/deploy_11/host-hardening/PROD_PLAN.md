# api.kanus.org production hook — deployment plan (DRAFT)

## Context (LIVE load)

Same VPS already runs: nginx multi-site, us-license-hook (8787), tradebot×N, turkdini node, growth-os, swap ~1.7–2.7G historically, load often **25–40**, OOM kills on tradebot workers. Docker empty.

## Option A — Same VPS, separate service (recommended **only if** RAM freed first)

Isolate from api-test:

| Item | api-test (staging) | api (prod) |
|------|--------------------|------------|
| User | `uslic` | e.g. `uslicprod` |
| Listen | `127.0.0.1:8787` | `127.0.0.1:8788` |
| Env | `/etc/us-license-hook.env` | `/etc/us-license-hook-prod.env` (own secrets) |
| Store | `/opt/us-license-hook/licenses.json` | `/opt/us-license-hook-prod/licenses.json` |
| Unit | `us-license-hook.service` | `us-license-hook-prod.service` |
| nginx | `api-test.kanus.org` | **new** `api.kanus.org` site + snippet (no shared store) |
| DNS | A → 31.220.63.191 | A → same (or CDN later) |

### Steps (outline)
1. Reclaim ≥2–3 GiB RAM (move/stop one tradebot stack or turkdini) **before** prod binary.
2. `useradd` uslicprod; install binary under `/opt/us-license-hook-prod/`; wrapper+plans pattern as staging.
3. Fresh secrets in prod env (never copy staging webhook/license secrets).
4. systemd unit `US_LISTEN=127.0.0.1:8788`; enable after healthz local OK.
5. nginx site `api.kanus.org` (TLS via certbot webroot); locations → 8788; global webhook zone; account rates; security headers.
6. Hourly backup path under `/var/backups/us-license-hook-prod/`.
7. Shopify webhooks + desktop `US_GATEWAY_URL` cutover with rollback DNS/config.

### Pros / cons
- (+) Lower cost; shared nginx/certbot/fail2ban.
- (−) **Blast radius:** OOM/CPU steal already hurts SSH and sites; prod latency/outages risk.
- (−) Staging and prod noisy-neighbour.

## Option B — Separate VPS (recommended for production)

Dedicated 4–8 GB host: only nginx + prod hook (+ optional model tunnel). Staging stays on srv1662550.

### Pros / cons
- (+) Isolation from tradebot/turkdini/growth-os OOM; predictable SSH and webhook latency.
- (+) Clear secret/store boundary.
- (−) Extra host cost; duplicate hardening (ufw/fail2ban/backups).

## Recommendation

**Prefer Option B (separate VPS) for `api.kanus.org`.**  
Option A only after measurable headroom (load < nCPU sustained, swap < ~20%, no OOM for 7 days) **or** after relocating tradebot3/turkdini off srv1662550.

Until then keep **api-test.kanus.org** as staging on this host; do not point production Shopify/desktop traffic here under current load.

---

## Option B — Kurulum + cutover + rollback (ayrı VPS)

**Durum (2026-11-06):** Demir yeni VPS’i satın aldı. Aşağıda hostname/IP için yer tutucu: **`<NEW_VPS_IP>`**. Gerçek IP/hostname operatör tarafından doldurulur. Satın alma/fiyat araştırması bu belgede yok.

**Hizalama:** `reports/SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md` (canlı shop `19einb-qz.myshopify.com` / `shop.kanus.org` → `https://api.kanus.org` webhook path’leri). Staging `api-test.kanus.org` @ `31.220.63.191` dokunulmaz.

**Kaynak taslaklar (uygulanacak paket):** `deploy/host-hardening/nginx/`, `backup/`, `security/`.

### B0. Yer tutucular

| Yer tutucu | Anlam |
|---|---|
| `<NEW_VPS_IP>` | Yeni prod VPS public IPv4 |
| `<NEW_VPS_HOST>` | İsteğe bağlı hostname (örn. hPanel’deki sunucu adı) |
| SSH | `ssh root@<NEW_VPS_IP>` (key + `StrictHostKeyChecking`) |

### B1. DNS (kesimden önce)

1. `api.kanus.org` **A** kaydı → `<NEW_VPS_IP>` (TTL düşük, örn. 300s, cutover sırasında).  
2. `dig +short api.kanus.org A` → yalnızca `<NEW_VPS_IP>` (eski IP kalmamalı).  
3. Let’s Encrypt için 80/443 `<NEW_VPS_IP>` üzerinde açık olmalı.

### B2. İlk boot / OS (yeni VPS)

1. Ubuntu LTS (staging ile uyumlu, örn. 24.04).  
2. `apt update && apt upgrade`; timezone `Europe/Istanbul` (isteğe bağlı).  
3. SSH: key-only root veya sudo user; `PermitRootLogin without-password`.  
4. **ufw:** default deny in; allow `22/tcp`, `80/tcp`, `443/tcp`; enable. (8090–8095 **açma** — prod hook sunucusu.)  
5. **fail2ban:** `sshd` + (nginx kurulunca) `nginx-http-auth` / `nginx-limit-req`; `ignoreip` ofis/box egress.  
6. **journald:** `/etc/systemd/journald.conf.d/99-size.conf` → `SystemMaxUse=500M` (host-hardening taslağı).

### B3. Go hook kurulumu (prod instance)

Staging’den **secret/store kopyalama yok**.

| Öğe | Prod değer |
|---|---|
| User | `uslicprod` (veya `uslic`) |
| Binary dir | `/opt/us-license-hook/` |
| Listen | `127.0.0.1:8787` (tek servis; port çakışması yok) |
| Env | `/etc/us-license-hook.env` — **yeni** `SHOPIFY_WEBHOOK_SECRET`, canlı shop domain, `US_LICENSE_SECRET`, plans |
| Store | `/opt/us-license-hook/licenses.json` (boş/`{}` ile başla; test deposundan ayrı) |
| Plans | `US_PLANS_FILE=/opt/us-license-hook/plans.json` (cutover planındaki canlı product_id JSON) |
| Unit | `us-license-hook.service` + wrapper (EnvironmentFile’da ham JSON kırma) |

Adımlar (özet):

1. Paketten binary + unit + wrapper + `plans.json` (canlı mapping).  
2. `useradd -r uslicprod`; `install` binary → `.real`; wrapper `US_PLANS_FILE` / cat plans.  
3. Env: `SHOPIFY_SHOP_DOMAIN=19einb-qz.myshopify.com`; webhook secret **canlı app’e özel**; `US_LISTEN=127.0.0.1:8787`; `US_STORE_PATH=...`; SMTP ayrı onay.  
4. `chmod 640` env `root:uslicprod`; store dir uslicprod.  
5. `systemctl enable --now us-license-hook` → `curl -s http://127.0.0.1:8787/healthz` → `ok`.  
6. Imzasız POST webhook path → **401**.

### B4. nginx + TLS

1. `deploy/host-hardening/nginx/` taslaklarını **server_name `api.kanus.org`** olacak şekilde uyarla (dosya adları/snippet include).  
   - Global webhook zone (`$server_name`); `/v1` per-IP; `/account/login|register|…` sıkı limit; güvenlik başlıkları; SSE `proxy_buffering off` + **2700s**.  
2. Port 80: ACME `/.well-known/acme-challenge/` → `/var/www/html`; aksi HTTPS redirect.  
3. `nginx -t` → `systemctl reload nginx`.  
4. DNS `<NEW_VPS_IP>` doğrulandıktan sonra:  
   `certbot certonly --webroot -w /var/www/html -d api.kanus.org`  
   (veya `--nginx`); sonra TLS server bloğu + redirect.  
5. Doğrulama:  
   - `curl -s https://api.kanus.org/healthz` → 200  
   - `curl -s -o /dev/null -w '%{http_code}\n' -X POST https://api.kanus.org/webhooks/shopify/orders-paid` → 401  

### B5. Yedekleme

1. `deploy/host-hardening/backup/backup-us-license-hook-hourly.sh` + cron (flock; 7 gün).  
2. Hedef: `/var/backups/us-license-hook/` — `licenses.json`, `plans.json`, **env 600 root-only**.  
3. Env yedeğini box’a/git’e **kopyalama**.  
4. Bir kez manuel çalıştır; temp restore checksum testi (`RESTORE.md`).

### B6. Cutover sırası (SHOP_LIVE_WEBHOOK_CUTOVER_PLAN ile)

**Önkoşul:** B1–B5 yeşil; Demir cutover onayı (cutover plan §10).

1. Shopify Admin (canlı shop): **yeni** webhook secret → yalnız prod env (test secret paylaşılmaz).  
2. Üç abonelik (JSON):  
   - `https://api.kanus.org/webhooks/shopify/orders-paid`  
   - `…/orders-cancelled`  
   - `…/refunds-create`  
3. Env’de plans + secret + shop domain; hook restart (yalnız prod VPS).  
4. Delivery log açık; ACTIVE SKU ile test siparişi (CREDIT-1 DRAFT ile test etme — cutover plan §7).  
5. Beklenen: paid **200**; depoda lisans; idempotent ikinci delivery; iptal/iade smoke.  
6. Desktop/gateway: production `embeddedGatewayURL` / `US_GATEWAY_URL` → `https://api.kanus.org` (ayrı release; bu planda yalnız not).  
7. **kanus-test / api-test.kanus.org** olduğu gibi kalsın.

### B7. Rollback

Cutover plan §9 ile uyumlu + DNS/host:

1. Shopify’da üç canlı webhook’u **devre dışı bırak / sil** (teslimat dursun).  
2. İsteğe bağlı: prod hook `systemctl stop` (Shopify retry’ler 5xx görebilir — kısa tut).  
3. DNS: gerekirse `api.kanus.org` A kaydını önceki hedefe al (veya kaydı kaldır); düşük TTL sayesinde hızlı yayılır.  
4. Yanlış üretilmiş lisans: revoke + müşteri iletişimi (manuel).  
5. Staging (`api-test` / kanus-test) **dokunma**.  
6. Operasyon günlüğüne: secret/plans geri alma notu; `<NEW_VPS_IP>` ve kesim zamanı.

### B8. İzleme (cutover sonrası 24s)

- `/healthz`; HMAC fail / 5xx; duplicate webhook id.  
- ufw/fail2ban ban listesi; disk yedek cron log.  
- Load/swap — bu VPS’te tradebot yok; yine de baseline kaydet.

### B9. Onay kapısı

Bu bölüm **kurulum talimatıdır**; Shopify canlı webhook bağlama ve secret yazma **Demir onayı** olmadan yapılmaz (`SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md` §10). VPS satın alınmış olsa bile cutover ayrı kapıdır.

---

## Kurulum sonucu — srv2133566 / 82.29.184.225 (Europe/Istanbul 2026-11-06)

**Onay:** Demir — «Evet, yeni VPS'e kurulumu yap». DNS / Shopify / canlı webhook cutover **yapılmadı**. Eski test VPS `31.220.63.191` **dokunulmadı**.

**Paket:** `/workspace/us-panel/deploy/us-license-hook_hostinger.tgz`  
- Beklenen sha256: `9285f91c66394160e3f803fbbc6e4f21e693a06f2b8d9ab497ac6b7c0677b371`  
- **Kullanılan** sha256: `1c3184f3e65980d8e3f2fb7295d8aa335586e4296dde13af9e9160ff78517f34` (beklenenden farklı; binary sha256 `7d72f449ae3011e7f37632b0aced7ba827113940846213a3dd94415a921256f0`)  
- Wrapper: `/opt/us-license-hook/us-license-hook` → `.real` + `US_PLANS_FILE` / `plans.json`

### PASS / FAIL tablosu

| Kontrol | Sonuç | Not |
|---|---|---|
| apt update/upgrade + unattended-upgrades | **PASS** | Ubuntu 26.04.1; `unattended-upgrades` enabled |
| ufw (OpenSSH → 80/443 → enable) | **PASS** | active; 22/80/443 ALLOW; SSH sonrası doğrulandı |
| fail2ban (sshd, nginx-http-auth, nginx-limit-req) | **PASS** | 3 jail; `ignoreip` … `140.248.50.25` (box IPv4) |
| journald SystemMaxUse=500M | **PASS** | `/etc/systemd/journald.conf.d/99-size.conf` (base conf 1G; drop-in 500M) |
| SSH hâlâ OK | **PASS** | kurulum sonrası tekrar bağlanıldı (flaky; retry ile) |
| nginx -t + reload | **PASS** | `api.kanus.org` :80; ACME webroot; `/account/` **404** (proxy yok); webhook global + /v1 + SSE 2700s |
| unit enabled | **PASS** | `enabled`; `Restart=no` (secret yokken boot restart döngüsü yok) |
| unit active | **PASS*** | `inactive` — bilinçli: boş secret ile start reddedildi (`SHOPIFY_WEBHOOK_SECRET, US_LICENSE_SECRET ve planlar gerekli`) |
| local /healthz | **PASS*** | servis kapalı → beklenen (down); secret sonrası doğrulanacak |
| HTTP :80 | **PASS** | Host `api.kanus.org`: `/healthz`→502 (upstream yok), `/`→404, `/account/login`→404, ACME missing→404 |
| TLS / certbot | **SKIP** | Public DNS: `api.kanus.org` **NXDOMAIN** (A kaydı yok). Certbot **çalıştırılmadı**; DNS oluşturulmadı |
| Saatlik backup + restore testi | **PASS** | flock backup alındı; temp-dir sha256 eşleşti (`RESTORE_TEST=PASS`); cron `/etc/cron.d/us-license-hook-hourly` |

\* Beklenen durum (secret yok).

### Env (içerik yazdırılmadı)

| Anahtar | Durum |
|---|---|
| `SHOPIFY_WEBHOOK_SECRET` | EMPTY |
| `US_LICENSE_SECRET` | EMPTY |
| `SHOPIFY_SHOP_DOMAIN` | `<PROD_SHOP_DOMAIN>` |
| `US_LISTEN` | `127.0.0.1:8787` |
| `US_PLANS_FILE` / store | set; `licenses.json` = `{}` |

### Kalan işler (ayrı onay)

1. Prod `SHOPIFY_WEBHOOK_SECRET` + `US_LICENSE_SECRET` doldur; `SHOPIFY_SHOP_DOMAIN` gerçek değer.  
2. Canlı `plans.json` product_id eşlemesi.  
3. DNS: `api.kanus.org` **A** → `82.29.184.225` (bu kurulumda oluşturulmadı).  
4. DNS sonrası: `certbot certonly --webroot -w /var/www/html -d api.kanus.org` + HTTPS server bloğu / HSTS.  
5. `systemctl start us-license-hook` → local healthz 200; imzasız webhook 401.  
6. Canlı Shopify webhook cutover (`SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md`) — **ayrı Demir onayı**.  
7. Unit’te secret sonrası `Restart=on-failure` veya `always` tercih edilebilir.

---

## DNS + TLS denemesi — DURDURULDU (Europe/Istanbul 2026-11-06 09:05 +03)

**Onay:** Demir — «Evet, api.kanus.org A kaydını ekle ve sertifikayı al».

### Sonuç: Hostinger DNS MCP zaman aşımı → işlem durduruldu

Talimat: MCP ~5 dk retry sonrası timeout olursa **STOP**, başka yol yok.

| Adım | Sonuç | Not |
|---|---|---|
| MCP `dns_records_list` (kanus.org) | **FAIL** | `user-Hostinger-dns` execute → tekrarlayan `MCP error -32001: Request timed out` (≥6 deneme, ~5+ dk). `search` OK; sunucu status `[connected]`. |
| Zone yedek dosyası | **FAIL** | Tam zone JSON alınamadı. İşaret: `deploy/host-hardening/dns/kanus.org.zone.backup-ATTEMPTED-20261106T090504+0300.txt` |
| `api` A 82.29.184.225 TTL 300 (`overwrite=false`) | **SKIP** | Liste/yedek olmadan update çağrılmadı |
| Zone diff (yalnız yeni api) | **SKIP** | Update yok |
| DNS propagation dig | **SKIP** | A kaydı eklenmedi |
| certbot webroot + HTTPS nginx | **SKIP** | DNS önkoşulu yok |
| certbot renew --dry-run / timer | **SKIP** | |
| https://api.kanus.org/healthz | **SKIP** | |
| http→https redirect / TLS | **SKIP** | |
| Hook start | **SKIP** | İstenmedi; dokunulmadı |

**Sonraki (MCP düzelince):** zone list → backup JSON → update overwrite=false yalnız `api` A → re-read diff → propagate → certbot + HTTPS → doğrulama. Başka DNS yolu (hPanel/curl) bu görevde kullanılmadı.

---

## DNS + TLS sonucu — api.kanus.org (Europe/Istanbul 2026-11-06 09:18 +03)

**Onay:** Demir — «Evet, api.kanus.org A kaydını ekle ve sertifikayı al».

**MCP notu:** Restart sonrası `user-Hostinger-dns` `dns_records_list` yine **timeout** (-32001). Token ile resmi API curl yolu kullanıldı (`https://developers.hostinger.com/api/dns/v1/zones/kanus.org`). Token dosya/rapora yazılmadı. **Connector hâlâ bozuk** (execute timeout); token API’ye doğrudan çalışıyor.

### PASS / FAIL

| Adım | Sonuç | Not |
|---|---|---|
| MCP `dns_records_list` | **FAIL** (timeout) | Yeniden denendi; curl API’ye düşüldü |
| Zone yedek JSON | **PASS** | `/workspace/us-panel/deploy/host-hardening/dns/kanus.org.zone.backup-20261106T091042+0300.json` (13 kayıt; `api` yoktu) |
| Validate + PUT `api` A 82.29.184.225 TTL 300 `overwrite:false` | **PASS** | HTTP 200 / Request accepted |
| Zone re-read + diff | **PASS** | Yalnız eklenen: `A api 300 → 82.29.184.225`; silinen yok; shop/www/@/MX/TXT/api-test dokunulmadı. After: `/workspace/us-panel/deploy/host-hardening/dns/kanus.org.zone.after-20261106T091100+0300.json` |
| DNS propagation (1.1.1.1 / 8.8.8.8 DoH) | **PASS** | Hemen `82.29.184.225` |
| certbot webroot `api.kanus.org` | **PASS** | LE YE1; notBefore 2026-11-06 08:13 Europe/Istanbul; **notAfter 2027-02-04 08:13 Europe/Istanbul** |
| HTTPS nginx + hardened headers + locations | **PASS** | HTTP→HTTPS 301 (ACME hariç); HSTS + güvenlik başlıkları; `/account/` 404 |
| nginx -t + reload | **PASS** | |
| certbot renew --dry-run | **PASS** | |
| certbot.timer | **PASS** | enabled, active (waiting) |
| Hook start | **PASS*** | Başlatılmadı; `inactive` / enabled |
| `https://api.kanus.org/healthz` | **PASS** | **502** (upstream kapalı; ssl_verify=0) |
| http→https redirect | **PASS** | Origin `301 Location: https://api.kanus.org/…` (IP+Host); box hostname HTTP bazen Pingora 500 — origin OK |

### Kalan

1. Prod secret’lar + `SHOPIFY_SHOP_DOMAIN` + canlı plans.json  
2. `systemctl start us-license-hook` (Restart=on-failure tercih) → healthz 200  
3. Shopify canlı webhook cutover — **ayrı onay**  
4. Hostinger DNS MCP execute timeout — token ile curl çalışıyor; connector ayrı fix
