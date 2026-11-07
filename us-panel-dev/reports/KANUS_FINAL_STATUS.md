# KANUS — 2026.11.06.5 son durum

Güncelleme: 06.11.2026 07:14 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme veya adlı kanıt dosyasıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Go VPS hook tek issuer ve sunucu taraflı model proxy (Round 5 VPS’te yollar canlı). .5 EXE hedefi https://api-test.kanus.org/v1/chat/completions. Python/Cloud Run issuer yok. Kanıt: HOOK_VPS_STATUS_20261106.md Round 5; KANUS_TEST_RESULTS.md.
- PASS — SSE/JSON proxy: reserve → settlement / release; müşteri lisansı upstream’e gitmez. Yerel suite PASS.
- PASS — Go tam suite/vet/race ve çapraz derleme; artifact hash. KANUS_TEST_RESULTS.md.
- PASS — kanus-test E2E: ödeme, idempotency, geçersiz HMAC, gerçek iade/iptal teslimatı, kredi reserve/release. E2E_HOOK_KANUS_TEST_20261106.md + KANUS_TEST_RESULTS.md 06:23.
- BLOCKED — VPS → ds-a100 tünel (127.0.0.1:18000) ve Go proxy → GCP start/idle-stop/watchdog. Native Windows smoke. SMTP kapalı.
- NOT TESTED — Gerçek DeepSeek/45 dk cold-start ölçümü; shop.kanus.org; ücretli müşteri siparişi.

```text
1ef0f70429953a79de7ea95f21b028f9fb0ad22412e9b399d792bc388423f863  /workspace/us-panel/dist/US_Ticari_Aday_2026_11_06_5.exe
7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf  /workspace/us-panel/dist/us-license-hook_linux
7061b4d831e8e370a3303c89427e9049bc236c38b416ada857d588fa5cd2ccaf  /workspace/us-panel/deploy/us-license-hook/us-license-hook
0e599d2877087eb6efaf97f34fa711ad3263a18c272897d89cc52af88a6fa39a  /workspace/us-panel/deploy/us-license-hook_hostinger.tgz
```

---

## Canlı lansman Go / No-Go listesi

**Karar (06.11.2026 07:14 TSİ): NO-GO / PENDING** — Demir’in gerçek ücretli sipariş onayı ve kalan kapılar PASS olmadan canlı lansman yok.

| # | Kapı | Durum | Kanıt / not |
|---|---|---|---|
| G1 | Yerel Go suite / vet / race | PASS | KANUS_TEST_RESULTS.md |
| G2 | Windows cross-build + deploy tgz hash | PASS | BUILD_SHA256 / KANUS_TEST_RESULTS.md (native smoke değil) |
| G3 | kanus-test paid + idempotency + HMAC | PASS | E2E_HOOK_KANUS_TEST_20261106.md |
| G4 | kanus-test gerçek refund + cancel teslimatı | PASS | KANUS_TEST_RESULTS.md 06:23 |
| G5 | Round 5 sunucu proxy VPS (yollar/health/401/status) | PASS | HOOK_VPS_STATUS_20261106.md Round 5 |
| G6 | Gerçek GPU / model yanıtı (upstream 18000) | BLOCKED | HOOK_VPS_STATUS Round 5; KANUS_GPU_REPORT |
| G7 | Go → GCP start + 15 dk idle-stop + watchdog | BLOCKED | KANUS_GPU_REPORT / FINAL |
| G8 | Cold-start / TTFT / inference süre ölçümü | NOT TESTED | **KAN ölçümü bekleniyor** (KANUS_GPU_REPORT) |
| G9 | SMTP ile gerçek lisans e-postası | BLOCKED | US_SMTP_ENABLE=0 |
| G10 | Native Windows EXE smoke (Demir PC) | NOT TESTED / BLOCKED box | EXE_SMOKE_CHECKLIST_TR.md |
| G11 | Birim ekonomi / fiyat kabulü (Demir) | NOT TESTED | UNIT_ECONOMICS.md (hesap var; onay yok) |
| G12 | shop.kanus.org / ücretli gerçek sipariş | NOT TESTED | **Demir onayı beklenir** |
| G13 | Güvenlik taraması / backup-restore kabulü | NOT TESTED | KANUS_SECURITY_REPORT vb. |
| G14 | TLS / nginx prod kabul (shop + api) | PARTIAL | api-test healthz PASS; tam prod kabul NOT TESTED |

### Özet sayımlar (bu Go/No-Go tablosu)

| Etiket | Adet |
|---|---:|
| PASS | 5 |
| BLOCKED | 3 |
| NOT TESTED | 5 |
| PARTIAL | 1 |
| **Karar** | **NO-GO / PENDING** |

Go-Live için asgari: G6+G7 PASS, G10 native smoke, G12 Demir onaylı gerçek sipariş, G9 veya bilinçli SMTP istisnası, G11 fiyat onayı.

Shopify / hook / VPS’e bu turda dokunulmadı (yalnız rapor).

## Round 6 — 06.11.2026 07:20 TSİ (Europe/Istanbul)
- PASS — Yerel fake SSE HTTP yükü:50 ayrı lisans, queue32+active1;33 success/17 beklenen429, unexpected0; p50=104.970ms,p95=334.002ms. Tek kredi50 paralel reserve:1 success/49 beklenen402; negatif/double-spend yok,p50=5.084ms,p95=6.136ms. Expired/revoked401;1.1sn hourly2sn debit. Raw LOAD_TEST_20261106.txt (mock, gerçek GPU performansı değil).
- PASS — --version/--selftest/--check-license file/--check-server erken GUI'siz dispatch; log-first Windows console fallback, maskeli key, GET status/health ve redirect engeli unit testleri. Gerçek Linux version/selftest main çağrısı geçti.
- PASS — Go289+37,6 skip,0 fail; full race ve vet exit0; v3 raw ve .6 cross-build güncel. EXE SHA256 c1417c544ad2777609c47bf738ea28897bf38fa96d3747e6003e83ab1263a327.
- PASS — Hook production kodu/binary/tgz değişmedi, yeniden build/deploy yapılmadı. Lisanslı model proxy ve Go issuer mimarisi korundu.
- BLOCKED — Native Windows ve VPS model upstream127.0.0.1:18000 tunnel; gerçek GPU start/idle-stop entegrasyonu. Gerçek rental E2E env unset temiz skip; canlı model çalıştırılmadı.
- NOT TESTED — VPS Round5 deploy bilgisi aktarıldı, bu yürütücü doğrulamadı; mevcut KAN.US E2E kayıtları korunmuştur. Tam komutlar, metrikler, exit kodları ve açık testler ROUND6_RAPOR_TR.md.

### Fiyat kararı (06.11.2026 07:22 TSİ)
- Demir tek istem zarar riski için **Seçenek A (min paket 5 istem, 695 TL)** seçti. Canlıya uygulanmadı: Shopify 1-istem varyantının kaldırılması/gizlenmesi + `US_CREDIT_PRODUCTS` güncellemesi bekliyor → Go/No-Go maddesi **PENDING**.
- **07:26 TSİ:** 1 istem ürünü canlıda DRAFT (gid://shopify/Product/15428080632036, vitrinde görünmüyor, silinmedi). Fiyat kararı maddesi → **PASS (SHOP raporu)**. Allowlist bilinçli olarak korunuyor.
