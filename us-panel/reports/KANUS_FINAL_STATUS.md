# KANUS — 2026.11.06.5 son durum

Güncelleme: 06.11.2026 08:24 TSİ (Europe/Istanbul)

PASS yalnız yerel test/dosya/derleme veya adlı kanıt dosyasıdır; canlı kabul değildir. Etiketler: PASS, FAIL, BLOCKED, NOT TESTED.

- PASS — Go VPS hook tek issuer ve sunucu taraflı model proxy (Round 5 VPS’te yollar canlı). .5 EXE hedefi https://api-test.kanus.org/v1/chat/completions. Python/Cloud Run issuer yok. Kanıt: HOOK_VPS_STATUS_20261106.md Round 5; KANUS_TEST_RESULTS.md.
- PASS — SSE/JSON proxy: reserve → settlement / release; müşteri lisansı upstream’e gitmez. Yerel suite PASS.
- PASS — Go tam suite/vet/race ve çapraz derleme; artifact hash. KANUS_TEST_RESULTS.md.
- PASS — kanus-test E2E: ödeme, idempotency, geçersiz HMAC, gerçek iade/iptal teslimatı, kredi reserve/release. E2E_HOOK_KANUS_TEST_20261106.md + KANUS_TEST_RESULTS.md 06:23.
- BLOCKED — VPS → ds-a100 tünel (127.0.0.1:18000) ve Go proxy → GCP start/idle-stop/watchdog.
- PASS — Native Windows CLI 4/4; SMTP Hostinger test PASS (ayrı kanıt dosyaları).
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
| G8 | Cold-start süreleri (~28 dk / ~3:52 SSD) | PARTIAL | `KANUS_GPU_REPORT.md` SRE; **KAN teyidi bekleniyor** — UNIT_ECONOMICS §10 |
| G9 | SMTP Hostinger test mail | PASS | `SMTP_HOSTINGER_20261106.md` (müşteri maili ayrı) |
| G10 | Native Windows **CLI** smoke 4/4 | PASS | `WINDOWS_NATIVE_SMOKE_20261106.md` |
| G10b | Windows **GUI** | NOT TESTED | aynı dosya (kapsam dışı) |
| G11 | Birim ekonomi / fiyat kabulü (Demir) | NOT TESTED | UNIT_ECONOMICS.md (hesap var; onay yok) |
| G12 | shop.kanus.org / ücretli gerçek sipariş | NOT TESTED | **Demir onayı beklenir** |
| G13 | Güvenlik taraması / backup-restore kabulü | NOT TESTED | KANUS_SECURITY_REPORT vb. |
| G14 | TLS / nginx prod kabul (shop + api) | PARTIAL | api-test healthz PASS; tam prod kabul NOT TESTED |
| G15 | Billing BigQuery export (standard+detailed) | PASS | **KAN.US 08:21 TSİ**; dataset `billing_export` (`GCP_CLEANUP` §3). İlk veri ≤24 sa |

### Özet sayımlar (bu Go/No-Go tablosu)

| Etiket | Adet |
|---|---:|
| PASS | 7+ (G9 SMTP, G10 CLI eklendi) |
| BLOCKED | 2 (G6, G7) |
| NOT TESTED | 4+ (G10b GUI, G11–G13…) |
| PARTIAL | 2 (G8, G14) |
| **Karar** | **NO-GO / PENDING** |

Go-Live için asgari: G6+G7 PASS, G12 Demir onaylı gerçek sipariş, G11 fiyat onayı; GUI smoke isteğe bağlı.

Shopify / hook / VPS’e bu turda dokunulmadı (yalnız rapor).

