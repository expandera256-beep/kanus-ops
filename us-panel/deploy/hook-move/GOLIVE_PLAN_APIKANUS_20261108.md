# api.kanus.org canlı mağaza geçiş planı — 08.11.2026

**PLAN / UYGULANMADI.** Kontrol yaklaşık 01:20 TSİ. Yalnız salt okunur SSH ve genel DNS sorguları yapıldı. us-license-hook (:8787) başlatılmadı/değiştirilmedi; kurulum/restart/nginx reload/DNS değişikliği/Shopify işlemi/mail gönderimi yok. Aşağıdaki işlemler gelecek onay kapılarıdır, yürütme yetkisi değildir.

## Kaynak ve güncel durum
İncelenenler: `reports/HOOK_MOVE_PROD_VPS_20261106.md` (özellikle §15); `deploy/hook-move/` R11/R12/R13 planları, cutover/rollback referansları ve nginx snippetleri; `deploy/host-hardening/PROD_PLAN.md`, backup v2 scripti ve cron, nginx referansları; `reports/SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md`; PANEL DEV `.12` incelemesi. PROD_PLAN'ın ilk kısmındaki ters portlar/eski test IP'si ve ağır eski VPS yükü tarihsel bilgi; aşağıdaki güncel ölçümler esas alınır.

- VPS: srv2133566 / 82.29.184.225. Prod servis kurulu, enabled fakat **inactive/dead**, Restart=no; :8787 dinleyici yok. Test active/running, 127.0.0.1:8788.
- Prod binary `/opt/us-license-hook/us-license-hook.real`: 7,454,880 byte; SHA256 `7d72f449ae3011e7f37632b0aced7ba827113940846213a3dd94415a921256f0`. Test binary hash'i aynı. Exact release/buildVersion kesinleştirilmedi; binary çalıştırılmadı. Eski kaynak manifesti hâlâ BLOCKED; bu SHA .11/.12/.13 değildir.
- Prod dir: wrapper 512 B, plans.json 571 B, licenses.json 3 B (geçerli boş `{}`), binary. JSON içerikleri dışarı çıkarılmadı.
- Prod env: Shopify webhook ve license secret EMPTY, shop domain PLACEHOLDER, SMTP_ENABLE=0. Test secrets SET. Ayrı env dosyaları root:uslic 640 ve root:uslic-test 640. Secret eşitliği karşılaştırılmadı; ayrı yol tek başına farklı secret kanıtı değildir.

| Sınır | Prod | api-test |
|---|---|---|
| Unit/user | us-license-hook / uslic | us-license-hook-test / uslic-test |
| Loopback port | 8787 (kapalı) | 8788 (açık) |
| Store/plans | /opt/us-license-hook/ | /opt/us-license-hook-test/ |
| Env | /etc/us-license-hook.env | /etc/us-license-hook-test.env |
| Backup | /var/backups/us-license-hook | /var/backups/us-license-hook-test |

Prod dir root:uslic 755; test dir uslic-test:uslic-test 755. Bu yüzden dosya bazında store/plans izinleri ve servis yazabilirliği uygulama öncesi denetlenmeli; “tam izolasyon” iddiası yok. Farklı UID/store/env olsa da kernel, CPU, RAM, nginx ve disk ortaktır. Yeni kullanıcı/izin değişikliği bu planda uygulanmadı; gerekiyorsa ayrıca onay.

- Nginx api.kanus.org site hazır: HTTP→HTTPS, ACME webroot, HSTS/nosniff/frame/CSP/referrer başlıkları; prod-only snippet →127.0.0.1:8787; /account/ halen 404. Hesap/login akışı isteyen ürün için bu ek bir **bloker**; yalnız prod özel config değişikliği, rate-limit/Retry-After ve test gerekir. api-test'e dokunulmaz.
- LE cert SAN api.kanus.org; başlangıç 06.11.2026 08:13:21 TSİ, bitiş **04.02.2027 08:13:20 TSİ**. Cert dosyası metadata'sı doğrulandı; bu tur handshake/renew dry-run yapılmadı.
- Anlık kaynaklar: 2 CPU, 7935 MiB RAM / 7423 MiB available / 511 MiB used, swap 0; load 0.00/0.00/0.00; root 96G, 1.8G used, 95G available (%2). Test servis MemoryCurrent yaklaşık 4 MiB. Tek anlık ölçüm CPU steal/uzun süre kapasite SLA kanıtı değildir; kısa baseline önkoşul.

