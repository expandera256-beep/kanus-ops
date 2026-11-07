# Kalan işler — .5

Güncelleme: 06.11.2026 06:10 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

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

- NOT TESTED — Yeni mail order.name kullanır; önceki pending outbox kayıtlarında orderName yoksa numeric id üzerinden insan sipariş adı üretilemez. Eski kayıtların uygun bakımda gözden geçirilmesi gerekir; bu tur canlı depo değiştirilmedi.

## Round6 .6 kabul sınırı
- BLOCKED — Yeni Windows GUI headless flag/AttachConsole davranışı native Windows üzerinde henüz çalıştırılmadı; Linux main/unit ve PE build PASS bunun yerine geçmez.
- BLOCKED — Model upstream yeni plan127.0.0.1:18000 tunnel bekliyor; gerçek rental test US_REAL_MODEL_E2E opt-in kapalı SKIP. HOST runtime endpoint kabulü ve ayrı canlı test onayı gerekli.
- NOT TESTED — Fake SSE50-lisans testinden gerçek GPU throughput/TTFT veya uzun soak sonucu çıkarılamaz; önceki platform açıkları korunuyor.