## Round 6 — 06.11.2026 07:20 TSİ (Europe/Istanbul)
- PASS — Yerel fake SSE HTTP yükü:50 ayrı lisans, queue32+active1;33 success/17 beklenen429, unexpected0; p50=104.970ms,p95=334.002ms. Tek kredi50 paralel reserve:1 success/49 beklenen402; negatif/double-spend yok,p50=5.084ms,p95=6.136ms. Expired/revoked401;1.1sn hourly2sn debit. Raw LOAD_TEST_20261106.txt (mock, gerçek GPU performansı değil).
- PASS — --version/--selftest/--check-license file/--check-server erken GUI'siz dispatch; log-first Windows console fallback, maskeli key, GET status/health ve redirect engeli unit testleri. Gerçek Linux version/selftest main çağrısı geçti.
- PASS — Go289+37,6 skip,0 fail; full race ve vet exit0; v3 raw ve .6 cross-build güncel. EXE SHA256 c1417c544ad2777609c47bf738ea28897bf38fa96d3747e6003e83ab1263a327.
- PASS — Hook production kodu/binary/tgz değişmedi, yeniden build/deploy yapılmadı. Lisanslı model proxy ve Go issuer mimarisi korundu.
- PASS — Native Windows CLI smoke 4/4 (`WINDOWS_NATIVE_SMOKE_20261106.md`). GUI NOT TESTED.
- BLOCKED — VPS model upstream 127.0.0.1:18000 tunnel; Go→GCP start/idle-stop. Gerçek rental E2E canlı model çalıştırılmadı.
- NOT TESTED — VPS Round5 deploy bilgisi aktarıldı, bu yürütücü doğrulamadı; mevcut KAN.US E2E kayıtları korunmuştur. Tam komutlar, metrikler, exit kodları ve açık testler ROUND6_RAPOR_TR.md.

### Fiyat kararı (06.11.2026 07:22 TSİ)
- Demir tek istem zarar riski için **Seçenek A (min paket 5 istem, 695 TL)** seçti. Canlıya uygulanmadı: Shopify 1-istem varyantının kaldırılması/gizlenmesi + `US_CREDIT_PRODUCTS` güncellemesi bekliyor → Go/No-Go maddesi **PENDING**.
- **07:26 TSİ:** 1 istem ürünü canlıda DRAFT (gid://shopify/Product/15428080632036, vitrinde görünmüyor, silinmedi). Fiyat kararı maddesi → **PASS (SHOP raporu)**. Allowlist bilinçli olarak korunuyor.

## Round 7 — 06.11.2026 08:20 TSİ (Europe/Istanbul)
- PASS — Go hesap modu yerelde tamamlandı: register/login/balance/redeem/requests/workers ve verification/reset API'leri; bcrypt12, HMAC-expiry session, IP limit, aynı locked atomic JSON. us1 tek hesaba bağlanır; aynı hesap tekrar200, başka hesap409. Kredi kopyası/ikinci ledger YOK; raw key ve session aynı license/reservation/settlement'ı tüketir.
- PASS — Yerel fake SSE E2E register→login→redeem→balance→chat reserve/settle; hata/cancel release, hourly ceil saniye, karma session/raw son-kredi yarışı, refund ile bakiye azalması, eski JSON/restart testleri. SMTP allowlist fake sender testleri; gerçek SMTP kullanılmadı.
- PASS — Go 304 üst + 37 alt PASS, 6 beklenen skip,0 FAIL; full race/vet exit0. Ham çıktılar go_test/go_race/go_vet_topyekun_20261106_v4.txt. Linux gerçek headless version/selftest exit0; native Windows değildir.
- PASS — §8 tekrar (LOAD_TEST_20261106.txt son ROUND7):50lisans33HTTP200+17beklenen429; p50=102.374ms,p95=366.031ms; tek-kredi50reserve1HTTP200+49beklenen402,p50=3.491ms,p95=4.458ms; negatif/double-spend0,unexpected0. Expired/revoked401;1.1sn hourly→2sn. Bunlar localhost/fake SSE, GPU kapasite ölçümü değil.
- PASS — .7 accounts EXE ve yeni hook/tgz yerelde üretildi; PE amd64 GUI,9 arşiv dosyası kaynakla byte-eşit, binary kopyaları eşit, webhook secret template placeholder. Hashler BUILD_SHA256_20261106_v4.txt.
- BLOCKED — Native Windows, canlı accounts route kabulü (HOST yetkili deploy yapmadı), VPS18000 model tüneli, gerçek DeepSeek/proje kabulü ve Go GPU start/15dk idle-stop entegrasyonu.
- NOT TESTED — Gerçek SMTP/posta kutusu sahipliği, auth bağımsız güvenlik/vuln audit, Windows AV/code-sign. Önceki KAN.US/HOST canlı E2E girdileri korunmuştur; bu yürütücü canlı doğrulamadı. Ayrıntı ROUND7_RAPOR_TR.md.

## Maliyet / Spot (06.11.2026 08:24)
- PASS — ds-a100 Spot **1,848238 USD/sa** (Catalog JSON + §4a doğrulandı). On-demand 3,748 üst sınır.
- PASS — Billing export etkin (**KAN.US 08:21 TSİ**); ilk BigQuery satırları ≤24 sa.