## Seçim: A
Mevcut 82.29.184.225 üzerinde ayrı uslic/:8787 prod instance kullanılacak. B (ayrı VPS) bırakıldı. Yeni IP, satın alma ve DNS taşıma adımı yok. Kaynaklar uygun anlık görünüyor; paylaşılan CPU/nginx/host riski izlenecek. PASS sonrası tahmin: 45–90 dk hazırlık ve 10–20 dk kontrollü aktivasyon; kabul bekleme süresi hariç, taahhüt değil.

## Sıralı uygulama kapıları (şimdi uygulanmayacak)
### 1. DNS
`dig @1.1.1.1 +noall +answer` ile api A **82.29.184.225 TTL300**, AAAA ve CNAME yanıtı yok. A seçeneğinde DNS değişikliği gerekmez. B seçilirse tam zone yedeği, yalnız api A değişimi ve propagation kontrolü ayrı onayla; diğer kayıtlar korunur. Recursive TTL cache yaşına bağlı olabilir; yetkili zone final kesim öncesi yeniden kontrol edilir.
Genel MX: mx1.hostinger.com prio5 ve mx2.hostinger.com prio10 TTL14400. api-test veya shop kayıtlarını değiştirme.

### 2. Canlı secrets, plans ve ledger
- **Demir/SHOP** canlı app'in gerçek webhook signing secret'ını ve doğrulanmış myshopify shop domain'ini sağlar; taslak domain `19einb-qz.myshopify.com` SHOP tarafından yeniden doğrulanır. Test app secret/domain/store asla prod'a kopyalanmaz. Secret yalnız güvenli kanal/env üzerinden; raporda yok.
- SHOP canlı product_id→plan/SKU/kredi/expiry/refund politikasını doğrular. `deploy/host-hardening/plans.live.json` ve mevcut plans taslaktır; 8 kayıt JSON-valid olması canlı ticari doğruluk değildir. DRAFT CREDIT-1 veya test ID'si otomatik aktive edilmez.
- Demir canlı US_LICENSE_SECRET ve US_SESSION_SECRET üretim/saklama/rotasyon politikasını onaylar; mevcut müşteri/lisans varsa anahtar değişimi geçersizleştirebilir. Sıfırdan üretme kararı sessizce alınmaz.
- Prod store şu an boş. “Canlı mağazayı taşımak” Shopify ürünlerini veya api-test ledger'ını otomatik taşımak değildir. SHOP mevcut gerçek ücretli sipariş/lisansların nerede olduğunu, önceki webhook aboneliklerini ve outstanding refund/fulfillment olaylarını envanterler. Backfill gerekiyorsa sipariş listesi, idempotency ve mali etkisi için ayrı onay; paid replay kör yapılmaz.
- Gerçek ledger üzerinde forward/backward compatibility, exact eski binary rollback sınırı ve tutarlı snapshot drill onaylı scratch'ta kanıtlanmadan açılış yok. Round7 sentetik testleri gerçek 7d72 uyumu değildir.

