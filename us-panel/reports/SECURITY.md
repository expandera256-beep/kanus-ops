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

## Round 7 — 06.11.2026 08:20 TSİ (Europe/Istanbul)
- PASS — bcrypt12 (UTF8byte12..72), unknown-user dummy compare, signed HMAC session expiry+nonce+version; reset eski token'ları invalid eder. IP15dk register5/login10/reset5/verify10;4096bucket,4bcrypt worker; güvenilmeyen X-Real-IP spoof testi reddedilir.
- PASS — Aynı flock+fsync+rename JSON'da metadata accounts+licenseOwners; ikinci kredi ledger yok; backward-open/no rewrite ve restart persistence testleri;50session/raw yarışında tek debit. SMTP allowlist set-empty deny-all, dış alıcı fake sender'a dahi verilmez. Multi-candidate n/best_of/max_completion_tokens override400/no-debit testli.
- NOT TESTED — Bunlar bağımsız security audit değildir; dependency vuln taraması yapılmadı (bcrypt golang.org/x/crypto v0.33.0). Register mailbox proof değildir; raw key reset ile revoke edilmez; local IP limiter restart'ta sıfırlanır, nginx koruması ve8787loopback gerekir.
- BLOCKED — Canlı auth/TLS/nginx/SMTP/Windows kabulü yok. Yeni secret yalnız placeholder, SMTP0; gerçek secret/credential dosyaları bu tur okunmadı.
