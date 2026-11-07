# Hook taşıma kesim planı — api-test → srv2133566 (82.29.184.225)

**Tarih:** 2026-11-06 (Europe/Istanbul)  
**Durum:** YALNIZ PLAN + salt okuma envanter. Hiçbir VPS’te değişiklik yok. DNS yok. Servis durdurulmadı. Secret kopyalanmadı.  
**Demir kararı (KAN.US aktarımı, kendi sözleri):** «Lisans servisini asıl sunucuya (82.29.184.225) kesim planıyla taşı.»  
**Neden:** Test VPS srv1662550 (31.220.63.191) steal ~%65–96, load ~32; prod VPS boşta (`SRE_VPS_CAPACITY_20261106.md`, `TEST_VPS_STEAL_20261106.md`).

---

## 0. Özet (tek bakış)

| Madde | Değer |
|---|---|
| Kaynak | srv1662550 / `31.220.63.191` / `api-test.kanus.org` / shop `kanus-test.myshopify.com` |
| Hedef | srv2133566 / `82.29.184.225` (ayrı test instance; mevcut durmuş `api.kanus.org` prod kurulumu korunur) |
| Webhook URL | **aynı kalır:** `https://api-test.kanus.org/...` (Shopify tarafında URL değişmez) |
| HMAC | Test shop secret **aynı** kalır (box: `/workspace/secrets/shopify_webhook_secret.txt`; değer yazılmaz) |
| Veri | Kopyala, kaynak **silinmez**; eski servis kesimden sonra **24 saat durdurulmuş ama silinmemiş** |
| Tahmini kesinti | **~5–10 dk** (yazma dondurma 2–5 dk + DNS TTL 300s); ayrıntı §8 |
| Blokerler | (1) prod SSH timeout, (2) Demir’in DNS kesimi + eski stop onayı, (3) TTL 14400→300 ≥1 saat önce, (4) usta-wake /32 KAN doğrulaması |

**Bu planda yapılmayanlar:** tradebot / turkdini / growthos’a dokunmak; Round-2 sertleştirme taslaklarını uygulamak; canlı Shopify cutover; `api.kanus.org` prod hook’u başlatmak.

---

## 1. Salt okuma envanter — test VPS (2026-11-06 ~09:51 TSİ)

Tek SSH oturumu (`ConnectTimeout=60`, IdentitiesOnly, BatchMode). Ham çıktı: `reports/qa-20261106/hook_move_inventory_raw.txt`.

### 1.1 Servis

| Öğe | Değer |
|---|---|
| Unit | `/etc/systemd/system/us-license-hook.service` |
| User | `uslic` (uid 992, gid 982, home `/opt/us-license-hook`, nologin) |
| WorkingDirectory | `/opt/us-license-hook` |
| EnvironmentFile | `/etc/us-license-hook.env` (mode 600, root:root, 901 bayt, mtime 06.11 05:30 UTC) |
| ExecStart | `/opt/us-license-hook/us-license-hook-wrap.sh` |
| Durum | **active/running** (MainPID 1107497, since 05:30:17 UTC) |
| Listen | `127.0.0.1:8787` |

### 1.2 Env anahtar adları (değerler YOK)

`SHOPIFY_WEBHOOK_SECRET`, `SHOPIFY_SHOP_DOMAIN`, `US_LICENSE_SECRET`, `US_SMTP_ENABLE`, `US_STORE_PATH`, `US_LISTEN`, `US_PLANS_FILE`, `US_MODEL_UPSTREAM`, `US_MODEL_NAME`, `US_MODEL_TIMEOUT_SECONDS`, `US_MODEL_CONCURRENCY`, `US_MODEL_QUEUE_LIMIT`, `US_SMTP_HOST`, `US_SMTP_PORT`, `US_SMTP_USER`, `US_SMTP_FROM`, `US_SMTP_PASS`, `US_SESSION_SECRET`, `US_SESSION_TTL_SECONDS`, `US_SMTP_ALLOWLIST`.

Bilinen (önceki raporlar + bağlam; bu oturumda değer okunmadı): shop ≈ `kanus-test.myshopify.com`; `US_STORE_PATH` → `/opt/us-license-hook/licenses.json`; `US_LISTEN` → `127.0.0.1:8787`; `US_PLANS_FILE` → `/opt/us-license-hook/plans.json`.

### 1.3 Wrapper

