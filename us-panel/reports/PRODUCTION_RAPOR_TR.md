# KANUS — 2026.11.06.5 son durum

Güncelleme: 06.11.2026 06:10 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Go VPS hook tek issuer ve artık güvenilir model proxy'sidir. .5 EXE yalnız https://api-test.kanus.org/v1/chat/completions üzerinden lisansla çağrı yapar; desktop Reserve/Settle çift ücretlendirmesi ve ticari SSH/model doğrudan yolu kaldırıldı. Python/Cloud Run issuer yok.
- PASS — SSE/JSON proxy: modelden önce atomik reserve, doğrulanmış bitişte settlement, hata/kesik akış/timeout/iptalde release. Upstream'e müşteri lisansı gönderilmez. Saatlik kullanım saniyeyle ölçülür.
- PASS — 50 dk istemci, 45 dk sunucu timeout; GPU ısınıyor/kuyruk durumu başarı sayılmaz; SSE heartbeat. Terminal gerçek stdout/stderr/exit-code/timeout/iptal ve Go launch; lisanslı Go bakiye/son kullanım özeti; bounded istemci ve sunucu kuyrukları eklendi.
- PASS — OpenStore 0-byte davranışı korundu ve nil-store döndürme açığı düzeltildi. US_PLANS_FILE önceliği korundu/testlendi. Lifetime flock/LockFileEx ve atomik fsync/rename; Linux ayrı süreç testi geçti.
- PASS — Go tam suite/vet/race ve çapraz derleme geçti; artifact/paket byte eşitliği ve hash doğrulandı. Ayrıntılar KANUS_TEST_RESULTS.md.
- PASS — 15 dk idle (900 sn), 10 dk boot (600 sn) mevcut Python lifecycle config/logic fake-provider testleriyle doğrulandı; bu kaynak tek başına aktif Go model proxy'sinin VM yönetimi değildir.
- BLOCKED — VPS -> ds-a100 özel erişim/tünel ve aktif Go proxy -> GCP start/idle-stop/watchdog entegrasyonu kurulmadı. .5 VM'yi kendisi açıp kapatmaz. Native Windows bu Linux box'ta BLOCKED; gerçek model/ödeme kabulü yok. Üretim hazır onayı verilmedi.
- PASS — Deploy/Shopify/GCP/SSH/SMTP/canlı çağrı yapılmadı; gerçek sipariş/ödeme Demir'in yeni onayını gerektirir. Sırlar açılmadı, env placeholder korundu.


| Durum | İş | Neden / sonraki kabul |
|---|---|---|
| BLOCKED | Native Windows EXE | Box Linux; gerçek Windows başlatma/WebView, terminal process-tree, dosya izinleri/LockFileEx, SSH'siz model ve güncelleme kabulü yapılmadı. EXE_SMOKE_CHECKLIST_TR.md uygulanmalı. |
| BLOCKED | VPS -> ds-a100 upstream | Go sadece VPS'deki US_MODEL_UPSTREAM'e gider. HOST özel ağ/SSH tüneli, IAM/firewall ve upstream erişimini yetkili ortamda kurup doğrulamalı; PC tüneli VPS tüneli değildir. |
| BLOCKED | Aktif otomatik GPU start/15 dk stop | 900 sn policy mevcut Python controller'da fake-provider testli; Go proxy ile worker lease/kuyruk/watchdog/provider adaptörü bağlı değil. Health/metadata GCP başlatmaz; VM kapalıysa proxy kendisi uyandıramaz. |
| BLOCKED | Çok kullanıcılı OS izolasyonu ve CPU/RAM sınırı | Terminal güvenilir kullanıcının proje kodunu OS hesabıyla çalıştırır; cwd/env/output/time/concurrency sınırı vardır, container/cgroup/Windows Job sandbox yoktur. Proje kodu cwd dışına programatik erişebilir. |
| BLOCKED | Tam Go kullanıcı hesabı/oturum, birleşik cüzdan | .5 lisans-scoped Bearer kullanır; Python account formu kapalı, ayrı Go hesap/şifre/hesaplar arası kredi birleştirme tasarımı yok. Lisans kredi özeti hesap API'si diye sunulmaz. |
| NOT TESTED | Canlı model ve 45 dk cold-start | 50 dk timeout/heartbeat durum makinesi testli; gerçek 45 dk GPU akışı çalıştırılmadı. |
| NOT TESTED | Shopify Bogus Gateway / iptal / iade | Gerçek sipariş, ödeme veya iade için Demir'in yeni onayı yok; yalnız dry-run ve signed fixture testleri. |
| NOT TESTED | TLS/nginx/Caddy/systemd/backup-restore | Dosyalar hazır; canlı servis, config binary lint ve restore kabulü yok. Aynı JSON'a ikinci writer fail-closed; rolling çok örnek yok. |
| NOT TESTED | Modelün doğrudan ağ erişimini kapatma | EXE yolu kapalı; müşteriye eskiden verilmiş SSH/VM yetkilerini ve firewall'ı HOST incelemeli. Bu tur yetki değişmedi. |
| BLOCKED | SMTP teslimatı | Bilerek US_SMTP_ENABLE=0; posta/SPF/DKIM/DMARC kabulü ve müşteri teslimatı yapılmadı. |
| NOT TESTED | Crash/retry uzlaşması | Commit sonrası terminal yanıtın teslimi kaybolabilir; reserve sonrası crash TTL'e kadar hold bırakabilir. Aynı request_id için kalıcı sonuç replay/uzlaşma yok; kör tekrar çift model çağrısı olabilir. |
| NOT TESTED | Finansal-only refund / legacy saatlik | Miktarsız iade manuel inceleme; eski use anahtarında belirsiz hareket varsa otomatik migration reddedilir. Yeni saate bitmiş bakiye eklenmez. |
| NOT TESTED | PostgreSQL/Gunicorn arşiv yolu | 40 opt-in test çalışmadı; aktif Go JSON yolu SQL/Cloud Run değildir. |
| NOT TESTED | Güvenlik denetimi / tokenizasyon / gerçek yük | govulncheck/DAST ve webhook topic-body semantik/replay incelemesi; gerçek tokenizer, TTFT/GPU maliyeti ve 1/5/10/25/50 GPU benchmark yok. |
| NOT TESTED | TypeScript/YAML/SQL runtime kabulü | Dosya düzenleme var; gerçek TS derleme, YAML semantik ve PostgreSQL/SQL çalışma kabulü yapılmadı. Bağımlılık kurulumu kullanıcı onayı dışında otomatik yapılmaz. |

