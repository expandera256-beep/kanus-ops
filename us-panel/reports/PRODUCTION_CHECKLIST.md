# Üretim kontrol listesi — .5 / Round 6+

Güncelleme: 06.11.2026 09:13 TSİ (Europe/Istanbul)

PASS yalnız adlı kanıt dosyasıdır. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

## PASS (dosya kanıtlı)

- PASS — Billing BigQuery export (standard + detailed) etkin; dataset `billing_export` / `project-27082a40-6083-4952-9db`. **KAN.US bildirimi 08:21 TSİ** (dataset: `GCP_CLEANUP_20261106.md` §3). **İlk veri ≤24 sa bekleniyor.**

- PASS — Go suite/vet/race, cross-build. Kanıt: `KANUS_TEST_RESULTS.md`.
- PASS — kanus-test E2E paid/idempotency/HMAC/iade/iptal. Kanıt: `E2E_HOOK_KANUS_TEST_20261106.md`, `KANUS_TEST_RESULTS.md`.
- PASS — Round 5 VPS proxy yolları. Kanıt: `HOOK_VPS_STATUS_20261106.md`.
- PASS — Hostinger Mail SMTP (ENABLE=1, test gönderim). Kanıt: `SMTP_HOSTINGER_20261106.md`.
- PASS — 1 istem ürün DRAFT (15428080632036, 07:26). Kanıt: `KANUS_SHOPIFY_LIVE_AUDIT.md`.
- PASS — Accounts mode kararı = Go hook gateway. Kanıt: `ACCOUNTS_MODE_GAP_20261106.md`.
- PASS — Round 6 yük testi (mock). Kanıt: `LOAD_TEST_20261106.txt`.
- PASS — Headless 4-flag smoke (unit/Linux). Kanıt: `ROUND6_RAPOR_TR.md`, `SMOKE_CLI_LOCAL_20261106.txt`, `SMOKE_RUN_LOG_20261106.txt`.
- PASS — **Native Windows CLI smoke 4/4** (--version/--selftest/--check-server/--check-license, Demir PC, .6 EXE). Kanıt: `WINDOWS_NATIVE_SMOKE_20261106.md`.

## NOT TESTED

- **Windows GUI** (panel ekranları, hesap UI) — native smoke kapsamı dışı. Kanıt notu: `WINDOWS_NATIVE_SMOKE_20261106.md`.
- Canlı TLS/backup tam kabul, security scan bağımsız, Shopify plan fatura satırı.
- Operatör deploy yetkisi bu rapor değildir.

## BLOCKED (sahip)

| Madde | Sahip |
|---|---|
| GPU upstream tünel 18000 + gerçek model yanıtı | KAN / SRE |
| Go→GCP start/idle-stop/watchdog | PANEL DEV + KAN |
| Accounts route implementasyonu (.7) | PANEL DEV |
| Güçlü OS sandbox | PANEL DEV |
| shop.kanus.org ücretli Go-Live | Demir |
| Cold-start/SSD kâr modelinde KAN teyidi | KAN |
| e-Fatura entegrasyon kurulumu | Ticari |
| usta-wake güvenli kapatma uygulaması | SRE + Demir |
| Hostinger VPS / Shopify plan fatura satırı | SRE / SHOP / Ticari |