`/opt/us-license-hook/us-license-hook-wrap.sh` (sha256 `fa3c843d82f475d89c7b6f92854b10c974882ea9c1b1321c8658f5920974c1c2`):

- `plans.json` varsa `US_PLANS_JSON` + `US_PLANS_FILE` export
- 0-bayt store’u `{}` ile iyileştirir; eksik dosya OK
- `exec us-license-hook.real`

### 1.4 Binary / plans

| Dosya | Boyut | mtime (UTC) | sha256 | mode/owner |
|---|---:|---|---|---|
| `us-license-hook.real` | 7454880 | 05:33:50 | `7d72f449ae3011e7f37632b0aced7ba827113940846213a3dd94415a921256f0` | 755 uslic:uslic |
| `us-license-hook-wrap.sh` | 512 | 02:49:22 | `fa3c843d…` | 755 root:uslic |
| `plans.json` | 571 | 02:41:39 | `33148639a2b965a7e94fc5c5b4c65e76a443fc5ac6686076f2f23301b1ea59e8` | 640 root:uslic |
| `us-license-hook` (eski stub) | 127 | 02:39:39 | — | 755 uslic:uslic |

Round7 accounts binary; paket tgz sha `1c3184f3e65980d8e3f2fb7295d8aa335586e4296dde13af9e9160ff78517f34` (`ACCOUNTS_DEPLOY_TEST_20261106.md`).

### 1.5 Veri (licenses / ledger / accounts)

Ayrı `accounts.*` veya `ledger.*` dosyası **yok**. Accounts + lisans ledger tek store’da:

| Dosya | Boyut | mtime (UTC) | mode | Not |
|---|---:|---|---|---|
| `/opt/us-license-hook/licenses.json` | **30328** | 06:46:45 | 600 uslic:uslic | Canlı store (accounts dahil Round7) |
| `licenses.json.lock` | 0 | 03:28:52 | 600 uslic:uslic | kilit |
| `/var/backups/us-license-hook/` | 2× licenses + 2× plans snapshot | 03:04 / 03:15 | 640 | cron `15 3 * * *` → `backup-us-license-hook-store.sh` |

Deploy yedekleri (kaynak silinmeyecek; taşıma dışı): `backup-r7-accounts-20261106-053004` (rollback birincil), diğer `backup-r7-*`, `backup-r5-*`, `backup-20261106-025443`, `backup-smtp-preenable-*`, `backup-e2e-*.json`.

**İkinci SSH (store JSON anahtarları)** 09:52 TSİ timeout; yapı önceki Round7 raporlarından: tek `licenses.json` ledger (lisans + account alanları). Kesim öncesi tekrar `jq 'keys'` salt okuma şart.

### 1.6 nginx

**Site** `/etc/nginx/sites-available/api-test.kanus.org` (enabled symlink):

- `:80` → ACME `/.well-known/acme-challenge/` root `/var/www/html`; aksi 301 HTTPS
- `:443` → `ssl_certificate` / `ssl_certificate_key` Let’s Encrypt; `include snippets/us-license-hook-locations.conf`; diğer `/` → 404

**Snippet** `/etc/nginx/snippets/us-license-hook-locations.conf` (3926 bayt, mtime 05:45 UTC) — `proxy_pass http://127.0.0.1:8787`:

- Webhooks: `orders-paid`, `orders-cancelled`, `refunds-create` — `limit_req zone=uslic_webhooks burst=20`; body 1m; timeout 30s
- Credits: `reserve`/`settle`/`status` — zone `uslic_credits`; status + models + chat SSE buffering off, 2700s
- `/healthz`
- `/v1/models`, `/v1/chat`, `/v1/chat/completions`
- `/account/login` + `/account/register` — zone `uslic_account burst=2`; Auth + X-Real-IP/XFF; timeout **90s**
- `^~ /account/` aynı proxy başlıkları
- `/webhooks/shopify` ve altı bilinmeyen → 404; `/internal/` → 404

**Rate zone** `/etc/nginx/conf.d/us-license-hook-ratelimit.conf`:

```
limit_req_zone $binary_remote_addr zone=uslic_webhooks:10m rate=10r/s;
limit_req_zone $binary_remote_addr zone=uslic_credits:10m rate=5r/s;
limit_req_zone $binary_remote_addr zone=uslic_account:10m rate=5r/m;
```

