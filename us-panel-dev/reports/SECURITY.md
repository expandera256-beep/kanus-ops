# Güvenlik — .5

Güncelleme: 06.11.2026 07:19 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme kanıtıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Ticari istemci yalnız Go lisanslı model proxy kullanır; server reservation ID dışarı verilmez. Desktop signing secret yok; license Bearer upstream'e taşınmaz; redirect kapalı; upstream HTTP yalnız loopback, uzak hedef HTTPS gerektirir.
- PASS — SSE terminal ve JSON başarı cevabı settlement diske yazılmadan verilmez. Eksik/error/length stream ve timeout/cancel release testli. Model/body/token/bounded queue sınırları vardır.
- PASS — JSON tek writer flock; temp+fsync+rename; Linux directory fsync. Ayrı süreç lock testi ve tam race temiz. Anahtar/e-posta kredi status cevabında yok; kullanım yalnız o lisansa scoped, no-store.
- PASS — Terminal yalnız açık kullanıcı komutu: JSON argv, shell yok, proje cwd, env secret süzme, timeout/tree-cancel ve bounded output. UI güvenilir proje kodunun OS yetkileriyle çalıştığını açıklar.
- BLOCKED — OS CPU/RAM/container sandbox ve host kullanıcıları arası güçlü izolasyon yok. Model/SSH doğrudan ağ erişiminin kapatılması HOST firewall/IAM kabulüdür; bu tur yapılmadı.
- NOT TESTED — Commit sonrası yanıt kaybı/crash uzlaşması, gerçek TLS/rate-limit, secret rotation, Windows native LockFileEx, restore ve bağımsız security scan/topic-body replay audit açık.
- PASS — Bu tur gerçek secret dosyaları açılmadı/kopyalanmadı; env webhook/license sırları placeholder kaldı. SMTP ve canlı servis/deploy yok.

- NOT TESTED — SMTP sağlayıcı seçimi ve kanus.org SPF/DKIM/DMARC şablonları: ayrıntı **KANUS_SECURITY_REPORT.md** (Hostinger MX mevcut; DNS bu turda değiştirilmedi).