### 3. SMTP üretim teslimi
Kaynak dar okuma: `work2/kiyas/term/cmd/us-license-hook/main.go:56–61`: **US_SMTP_ENABLE=1** sender/flush'ı etkinleştirir. `licensesrv/mail.go:136–152`: **US_SMTP_ALLOWLIST hiç tanımlı değilse sınırsız**, tanımlı boş değer ise **herkesi engeller**. Dolayısıyla `US_SMTP_ALLOWLIST=` üretime açmaz! Ayrı bir “production mode” flag'i bu okunan yolda yok. Test allowlist'i korunur. Exact nihai PASS binary ile bu semantik ayrıca doğrulanır; çalışan eski binary'ye kaynak eşitliği varsayılmaz.
Prod için onayla ayarlanacaklar: US_SMTP_ENABLE=1, US_SMTP_HOST=smtp.hostinger.com, US_SMTP_PORT=465, US_SMTP_USER/US_SMTP_PASS (Demir'in yetkili mailbox bilgileri), US_SMTP_FROM (yetkili support mailbox), prod env/systemd inherited kaynaklarından US_SMTP_ALLOWLIST tamamen kaldırılması. İlk kontrollü teslimde dar allowlist, doğrulama sonrası ayrıca genel açma önerilir. Kuyruktaki skipped/dead/sent satırların yeniden gönderim davranışı ve çift mail riski kontrol edilmeden toplu flush yok. Mesaj/metin/gerçek alıcı Demir onayına tabi.
Genel DNS dig sonucu:
- SPF TTL3600: `v=spf1 include:_spf.mail.hostinger.com ~all` (softfail).
- DMARC TTL3600: `v=DMARC1; p=none; rua=mailto:s***@kanus.org; fo=1` (izleme, enforcement yok).
- hostingermail-a/b/c._domainkey CNAME TTL300 → karşılık gelen hostingermail-a/b/c.dkim.mail.hostinger.com.
SPF/DKIM/DMARC kayıtlarının varlığı gerçek alignment/deliverability PASS değildir. DKIM hedef public-key TXT ve yetkili test alıcıdaki Authentication-Results (spf/dkim/dmarc), From alignment, inbox/spam kontrolü gerekir. DNS/mail politika değişimi ayrı onay.

### 4. Backup
**v2 kapsamı zaten genişletilmiş; yeniden cron kurmak gerekmez.** Kurulu SHA `d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0`.
`5 * * * * root /usr/local/sbin/backup-us-license-hook-hourly.sh prod >>/var/log/us-license-hook-backup-hourly.log 2>&1`
Prod dalı BASE=/opt/us-license-hook, ENVF=/etc/us-license-hook.env; licenses.json+plans.json+env, manifest SHA ve JSON parse; /var/backups/us-license-hook. Test :07 aynen kalır. Önceki ölçüm LAST_SUCCESS prod 08.11 00:48:05 TSİ, test 00:48:06 TSİ; bu tur LAST_SUCCESS yenilenmedi.

### 5. Shopify abonelikleri — yalnız SHOP/Demir
Canlı app/store kimliği ve mevcut abonelik envanteri doğrulandıktan, host sağlıklı olduktan sonra **ayrı açık onayla** JSON abonelikler:
- orders/paid → https://api.kanus.org/webhooks/shopify/orders-paid
- orders/cancelled → https://api.kanus.org/webhooks/shopify/orders-cancelled
- refunds/create → https://api.kanus.org/webhooks/shopify/refunds-create
Aynı topic için çift abonelik/yanlış app secret riski kapatılır; önceki config ve subscription ID'leri rollback için kaydedilir. Test shop ve api-test aboneliklerine dokunulmaz. Oluşturma/silme/değiştirme bu görevde yapılmadı.

### 6. Aktivasyon ve smoke
Demir'in ayrı prod start/deploy onayı sonrası: yedek -> PASS binary hash doğrulama -> canlı env/plans/PEM -> yalnız prod özel nginx account/proxy gerekliyse diff -> nginx -t -> onaylı reload -> yalnız prod servis start. Restart= ayarı eklenmez, VM restart yok.
Önce loopback ve public normal TLS healthz200; imzasız/yanlış HMAC401 ve yanlış-shop red (yazmayan istek). Gerçek HMAC/sipariş testi ayrı onayla: bir kontrollü sipariş paid→tek lisans/kredi, aynı delivery idempotent, cancel/refund doğru revoke/azaltma, store/outbox ve tek mail teslimi. Satın alma/ödeme/iade maliyeti ve alıcısı ayrıca onaylanır. GPU'suz smoke GPU readiness/satılabilir model kanıtı değildir: ücretli gerçek model/wake/start/status ve 600sn idle kapanışı KAN/PANEL DEV ile ayrı onaylı kabul penceresinde, bütçe sınırıyla yapılmalı. PASS yoksa müşteri satış trafiği açılmaz.
Masaüstü gömülü api-test URL'si veya /account 404 kalması müşteri akışını kırabilir; canlı client/gateway URL, auth ve model route uçtan uca doğrulanmadan “go-live” denmez.

### 7. İzleme
İlk 30–60 dk operatör gözetimi, ardından 24 saat: health/5xx/HMAC red/429+Retry-After; webhook delivery ID/order/refund eşleştirme ve idempotency; kredi/rezerv/settlement tutarlılığı; outbox skipped/dead/teslim; backup LAST_SUCCESS/manifest; RAM/CPU-steal/load/disk/OOM; TLS expiry. Ham auth/query/PII loglanmaz. Sürekli alarm/rutin kurma ayrıca onay; bu belgede kurulmadı. OOM, artan5xx, yanlış kredi/çift mail/yanlış shop veya GPU bütçe ihlalinde yeni satış/cutover durdurma kapısı.

### 8. Rollback
SHOP eski abonelik URL/ID durumunu geri getirir veya onaylı duraklatır; retry/backlog kaybolmamalı. Gerekirse onayla yalnız prod servis durur; test etkilenmez. Binary/env/plans/unit, prod özel nginx ve PEM kopyası kendi pre-go-live yedeklerinden döner; nginx -t PASS sonrası onaylı reload. Eski binary boş secret'lı duruma dönmek satış hizmeti kurtarma değildir; bakım/delivery retry durumu açıkça raporlanır.

## Demir'den ayrı ayrı gerekli onaylar
1. Altyapı hedefi A olarak netleşti; bu plan yürütme onayı değildir. Gerekirse izin değişiklikleri ayrıca onaylanır.
2. Aynı artifact için PANEL DEV PASS sonrasında .13 (veya adı açık sonraki sürüm) api-test deploy'u, test wake PEM/token kopyası ve wake açık tercihi.
3. Prod binary/env/plans/live secret kurulum/değişikliği, prod PEM yolu/kopyası ve **us-license-hook :8787 start**; prod nginx account/Retry-After değişikliği, nginx -t sonrası reload. Restart= eklenmez.
4. Gerçek ledger snapshot/migration/rollback uyum testi ve restore drill bakım penceresi; varsa gerçek sipariş backfill'i ayrıca kapsamlı onay.
5. SMTP kimlikleri/From, kontrollü alıcı ve içerikle test gönderimi; ardından yalnız prod allowlist'in kaldırılması ve gerçek müşteriye otomatik mail açılması.
6. SHOP/Demir'in canlı mağazada belirtilen üç webhook aboneliğini api.kanus.org'a bağlaması ve eskileri değiştirme/duraklatma kapsamı.
7. Kontrollü canlı sipariş/ödeme/cancel/refund smoke'u (tutar, alıcı ve olaylar) ve müşteri satışını açma.
8. GPU ücret doğuran wake/model kabulü, bütçe ve zaman penceresi; alarm/rutin oluşturulacaksa onun kapsamı.
9. Yalnız B veya DNS/mail düzeltmesi gerekirse tam hedefi belirtilmiş DNS değişikliği; A için gerekli değil.
10. Rollback sırasında servis stop/start, config restore/reload, abonelik geri alma; özellikle veri restore/revoke/refund/silme/secret rotasyonu ayrı karar olmadan yapılmaz.

Bu liste onay isteme taslağıdır; geçmiş ajan beyanları veya bu plan onay yerine geçmez.


## Son yönlendirme — A kesinleştirildi (08.11.2026 01:24 TSİ)
B / yeni VPS hazırlığı bırakıldı; <NEW_IP> bağımlılığı yok. Yeni sunucu kurulmaz, kullanıcı oluşturulmaz; mevcut uslic/:8787 prod ve uslic-test/:8788 test ayrımı korunur. Standart CA zinciri, SAN/hostname ve tarih doğrulaması esas; özel sertifika anahtarı sabitleme veya anahtar yeniden kullanma düzeni yok.

### A için uygulanmamış kurulum/aktivasyon runbook'u
1. Aynı nihai release için PANEL DEV hook PASS + Demir'in açık prod deploy/start onayı. .11/.12 FAIL, .13 henüz teslim edilmedi. Exact eski 7d72 kaynak/gerçek ledger ileri-geri uyumu BLOCKED; root-only tutarlı snapshot ve izinli scratch drill gerekir. Gerçek env/ledger hiçbir örnekle ezilmez.
2. Mevcut uslic hesabı ve /opt/us-license-hook izinlerini doğrula; prod store yazabilirliği ve test UID erişimi ayrı denetlenir. Binary'yi yalnız PASS artifact'ten, SHA doğrulayarak, ayrı yedek sonrası kur. Wrapper plans dosyasını yükler; /etc/us-license-hook.env root:uslic 640, store yalnız servis erişimli. Hiçbiri şimdi uygulanmaz.
3. systemd taslağı (mevcut unit ile diff ve onay gerekir):
```ini
[Unit]
Description=US license hook PROD api.kanus.org
After=network-online.target
[Service]
User=uslic
Group=uslic
WorkingDirectory=/opt/us-license-hook
EnvironmentFile=/etc/us-license-hook.env
ExecStart=/opt/us-license-hook/us-license-hook
[Install]
WantedBy=multi-user.target
```
Restart direktifi eklenmez; mevcut Restart=no korunur, varsayılan da no'dur. Enable/start veya daemon-reload bu tur yok.
4. OS/firewall/fail2ban: mevcut hardening'i salt okunur denetle; yeni uygulama olursa ayrıca onay. Referans security/TLS_AND_FIREWALL.md ve PROD_PLAN: güvenlik güncellemeleri, key-only SSH, journald SystemMaxUse500M; UFW deny incoming/allow outgoing, yalnız SSH22+HTTP80+HTTPS443 gerekli; 8787/8788 loopback kalır, 8090–8095 açılmaz. Fail2ban sshd/nginx-http-auth/nginx-limit-req; yönetici allowlist ve Shopify retry/ban etkisi incelenir. SSH erişimi ikinci oturumdan kanıtlanmadan firewall/SSH değişmez. VM restart yok; round-2 otomatik uygulanmaz.
5. SMTP env/prod mod, canlı secrets/plans, backup v2 :05 ve SHOP abonelikleri yukarıdaki sırayla; kapsam zaten /opt/us-license-hook. Backup eski arşivleri elle değiştirmez; root-only bin/env/unit/site yedeği ayrıca gerekir.
6. Nginx önerisi: `deploy/hook-move/nginx/api.kanus.org-account.proposal.diff`. Salt okunur mevcut prod site ve api-test'in login/register exact + /account/ fallback proxy düzeni incelendi. Hook main.go:104–108 /account/ handler'ını bağlar; accounts.go:327 route listesinde register/login/balance/redeem/requests/workers/verify-email/verification-request/password-reset-request/confirm var. Nihai release route listesi yeniden doğrulanır.
   Diff yalnız api.kanus.org server block'unu değiştirir; prod include içeriğini inline ederek eski /account/404'ü kaldırır, başka site/snippet'e yazmaz. Auth uçlarında mevcut uslic_account_auth zone ve 429/Retry-After900 önerisi, kalan /account/ yolunda doğru Host/Authorization/X-Real-IP/X-Forwarded-* ve :8787 proxy. 900 konservatif beklemedir; upstream429 body/başlığını normalleştirdiği için PANEL DEV/PERF incelemesi şart. Güvenlik başlıkları named location'da korunur. Backup -> nginx -t -> yalnız onayla reload; FAIL'de yedeğe dön, reload yok.
7. Mevcut LE cert/standart CA doğrulaması kullanılır. Certbot timer kontrolü aşağıda; yeni sertifika veya key üretimi yok. A DNS zaten doğru, TTL300; “TTL düşür” işlemi gereksiz. DNS değişikliği yalnız hedef gerçekten değişirse yeni onayla.
8. Yalnız onaylı prod start, GPU'suz smoke; ücretli GPU/model kabulü ve SMTP/sipariş/iade deneyi ayrı izinle. Başarılı test healthz tek başına satılabilir model kanıtı değil. İzleme ve rollback önceki veri-korumalı sırayla; test servisine dokunma.

### Salt okunur PEM SAN / zincir kanıtı
VPS'te her domain için `openssl s_client -connect 127.0.0.1:443 -servername <domain> -showcerts`; leaf ve untrusted zincir RAM/process substitution üzerinden `openssl verify -verify_hostname <domain> -CAfile /etc/ssl/certs/ca-certificates.crt -untrusted <chain> <leaf>` ile kontrol edildi; sunucuya geçici dosya yazılmadı.

| Domain | SAN | Leaf notAfter (TSİ) | Sunulan zincir | verify |
|---|---|---|---|---|
| api.kanus.org | DNS:api.kanus.org | 04.02.2027 08:13:20 | 4 cert: leaf + YE1 + Root YE + cross-signed ISRG Root X2 | exit0 / OK |
| api-test.kanus.org | DNS:api-test.kanus.org | 04.02.2027 03:48:11 | aynı 4 cert düzeni, leaf+YE1 dahil tam doğrulanabilir zincir | exit0 / OK |

Certbot.timer enabled/active/waiting; son tetik 08.11.2026 **01:15:03 TSİ**, sonraki 08.11.2026 **13:25:27 TSİ**. Timer active olmak son renewal başarı kanıtı değildir.
**renew --dry-run ÇALIŞTIRILMADI:** Certbot dry-run salt okunur değildir; log/lock/ACME hesap state yazabilir ve renewal hook çalıştırabilir. “Sunucuda hiçbir değişiklik yok” kuralıyla çeliştiğinden duruldu. Ayrı dar kapsamlı onay + hook/config incelemesi sonrası staging renewal testi planlanabilir; PASS iddia edilmez.

Wake `/etc/us-model/wake.crt`: subject=issuer CN34.41.186.170 (self-signed); SAN IP34.41.186.170; geçerlilik 06.11.2026 **09:46:48 TSİ** – 05.11.2028 **09:46:48 TSİ**. Sistem CA + verify_ip ile **FAIL (exit2, error18 self-signed)**. Bu dosya metadata'sıdır; uzak8443'ün aynı leaf'i sunduğu bu tur ölçülmedi. Standart doğrulama hedefiyle wake **BLOCKED**: yetkili CA imzalı, SAN uyumlu tam zincir sağlanmalı veya ayrı onaylı özel trust-anchor politikası belirlenmeli. TLS doğrulamayı kapatmak çözüm değildir; leaf dosyasını bulmak sistem CA trust kanıtı sayılmaz. Nihai binary'nin TLS yapılandırması standart doğrulama hedefiyle uyumlu olduğu PANEL DEV tarafından doğrulanmalı.

### Güncel ayrı onay kapıları
- PASS sonrası adı/hash'i belirli release için prod binary/env/plans/secrets ve :8787 start; gerekiyorsa uslic dosya izinleri.
- Yalnız prod nginx account diff'i + nginx -t sonrası reload; firewall/SSH/fail2ban değişikliği varsa ayrı kapsam.
- Gerçek ledger snapshot/scratch uyum ve restore drill bakım penceresi; backfill/veri restore ayrı karar.
- SMTP yetkileri, kontrollü test mesajı/alıcı/içerik ve sonra genel teslim açılması.
- Canlı Shopify üç abonelik, kontrollü sipariş/cancel/refund ve satış açılması.
- Wake CA çözümü, ücretli GPU kabulü/bütçe ve sürekli alarm kurulumu.
- Certbot dry-run'ın olası yazma/hook etkileri için ayrı izin; DNS değişikliği A'da gerekmiyor.
- Rollback'in servis/config/abonelik adımları; veri restore/revoke/iade/müşteri maili ayrıca açık onay.