## Round 7 açık kabul işleri — 06.11.2026 08:20 TSİ (Europe/Istanbul)
- BLOCKED — Yeni .7 hook + /account/ nginx prefix + US_SESSION_SECRET HOST tarafından yetkili kurulmadan .7 EXE'nin canlı login'i hazır sayılmaz; mevcut VPS sürümü sadece diğer ajan bildirimiyle Round5. Bu tur deploy yok.
- BLOCKED — Native Windows GUI/console-attach/CLI/Job/process/network kabulü Linux box'ta çalıştırılmadı. EXE_SMOKE_CHECKLIST_TR.md ve WINDOWS_SMOKE_COMMANDS_20261106_TR.md yeni .7 komutlarını içerir.
- BLOCKED — VPS-origin 127.0.0.1:18000→ds-a100 private model erişimi/tüneli yok (aktarılmış durum, uzaktan kontrol edilmedi); Go GCP start/15dk idle-stop controller entegrasyonu yok. Workers metadata15dk yalnız policy, VM kapatma kanıtı değil.
- NOT TESTED — Gerçek model çok-dosyalı onarım/rental kalite kabulü, approved Bogus Gateway account E2E, SMTP delivery ve gerçek mailbox verification/reset. Register email sahipliği kanıtı değildir; doğrulama opsiyonel, SMTP0. Reset/verify GUI kontrolleri henüz yok, endpoint/local test var.
- NOT TESTED — Bağımsız auth/penetrasyon ve dependency vulnerability audit; persistent/distributed IP limiter, per-session logout/refresh yok (reset tüm sessionVersion'ı değiştirir). Reset raw us1 anahtarını iptal etmez.
- NOT TESTED — Crash/result replay ve büyük JSON performansı; hold expiry cleanup sonraki reserve/redeem gibi yazan işlemde yapılır, salt bakiye eski reserve gösterebilir. Tek writer şart; eski R5/R6 writer yeni accounts alanlarını düşürebilir: kör rollback yasak, tutarlı snapshot/reconciliation gerekir.
- NOT TESTED — Tek model çağrısı kaynak paketler arasında saniye bölmez; aggregate seconds tek uzun çağrı garantisi değildir. Hourly seçim en yüksek kalan saniyeyi önce kullanır (eski1s paketin yeni1h paketi gölgelemesi testle düzeltildi). Sub önce seçilir; başarıda cold-start dahil upstream wall süre ceil-saniye, queue hariç; failure/cancel0debit.
- PASS — Önceki “Go account endpoint yok / local404” kod açığı yerelde kapandı; canlı gap henüz kapandı diye sunulmaz. Tüm son testler0FAIL; altı skip nedeni ROUND7_RAPOR_TR.md.

## Go-live adım listesi (KAN.US p6) — 06.11.2026 09:13 TSİ (Europe/Istanbul)

Canlı / geri alınamaz işlemler **Demir onayı** ister. Salt rapor; bu turda cutover yok.

| # | Adım | Sahip | Gerekli onaylayan | Durum | Kanıt / not |
|---|---|---|---|---|---|
| G1 | Ödeme sağlayıcı (Shopify Payments / 3P gateway, test↔canlı) | Ticari + SHOP | **Demir** | **NOT KNOWN** | `SHOP_LIVE_READINESS_20261106.md`: Payments account Access denied; Admin’de sağlayıcı listesi bu ajan tarafından doğrulanmadı. Gateway komisyonu **NOT KNOWN**. |
| G2 | Politikalar: iade / gizlilik / şartlar / KVKK | Ticari + Hukuk | **Demir** | **PARTIAL** | Canlı mağaza: yalnız Gizlilik dolu; iade/şartlar/kargo API’de yok (`SHOP_LIVE_READINESS_20261106.md`). Yerel taslaklar: `docs/legal/CAYMA_VE_IADE_POLITIKASI.md`, `KULLANIM_SARTLARI.md`, `KVKK_AYDINLATMA_METNI.md`, `MESAFELI_SATIS_SOZLESMESI.md` — yayında / Admin’e işlendi **NOT KNOWN**. |
| G3 | Webhook cutover (orders/paid|cancelled, refunds/create → api.kanus.org) | SHOP + HOST | **Demir** | **BLOCKED** (plan hazır) | Plan: `SHOP_LIVE_WEBHOOK_CUTOVER_PLAN.md`. Canlı abonelik sayısı bu bağlayıcıda 0 (`SHOP_LIVE_READINESS`). Uygulanmadı. |
| G4 | Prod secrets (`SHOPIFY_WEBHOOK_SECRET`, `US_LICENSE_SECRET`, shop domain, plans) | HOST | **Demir** | **BLOCKED** | `deploy/host-hardening/PROD_PLAN.md` §Kurulum sonucu: secret’lar **EMPTY**; unit inactive (bilinçli). |
| G5 | DNS `api.kanus.org` A → 82.29.184.225 | HOST / DNS | **Demir** (yapıldıysa teyit) | **PASS** (dosya) | Zone after: `deploy/host-hardening/dns/kanus.org.zone.after-20261106T091100+0300.json` (`api` A 82.29.184.225). İlk MCP denemesi 09:05 timeout ile durdurulmuştu; 09:11 after dosyası kayıtlı. Canlı dig bu turda **NOT TESTED**. |
| G6 | Certbot TLS + HTTPS nginx `api.kanus.org` | HOST | **Demir** | **NOT KNOWN** / son kayıt **SKIP** | `PROD_PLAN.md`: certbot DNS önkoşulu yokken SKIP. DNS after dosyasından sonra certbot başarı kanıtı bu turda **yok**. |
| G7 | Hook start (`systemctl start us-license-hook`) + /healthz | HOST | **Demir** | **BLOCKED** | Secret yok → start reddedildi; local healthz down (`PROD_PLAN.md`). |
| G8 | EXE release (prod gateway `api.kanus.org`, accounts) | PANEL DEV + Demir | **Demir** | **PARTIAL** | Native Win smoke **PASS** (.6, api-test): `WINDOWS_NATIVE_SMOKE_20261106.md`. Prod `US_BUILD_GATEWAY=https://api.kanus.org` build/yayın **NOT TESTED**. .7 accounts canlı login secret/nginx önkoşulu **BLOCKED**. |

### Go-live öncesi özet
- Ücretli sipariş / canlı webhook / secret yazma / EXE prod yayın: **Demir onayı zorunlu**.
- Maliyet senaryoları: `COST_REPORT_20261106.md`.
