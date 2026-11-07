# KANUS Global Readiness

<!-- BEGIN i18n (GLOBAL) -->
## i18n (§7.1–7.2) — güncelleme 08.11.2026 ~00:45 TSİ (taban pinned/v10, yayın yok)

Ayrıntı: /workspace/us-panel-dev/global-work/GLOBAL_I18N_REPORT.md · Site notu: /workspace/us-panel-dev/global-work/SITE_STATIC_VS_PRERENDER.md · Yayın paketi taslağı: /workspace/us-panel-dev/global-work/site/release/


> **DÜZELTME NOTU (Tur 4):** Önceki "r10_base" tabanı sabit .10 manifestiyle eş değil (33/181 fark); taban artık `pinned/v10`. 02: gui.go **+2 satır/2 hunk**, **37 data-i18n** + 3 placeholder (önce "1 satır", "36" yazılmıştı). Python: Ran 119 = **78 PASS + 41 SKIP** (US_GO_LOCALES ile 79 + 40), "119 PASS" değil. `licensesrv TestParsePlansBothExportSchemas` FAIL'i kod hatası değil, fixture yolu; fixture'lı koşuda PASS.
> PANEL DEV G-01…G-04 (03) kapatıldı: 8 regresyon testi önce FAIL, sonra PASS. Yeni 03 sha256 `171bf6fc4a45e72fea90ce73ef680d957abf21c3d3af083ae60f9679e091a41a`; 04 yalnız bağlamdan yeniden üretildi; core/02/05 değişmedi.

> **Tur 5 (PANEL DEV R2):** symlink'li HOME/config desteklendi (kök bir kez çözülüp fd/handle ile sabit), kök açılamazsa bellek + 503, GUI başarısız kayıtta geri alır + TR/EN mesaj; Windows handle tabanlı (NtCreateFile/RootDirectory, FILE_OPEN_REPARSE_POINT, GetFileInformationByHandle, FileRenameInformation). 03 sha256 `425a76afd9e07b9bdc711b5496a1539046d975e63b54380d6e4f503f2c2c40a0`, 04 (yalnız bağlam) `1c937497ee30ef7de86245b2fa4e2b6668b8660e2db62017bd7a7c2db5864991`.

> **Tur 7 (gerçek Windows, 08.11):** DESKTOP-48D4KOL, Win10 19045, race'siz binary 548944ea: **3 PASS / 1 SKIP / 1 FAIL** (TestWinPref 3 PASS + SymlinkFileRejected SKIP; FAIL TestPrefStoreRoundTrip 'perm -rw-rw-rw-'). Kök neden: Go'da Windows FileMode yalnız read-only bitini yansıtır; test Unix varsayımıydı. Güvenlik: eski kod lang.json'u devralınan ACL ile oluşturuyordu (profil dışı paylaşımlı kökte Users/Everyone yazabilirdi). Düzeltme: temp dosya ve oluşturulan alt dizinler NtCreateFile OBJECT_ATTRIBUTES ile korumalı DACL alır (sahip=geçerli kullanıcı; yalnız kullanıcı+SYSTEM+Administrators FA; devralım kapalı); okuma NumberOfLinks≠1 dosyayı reddeder. Test platforma ayrıldı (Unix 0600 / Windows DACL). Yeni 03 `ad63ce899ab221444ccc4b9228374f6cb09d5134f8d028f778fa4e5d67af0ce7`; 04 değişmedi. Yeni binary `ci/i18n_windows_amd64.test.exe` sha256 `ca7a79d56865eb24adfb338d86770c30cfc6e5f813aef599905f7c5d7e4e9a46` (6 TestWinPref) — Windows'ta yeniden koşum bekliyor.

> **Tur 7b (R3-01):** kök handle önbelleği sınırlı (maxRoots=1, refcount ile güvenli kapatma, Unix/Windows aynı); 64 kök sonrası fd 7→71 idi, şimdi sabit. 03 `f66d4416…`, 04 değişmedi; Windows binary `d866b9aa…` (NOT TESTED on Windows).

> **Tur 8 (§9):** kanus.org/us sözlükleri 4 dilde 126/126 tam; eksikler tarih (ham ISO), gönderim hatalarında mesaj yok, `<head>` yalnız EN, fiyat gösterimi yok. Yerel paket `site/release_us/` (YÜKLENMEDİ) tarih + hata mesajını 4 dilde düzeltir (Chrome testi 12/12; canlı kopya 4/12). Panel DE/RU (yama 06 `83aaca5d…`), e-posta DE/RU (yama 07 `70acc9b0…`). Race 420/0/7, S2 36/36.