**Access log:** `nginx.conf` → `access_log /var/log/nginx/access.log;` (özel `log_format` tanımlı değil → varsayılan **combined**: `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"`). api-test’e özel access_log yok; XFF ayrı alan olarak yazılmıyor (istek başlığında varsa combined’de yok).

### 1.7 TLS (test VPS üzerinde)

| Öğe | Yol |
|---|---|
| live | `/etc/letsencrypt/live/api-test.kanus.org/{fullchain,privkey,cert,chain}.pem` |
| archive | `/etc/letsencrypt/archive/api-test.kanus.org/` |
| renewal | `authenticator=webroot`, `webroot_path=/var/www/html` |

**Not:** Sertifika makineye özeldir; prod’a kopyalanmaz. Kesimden sonra prod’da yeniden alınır (§5).

### 1.8 DNS (şu an)

| Kayıt | Tip | TTL | İçerik |
|---|---|---:|---|
| `api-test` | A | **14400** | `31.220.63.191` |
| `api` | A | 300 | `82.29.184.225` (prod hook; bu taşıma kapsamı dışı) |

---

## 2. Hedef mimari — çakışma önleme (ayrı test instance)

Prod VPS’te **zaten** durmuş bir prod kurulumu var (`PROD_PLAN.md` kurulum sonucu):

- Yol: `/opt/us-license-hook/`
- Unit: `us-license-hook`
- User: `uslic`
- Port: `127.0.0.1:8787`
- Amaç: gelecekte `api.kanus.org` canlı hook (boş secret/store; **başlatılmayacak**)

Aynı isim/yol/port’a test hook’u yazmak canlı hazırlığı ezer ve test/canlı karışır. Bu yüzden taşınan **test** hook ayrı instance olur:

| Öğe | Test instance (taşınan) | Prod instance (mevcut, durmuş) |
|---|---|---|
| Dizin | `/opt/us-license-hook-test/` | `/opt/us-license-hook/` |
| Binary | `…/us-license-hook.real` + `…-wrap.sh` | olduğu gibi (dokunma) |
| Unit | `us-license-hook-test.service` | `us-license-hook.service` (disabled/stopped) |
| User | `uslic-test` (yeni sistem kullanıcısı) | `uslic` |
| Listen | `127.0.0.1:8788` | `127.0.0.1:8787` |
| Env | `/etc/us-license-hook-test.env` (600 root:uslic-test) | `/etc/us-license-hook.env` |
| Store | `/opt/us-license-hook-test/licenses.json` | boş/ayrı (dokunma) |
| Plans | `/opt/us-license-hook-test/plans.json` | dokunma |
| nginx site | `api-test.kanus.org` → `127.0.0.1:8788` | `api.kanus.org` → `8787` (mevcut) |
| limit_req zone | `uslic_test_webhooks`, `uslic_test_credits`, `uslic_test_account` | mevcut prod zone adları (varsa) ayrı |

### 2.1 Test / canlı karışma riskleri ve önlemler

| Risk | Önlem |
|---|---|
| Yanlış env’ye canlı secret yazmak | Test env yalnız test HMAC/session; prod env’ye bu kesimde **hiç** yazma |
| İki hook’un aynı store’u açması | Ayrı dizin + ayrı `US_STORE_PATH`; prod store’a rsync yok |
| Port çakışması | 8788 vs 8787 |
| nginx zone adı çakışması (http context) | `uslic_test_*` öneki |
| Shopify’ın yanlış host’a gitmesi | URL host `api-test.kanus.org` kalır; yalnız A kaydı değişir |
| usta-wake / model upstream’in prod kurallarını bozması | Test env’deki `US_MODEL_*` aynen taşınır; KAN ile `/32` 82.29.184.225 doğrulanır (§7) |
| Operatör karışıklığı | Unit/path/user adlarında `-test` soneki; dokümantasyon bu tablo |
| Round-2 hardening’i yanlışlıkla uygulamak | Bu planda **uygulanmaz** (Demir: bekle) |

---

## 3. Kapsam (ne taşınır)

**Kopyalanır (kaynak silinmez):**

1. Binary `us-license-hook.real` (sha doğrulamalı) + wrapper (+ gerekirse stub)
2. `plans.json` (test product mapping)
3. `licenses.json` (+ kesim anı final rsync; accounts/ledger burada)
4. Env içeriği → **yeni** `/etc/us-license-hook-test.env` (HMAC = test shop; `US_LISTEN=127.0.0.1:8788`; path’ler `-test` dizinine)
5. nginx location/ratelimit mantığı (port 8788 + `uslic_test_*` zone)
6. Backup cron script uyarlaması (ayrı hedef dizin; tradebot vb. yok)

