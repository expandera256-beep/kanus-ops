# Round 8 — §2 yerel kodlama raporu (06.11.2026, TSİ)

Kapsam: yalnız US box'ta yerel kod + test. Canlı sistem, deploy, servis yeniden başlatma, GPU, Shopify, VPS: **yok**. licensesrv, gateway/account protokolleri, gömülü anahtarlar: **dokunulmadı**.

## Sonuç
| Madde | Durum | Yama |
|---|---|---|
| 1 Ortak güvenli çalıştırma (Terminal+Run) | PASS (Windows Job Object canlı: NOT TESTED) | /workspace/us-panel-dev/patches/s2_item1_safe_exec.patch |
| 2 Kök içi atomik, sınırlı dosya araçları | PASS | /workspace/us-panel-dev/patches/s2_item2_file_tools.patch |
| 3 Bütçeli build→hata→düzeltme→retest (httptest mock) | PASS (gerçek model/GUI bağlantısı: NOT TESTED) | /workspace/us-panel-dev/patches/s2_item3_repair_loop.patch |
| 4 Kalıcı atomik görev durumu + çökme sonrası devam | PASS | /workspace/us-panel-dev/patches/s2_item4_task_state.patch |
Birleşik yama: /workspace/us-panel-dev/patches/s2_all_items_combined.patch. Ayrıntı: /workspace/us-panel-dev/PANEL_DEV_CHANGES.md.

## Testler (ana kaynak, birleştirme sonrası)
- go vet ./...: PASS · GOOS=windows go vet: PASS · GOOS=windows go build: PASS
- go test -race -count=1 ./...: PASS — 361 PASS / 0 FAIL / 7 SKIP (alt testler dahil; baseline 338 PASS). Yeni üst düzey test: 20 (+1 yalnız alt süreç olarak çalışan çökme yardımcısı, normalde SKIP).
- SKIP edilenler: gerçek model/kapı/REA/PHP testleri ve alt-süreç yardımcıları (opt-in). Gerçek model, Windows çalışma, canlı entegrasyon: NOT TESTED.

## EXE .8
- /workspace/us-panel/dist/US_Ticari_Aday_2026_11_06_8.exe — sürüm 2026.11.06.8
- sha256: e7e54a42d135abde6c817613ff7a8486eae76b2f0a7b23a86df39cc3cd11b26b
- .7 ile aynı yöntem: CGO_ENABLED=0 GOOS=windows GOARCH=amd64, -trimpath -H=windowsgui, releaseChannel=commercial, releaseAuthority=accounts, gateway/license URL https://api-test.kanus.org, embeddedKey/embeddedLicenseSecret boş.
- Windows'ta çalıştırma/smoke: NOT TESTED.

## Açık notlar
- Linux'ta RLIMIT_NPROC bilerek yok (UID bazlı; thread'li sunucuları bozuyor). Çok kullanıcılı izolasyon cgroup/container ister — bu bir sandbox değildir.
- Run yolu program ömrü 4 saat ile sınırlı; Terminal 1–600 sn.
- go.mod: x/sys satırındaki `// indirect` yorumu kalktı (sürüm aynı).
