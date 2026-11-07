# Maliyet — .5 / Round 6+ güncelleme

Güncelleme: 06.11.2026 08:24 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya kanıtıdır; canlı fatura kabulü değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

Sabit maliyetler: **[FIXED_COSTS.md](./FIXED_COSTS.md)** · Birim ekonomi / kâr modeli: **[UNIT_ECONOMICS.md](./UNIT_ECONOMICS.md)** · Vergi: **[TR_VERGI_EFATURA.md](./TR_VERGI_EFATURA.md)**

## Kanıtlı güncellemeler

- PASS — Hostinger Mail SMTP (VPS 465, test mail Demir adresine, revoke). Kanıt: `SMTP_HOSTINGER_20261106.md`.
- PASS — 1 istem paketi DRAFT (gid://shopify/Product/15428080632036, 07:26 TSİ; vitrinde yok, silinmedi). Kanıt: `KANUS_SHOPIFY_LIVE_AUDIT.md` §«1 istem paketinin satıştan kaldırılması»; `KANUS_FINAL_STATUS.md`.
- PASS — Accounts mode kararı: gateway **Go hook** (Cloud Run issuer yok). Kanıt: `ACCOUNTS_MODE_GAP_20261106.md` (karar §D/E). Implementasyon tamamı hâlâ BLOCKED (ayrı satır).
- PASS — Yerel yük testi Round 6 (50 lisans queue32; reserve race; hourly). Kanıt: `LOAD_TEST_20261106.txt` (mock SSE; gerçek GPU değil).
- PASS — Headless smoke 4 bayrak (version/selftest/check-license/check-server) unit + Linux version/selftest. Kanıt: `ROUND6_RAPOR_TR.md`, `SMOKE_CLI_LOCAL_20261106.txt`, `SMOKE_RUN_LOG_20261106.txt` (version+selftest), `KANUS_TEST_RESULTS.md`.
- **Native Windows 4/4 smoke:** KAN.US bildirdi — dosya kanıtı yok (`SMOKE_RUN_LOG` yalnız 2 satır Linux; `WINDOWS_SMOKE_COMMANDS` / `EXE_SMOKE_CHECKLIST` native’i BLOCKED tutuyor).

## GPU / paket

- PASS — a2-highgpu-1g europe-west4 **3,748 USD/sa** resmi Catalog SKU kırılımı (önceki tur). Kanıt: `UNIT_ECONOMICS.md` §1.
- PASS — Min satış paketi politikası **5 istem / 695 TL** (Demir Seçenek A); 1 istem DRAFT. Canlı fiyat diğer SKU’larda değiştirilmedi.

## BLOCKED (sahip)

| Madde | Sahip |
|---|---|
| VPS→127.0.0.1:18000 GPU tüneli + gerçek model | KAN / SRE |
| Go proxy → GCP start / 15 dk idle-stop / watchdog | PANEL DEV + KAN |
| Accounts Go endpoint implementasyonu (karar PASS, kod BLOCKED) | PANEL DEV |
| Cold-start / TTFT / Local SSD ölçümü (kâr modelinde boş) | KAN |
| shop.kanus.org ücretli gerçek sipariş / Go-Live | Demir |
| Shopify plan + TR gateway sözleşme rakamı | SHOP / Ticari |
| Hostinger VPS abonelik planı (MCP timeout; fatura yok) | SRE / Ticari |
| usta-wake decommission (taslak) | SRE + Demir |
| e-Fatura/e-Arşiv uygulama kurulumu | Ticari + muhasebe |

## NOT TESTED

- Canlı GCP/Hostinger/Shopify fatura CSV.
- Disk/IP/egress tam amortisman.
- PASS — Native Windows CLI smoke 4/4. Kanıt: `WINDOWS_NATIVE_SMOKE_20261106.md`. GUI NOT TESTED.

- PASS — Billing BigQuery export (standard + detailed) etkin, dataset `billing_export`, proje `project-27082a40-6083-4952-9db`. Kanıt/bildirim: **KAN.US 08:21 TSİ** (dataset hazırlığı `GCP_CLEANUP_20261106.md` §3). **İlk veri ≤24 sa bekleniyor.**
- PASS — ds-a100 beklenen Spot liste **1,848238 USD/sa** (local SSD dahil); on-demand üst sınır 3,748. Kanıt: `sre/billing_catalog_ds-a100_20261106.json`, `SRE_WATCHDOG_DESIGN_20261106.md` §4a.
