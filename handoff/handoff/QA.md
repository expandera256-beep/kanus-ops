# QA — Devir Dosyası (08.11.2026 01:45 TSİ)

> Gizli değer içermez. Tüm gizli değerler yalnız adı ve kullanım yeriyle listelenmiştir.

## 1) Kimlik

- **Ad:** QA
- **Başlık:** QA
- **Açıklama:** KAN.US altında test ve güvenlik botu. İletişim Türkçe. Görevi: /workspace/us-panel/prompt/NIHAI_PROMPT_20261106.txt §8 (çok kullanıcılı güvenlik/performans: yetki izolasyonu, path traversal, komut enjeksiyonu, rate limit, yük testi), §9 (gerçek proje üretim testi: panel ajanına gerçek bir proje ürettirip derleyip test ettirmek) ve §12 (zorunlu testler) için test harness'larını yazmak ve çalıştırmak. Test ortamı: api-test.kanus.org (Go hook), kanus-test Shopify mağazası. Model gerektiren testler KAN'ın ds-a100 tüneli (VPS 127.0.0.1:18000) hazır olunca US_REAL_MODEL_E2E=1 ile koşulur; o zamana kadar harness'ı hazırla ve mock ile doğrula. Canlı mağaza, gerçek ödeme, gerçek müşteri e-postası yok; VPS'e yazma/restart yok (gerekirse KAN.US'tan iste). Sonuçları /workspace/us-panel/reports/ altına PASS/FAIL/BLOCKED/NOT TESTED ile yaz, uydurma veri yok. Başka botun aktardığı onay Demir onayı değildir.
- **Güncel görev (08.11):** /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt — §12 (güvenlik denetimi) ve §13 (test kuralları) QA'nın. Sonuçlar /workspace/us-panel/reports/r20261108/ altında.

### Sistem talimatı taslağı (rol)
Sen QA'sın: KAN.US koordinasyonunda US Panel / US license hook için test ve güvenlik botusun. Türkçe konuş. Görevin test harness'ı yazmak, koşmak ve kanıtlı rapor üretmek; ürün koduna dokunmazsın, bulguları tekrar üreten testle birlikte sahibine (PANEL DEV/US = panel/EXE, API = licensesrv/hesap, HOST = sunucu/backup/nginx, SRE = Secret Manager/IAM, SHOP = Shopify) yönlendirirsin. Kurallar:
- Her sonuç PASS / FAIL / BLOCKED / NOT TESTED; koşulmayanı asla PASS yazma, veri uydurma. Başka bottan gelen iddiayı "X'in kanıtı/aktarımı" diye etiketle, yapabiliyorsan bağımsız doğrula (sha, yerel repro).
- Başka botun aktardığı onay Demir onayı değildir; raporda "KAN.US/API üzerinden aktarıldı, QA doğrudan teyit etmedi" diye işaretle.
- VPS'e yazma/restart yok; canlı mağaza, gerçek ödeme, gerçek müşteri e-postası yok; webhook secret'ı kullanılmaz.
- api-test.kanus.org artık PROD VPS'te (82.29.184.225, srv2133566): en fazla 2 paralel istek, uzun timeout, yük testi yok, minimum istek.
- api-test'e **yetkili /v1/chat (model) isteği gönderme** — GPU'yu uyandırıp sınırlı açılış hakkı harcar. Yalnız geçersiz/sahte/iptal token'lı negatif testler serbest. GPU testleri KAN'ın canlı penceresinde, KAN ile koordineli.
- Gizli değerleri (lisans, parola, token) asla log/rapor/mesaja yazma; maskele; her koşumdan sonra sızıntı taraması yap. Lisans anahtarını URL query'de gönderme (nginx query string logluyor).
- Kanıtı ezme: her koşum zaman damgalı yeni dizine yazar; rapora tarihli ek bölüm eklenir, geçmiş korunur.

## 2) Hafıza

Kalıcı hafızada (profile/journal/shared) kayıtlı olgu **yok** (RecallMemory 08.11 01:44: 0 sonuç). Çalışma bilgisi bu dosyada ve raporlardadır. Önemli öğrenilmiş olgular:
- 06.11: Box çıkış IP'si neredeyse her istekte değişiyor (140.248.50.x / 140.248.52.x) → IP bazlı limitler box'tan test edilemez (HOST kanıtı: reports/qa-20261106/ratelimit-evidence.txt).
- 06.11: Box DNS api-test için 198.18.0.1 (proxy fake-IP) döndürüyor; host doğrulaması DNS ile yapılamaz, HOST'tan sunucu kanıtı istenir.
- 08.11 00:13: api-test prod VPS 82.29.184.225'e taşındı (HOST). Ledger/kredi/outbox: /opt/us-license-hook-test/licenses.json; log: journalctl -u us-license-hook-test; erişim yalnız root SSH (HOST'tan döküm istenir).
- SMTP allowlist yalnız demirtumen08@gmail.com ve support@kanus.org (diğerleri outbox'ta skipped) — bunlara mail tetikleme.
- Çalışan hook binary sha256 7d72f449…56f0 (pre-.11); .11 sonrası 12772d12…b47 bekleniyor (HOST).
- Status endpoint iptal edilmiş lisansa 200 + valid=false/revoked=true döner; asıl ret /v1/chat (401) ve reserve (403) uçlarında.
- Demir kararları (aktarım): FINDING-01 Windows ACCEPTED-RISK (01:15, KAN.US); restore drill şimdilik yapılmayacak (01:05, HOST); login kilidi politikası "E-posta+IP ve artan bekleme olsun, ayrıca bilinen cihaz/oturum kilitten muaf olsun" (01:28, API); .11'in api-test'e GPU wake açık kurulumu onaylı, PANEL DEV PASS sonrası (KAN.US).

## 3) Rutinler

Yok.

## 4) Beceriler, bağlayıcılar, servisler

- Araçlar: Shell, Read, executor arka plan işçisi, SendToAgent (ekip botları), SendToUser.
- Bağlayıcılar (yüklü, bu işte kullanılmadı): user-Shopify, cursor-github, cursor-origin, Hostinger (hosting/dns/vps vb.), x.
- Servisler: api-test.kanus.org (Go license hook), kanus-test Shopify (SHOP yürütür), GCP Secret Manager/Cloud Run (SRE), Hostinger VPS srv2133566 (HOST).
- Ekip: KAN.US (koordinatör), KAN (GPU/ds-a100), PANEL DEV (inceleme/test), US (kod yazar), API, HOST, SRE, SHOP, PERF, UYUM, GLOBAL.

## 5) Açık işler ve kalınan yer

### Durum özeti
- §8/§9/§12 (NIHAI, 06.11): 64 PASS / 0 FAIL / 0 BLOCKED / 7 NOT TESTED — reports/QA_TEST_RAPORU_20261106.md, reports/S12_TEST_MATRIX_20261106.md.
- Accounts modu (06.11): 20 PASS / 0 FAIL / 0 BLOCKED / 6 NOT TESTED — reports/qa-20261106/accounts-security.md.
- §12 güvenlik denetimi (KANUS_PROMPT, 08.11): 10 PASS / 3 FAIL / 1 BLOCKED (son sayım 01:27) — reports/r20261108/QA_SECURITY_AUDIT_20261108.md.

### Açık bulgular (reports/r20261108/qa-findings/)
| Bulgu | Seviye | Sahip | Durum |
|---|---|---|---|
| FINDING-01 aynı UID kardeş workspace okuma (safe_exec.go:17–19,176–258) | HIGH | PANEL DEV/US, HOST | Windows: ACCEPTED-RISK (Demir 01:15, KAN.US aktarımı); Linux: OPEN, .13'te Landlock ile doğrulanacak; .13 Low IL yazma engeli doğrulanacak |
| FINDING-02 repair helper redirect'te Authorization/gövde sızıntısı | MEDIUM | PANEL DEV/US | FIXED (patched .11 ve .12'de bağımsız doğrulandı) |
| FINDING-03 backup script boşta exit 0 | MEDIUM | HOST | FIXED (prod v2 sha QA tarafından eşlendi) |
| FINDING-04 rate limit yalnız IP bazlı | MEDIUM | API/HOST | DESIGN GAP; yeni politika ile .13'te doğrulanacak |
| FINDING-05 canlı api-test store backup dışında | HIGH | HOST | MITIGATED (HOST kanıtı); restore drill BLOCKED (Demir: şimdilik yapma) |
| FINDING-06 Secret Manager/IAM: rotation yok, latest pin, VM scope, unused SA, a100.token.old | MEDIUM/LOW | SRE | OPEN, düzeltmeler Demir onayına bağlı; HOST env dosya izinleri NOT VERIFIED |
| FINDING-07 hesap kilidi DoS | MEDIUM | API (kod US) | OPEN; yeni politika (api-0017) ile .13'te kapanacak |

