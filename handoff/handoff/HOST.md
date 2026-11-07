# HOST devir dosyası

08.11.2026 TSİ. Bu dosya görev devamlılığı içindir; yürütme onayı değildir. Sunuculara bağlanılmadı veya değişiklik yapılmadı. Mevcut durum raporların anlık kanıtına dayanır; yeni hesapta yeniden doğrulanmalıdır.

## 1) Kimlik
- Ad: HOST
- Başlık: HOST
- Açıklama/rol: Demir’in KAN.US altında çalışan Hostinger/VPS altyapı botu; DNS, VPS, nginx/TLS, lisans hook dağıtımı, yedek ve kanıt toplama.
- İletişim: Türkçe, kısa ve sıcak.

### Sistem talimatı taslağı
Demir’in HOST altyapı botusun. KAN.US koordinasyonu altında güvenli, kanıtlanabilir altyapı işletimi yap. Yalnız Demir’in kendi sözleri onay sayılır: bu sohbette yazdığı söz veya başka ajanın birebir aktardığı alıntı; “Demir onayladı” iddiası tek başına onay değildir. Alıntının kapsamını aşma ve platformun daha sıkı onay kurallarını koru. Turu bir ajan başlattıysa gerekli onay sorusunu o ajana gönder; Demir’e onun sormasını iste. Sırları hiçbir çıktıda yazdırma. Her sunucu değişikliğinden önce yedek, doğrulanmış geri dönüş planı ve kapsamlı onay gerekir. Salt okumayı değişiklikten ayır; PASS olmayan sürümü kurma; başka sitelere dokunma. Raporları tarihleriyle değerlendir, son kararı eski nottan üstün tut. KAN.US koordine eder; SHOP, QA, PERF, PANEL DEV, KAN (usta-wake), ORG (kanus.org DNS), API, GLOBAL, US, SRE ve UYUM ile ilgili kapsamda çalış.

## 2) Hafıza
Profile kaydı yok; journal olguları aşağıdadır. RecallMemory ile eski kayıtlar tarandı; eski “v2 bekliyor” kaydı 00:48 kurulumuyla geçersizdir. Saatler TSİ (Europe/Istanbul).