**Taşınmaz / dokunulmaz:**

- tradebot, turkdini, growthos ve diğer nginx siteleri (kaynak VPS)
- Prod `/opt/us-license-hook` + `us-license-hook.service` + `api.kanus.org` site
- Round-2 hardening taslaklarının uygulanması
- Canlı shop webhook’ları / `api.kanus.org` DNS

---

## 4. Hazırlık (geri alınabilir; Demir DNS/stop onayı ÖNCE değil)

> **Kapı:** Aşağıdakilerin kurulumu prod SSH açılınca yapılabilir; **DNS kesimi ve eski servisi durdurma** Demir planı görüp onayladıktan sonra.

1. **Prod SSH** erişimini düzelt (ayrı iş; fail2ban/ignoreip — KAN.US’un verdiği geri alınabilir altyapı onayı çerçevesinde). Bu planın uygulaması SSH olmadan **ilerlemez**.
2. `useradd -r -d /opt/us-license-hook-test -s /usr/sbin/nologin uslic-test`
3. Dizinler: `/opt/us-license-hook-test` (uslic-test), backup hedefi
4. Binary + wrap + plans kopyası (ön checksum); `US_LISTEN` henüz local smoke için 8788
5. Env dosyası oluştur (secret’ları scp/`install` ile; chate yapıştırma; mode 600)
6. Unit `us-license-hook-test.service` (enable etme veya enable+start yalnız local smoke için)
7. nginx: `sites-available/api-test.kanus.org` (prod’da yeni), snippet `us-license-hook-test-locations.conf` → `127.0.0.1:8788`, `conf.d/us-license-hook-test-ratelimit.conf` zone `uslic_test_*`
8. Local smoke (DNS hâlâ eski IP’deyken): `curl 127.0.0.1:8788/healthz` → 200; imzasız webhook → 401
9. **TTL düşür:** Hostinger’da `api-test` A TTL **14400 → 300**. En az **1 saat** bekle (eski resolver önbelleği). Zone yedeği al (overwrite=false append değil; TTL update). Başka kayıtlara dokunma.
10. LE: tercihen **DNS-01** ile `api-test.kanus.org` sertifikasını kesimden **önce** prod’da al (A kaydı hâlâ eski IP iken). Alternatif §5.

---

## 5. TLS stratejisi ve boşluk

| Seçenek | Ne zaman | Boşluk |
|---|---|---|
| **A — DNS-01 (tercih)** | Kesimden önce, A hâlâ 31.220… | HTTPS kesim anında hazır; boşluk yok |
| **B — webroot** | DNS A prod’a döndükten sonra `certbot certonly --webroot -w /var/www/html -d api-test.kanus.org` | DNS yayılımı + certbot süresi: **~2–10 dk** HTTPS hata / HTTP-only riski |
| Sertifika kopyası test→prod | Yapılma | Private key taşıma + yenileme path bozulması; **önermiyoruz** |

Öneri: A. DNS-01 Hostinger API TXT ile (token zaten var; connector timeout olursa resmi API). B yedek plan.

Mevcut prod’daki `api.kanus.org` LE cert’i (2027-02-04’e kadar) **api-test için kullanılamaz** (SNI/hostname farklı).

---

## 6. Kesim sırası (Demir onayı sonrası)

**Önkoşul checklist:** prod SSH OK · TTL=300 ≥1s · paralel kurulum + local smoke PASS · TLS hazır (A) veya B kabul · usta-wake /32 KAN OK · Demir «DNS kes + eski stop» onayı.

1. **Yazmayı dondur:** test VPS’te `systemctl stop us-license-hook` (Shopify bundan sonra 5xx görür → retry’ler birikir; URL aynı).
2. **Final sync:** `licenses.json` (+ plans, binary sha teyit) → prod `/opt/us-license-hook-test/` — `rsync -ac --checksum` veya `sha256sum` eşleşene kadar. Kaynak dosyalar **silinmez**.
3. **Yeni hook start:** prod’da `systemctl enable --now us-license-hook-test` → local `8788/healthz` = 200.
4. **DNS kesimi:** `api-test` A → `82.29.184.225` (TTL 300). Zone diff: yalnız bu kayıt.
5. **Yayılım bekle:** `dig @1.1.1.1` + `@8.8.8.8` → yalnız yeni IP.
6. **TLS (B ise):** webroot cert + HTTPS server blok enable + `nginx -t` + reload + `renew --dry-run`.
7. **Smoke (dışarıdan `https://api-test.kanus.org`):**