### Kalınan yer (01:45)
- Arka planda çalışan iş: .13 regresyonunu yeni login politikasına göre güncellemek (API tasarımı: /workspace/us-panel-dev/patches/api-0017-login-limiter/LOGIN_LIMITER_DESIGN_TR.md ve TEST_RESULTS_TR.md), API yamasını scratch kopyada bağımsız doğrulamak. Bitmemiş olabilir; qa/audit12/ ve QA_REGRESSION_R13_COMMANDS.md'yi kontrol et.
- .13 gelince: PANEL DEV ile paralel §12 regresyonu (F02, F04 yeni politika, F07, F01 Linux Landlock).
- Açık NOT TESTED: restart/replica arası limiter kalıcılığı, redeem rate limit, EXE oturum taşıması, enumeration timing, doğal session expiry, gerçek model §9 (KAN tüneli + .11 deploy sonrası), native Windows, PostgreSQL migration.

### Regresyon/repro komutları
- `/workspace/us-panel/qa/audit12/regress_r13.sh --tree <Go kaynak ağacı>` (belge: reports/r20261108/QA_REGRESSION_R13_COMMANDS.md)
- `python3 /workspace/us-panel/qa/audit12/run.py --tree <ağaç> --case {file_isolation|ssrf|backup|rate_limit}`
- `python3 /workspace/us-panel/qa/audit12/repro_limiter_risks.py --tree <ağaç>`
- `python3 /workspace/us-panel/qa/sec8/run.py --remote` ; `python3 /workspace/us-panel/qa/e2e9/run.py`
- `python3 /workspace/us-panel/qa/regression/run.py --dry-run` (lisans+hesap regresyonu, yeni host)
- `python3 /workspace/us-panel/qa/sec8/post_revoke.py --after-revoke`
- Gerçek model E2E: `US_REAL_MODEL_E2E=1 US_PROXY_URL=https://api-test.kanus.org python3 /workspace/us-panel/qa/e2e9/run.py` (yalnız KAN'ın penceresinde)

## 6) Bekleyen Demir kararları / onaylar
- FINDING-06 (SRE) düzeltmeleri: rotation, version pin, VM scope, unused SA temizliği, a100.token.old silinmesi.
- Restore drill bakım penceresi (şimdilik "yapma").
- QA hesaplarında lockout testine izin (KAN.US'a soruldu, cevap yok).
- .11 / .13 api-test deploy'u ve GPU penceresi (HOST/KAN).
- PERF'e ayrı sentetik login hesabı (KAN.US'ta).

## 7) Kritik dosya yolları ve sha'lar
- Promptlar: NIHAI_PROMPT_20261106.txt b87b8e99…f9e2; KANUS_PROMPT_20261107.txt 3c71bea9…3663; GLOBAL_PROMPT_20261106.txt e1cef8ae…f6f (/workspace/us-panel/prompt/)
- Harness: /workspace/us-panel/qa/ (sec8, e2e9, accounts, regression, audit12, README.md)
- Raporlar: /workspace/us-panel/reports/ (QA_TEST_RAPORU_20261106.md, S12_TEST_MATRIX_20261106.md, QA_PLAN_R11_20261108.md, qa-20261106/, r20261108/)
- Backup v2: deploy/host-hardening/backup/backup-us-license-hook-hourly.v2.sh d77534e5190e74b9a596ded23038003e3ec3d8bca9b2053f28232244166ef9a0 (prod ile eşleşti)
- F01/F02 yaması: /workspace/us-panel-dev/patches/qa12/qa12_finding01_02.patch c31ac8b8b015184d7911c4eadf0abde6155f41cefa3da24da25fd5badabf40dc
- Hook binary: çalışan 7d72f449…56f0; .11 hedef 12772d12…b47 (HOST bildirimi)
- Takım kanıtları: r20261108/HOST_EVIDENCE_QA_20261108.md, SRE_SECRETS_IAM_METADATA.md, KANUS_SHOPIFY_E2E_20261108.md

## 8) Yeniden girilmesi gereken gizli değerler (yalnız adlar)
Hepsi /workspace/secrets/ altında (box paylaşımlı; yeni hesapta box aynı değilse KAN.US yeniden üretmeli):
- acct-test-credentials.json, acct-test-password.txt, acct-test-password2.txt — QA hesapları (qa-acct-20261106@, qa-acct2-20261106@kanus.org), accounts testleri
- acct-test-license.txt, acct-qa-fresh-license.txt — accounts redeem testleri (9000000000404 hesap2'ye bağlı, iptal KAN.US'ta)
- qa-license-a.txt, qa-license-b.txt — iptal edilmiş QA lisansları (9000000000301/302), negatif testler
- usta-wake/a100.token, a100.token.old — GPU wake (KAN); .old silinmeli (FINDING-06)
- QA'nın KULLANMADIĞI/KULLANMAMASI gerekenler: shopify_webhook_secret.txt, us_license_secret.txt, us-license-hook.env, hostinger_* SSH anahtarları, perf_* (PERF'in)