## Round 7 açık kabul işleri — 06.11.2026 08:20 TSİ (Europe/Istanbul)
- BLOCKED — Yeni .7 hook + /account/ nginx prefix + US_SESSION_SECRET HOST tarafından yetkili kurulmadan .7 EXE'nin canlı login'i hazır sayılmaz; mevcut VPS sürümü sadece diğer ajan bildirimiyle Round5. Bu tur deploy yok.
- BLOCKED — Native Windows GUI/console-attach/CLI/Job/process/network kabulü Linux box'ta çalıştırılmadı. EXE_SMOKE_CHECKLIST_TR.md ve WINDOWS_SMOKE_COMMANDS_20261106_TR.md yeni .7 komutlarını içerir.
- BLOCKED — VPS-origin 127.0.0.1:18000→ds-a100 private model erişimi/tüneli yok (aktarılmış durum, uzaktan kontrol edilmedi); Go GCP start/15dk idle-stop controller entegrasyonu yok. Workers metadata15dk yalnız policy, VM kapatma kanıtı değil.
- NOT TESTED — Gerçek model çok-dosyalı onarım/rental kalite kabulü, approved Bogus Gateway account E2E, SMTP delivery ve gerçek mailbox verification/reset. Register email sahipliği kanıtı değildir; doğrulama opsiyonel, SMTP0. Reset/verify GUI kontrolleri henüz yok, endpoint/local test var.
- NOT TESTED — Bağımsız auth/penetrasyon ve dependency vulnerability audit; persistent/distributed IP limiter, per-session logout/refresh yok (reset tüm sessionVersion'ı değiştirir). Reset raw us1 anahtarını iptal etmez.
- NOT TESTED — Crash/result replay ve büyük JSON performansı; hold expiry cleanup sonraki reserve/redeem gibi yazan işlemde yapılır, salt bakiye eski reserve gösterebilir. Tek writer şart; eski R5/R6 writer yeni accounts alanlarını düşürebilir: kör rollback yasak, tutarlı snapshot/reconciliation gerekir.
- NOT TESTED — Tek model çağrısı kaynak paketler arasında saniye bölmez; aggregate seconds tek uzun çağrı garantisi değildir. Hourly seçim en yüksek kalan saniyeyi önce kullanır (eski1s paketin yeni1h paketi gölgelemesi testle düzeltildi). Sub önce seçilir; başarıda cold-start dahil upstream wall süre ceil-saniye, queue hariç; failure/cancel0debit.
- PASS — Önceki “Go account endpoint yok / local404” kod açığı yerelde kapandı; canlı gap henüz kapandı diye sunulmaz. Tüm son testler0FAIL; altı skip nedeni ROUND7_RAPOR_TR.md.