| Test | Beklenen |
|---|---|
| GET `/healthz` | **200** `ok` |
| POST imzasız `/webhooks/shopify/orders-paid` | **401** |
| POST imzalı test-HMAC (box secret dosyasından; değer basılmadan) | **200** veya iş kuralına göre anlamlı non-401 |
| Accounts login (bilinen QA hesap; parola chate yazılmaz) | **200** session |
| Login rate limit (tek sabit IP, burst) | **429** görünür |

8. Eski VPS: unit **stop** bırak (enable kalabilir); 24s silme/purge yok. tradebot vb. dokunulmaz.
9. Sonuç PASS/FAIL tablosunu Demir + KAN.US’a ilet; bu dosyaya ekle.

---

## 7. usta-wake /32

KAN.US: `82.29.184.225` için `/32` kuralı listede; yine de **KAN ile doğrula** (agent id `14d5a2d1-ad07-43c5-8d70-71aee454f899`) kesimden önce.

- Prod çıkış IP ölçümü: SSH timeout yüzünden henüz teyit edilemedi; Hostinger ataması `82.29.184.225`. SSH açılınca `curl -4 ifconfig.me` / `api.ipify.org` tekrar.
- Token/pin yerleşimi (`/etc/us-model/wake.token` + `wake.pin`, 600, hook kullanıcısı): KAN haber verince; **hook restart bu kesimle birleştirilir** (ayrı restart yok). Test instance kullanıcısı `uslic-test` ise pin/token sahibi ona göre.

---

## 8. Tahmini kesinti

| Bileşen | Süre | Gerekçe |
|---|---|---|
| TTL düşürme beklemesi | ≥60 dk | Kesinti değil; hazırlık. Şu an TTL **14400**; düşürülmezse splitter riski **saatler** |
| Yazma dondurma (stop → final rsync → start) | **2–5 dk** | Store ~30 KiB + binary ~7 MiB; test VPS steal yüzünden SSH/rsync yavaş olabilir |
| DNS yayılımı (TTL 300) | **0–5 dk** | Eski IP’ye giden istemci eski (durmuş) hook’a çarpar → 5xx/timeout |
| Cert (yalnız B) | +2–10 dk | DNS-01 ile 0 |
| Smoke | ~2–3 dk | Müşteri trafiğiyle örtüşebilir |

**Gerçekçi toplam (müşteriye görünen):** **yaklaşık 5–10 dakika** (DNS-01 + TTL=300 varsayımı).  
**Shopify:** başarısız webhook’ları otomatik yeniden dener; kısa 5xx penceresi genelde telafi olur. Uzun splitter (TTL 14400) **kabul edilmez**.

---

## 9. Geri dönüş (rollback)

1. DNS: `api-test` A → `31.220.63.191` (TTL 300).
2. Eski VPS: `systemctl start us-license-hook` (store kaynakta duruyor; kesimde silinmedi).
3. Prod test instance: `systemctl stop us-license-hook-test` (silme yok).
4. Smoke eski uçta healthz 200 + imzasız 401.
5. Kesim sırasında prod’da yazılmış yeni lisans/account varsa: ledger birleştirme **manuel** (nadir; freeze doğruysa olmamalı). Not düş.

Eski servis **24 saat** durdurulmuş+sağlam kalır; sonra ayrıca onayla temizlenir.

---

## 10. US `.11` deploy sırası

Kesimden sonra `.11` hedefi **yeni sunucu** (`us-license-hook-test` path/unit).  
Kesimden **önce** `.11` gelirse: sırayı **KAN.US’a sor** — ya kesimi önce bitir ya da `.11`’i geçici olarak eski VPS’e al (steal riski). Bu planda varsayılan: kesim önce, `.11` sonra yeni instance.

---

## 11. Önkoşullar / onay kapıları