> **Tur 9:** Windows r3 (d866b9aa) 18 PASS / 1 FAIL / 1 SKIP; FAIL OS-genel handle sayısının sıfır toleranslı karşılaştırması (ölçüm gürültüsü, kod sızıntısı bulunmadı). Test dahili kök sayaçları + 64/512 ölçekleme + kapalı-handle geçersizliği ile yeniden yazıldı (mutasyonla doğrulandı). 03 `ba8845c5…`; binary `166d869a…` (06 dahil); betik exit kodları -File/-Command'da 0/1/2/3. Race 422/0/7, S2 36/36. **Windows r4 gerçek: 23 PASS / 0 FAIL / 1 SKIP (symlink yetkisi yok, belgeli), exit 2 PARTIAL (KAN.US beyanı, logda yok); HandleScaling 64:+6 / 512:+2 (log).** Binary 166d869a yeniden derlemede birebir aynı (`ci/BUILD_INFO_166d869a.md`); betik `32c3034e…` artık result.txt yazar.

Yamalar (sırayla): global_i18n_panel_core → _02_gui_strings → _03_pref → _04_errors (Go, r10_base) ; _05_email_loader (kanus/ Python). Hook ve ledger/ödeme dosyalarına dokunulmadı.

| Madde | Durum | Test | Sonuç |
|---|---|---|---|
| Dil dosyaları TR/EN/DE/RU (panel 93 anahtar parite) | IMPLEMENTED (DE/RU anadil kontrolü NOT TESTED) | i18n_test (anahtar eşitliği, placeholder, yasak slogan) | PASS |
| Dinamik dil değiştirme (36 statik metin, lang/dir) | IMPLEMENTED | gui_i18n_test, check_ui.mjs | PASS |
| Kullanıcı dil tercihi (yerel, lang.json; nofollow/regular/4 KB, dirfd yazım, Origin, MaxBytes) | IMPLEMENTED | pref_test, pref_security_unix_test, gui_lang_test, gui_lang_security_test | PASS |
| Tarih/saat, para biçimi TR/EN/DE/RU (ödeme TRY-only; kur kaynağı yok, döviz gösterimi yok) | IMPLEMENTED | TestFormatting, TestFormattingDeRu | PASS |
| Çok dilli hata mesajları | PARTIAL (gui.go 28 metin) | gui_errors_i18n_test (TR byte-eşit) | PASS |
| Çok dilli e-posta | PARTIAL (yükleyici; konu TR "US Panel lisans anahtarın", EN "Your US Panel license key" kabul; gönderime bağlı değil) | test_i18n_mail (10); discover 78+41 skip (US_GO_LOCALES ile 79+40) | PASS |
| Yardım merkezi | PARTIAL | — | NOT TESTED |
| DE/FR/ES/AR + RTL genişletilebilirlik | IMPLEMENTED | Fallback/RTL testleri (Go+Py) | PASS |
| Yamalar pinned/v10'a sıralı uygulanır | — | git apply --check+apply 1→4; 05 on work2 | PASS |
| go build / go vet ./... / GOOS=windows vet+build / gofmt -l | — | | PASS |
| go test -race -p=1 ./... (fixture'lı) | — | 415/0/7 (taban 386/0/7), race 0; S2 36/36 | PASS |
| Windows i18n testleri (gerçek Win10 19045, binary 166d869a, race'siz) | — | 23 PASS / 0 FAIL / 1 SKIP (SymlinkFileRejected, yetki yok) | PASS |
| Windows DACL koruması + hardlink senaryosu | IMPLEMENTED | TestWinPrefDACLSharedRoot, TestWinPrefHardlinkAliasRejected, checkPrivateFile (DACL) | PASS (Win10 19045, r4) |
| Yığın .11→UYUM→API→GLOBAL | — | check+apply 4/4, build+vet | PASS (birleşik suite NOT TESTED) |
| GUI kayıt hatası geri alma | IMPLEMENTED | gui_lang_ui_test.mjs 4/4 | PASS |
| /tr, /en + 7 sayfa (yerel paket, 16 dosya) | IMPLEMENTED; hukuki sayfalar Demir kararıyla paket dışı → MISSING | validate.py 14/14, sha256sum -c 16/16 | PASS |
| Hukuki sayfa yok / iç link yok / TASLAK yok | — | validate.py + negatif test | PASS |
| Slogan tek anahtar + tek komut yeniden üretim | IMPLEMENTED (metin Demir'de) | make_release.py test sloganı | PASS |
| VERIFY_AFTER_UPLOAD (yerel .htaccess benzetimi) | — | 41/41 | PASS |
| hreflang/canonical/sitemap/robots/OG/JSON-LD (yerel) | IMPLEMENTED | validate.py + negatif test | PASS |
| Canlı site çok dil | PARTIAL: sözlük 4 dil tam; tarih ham ISO, form hataları sessiz, head yalnız EN; URL/SEO yok | curl + Chrome baseline 4/12 | BROKEN (canlı) |
| kanus.org/us düzeltme paketi (release_us) | IMPLEMENTED, YÜKLENMEDİ | Chrome 12/12, validate_us 21/21, VERIFY sim 13/13 | PASS (yerel) |
| Canlı yayın | BLOCKED (yüklemeyi KAN.US yapacak; slogan bekleniyor) | CONFLICT_CHECK 16/16 fallback | NOT TESTED |
| W3C/Lighthouse/gerçek mobil, Windows EXE | — | | NOT TESTED |
<!-- END i18n (GLOBAL) -->
