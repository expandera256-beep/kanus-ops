# HOST → QA salt-okuma kanıtları (2026-11-08 00:46 TSİ)
Kaynak: root SSH srv2133566 (82.29.184.225), salt okuma. Değişiklik yapılmadı.

## a) Yeni host kanıtı
- hostname: srv2133566; us-license-hook-test MainPID 21521, /proc/21521/exe → /opt/us-license-hook-test/us-license-hook.real, sha256 7d72f449ae3011e7f37632b0aced7ba827113940846213a3dd94415a921256f0.
- Bu binary'de version ucu yok (/version, /v1/version 404); healthz gövdesi "ok". Sürüm başlığı eklemek nginx değişikliği → onay gerektirir.
- Box çıkış IP'leri 140.248.52.x (DNS 198.18.0.1 sahte-IP proxy'si; trafik gerçekte 82.29.184.225'e gidiyor). Prod nginx access.log'da 21:45:14–21:45:25 UTC (00:45 TSİ) box kaynaklı istekler, örnek:
  140.248.52.61 [07/Nov/2026:21:45:14 +0000] "POST /account/login" 200
  140.248.52.120 [07/Nov/2026:21:45:14 +0000] "GET /v1/credits/status" 200
  140.248.52.57 [07/Nov/2026:21:45:19 +0000] "POST /account/login" 429
  140.248.52.69 [07/Nov/2026:21:45:18 +0000] "GET /healthz" 200
  (Probe'unun zaman damgasını/yolunu verirsen birebir satırı çıkarırım.)
- Dış bakış: srv1662550 (31.220.63.191) → api-test 82.29.184.225, healthz 200, TLS verify=0 (00:14 ve 00:16 TSİ, access.log'da 31.220.63.191 satırları).

## b) Yedekleme / restore
- Kurulu: /usr/local/sbin/backup-us-license-hook-hourly.sh sha256 a47c1142…12b2 = deploy/host-hardening/backup/ DRAFT ile birebir aynı (FINDING-03 prod'da geçerli). cron.d/us-license-hook-hourly: her saat :05, root, log /var/log/us-license-hook-backup-hourly.log.
- Son başarılı koşum 2026-11-07 21:05:01 UTC (00:05 TSİ); /var/backups/us-license-hook/ 123 dosya, 7 gün retention.
- **YENİ BULGU (HIGH, HOST):** script yalnız /opt/us-license-hook (durdurulmuş prod api.kanus.org) store'unu yedekliyor. Canlı api-test store'u /opt/us-license-hook-test/licenses.json **saatlik yedeğe dahil değil**. Mevcut api-test yedekleri yalnız elle: box hook-move/backup/licenses.json.cutover-20261108T001140+0300 (sha c2a15022…), prod licenses.json.bak-precutover-20261108 ve .pristine.
- Restore: planlı restore denemesi kaydı yok. Tek fiili restore benzeri işlem 2026-11-08 00:12 TSİ taşıma: store srv1662550'den kopyalandı, sha256 birebir doğrulandı, servis bu dosyayla açıldı, mevcut lisanslar 200.
- Düzeltme v2 KURULDU 2026-11-08 00:48 TSİ (Demir onayı 00:47; sha d77534e5…a0): cron :05 prod, :07 test; ilk koşum ikisi de OK, test yedeği sha = canlı store a0daf1ea…; v1 script/cron yedekleri /root/*.v1-20261107T2148*. Önceki taslak notu: deploy/host-hardening/backup/backup-us-license-hook-hourly.v2.sh — store/plans eksikse exit 2, bozuk JSON exit 3, manifest sha256 doğrulaması, LAST_SUCCESS dosyası, `prod|test` parametresiyle iki instance. Yerel testler: eksik store 2, eksik plans 2, bozuk JSON 3, normal 0.

## c) Journal kesiti (journalctl -u us-license-hook-test -n 300, maskelendi)
Tüm çıktı 7 satır; istek başına log yok:
2026-11-06T07:06:49Z systemd: Started us-license-hook-test.service
2026-11-06T07:06:50Z us-license-hook-wrap.sh[11289]: dinleniyor 127.0.0.1:8788
2026-11-06T07:07:26Z systemd: Stopping / Deactivated successfully / Stopped
2026-11-07T21:12:20Z systemd: Started us-license-hook-test.service
2026-11-07T21:12:20Z us-license-hook-wrap.sh[21521]: dinleniyor 127.0.0.1:8788
Token/lisans/parola/e-posta görünmüyor. Not: nginx access.log query string'leri loglar (ör. /v1/credits/status?license=… kullanılırsa anahtar loga düşer) — QA testlerinde lisansı query'de göndermeyin.

## FINDING-01 (host tarafı)
Sunucuda hook ayrı sistem kullanıcılarıyla çalışıyor (uslic, uslic-test). Panel'in güvenilmeyen kodu per-user UID/container ile koşması PANEL DEV tasarımına bağlı; sunucuda henüz böyle bir yürütme yok. Host tarafı politika (per-user UID, systemd sandboxing, nspawn/container) PANEL DEV tasarımı netleşince önerilecek; şu an değişiklik yok.


## QA kapanış kanıtı (01:0x TSİ)

Salt okunur toplama: 08.11.2026 01:03:26 TSİ. Sunucuda değişiklik/kurulum/restart yapılmadı.

Kurulu script box v2 ile birebir SHA256 eşleşti. LAST_SUCCESS zamanları aşağıda TSİ’ye çevrildi; dosya adlarındaki Z kimlikleri değiştirilmedi. Manifestler yalnız checksum/dosya listesi; store/env içeriği okunmadı.

```text
SCRIPT_SHA256 d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0
CRON
# us-license-hook hourly backups (v2, 2026-11-08). v1 backup: /root/us-license-hook-hourly.cron.v1-20261107T214805Z
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
5 * * * * root /usr/local/sbin/backup-us-license-hook-hourly.sh prod >>/var/log/us-license-hook-backup-hourly.log 2>&1
7 * * * * root /usr/local/sbin/backup-us-license-hook-hourly.sh test >>/var/log/us-license-hook-test-backup-hourly.log 2>&1

BACKUP /var/backups/us-license-hook LAST_SUCCESS 2026-11-08 00:48:05 TSİ (dosyadaki UTC değerinin dönüşümü)

MANIFEST /var/backups/us-license-hook/manifest.20261107T214805Z.sha256 mtime_TSI 2026-11-08T00:48:05.939549+03:00 bytes 294 sha256 6f6f306cf2f289964ce1b579ea3b94694100c55c34b7edb5a59ee2812e711f30
MANIFEST_METADATA_AND_FILE_LIST
ca3d163bab055381827226140568f3bef7eaac187cebd76878e0b63e9e442356  licenses.json.20261107T214805Z
33148639a2b965a7e94fc5c5b4c65e76a443fc5ac6686076f2f23301b1ea59e8  plans.json.20261107T214805Z
6939a89c603ef062ee81697ee8f89f5c92d015a5bf006f283fb3a5a1006fb4d1  us-license-hook.env.20261107T214805Z

LOG_LAST_15
backed /opt/us-license-hook/plans.json -> /var/backups/us-license-hook/plans.json.20261107T180501Z
backed /etc/us-license-hook.env -> /var/backups/us-license-hook/us-license-hook.env.20261107T180501Z
backed /opt/us-license-hook/licenses.json -> /var/backups/us-license-hook/licenses.json.20261107T190501Z
backed /opt/us-license-hook/plans.json -> /var/backups/us-license-hook/plans.json.20261107T190501Z
backed /etc/us-license-hook.env -> /var/backups/us-license-hook/us-license-hook.env.20261107T190501Z
backed /opt/us-license-hook/licenses.json -> /var/backups/us-license-hook/licenses.json.20261107T200501Z
backed /opt/us-license-hook/plans.json -> /var/backups/us-license-hook/plans.json.20261107T200501Z
backed /etc/us-license-hook.env -> /var/backups/us-license-hook/us-license-hook.env.20261107T200501Z
backed /opt/us-license-hook/licenses.json -> /var/backups/us-license-hook/licenses.json.20261107T210501Z
backed /opt/us-license-hook/plans.json -> /var/backups/us-license-hook/plans.json.20261107T210501Z
backed /etc/us-license-hook.env -> /var/backups/us-license-hook/us-license-hook.env.20261107T210501Z
backed /opt/us-license-hook/licenses.json -> /var/backups/us-license-hook/licenses.json.20261107T214805Z
backed /opt/us-license-hook/plans.json -> /var/backups/us-license-hook/plans.json.20261107T214805Z
backed /etc/us-license-hook.env -> /var/backups/us-license-hook/us-license-hook.env.20261107T214805Z
backup[prod]: OK 3 files
BACKUP /var/backups/us-license-hook-test LAST_SUCCESS 2026-11-08 00:48:06 TSİ (dosyadaki UTC değerinin dönüşümü)

MANIFEST /var/backups/us-license-hook-test/manifest.20261107T214805Z.sha256 mtime_TSI 2026-11-08T00:48:06.081548+03:00 bytes 299 sha256 5ed6ef3e2236c814a49e91503a4c47503424d30287b01939055d1f93fa785cd0
MANIFEST_METADATA_AND_FILE_LIST
a0daf1ea35049551ff4b1584b4a604bf477fdda93dd4d471559b873f9b94ab9d  licenses.json.20261107T214805Z
33148639a2b965a7e94fc5c5b4c65e76a443fc5ac6686076f2f23301b1ea59e8  plans.json.20261107T214805Z
60f32b00f7062907a46e8ada6a9e5348737b105aa5dc2c66ccffd9a5a3606252  us-license-hook-test.env.20261107T214805Z

LOG_LAST_15
backed /opt/us-license-hook-test/licenses.json -> /var/backups/us-license-hook-test/licenses.json.20261107T214805Z
backed /opt/us-license-hook-test/plans.json -> /var/backups/us-license-hook-test/plans.json.20261107T214805Z
backed /etc/us-license-hook-test.env -> /var/backups/us-license-hook-test/us-license-hook-test.env.20261107T214805Z
backup[test]: OK 3 files
```

### Nginx korelasyonu ve query taraması
QA listesi: QA_QUERY_LOG_REVIEW_20261108.md:30–44. Sekiz probe için 8 eşleşme bulundu; command değeri de ihtiyaten maskelendi, credential değildir. key/token/license/pin/secret parametreli veya us1/usa1 biçimli query **0 satır**, anahtar görünmedi. Parametre adı sayımı boş. Tüm mevcut access.log ve access.log.1 tarandı; başka rotated/gz dosyası bulunmadı. Log ortak olduğundan Host alanı yok; saat/path/status ile korelasyon yapılır, vhost bağımsız kanıt değildir. Eski VPS logları bu sunucuda yok; 06.11 08:00–08:24 TSİ için prod logunda eşleşme olmaması eski VPS geçmişini kapatmaz. Restore drill halen onay/bakım penceresi bekler.

```text
PIN_FORMAT /etc/us-model/wake.pin hash-like bytes 53
SCANNED_LOG_FILES ["/var/log/nginx/access.log", "/var/log/nginx/access.log.1"]
SENSITIVE_QUERY_LINES 0 PARAM_COUNTS {}
SENSITIVE_QUERY_MATCHES

QA_PROBE_MATCHES
/var/log/nginx/access.log: 140.248.52.102 - - [08/Nov/2026:00:30:16 +0300 TSİ] "GET /healthz HTTP/1.1" 200 2 "-" "Python-urllib/3.13"
/var/log/nginx/access.log: 140.248.52.102 - - [08/Nov/2026:00:30:16 +0300 TSİ] "GET /v1/credits/status HTTP/1.1" 401 8 "-" "Python-urllib/3.13"
/var/log/nginx/access.log: 140.248.52.32 - - [08/Nov/2026:00:30:17 +0300 TSİ] "GET /v1/models HTTP/1.1" 401 8 "-" "Python-urllib/3.13"
/var/log/nginx/access.log: 140.248.52.32 - - [08/Nov/2026:00:30:18 +0300 TSİ] "POST /account/balance HTTP/1.1" 401 37 "-" "Python-urllib/3.13"
/var/log/nginx/access.log: 140.248.52.121 - - [08/Nov/2026:00:30:18 +0300 TSİ] "POST /webhooks/shopify/orders-paid HTTP/1.1" 401 13 "-" "Python-urllib/3.13"
/var/log/nginx/access.log: 140.248.52.72 - - [08/Nov/2026:00:30:18 +0300 TSİ] "POST /webhooks/shopify/orders-paid HTTP/1.1" 401 13 "-" "Python-urllib/3.13"
/var/log/nginx/access.log: 140.248.52.46 - - [08/Nov/2026:00:30:19 +0300 TSİ] "GET /v1/%2e%2e/%2e%2e/qa-synthetic HTTP/1.1" 400 166 "-" "-"
/var/log/nginx/access.log: 140.248.52.69 - - [08/Nov/2026:00:30:19 +0300 TSİ] "GET /v1/credits/status?command=[MASKELİ] HTTP/1.1" 401 8 "-" "Python-urllib/3.13"
OLD_WINDOW_MATCHES

```

/etc/us-model-test/wake.pin ve /etc altındaki başka wake.pin kopyası bulunmadı. Prod pin yalnız başlık eşitliği ve byte sayısıyla incelendi; içerik okunup raporlanmadı.