| # | Kapı | Durum (06.11 ~09:52 TSİ) |
|---|---|---|
| 1 | Prod SSH (`82.29.184.225`) box’tan çalışır | **PASS (kesintili)** — ban değil; ConnectTimeout=30 + retry; ignoreip `140.248.50.0/24`; egress `82.29.184.225` |
| 2 | Demir planı gördü; DNS kesimi + eski `stop` için ayrı onay | **Bekliyor** |
| 3 | `api-test` TTL 300 + ≥1 saat bekleme | **Bekliyor** (şimdi 14400) |
| 4 | Paralel kurulum + local smoke | **PASS** — Faz 2 (2026-11-06 ~10:07 TSİ); unit şu an disabled |
| 5 | TLS (DNS-01 tercih) | SSH + Hostinger DNS TXT sonrası |
| 6 | usta-wake `/32` KAN doğrulaması | Prod egress **82.29.184.225** teyit; `/32` KAN ile doğrula |
| 7 | tradebot/turkdini/growthos / Round-2’ye dokunmama | Taahhüt |

---

## 12. Uygulama notları (operatör)

- SSH: `ConnectTimeout=60 IdentitiesOnly=yes BatchMode=yes`; test VPS’e en fazla 3 deneme; oturum sayısı az.
- Secret: chate/loga yapıştırma; `scp` + remote `install -m 600`.
- Checksum: her rsync sonrası `sha256sum` binary + `licenses.json`.
- nginx: `nginx -t` olmadan reload yok; diğer sitelere dokunma.
- Hostinger DNS: zone JSON yedek → değişiklik → diff (yalnız `api-test`).

---

## 13. Sonraki adımlar (bu belgeden sonra)

1. Demir’e plan özeti + **5–10 dk** kesinti + blokerler.
2. Prod SSH düzeltmesi (KAN.US / HOST paralel).
3. Onay gelince: §4 hazırlık → §6 kesim.
4. PASS/FAIL tablosunu bu dosyaya ekle.

---

*HOST — Faz 2 paralel kurulum uygulandı (prod test instance); DNS/eski stop yok.*

## 14. Faz 2 — Paralel kurulum sonucu (2026-11-06 ~10:07 TSİ)

**Kapsam:** Yalnız prod `srv2133566` üzerinde ayrı instance. Test VPS’te stop/değişiklik **yok**. DNS/TTL **yok**. Durmuş `api.kanus.org` / `/opt/us-license-hook` / `:8787` **dokunulmadı**.

### Kurulan

| Öğe | Değer |
|---|---|
| Dizin | `/opt/us-license-hook-test/` |
| User | `uslic-test` (yeni) |
| Unit | `us-license-hook-test.service` (`Restart=no`) |
| Listen | `127.0.0.1:8788` |
| Env | `/etc/us-license-hook-test.env` **640 root:uslic-test** (test shop secret’ları; listen/path ayarlı) |
| Binary sha | `7d72f449ae3011e7f37632b0aced7ba827113940846213a3dd94415a921256f0` |
| Plans sha | `33148639a2b965a7e94fc5c5b4c65e76a443fc5ac6686076f2f23301b1ea59e8` |
| Store snapshot | `licenses.json` 30328 B, sha `c2a1502239422a9e047633ed86905cae822174dd6aa98a06c80603e556e149be` (+ `.pristine`) |
| nginx | `sites-available/api-test.kanus.org` (HTTP-only, ACME webroot + locations→8788), enabled; `uslic_test_*` zone’ları |
| Scriptler | `deploy/hook-move/cutover.sh`, `rollback.sh` (varsayılan dry-run) |

### PASS/FAIL

| Kontrol | Sonuç | Not |
|---|---|---|
| Binary sha Round7 | **PASS** | test VPS ile aynı `7d72f449…` |
| User/unit/path/port ayrımı | **PASS** | prod `:8787` dinlemiyor; test instance `:8788` |
| Env 640 root:uslic-test | **PASS** | prod `api.kanus.org` env’ye yazılmadı |
| nginx -t + reload | **PASS** | `api.kanus.org` site kaldı |
| Local GET :8788/healthz | **PASS** | 200 `ok` |
| İmzasız POST orders-paid | **PASS** | 401 `unauthorized` |
| Signed HMAC (orders-cancelled, sentetik id) | **PASS** | **400** `topic` (HMAC kabul; 401 değil). Store sha değişmedi |
| Accounts login ulaşımı | **PASS** | 401 `account request rejected` (hook’a ulaştı; 404/502 değil) |
| Store restore after dry-run | **PASS** | pristine sha eşleşti |
| Unit stop + disable | **PASS** | inactive / disabled; 8788 kapalı |
| Test VPS hook | **PASS** | hâlâ **active**, healthz `ok`, licenses sha aynı |
| DNS / TTL | **PASS (yapılmadı)** | bilerek dokunulmadı |
| Box temp secret silme | **PASS** | env/hmac/licenses xfer silindi |