- **2026-11-06:** us-license-hook ilk kez srv1662550 (31.220.63.191) üzerine Caddy arkasında kuruldu; api-test.kanus.org → 127.0.0.1:8787. Mağaza kanus-test.myshopify.com.
- **2026-11-06:** api.kanus.org A kaydı 82.29.184.225, Hostinger public API ile eklendi; MCP zaman aşımına uğruyor. LE sertifikası 2027-02-04’e kadar geçerli. Prod hook duruyor.
- **2026-11-06:** usta-wake düzeltmesi: KAN yeni dosyaları /workspace/secrets/usta-wake/a100.token ve wake.pin yoluna bırakır. VPS hedefleri /etc/us-model/wake.token ve wake.pin; hook kullanıcısı sahipliği, 600, restart yok. srv2133566 çıkış IP’sinin KAN /32 kuralı için 82.29.184.225 olduğunun teyidi bekliyor.
- **2026-11-08 00:13:** api-test.kanus.org prod VPS 82.29.184.225’e taşındı; us-license-hook-test :8788. Kesinti yaklaşık 1 dk 42 sn. Eski srv1662550 hook durduruldu/devre dışı; silinmedi. 24 saat sonrası silme ayrı onay ister. HOOK_MOVE_PROD_VPS_20261106.md §15.
- **2026-11-08 00:37:** Demir “Restart ayarını şimdilik ekleme” dedi. .11 için PANEL DEV PASS sonrası dağıtım izni vardı; .11 FAIL aldığı için uygulanamaz, sonraki sürüme taşınamaz.
- **2026-11-08 00:46:** .11 PANEL DEV FAIL: kredisi olmayan oturum GPU uyandırabiliyor.
- **2026-11-08 00:48:** Demir onayıyla backup v2 kuruldu; saatlik :05 prod /var/backups/us-license-hook, :07 test /var/backups/us-license-hook-test, 7 gün saklama. v1 yedekleri /root/*.v1-20261107T2148*.
- **2026-11-08:** PERF’e göre nginx login 429 yanıtında Retry-After yok; sonraki api-test dağıtımıyla planlandı.
- **2026-11-08 01:04:** Demir “Restore drill'i şimdilik yapma” dedi.
- **2026-11-08 01:10:** .12 PANEL DEV FAIL; kredi tükenmesi, revoke veya iptal sonrasında GPU wake mümkün; model_proxy.go:127–135, 535–550.
- **2026-11-08 01:22:** API’nin birebir aktarımı: Demir “Pin olmasın, yalnız standart sertifika doğrulaması kullanılsın”. Kapsam panel → lisans sunucusu FX bağlantısıdır. Wake 34.41.186.170:8443 PEM pin ile kalır. GOLIVE raporundaki wake standart CA blokeri bu kapsam ayrımıyla okunmalı; doğrulama kapatılmaz.
- **2026-11-08 01:22:** KAN.US aktarımına göre altyapı seçimi A: api.kanus.org mevcut 82.29.184.225’te ayrı üretim kurulumu. Yalnız seçim onaylı; kurulum/başlatma değil.
- **2026-11-08 itibarıyla:** SHOP #1003 (19002359972093) üretildi, sonra iptal/iade edildi; maskeli us1.…pLm8 lisansı revoked, 0 kredi. Test siparişi 990000000001 SİLİNMEYECEK.

## 3) Rutinler
Kaydedilmiş rutin yok; zamanlama/tetik/prompt yok. Sunucudaki backup cron’u bot rutini değildir.

## 4) Beceriler, bağlayıcılar ve servisler
- Hostinger hosting, dns, vps, domains, wordpress, agency-hosting, billing, reach, ecommerce
- Hostinger public API (curl; MCP bozuk/zaman aşımı)
- Shopify connector
- GitHub connector
- SSH
- openssl
- certbot
- nginx
- systemd
- cron
- Shopify webhooks
- Hostinger SMTP (smtp.hostinger.com:465)
- routines
- canvas
- unblock-research
- troubleshoot-website
- sign-in

## 5) Açık işler ve sunucular
### Prod VPS: srv2133566 — 82.29.184.225
- nginx 80/443; api-test.kanus.org HTTPS → 127.0.0.1:8788, us-license-hook-test, uslic-test, /opt/us-license-hook-test. Çalışan binary rapor SHA256: `7d72f449ae3011e7f37632b0aced7ba827113940846213a3dd94415a921256f0` (sunucuda bu görevde yeniden doğrulanmadı; .11/.12/.13 olduğu iddia edilmez).
- api.kanus.org → us-license-hook :8787, uslic, /opt/us-license-hook; kurulu ama inactive, dinleyici yok. Secret’lar boş, mağaza domain’i placeholder, SMTP kapalı. Env yolları /etc/us-license-hook.env ve /etc/us-license-hook-test.env.
- LE ECDSA sertifikaları: api 04.02.2027 08:13:20 TSİ, api-test 04.02.2027 03:48:11 TSİ. SAN/hostname ve tam zincir doğrulaması exit0/OK raporlu; certbot.timer aktif. Son plan turunda renew dry-run yapılmadı; eski cutover raporunda geçmiş başarılı dry-run ayrı tarihsel kanıttır.
- /etc/us-model/wake.* uslic:uslic, 600. wake.pin hash-benzeri, 53 byte; DEĞERİ YAZILMAZ. .12 PEM ister; mevcut dosyayı ezmeden ayrı uslic-test PEM kopyası için onay gerekir.
- Yedekler: v2 :05 prod, :07 test; licenses.json, plans.json ve env, manifest SHA/JSON kontrolü, LAST_SUCCESS, 7 gün. Son raporlu başarı 08.11.2026 00:48:05 ve 00:48:06 TSİ. Güncel başarı varsayılmaz.
- Rollback: dağıtım öncesi binary/env/plans/unit/nginx/PEM yedeği; yalnız onaylı servis/config geri dönüşü. Ledger uyumu kanıtlanmadan eski binary veya eski veri üzerine dönüş yok. Cutover nginx ve store yedek yolları §7’de. Backup v1 geri dönüş dosyaları /root/*.v1-20261107T2148*.

### Eski test VPS: srv1662550 — 31.220.63.191
- Caddy 80/443 → eski us-license-hook 127.0.0.1:8787. Hook stopped/disabled; /opt/us-license-hook ve ledger korunuyor.
- Yedekler: yerinde eski store ve box cutover kopyası. Rapor eski backup-r7-accounts-20261106-053004 ve backup-r7-*, backup-r5-*, backup-20261106-025443, backup-smtp-preenable-*, backup-e2e-*.json adlarını kaydeder; varlıkları yeni görevde doğrulanmalı.
- Tarihsel cutover rollback: api-test DNS → 31.220.63.191, eski hook start, prod test hook stop. Bu yürütme izni değildir; yeni ledger değişiklikleri korunmadan eski store’a dönülmez. 24 saat doldu diye silme yok.

### Wake: 34.41.186.170:8443
- Self-signed TLS; /etc/us-model/wake.crt metadata’sında IP SAN mevcut. Sistem CA doğrulaması başarısız; panel FX pin iptali wake’e uygulanmaz. PEM pin doğrulaması korunur. Wake sunucusunun yedek/rollback ayrıntıları doğrulanmadı; KAN’dan alınmalı.

### SSH ve güvenli çalışma
- root + ed25519. Prod anahtar yolu /workspace/secrets/hostinger_api_prod_ed25519; eski test /workspace/secrets/hostinger_srv1662550_ed25519. İçerik okunmaz/yazdırılmaz. Mevcut box üzerinde ajanlar ortak dosya sistemini kullanır; başka hesapta aynı box/erişim garanti değildir, erişilebilirliği kontrol et ve gerekirse güvenli yeniden sağlama iste.
- Bağlantılar kesintili: ConnectTimeout 40–90 ve sınırlı tekrar. Box DNS fake-IP proxy 198.18.0.1; gerçek DNS/TLS kontrollerini VPS’ten yap. Box çıkışı 140.248.52.x.
- E-posta a***@ biçiminde; lisans us1.…last4; mail gövdesi yalnız uzunluk. Query’de lisans/token kullanma; nginx query loglar. QA’nın raporlu taramasında hassas query 0, 8 probe eşleşmesi var.

### Devam noktası
- .13 henüz teslim edilmedi. DEPLOY_PLAN_R13_APITEST_20261108.md hazır; aynı artifact için PANEL DEV PASS ve Demir’in .13 onayı şart. .11/.12 FAIL; kurma.
- GOLIVE_PLAN_APIKANUS_20261108.md A seçeneği ve nginx/api.kanus.org-account.proposal.diff yalnız öneri; uygulanmadı.
- Blokerler: .13 PASS, canlı secret’lar/plans.json ticari eşlemesi, gerçek ledger/rollback uyumu, SMTP teslim kabulü. SMTP allowlist tanımsızsa genel teslim; boş tanımlıysa herkesi engeller. Nihai binary ile tekrar doğrula.
- QA FINDING-03 FIXED-PENDING-VERIFY; FINDING-05 api-test backup kapsamı mitigated; restore drill BLOCKED (Demir şimdilik istemiyor). FINDING-01 host sandbox politikası PANEL DEV tasarımını bekliyor.
- SPKI_PIN_PLAN_20261108.md iptal. SPKI/yedek key/reuse_key işi yapılmayacak.

## 6) Bekleyen Demir kararları
- PASS sonrası .13 api-test deploy; uslic-test için ayrı PEM wake kopyası ve Retry-After.
- srv1662550 eski hook’un 24 saat sonrası silinmesi.
- Hostinger CPU steal destek bileti açılması (yanıtsız).
- Prod go-live ayrı kapılar: binary/env/plans/PEM/nginx kurulumu ve :8787 start; SMTP kontrollü test mesajı/alıcı/içerik ve genel teslim için allowlist kaldırma; üç canlı Shopify webhook (orders/paid, orders/cancelled, refunds/create); smoke/test sipariş ve satış açılması; ücretli GPU bütçesi; rollback.
- Ledger/restore bakım penceresi; mevcut restore drill yasağı kalkmadan yapılmaz.
- certbot renew dry-run: KAN.US şimdilik çalıştırılmamasını söyledi; ayrı onay gerekir.
- Round-2 hardening uygulanmadı.
- Sürekli yasaklar: Restart=on-failure ekleme; test VPS fail2ban’a dokunma; yeni kullanıcı/izin oluşturma; sormadan VM restart yok; onay dışı DNS ve başka nginx sitelerine dokunma.

## 7) Kritik dosya yolları ve sha'lar
Aşağıdaki box dosyalarının SHA256 değerleri bu devir sırasında sha256sum ile doğrulandı. Raporların içeriği değiştirilmedi. Env/ledger içerikleri aktarılmadı.

- `/workspace/us-panel/deploy/hook-move/DEPLOY_PLAN_R11_APITEST_20261108.md` — `e044ac895be1687c5ec289aa64f43513b91b64b511aee23e94626896ad289b5c`
- `/workspace/us-panel/deploy/hook-move/DEPLOY_PLAN_R12_APITEST_20261108.md` — `6e5151fbe748adab0a299c9acf560be213b9ef516af8fd812048f00e4786099f`
- `/workspace/us-panel/deploy/hook-move/DEPLOY_PLAN_R13_APITEST_20261108.md` — `8f0dd4011ab934c26c53f38d9d11d0e577c989cf48e87ac542a596824e3c4faf`
- `/workspace/us-panel/deploy/hook-move/GOLIVE_PLAN_APIKANUS_20261108.md` — `5692eb6628dbb0b7f48498725f2b3053f8963cc56ea66222eb83129b586f9186`
- `/workspace/us-panel/deploy/hook-move/backup/licenses.json.cutover-20261108T001140+0300` — `c2a1502239422a9e047633ed86905cae822174dd6aa98a06c80603e556e149be`
- `/workspace/us-panel/deploy/hook-move/nginx/api.kanus.org-account.proposal.diff` — `fae8f7995a878e887ed9ada02ac6b16d6aece257749293bff49d06a9e1ebe478`
- `/workspace/us-panel/deploy/host-hardening/PROD_PLAN.md` — `828e23c7df4c98a9b16cb9099f2540ca4c66c7a8dbb8e3e35607d95695b28f9a`
- `/workspace/us-panel/deploy/host-hardening/README.md` — `20392709dadcbc4f7cf0f713cc188a0ff7e3ad30b23cb3a425cfc7bb90268d64`
- `/workspace/us-panel/deploy/host-hardening/backup/README.md` — `dee3c0929221058967df5185f1a879056bddcbc27208934c57371738fe23435a`
- `/workspace/us-panel/deploy/host-hardening/backup/RESTORE.md` — `8f241b30b22044248aa50a755d07a629f12fc0e46b88b8932d91481cf269731c`
- `/workspace/us-panel/deploy/host-hardening/backup/backup-us-license-hook-hourly.v2.sh` — `d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0`
- `/workspace/us-panel/deploy/host-hardening/dns/kanus.org.zone.backup-20261106T091042+0300.json` — `4bda4ce2f1c7916a5fcc05b0c4a007f0aecb106c879bff9bb37bd46cc4816b1e`
- `/workspace/us-panel/deploy/host-hardening/dns/kanus.org.zone.backup-20261106T095924+0300-after-apitest-ttl.json` — `73429e7003dbd62ec2869b4a52e2023815a2b280ff1b94bf1546551ac9669757`
- `/workspace/us-panel/deploy/host-hardening/dns/kanus.org.zone.backup-20261106T095924+0300-before-apitest-ttl.json` — `34c06ac2120bb2a7913b755bb217aeeda81be79dbaa9136ac357af3bddb739be`
- `/workspace/us-panel/deploy/host-hardening/dns/kanus.org.zone.backup-20261108T001241+0300-after-apitest-cutover.json` — `5c8c4ae4b252321761932ac221c9be43dfbe873fb31665b0549ca6c7a4579e26`
- `/workspace/us-panel/deploy/host-hardening/dns/kanus.org.zone.backup-20261108T001241+0300-before-apitest-cutover.json` — `73429e7003dbd62ec2869b4a52e2023815a2b280ff1b94bf1546551ac9669757`
- `/workspace/us-panel/deploy/host-hardening/dns/kanus.org.zone.backup-ATTEMPTED-20261106T090504+0300.txt` — `276ac5d6df67501276848d8280495c44db7ce4f6e504f49af999aa3016f08aac`
- `/workspace/us-panel/deploy/host-hardening/nginx/README.md` — `574d9cc72cc24d02149a7a35787fa70c3c7889b31c513617d75574e931b8cec8`
- `/workspace/us-panel/deploy/host-hardening/security/README.md` — `68434621ce2e5e11484f8b727799d7937e1a54c7cd9d00173bbbfcb84b53bd9b`
- `/workspace/us-panel/deploy/host-hardening/security/SSH_FLAKY_REPORT.md` — `40f8f8ad06c556dd2f81fdccf4223b37566753beab19cea287c0bbdf28018f3b`
- `/workspace/us-panel/deploy/host-hardening/security/TLS_AND_FIREWALL.md` — `8eb52d1dfa54363d4cec692e635b9c36f9f1b09cf3839fce4862f2aafce3224e`
- `/workspace/us-panel/deploy/host-hardening/tls/SPKI_PIN_PLAN_20261108.md` — `15801e6f84d044358bc81a5c6c9e84ebf8b10a8f400e94f4e124b8baebf442ee`
- `/workspace/us-panel/dist/2026.11.06.11/SHA256SUMS` — `8b8792f5171d680559f7e5dc58099e522927a8cc7cb26d61932b73ebc70d46fe`
- `/workspace/us-panel/dist/2026.11.06.11/us-license-hook_hostinger.tgz` — `9db1386ddf9bb2d128d4d2bcd689a8814bedcba6d6ea950e6d4479a0dcbd87f9`
- `/workspace/us-panel/dist/2026.11.06.11/us-license-hook_linux` — `12772d120e42a414b712cc22f6442e5e52404cbdb25725f36e1fae49e2662b47`
- `/workspace/us-panel/dist/2026.11.06.12/SHA256SUMS` — `38e96f7b3f147d06c28487d9ff6c9ba0859d73486285b9ca503bc591cf24b981`
- `/workspace/us-panel/dist/2026.11.06.12/us-license-hook_hostinger.tgz` — `b5489468efddfde838f940e64b40c2e512788d64acd880a860ea85a1822f539d`
- `/workspace/us-panel/dist/2026.11.06.12/us-license-hook_linux` — `b2d7ab28f65dc98b6b75373f867575b73c030547b7745e8cf71a527dd1ec2c72`
- `/workspace/us-panel/reports/HOOK_MOVE_PROD_VPS_20261106.md` — `f037dc0a90df26b047b038da2eeae1f6823aa3a997e6a13a33dd2277ecf34fcf`
- `/workspace/us-panel/reports/r20261108/GLOBAL_FX_UPLOAD_20261108.md` — `d1382e9951baac719de31b1bfa5be17d554f7db9e0585a6e4c4b6f5291680a8c`
- `/workspace/us-panel/reports/r20261108/GLOBAL_LIVE_I18N_TEST_20261108.md` — `9c23d82b9b02de4f2d487fa5d3e55a93827c2dce4a1dcbafe83d63df3b3ffd8a`
- `/workspace/us-panel/reports/r20261108/GLOBAL_SITE_UPLOAD_20261108.md` — `644aadc9225f1f25d4dcaeff9d70a0a34e9bb2ddd1513c338406d19d890fc306`
- `/workspace/us-panel/reports/r20261108/HOST_EVIDENCE_QA_20261108.md` — `2a50ea29a72a0c3b0cbe5ea3fefd56b544d75a67ed7c79f43881e9fc11315c3a`
- `/workspace/us-panel/reports/r20261108/KANUS_AGENT_ACCEPTANCE.md` — `40c317846adf929f5db098d45c064b54a747fe1f6e68ff6e57273e608550eefa`
- `/workspace/us-panel/reports/r20261108/KANUS_CLOUDRUN_DB_ACCEPTANCE.md` — `e2afd99579c29422cf311c93707a098e59ed9aa29654c5e622c63dcdc27d9961`
- `/workspace/us-panel/reports/r20261108/KANUS_CLOUD_COSTS.md` — `68b14170cdc8284ba502caff11b73ae893590a8f9e2feb421513751e960c439b`
- `/workspace/us-panel/reports/r20261108/KANUS_CLOUD_COSTS_DETAILED.md` — `7441e6c508a714dc1fcb23ac371c33584ae71c6d61a4a792f043810e8c743313`
- `/workspace/us-panel/reports/r20261108/KANUS_CURRENT_STATE.md` — `692ca86deae6ac98b527ffc0b6a74584d31f0dd2c31a6d1f350d58bfe3f78af4`
- `/workspace/us-panel/reports/r20261108/KANUS_FINAL_RELEASE.md` — `331f8713602dd54f44d16fa15779f2fa603f617e2d25725d214cf7092de3f9bd`
- `/workspace/us-panel/reports/r20261108/KANUS_FIXED_ISSUES.md` — `de2d54772198afd6f15ce7d3db4ece1ec8c328ddc5602ad442682fbb268b35e8`
- `/workspace/us-panel/reports/r20261108/KANUS_GLOBAL_READINESS.md` — `4a69e5dcc3c4279ce4bdfbeb310cebb42f7c237541ee5c3e57c878dafe02d5bc`
- `/workspace/us-panel/reports/r20261108/KANUS_GPU_ACCEPTANCE.md` — `c1387a3523509943275acc126ac0f350e1a810cb20b2ffa0724422b3311ae25e`
- `/workspace/us-panel/reports/r20261108/KANUS_ORG_PRICING_UPDATE_PLAN.md` — `2d6451558039ba0142041ecff0ee30d1f7b45c3c595d156e71060233166e824b`
- `/workspace/us-panel/reports/r20261108/KANUS_PERF_REPORT.md` — `95abb4ee04f70d365275239e76fd58d56344425973a7b7211c3c3e306e19b888`
- `/workspace/us-panel/reports/r20261108/KANUS_SECURITY_REPORT.md` — `7b743dae1f23129202c27b2b20e4129331f011f69e0fca5e2d303ad1459d9afe`
- `/workspace/us-panel/reports/r20261108/KANUS_SHOPIFY_ACCEPTANCE.md` — `e61201d7b11c5490ac1bee59579ca8eda659ac91847d84f6c48f94d8f397e4ed`
- `/workspace/us-panel/reports/r20261108/KANUS_SHOPIFY_E2E_20261108.md` — `ba84a8898f2878fe8d17ab885f3934e68fe99d7d805ac2bddef381356d137e09`
- `/workspace/us-panel/reports/r20261108/KANUS_WINDOWS_ACCEPTANCE.md` — `707a813e2be4bc404076695d850e97730e9e4380e0bdd379c8b1d901defeb53f`
- `/workspace/us-panel/reports/r20261108/KAN_BOOT6_PLAN.md` — `de08c07a84619f1842b0e7012d06b3d5c537d48e8ee11a1537e5fa5f319fc9f3`
- `/workspace/us-panel/reports/r20261108/KAN_FIXES_20261108.md` — `fa2842120016be4020bf662c0052905a2ad2cc8179daf87f6a0cd3fa413db047`
- `/workspace/us-panel/reports/r20261108/KAN_IMAGE_ENGINE_PLAN.md` — `6d44566c9c0f20679d3c7a7f0d95b1a625a602a95eea9e13da1d4053b1d70a38`
- `/workspace/us-panel/reports/r20261108/KAN_MEDIA_MODELS_PLAN.md` — `d587eeabf145e7ddee6060ad04caa5e0f48017e923b41358b829ddb9bd392c1b`
- `/workspace/us-panel/reports/r20261108/KAN_S4_GAP.md` — `840eb5a26d771f3aedcb3767529f900a2f0674728163bda7f75adced56d2a736`
- `/workspace/us-panel/reports/r20261108/KAN_S4_LIVE_PLAN.md` — `9967d01baf61aaf2091931804f34cefa0ad7b41bb9a3a6b9f0b4a9ddf1cb2edb`
- `/workspace/us-panel/reports/r20261108/KAN_S4_LIVE_RESULT_BOOT5.md` — `665a21501d6d7fdc853950fa581a6f62992ecb25ef1595400158f1fa0ad30821`
- `/workspace/us-panel/reports/r20261108/KAN_S4_SIM_RESULTS.md` — `e50b61f20353f5552dd1862452caf7bc49e8d5695c5d38351c12069abdfd1f82`
- `/workspace/us-panel/reports/r20261108/KAN_WATCHDOG_DISCARD_FIX.md` — `42014f7ddf37c7352b6aaad0d6a456f363f00f0ce9280dab2b1ea19cc810b284`
- `/workspace/us-panel/reports/r20261108/QA_QUERY_LOG_REVIEW_20261108.md` — `aa1e9bc67f81a807e55c3abf8c59f2f1c565991978f7d63710a331dc24774356`
- `/workspace/us-panel/reports/r20261108/QA_REGRESSION_R13_COMMANDS.md` — `daa7eded3e39210bffcd4b242b8aa0a9535e4cf98a3a8fab793f1d369d31e425`
- `/workspace/us-panel/reports/r20261108/QA_SECURITY_AUDIT_20261108.md` — `cc4aa7076a9327fb2616e32bc34762cc8cd12ef28c9c4be076c944c36a15dc1d`
- `/workspace/us-panel/reports/r20261108/REMOTE_PHASES_REPORT.md` — `e1ec249af91ea49942dbab19bf7559841f2787a836b0db3dc423251211730aa6`
- `/workspace/us-panel/reports/r20261108/RUNBOOK.md` — `847452803c697e4d4096b8b611663c0887cb989872a92f535e0162451b54be20`
- `/workspace/us-panel/reports/r20261108/S7_PROJECT_PROMPT.md` — `72cd0dc48dc40652c6b322c89ad4bcd7580e9ed671b8cfc52c478f08e7b5c38c`
- `/workspace/us-panel/reports/r20261108/SHOP_AUTOMATION_COPY_DRAFT_TR.md` — `3ce7fc738d23b587405c31489ff7207086b6d9bc774cafae01d8797a447a14ff`
- `/workspace/us-panel/reports/r20261108/SHOP_LIVE_PRESALE_CHECKLIST_20261108.md` — `889a700b737c24b737ca0acb34a5230b07fa4ccd07caa516cd20776ab323fca1`
- `/workspace/us-panel/reports/r20261108/SRE_BOOT5_VERIFY.md` — `0f212c69107d2090085299048a43faf63229df3a6c5d372aaed59dbbe086c16c`
- `/workspace/us-panel/reports/r20261108/SRE_CERT_MONITOR.md` — `32c3fbad791c5f1c1598b4d8867555fdc89aef9bafe55dfb19b175e45aef99cf`
- `/workspace/us-panel/reports/r20261108/SRE_SECRETS_IAM_METADATA.md` — `4d6350b8ac5117a56275a6cf8fb8c8f0b12833aad48da138459ddfaf96988525`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/E2E_HOOK_KANUS_TEST_20261106.md` — `b35894111b4f3d4c4786eb79377bd2df7c69a3d2dfcbe2d6f215cd24db7eb516`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/GLOBAL_I18N_REPORT.md` — `d1e62910e8e5cb15ca3eab1e87448c7de36a026cb7e4e0dc6950d696feb1ec66`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/KANUS_API_REPORT.md` — `78fe0b3bb74b1b5a366c3f569dcbae3ef7bad3dcc734c3ab046acd6c467c3064`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/KANUS_GPU_REPORT.md` — `58271bde97e9dc09e659f377060d41d2a564a5f9bf9e9fb39399d087f6e82707`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/KANUS_PERF_REPORT.md` — `fb633b7c2c9faf16ffeb8a3370e78d10e382995c585adeb39f6a6935f1bcc658`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/KANUS_SECURITY_REPORT.md` — `52ec92298da5a24f349feb9cebafa94b886c3d9171a62b5c2fcc3001fc144a99`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/QA_TEST_RAPORU_20261106.md` — `766b6d674b273a4948b80a91338bab2a6256b1fc7a3aa0136f7b60ae25290422`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/SMTP_HOSTINGER_20261106.md` — `c36a43f1e2e65dc89287676bb4657a702976922c8f6eb0e6708510a9bf4ffdf6`
- `/workspace/us-panel/reports/r20261108/evidence/referenced-report-snapshots/SRE_CLOUDRUN_20261106.md` — `3f29ed9dce0ab06991da81cd545c13720055db476a8af31fd68f6b51525586df`
- `/workspace/us-panel/reports/r20261108/idle600.diff` — `202fca425892ad212282a55a90790a76c3e03a7d067356683bbe1ed71e8c2649`
- `/workspace/us-panel/reports/r20261108/image_research/black-forest-labs__flux2__README.md` — `435753749320bbb395c15983e63e3dc73936c03f0f1870c628e7bf922c55eef7`
- `/workspace/us-panel/reports/r20261108/image_research/black-forest-labs__flux__README.md` — `d87b208a6a7b4669f08fa1882025c6fe825ee29605b0250d0b9f56b36ce27078`
- `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-01.md` — `9d5f0e16ff2a08545fa9f981626287f93bd101fc621a6ed054f577bc80c98981`
- `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-02.md` — `c30b3a1fccdd93ef7178ed78f3e4571b411b7479b96ee1b7c11f7a4b22559943`
- `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-03.md` — `72c628d844efa329dd7c37a72d8d900bac19dbb13e78d62c7540f14c42053e1e`
- `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-04.md` — `a5bfa3dbf9b68a8e35b1389406bedc456d1aef912557425d95e368deca1c7e15`
- `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-05.md` — `a2ff8fe09628c6018e3b597e41c53b6f937e3f962cf00f6a708f69386ce96a12`
- `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-06.md` — `84bd8d4c8763a87bed95ff3538eea55bb283a7be6afabeea30d6e329d64448bf`
- `/workspace/us-panel/reports/r20261108/qa-findings/FINDING-07.md` — `229484a0b0f1027c011c6f1d144c88f3eb83459b0ce2fa5b0bcc2ed4e4a54b59`
- `/workspace/us-panel/reports/r20261108/qa-findings/US_YANIT_FINDING_01_02_TR.md` — `8258ede97925c50acc33ff8171ebd16feea6b20e862d7343ae0aea57f99dda52`
- `/workspace/us-panel/reports/r20261108/qa-raw/QA_SECURITY_AUDIT_BEFORE_HOST_REASSESSMENT.md` — `81486cf5a1bd0887da9febcbf9b16be05ffa81859f8b42212705dba69adab8f5`
- `/workspace/us-panel/reports/r20261108/raw_cloudrun/boot5_verify/commerce_full_spec_without_image.diff` — `31645c8921790053cce6bddebe437db73c36a557c11422cd23d589d9d44fc339`
- `/workspace/us-panel/reports/r20261108/raw_cloudrun/boot5_verify/commerce_revision_to_revision_spec.diff` — `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`
- `/workspace/us-panel/reports/r20261108/raw_costs/parent_cost_report_before_merge.md` — `2ff59ad91b0c433e1b9bdfe3eb63df2333e39dc1599afca6d502b3b0222b2557`
- `/workspace/us-panel/reports/r20261108/reference/KANUS_PERF_REPORT_before_DoH.md` — `0ea222bcd0babea206d4d4e30f8c7bd467784641cf6f6d04aca4c1f39211210c`
- `/workspace/us-panel/reports/r20261108/reference/login_results.md` — `def5cf32df923596411df049c592ae9f0b8fb142f58b917541c126fa5076b865`
- `/workspace/us-panel/reports/r20261108/reference/mock_results.md` — `de6876ea9689b1558218e41da5726a72293816639bf7270a79ccb9a075db8291`
- `/workspace/us-panel/reports/r20261108/reference/network_results.md` — `d4671885093e8fdc7c962de5c824902431ea9c7cd528e4618495ccc3af36a475`
- `/workspace/us-panel/reports/r20261108/results/credit_status_results.md` — `f7c64f802eb37cab10f54982b70f0027c134da7915be8a3f6888319802ddc251`
- `/workspace/us-panel/reports/r20261108/results/idle_policy_v5.md` — `56e5456d9ff02a24a09e60d0ab922796bddc36dce7595adad626e5c278fd4e6b`
- `/workspace/us-panel/reports/r20261108/results/initial_fixture_config_failure/mock_results.md` — `51fdb25ff1389673e037be1f5702cd7f40dec6c4436ebb377128344e7af6c26e`
- `/workspace/us-panel/reports/r20261108/results/isolation_results.md` — `7ce1e6849b1f7f44006d9e9c5aa4a4e9eca147858a213418b486c4d1c86ebdd1`
- `/workspace/us-panel/reports/r20261108/results/login_results.md` — `11ada219eb86f30b0eedd3826d9f79fcc548ed2e432dc6461aadc23ff69ad09b`
- `/workspace/us-panel/reports/r20261108/results/mock_results.md` — `b80759b954482b2515ff0e25025b6e2aa064b864ef0976b59f78b42f03e62429`
- `/workspace/us-panel/reports/r20261108/results/prompt_requirements.md` — `15555343f15702a4c94fe7722100d2ffca5916b2727a0eef4cf9cd5b34e78f05`
- `/workspace/us-panel/reports/r20261108/results/remote_comparison_plan.md` — `ce223b01408615e1b087eec92b06baac7e6cea150c3001642aedc45bfadc8aab`
- `/workspace/us-panel/reports/r20261108/results/remote_local_smoke.md` — `6cf61fd05f8df2a7360987180e39472359a66c9827ee43ba8d9813f5a4139823`
- `/workspace/us-panel/reports/r20261108/results/remote_phase_unit.md` — `dddf55c4cec350c9b4bb6be7248e4be707aafcc97c21eabbcd83421f84eaad02`
- `/workspace/us-panel/reports/r20261108/results/remote_phases.md` — `30fb5fec766c28c805f8bc2727e6348dc9bdb069f847362ed32028c968e030dc`
- `/workspace/us-panel/reports/r20261108/results/remote_phases_local_dns_stop.md` — `390d6d5c4346380f8bcbb27aae0af90844a2a5d05cdf0d49bfc8455c75f29915`
- `/workspace/us-panel/reports/r20261108/results/remote_safety_tests.md` — `b599c1635e2a04910152c4710528e0fc06d9b867ee7d70f3c006fca5f2a92284`
- `/workspace/us-panel/reports/r20261108/results/remote_update_validation.md` — `294eeabafcc76e03d6022d05845c5a2032275d9445bb93a3dc7b23e11796f11e`
- `/workspace/us-panel/reports/r20261108/results/saas7_results.md` — `62476061609f38bc84b273a4ee5f2b1c87b928346ceef6c24fbd0e470119839c`
- `/workspace/us-panel/reports/r20261108/results/section11_results.md` — `3987dd7f6f57989bec0d205078446157019677157f62dfc4f01eb5f33d94c1bb`
- `/workspace/us-panel/reports/r20261108/results/validation_manifest.md` — `a6c149afcfd3bb1ad9861010371be5f51acf168c15e941a3f1db96f86785e2c6`
- `/workspace/us-panel/reports/r20261108/saas_mock_fixture/README.md` — `778579cf47d2402342e8f3e82af02b8c0ba9a0c9e79a17b576e748ba014f7731`

Sunucu yolları (bu görevde SSH yok; SHA doğrulanmadı):
- /usr/local/sbin/backup-us-license-hook-hourly.sh — raporlu v2 SHA aşağıdaki box v2 ile aynı; sunucuda yeniden doğrulanmadı.
- /etc/cron.d/us-license-hook-hourly — doğrulanmadı.
- /etc/nginx/sites-available/api-test.kanus.org — doğrulanmadı.
- /root/api-test.kanus.org.nginx.bak-precutover-20261108 — doğrulanmadı.
- /opt/us-license-hook-test/licenses.json.bak-precutover-20261108 — doğrulanmadı.

Beklenen v2 script SHA256: `d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0`.
.12 için verilen beklenen SHA256SUMS/tgz/binary değerleri sırasıyla `38e96f7b3f147d06c28487d9ff6c9ba0859d73486285b9ca503bc591cf24b981`, `b5489468efddfde838f940e64b40c2e512788d64acd880a860ea85a1822f539d`, `b2d7ab28f65dc98b6b75373f867575b73c030547b7745e8cf71a527dd1ec2c72`; fiili box değerleri yukarıda, release’ler karıştırılmaz.

## 8) Yeniden girilmesi gereken gizli değerler
Yalnız ad/yol; değer yok. Güvenli giriş kanalından sağlanmalı; rapora veya mesaja yazılmaz.
- HOSTINGER_API_TOKEN — env, Hostinger public API/DNS.
- SSH private key dosyaları — /workspace/secrets/hostinger_api_prod_ed25519 ve /workspace/secrets/hostinger_srv1662550_ed25519; root SSH.
- Hook env hedefleri — /etc/us-license-hook.env ve /etc/us-license-hook-test.env. Aşağıdaki adlar box deploy env örneklerinden ve GOLIVE planından çıkarıldı; sunucuda env okunmadığı için tam kurulu ad envanteri doğrulanmadı. Canlı ve test kimlikleri karıştırılmaz.
- `SHOPIFY_SHOP_DOMAIN`
- `SHOPIFY_WEBHOOK_SECRET`
- `US_BCRYPT_CONCURRENCY`
- `US_BCRYPT_WAIT_SECONDS`
- `US_GPU_STATUS_URL`
- `US_GPU_WAKE_DISABLE`
- `US_GPU_WAKE_HTTP_ALLOW`
- `US_GPU_WAKE_PIN_FILE`
- `US_GPU_WAKE_TOKEN_FILE`
- `US_GPU_WAKE_URL`
- `US_GPU_WAKE_USE_SYSTEM_CA`
- `US_GPU_WARMUP_TIMEOUT`
- `US_LICENSE_SECRET`
- `US_LISTEN`
- `US_LOGIN_ACCOUNT_LIMIT`
- `US_LOGIN_ACCOUNT_WINDOW_SECONDS`
- `US_MODEL_CONCURRENCY`
- `US_MODEL_NAME`
- `US_MODEL_QUEUE_LIMIT`
- `US_MODEL_TIMEOUT_SECONDS`
- `US_MODEL_UPSTREAM`
- `US_PLANS_FILE`
- `US_SESSION_SECRET`
- `US_SESSION_TTL_SECONDS`
- `US_SHOPIFY_ENABLED`
- `US_SMTP_ALLOWLIST`
- `US_SMTP_ENABLE`
- `US_SMTP_FROM`
- `US_SMTP_HOST`
- `US_SMTP_PASS`
- `US_SMTP_PORT`
- `US_SMTP_USER`
- `US_STORE_PATH`
- `US_STRICT_DONE`
- wake.token ve wake.pin — /etc/us-model; KAN sağlar. Kaynak yolları /workspace/secrets/usta-wake/a100.token ve /workspace/secrets/usta-wake/wake.pin. Değer, hash/pin içeriği veya özel anahtar asla bu dosyaya eklenmez.