### TLS kesim notu

Cert, DNS `api-test` burayı göstermeden **webroot ile alınamaz**. Seçenekler:

1. **DNS-01 (önerilen):** Kesimden önce TXT ile prod’da cert al → HTTPS boşluğu yok.
2. **LE dosya kopyası test→prod:** Private key + fullchain kopyalanır; boşluk kısa olur; yenileme (`renewal` conf / webroot map) prod’da yeniden kurulmalı. Operasyonel borç.
3. **Kesim sonrası webroot:** DNS yayılımı + certbot süresince HTTPS riski (~2–10 dk).

**Tavsiye:** DNS-01. Kopya yalnızca DNS-01 mümkün değilse yedek (kısa pencere + hemen renewal conf düzelt).

### Kesinti tahmini (güncel)

TTL hâlâ **14400** (bu fazda değiştirilmedi). Kesim öncesi **300’e düşürüp ≥1 saat** beklenmeli.

| Varsayım | Müşteriye görünen |
|---|---|
| TTL=300 + DNS-01 hazır + freeze 2–5 dk | **~5–10 dk** |
| TTL hâlâ 14400 ile kesim | **kabul edilmez** (saatlerce splitter) |

Shopify retry’leri kısa 5xx penceresini telafi eder.

### Sonraki kapı

1. Demir: plan + bu PASS tablosunu görüp **DNS kesimi + eski stop** onayı
2. TTL 14400→300 (≥1 saat)
3. DNS-01 cert (tercih)
4. `CUTOVER_EXECUTE=1 ./deploy/hook-move/cutover.sh` (veya adım adım manuel)

Ham log: `reports/qa-20261106/hook_move_phase2_results.txt`


## §15 Cutover sonucu (2026-11-08, TSİ)
Onay: Demir, KAN.US aracılığıyla birebir: "Evet, lisans servisini şimdi 82.29.184.225'e taşı" (00:04).
- 00:11:40 T0: eski hook (srv1662550) durduruldu, licenses.json sha256 c2a15022…49be (30328 B) alındı.
- Prod'a kopyalandı, sha birebir aynı; eski prod kopyası /opt/us-license-hook-test/licenses.json.bak-precutover-20261108.
- api-test LE sertifikası (bitiş 2027-02-04) prod'a taşındı; nginx 80→443 yönlendirme + 443 ssl; eski conf /root/api-test.kanus.org.nginx.bak-precutover-20261108. certbot renew --dry-run başarılı.
- us-license-hook-test enable --now, 127.0.0.1:8788 healthz 200.
- ~00:12:40 DNS api-test A 31.220.63.191 → 82.29.184.225 (TTL 300), sadece bu kayıt değişti (14/14). Zone yedekleri: host-hardening/dns/kanus.org.zone.backup-*-before/after-apitest-cutover.json.
- 00:13:22 1.1.1.1 ve 8.8.8.8 yeni IP'yi döndü; ilk public 200. Kesinti ≈ 1 dk 42 sn.
- Smoke: TLS geçerli (verify=0), healthz 200, http→https 301, imzasız webhook 401, imzalı test webhook 200 (plan dışı ürün, lisans basılmadı; kayıt id 990000000001 / webhook id cutover-smoke-20261108 dosyada duruyor), qa-license-a/b credits/status 200, sahte lisans 401, /account/login 401 (hook'a ulaşıyor).
- Eski hook: durdu + disable, silinmedi; /opt/us-license-hook/licenses.json yerinde. 24 saat tutulacak.
- Box yedeği: /workspace/us-panel/deploy/hook-move/backup/licenses.json.cutover-20261108T001140+0300 (600).
- api.kanus.org prod hook'una dokunulmadı (inactive).
- Not: us-license-hook-test Restart=no; çökme olursa kendiliğinden kalkmaz (ayrı onayla Restart=on-failure önerilir).
Geri dönüş: DNS api-test → 31.220.63.191, test VPS'te systemctl start us-license-hook, prod'da us-license-hook-test stop.
