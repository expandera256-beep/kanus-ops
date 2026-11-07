# GLOBAL — §7.1 / §7.2 i18n ayrıntılı rapor (06.11.2026, TSİ)

Esas ağaç: `/workspace/us-panel-dev/work2/kiyas/term` (Go modülü `kiyas`, go 1.22). Kaynak ağaçlara dokunulmadı; iş `global-work/panel` (kopya) ve `global-work/site` içinde. Hiçbir şey yayınlanmadı. Marka: Karar A, metinler `docs/BRAND_A_COPY_20261106.md`; "Yerli Teknoloji, Küresel Vizyon / Built in Türkiye. Designed for the World." hiçbir yerde kullanılmadı (testle denetleniyor).

## Önceki durum — kanıt
- Panel: i18n/locale altyapısı yok (`rg -i "i18n|locale|Accept-Language"` *.go → eşleşme yok). ~1455 Türkçe sabit string *.go içinde, `gui.html` `<html lang="tr">` ve 122 TR satır; hata mesajları sabit TR (ör. `continuation.go:93`, `source_range.go:124`). Tarih: `license.go:250` sabit `2006-01-02` UTC. Para: yalnız TRY (`kanus/shopify_ledger.py:30` currency!='TRY' reddi).
- E-posta: `kanus/mail_worker.py`, `delivery.py` — tek dilli (TR) şablon varsayımı; çok dilli şablon yok.
- Canlı site (curl, salt okuma): `/`, `/tr`, `/en` aynı 1600 baytlık SPA kabuğu, `<html lang="en">`, title `KAN.US) — Space for your next idea`, canonical hep `https://kanus.org/`, hreflang yok, JSON-LD yok, OG var (og:url hep `/`). JS içinde dil seçici (`allowed={tr,de,ru,en}`). `/sitemap.xml` ve `/robots.txt` → HTTP 200 ama `text/html` SPA kabuğu (gerçek sitemap/robots yok). Site kaynağı yerelde bulunamadı (yalnız `raw.html` kopyası).

## Sınıflandırma
| Madde | Önce | Şimdi (yerel) | Kanıt / not |
|---|---|---|---|
| 7.1 TR/EN dil dosyaları | MISSING | IMPLEMENTED (paket+test; UI'ya bağlanmadı) | `i18n/locales/{tr,en}.json` |
| Dinamik dil değiştirme | MISSING | PARTIAL | `Negotiate`/`For` var; GUI bağlanması yok |
| Tarih/saat biçimi | MISSING (sabit) | IMPLEMENTED | `Date/DateTime` testli |
| Para birimi biçimi | MISSING (yalnız TRY) | IMPLEMENTED (biçim) | TRY/USD/EUR/GBP/AED/SAR/INR; ödeme tarafı değişmedi |
| Çok dilli hata mesajları | MISSING | PARTIAL | `err.*` anahtarları; mevcut 1455 sabit string taşınmadı |
| Çok dilli e-posta | MISSING | PARTIAL | `email.license.*` anahtarları; Python mail_worker bağlı değil |
| Çok dilli yardım merkezi | MISSING | PARTIAL | `help.*` anahtarları + site docs/support iskeleti |
| Kullanıcı dil tercihi | MISSING | PARTIAL | tercih > Accept-Language > tr; kalıcı saklama yok |
| DE/FR/ES/AR genişletilebilirlik, RTL | MISSING | IMPLEMENTED | yeni json yeter; `_meta.dir=rtl` testli |
| 7.2 /tr, /en | BROKEN (aynı kabuk) | IMPLEMENTED yerel | `site/out/{tr,en}/` |
| 10 sayfa (ana, ürün, fiyat, indir, docs, API, destek, gizlilik, koşullar, iade) | MISSING | PARTIAL (iskelet; hukuki metinler TASLAK) | 20 sayfa |
| hreflang | MISSING | IMPLEMENTED yerel, çift yönlü + x-default | validate.py |
| canonical | BROKEN (hep /) | IMPLEMENTED yerel | |
| sitemap | BROKEN (HTML döner) | IMPLEMENTED yerel | `out/sitemap.xml` xhtml:link |
| Open Graph | PARTIAL | IMPLEMENTED yerel (og:locale+alternate) | |
| schema.org | MISSING | IMPLEMENTED yerel | WebPage/Organization/SoftwareApplication |
| Mobil uyum | PARTIAL (viewport var) | IMPLEMENTED (viewport+akışkan CSS); cihaz testi yok | NOT TESTED gerçek cihaz |
| Performans | UNKNOWN | IMPLEMENTED (statik, ~2.7KB/sayfa, harici font/JS yok); Lighthouse yok | NOT TESTED |
| Canlı yayın | — | BLOCKED (onay) | §14 |

## Değişiklikler / dosyalar
- Yama: `/workspace/us-panel-dev/patches/global_i18n_panel_core.patch` — yalnız yeni dosyalar: `i18n/i18n.go`, `i18n/i18n_test.go`, `i18n/locales/tr.json`, `i18n/locales/en.json`. Mevcut hiçbir dosyaya dokunulmadı → r9/.9 yamalarıyla çakışma yok.
- Slogan tek anahtar: `brand.slogan` (panel), `slogan` (site `content.json`); değer BRAND_A: "Fikirlerini yazılıma dönüştür." / "Turn your ideas into software." — Demir onayı bekliyor.
- Site: `global-work/site/{build.py,content.json,validate.py,out/}`.

## Testler (gerçek)
| Test | Komut | Sonuç |
|---|---|---|
| i18n unit + race (9 test: eksik anahtar, placeholder eşleşmesi, yasak slogan, negotiate, fallback, RTL, hatalı locale, biçim, eşzamanlılık) | `go test -race -count=1 ./i18n/` | PASS |
| go vet tüm modül | `go vet ./...` | PASS |
| go build | `go build ./...` | PASS |
| go test tüm modül | `go test -count=1 ./...` | FAIL — yalnız `kiyas/licensesrv TestParsePlansBothExportSchemas`; dokunulmamış `base` kopyada da aynı FAIL → önceden var, i18n kaynaklı değil |
| Yama uygulanabilirlik | `git apply --check` on `us-panel-dev/work2/kiyas/term` ve r9 uygulanmış `us-panel/work2/kiyas/term` | PASS (ikisi) |
| HTML yapı, canonical, hreflang çift yönlü, OG, JSON-LD parse, sitemap XML↔sayfa, yasak ifade | `python3 site/validate.py` | PASS (20 sayfa / 20 URL) |
| Doğrulayıcı negatif kontrol (bozuk hreflang) | elle bozup çalıştırma | FAIL beklendiği gibi yakalandı |
| W3C/vnu HTML doğrulama, Lighthouse, gerçek mobil | — | NOT TESTED (araç yok) |
| Windows EXE | — | NOT TESTED |

## Kalan riskler / sonraki adım
1. Paket henüz UI'ya bağlı değil; `gui.html`/Go sabit stringlerinin anahtarlara taşınması US'in .9 işi bittikten sonra küçük parçalarla yapılmalı.
2. Kullanıcı dil tercihinin kalıcı saklanması (yerel config veya hesap) ve e-posta (Python) için eş JSON yükleyici.
3. Hukuki sayfalar TASLAK; içerik hukuki onay ister.
4. Canlı site kaynağı/barındırma bilinmiyor (SPA; sitemap/robots SPA fallback'e düşüyor).

---
# TUR 2 (06.11.2026 ~09:45 TSİ) — taban r10_base (.10)

Çalışma kopyaları: `global-work/r10dev` (Go, git aşamalı), `global-work/kanusdev` (Python). r10_base ve US ağaçlarına yazılmadı; hook dosyalarına (cmd/us-license-hook, licensesrv) dokunulmadı. Ledger/ödeme kodu değişmedi; TRY dışı para birimleri ödeme tarafında kapalı.

## Yamalar (sıralı)
| # | Dosya | Kapsam | Dokunduğu mevcut dosyalar |
|---|---|---|---|
| 1 | `patches/global_i18n_panel_core.patch` | i18n paketi, tr/en, biçimleme | yok (yeni dosyalar) |
| 2 | `patches/global_i18n_02_gui_strings.patch` | 36 metin + 3 placeholder `data-i18n`, dil seçici, `/api/i18n`, dinamik değiştirme, `html lang/dir` | gui.html, gui.go (+1 satır, import) |
| 3 | `patches/global_i18n_03_pref.patch` | kalıcı dil tercihi `lang.json` (desk.json yanında, US_DESK'e uyar), `/api/lang` GET/POST (yalnız JSON, 0600, atomik yazma) | gui.go (+1 satır), gui.html (change handler) |
| 4 | `patches/global_i18n_04_errors.patch` | gui.go'daki 28 benzersiz / 34 çağrı kullanıcı hata metni → `err.gui.*`; TR birebir aynı | gui.go |
| 5 | `patches/global_i18n_05_email_loader.patch` (kanus/ ağacı) | Python `i18n_mail.py` + `kanus/locales/`; aynı JSON biçimi, fallback, placeholder; gönderim YOK, delivery.py bağlanmadı | yok |

## Testler (gerçek)
| Test | Komut | Sonuç |
|---|---|---|
| Sıralı uygulama 1→4 r10_base (güncel) | `git apply --check && git apply` sırayla | PASS (4/4) |
| 05 uygulanabilirlik | `git apply --check` on `us-panel-dev/work2` | PASS |
| Aşama başı bağımsız (1, 1+2, 1+2+3) | `go vet . ./i18n && go test . ./i18n` + `node check_ui.mjs` | PASS (3/3) |
| Tüm yamalarla build | `go build ./...` | PASS |
| vet | `go vet ./...` | PASS |
| Tüm testler | `go test -count=1 ./...` | FAIL yalnız `licensesrv TestParsePlansBothExportSchemas`; yamasız güncel r10_base'de de aynı FAIL → önceden var (hook/US alanı) |
| Race | `go test -race . ./i18n` | PASS |
| UI JS sözdizimi/DOM bağları | `node check_ui.mjs` | PASS |
| TR koruma: markup = tr.json (≥30 öğe), 28 hata metni byte-eşit, varsayılan hata "yöntem yok\n" | gui_i18n_test / gui_errors_i18n_test | PASS |
| TR/EN anahtar eşitliği + placeholder + yasak slogan | i18n_test, test_i18n_mail | PASS |
| Python e-posta yükleyici (8 test, Go locales ile senkron dahil) | `US_GO_LOCALES=... python3 -m unittest test_i18n_mail` | PASS |
| kanus Python regresyon | `python3 -m unittest discover` (kanusdev) | PASS (117, 41 skip — önceden var olan ortam atlamaları) |
| Site validate (+ kanus.lang yazımı) | `site/validate.py` | PASS 20/20 |
| Yayın paketi yerel doğrulama | `VERIFY_AFTER_UPLOAD.sh http://127.0.0.1:8799` | 25 PASS; 2 FAIL beklenen (yerelde SPA yok: `/`, SPA fallback) |
| Canlı çakışma kontrolü | curl 22 yol | 22/22 SPA_FALLBACK → üzerine yazma yok |

## Güncel sınıflandırma (yalnız değişenler)
- Dinamik dil değiştirme: PARTIAL → IMPLEMENTED (36 statik metin; JS'in dinamik ürettiği metinler TR kalır).
- Kullanıcı dil tercihi: PARTIAL → IMPLEMENTED (yerel panel; hesap düzeyinde senkron yok).
- Çok dilli hata mesajları: PARTIAL (gui.go HTTP hataları tamam; diğer *.go içindeki ~1400 TR metin ve JS mesajları bekliyor).
- Çok dilli e-posta: PARTIAL (yükleyici hazır+testli; delivery.py/mail_worker bağlantısı ve konu metni farkı — mevcut "US lisans anahtarınız" ≠ BRAND_A "US Panel lisans anahtarın" — ORG/Demir kararı).
- Yardım merkezi: PARTIAL (anahtarlar + site docs/support iskeleti).
- §7.2 çok dilli site (canlı): PARTIAL — istemci tarafı tr/de/ru/en var (localStorage `kanus.lang`, varsayılan en), URL bazlı /tr–/en, hreflang, gerçek sitemap/robots MISSING/BROKEN; yerel paket hazır, yayın BLOCKED (onay).

## Ek çıktılar
- `global-work/SITE_STATIC_VS_PRERENDER.md`
- `global-work/site/release/` (FILES.txt, SHA256SUMS.txt, CONFLICT_CHECK.md, VERIFY_AFTER_UPLOAD.sh, README_RELEASE.md, public_html/)
- Loglar: `global-work/r10_patched_tests.log`, `global-work/r10_base_tests.log`

---
# TUR 3 (06.11.2026 ~09:55 TSİ) — yerel, yükleme yok

## Site paketi
- Demir kararı (09:42 TSİ): hukuki sayfalar (privacy/terms/refund, tr+en) paketten tamamen çıkarıldı; "yakında" sayfası da yok. Taslaklar yalnız `site/content.json` içinde `"draft": true` + "TASLAK — YAYINLANMAZ" olarak durur, build'e girmez.
- Menü/footer'da bu yollara link yok; hreflang ve sitemap'ten çıkarıldı. docs/api metinlerindeki "Taslak iskelet" ifadesi "hazırlanıyor" ile değiştirildi.
- Slogan tek anahtar: `content.json` → `"slogan": {"tr","en"}`. Tek komut: `python3 make_release.py [--slogan-tr ..] [--slogan-en ..]` → build → validate → public_html → FILES.txt → SHA256SUMS.txt → `sha256sum -c`. Yansıdığı alanlar (yalnız /tr/ ve /en/ ana sayfaları): `<title>`, `og:title`, JSON-LD `WebPage.name` + `WebPage.slogan` + `Organization.slogan`, `<h1>`.
- Paket: 16 dosya (tr 7 + en 7 sayfa, robots.txt, sitemap.xml 14 URL). Önceki: 22.
- validate.py yeni kontroller: pakette olmayan sayfaya iç link yok; hukuki dizin yok; hreflang hukuki yola gitmiyor; çıktıda TASLAK/DRAFT/hukuki taslak işareti yok; sitemap = indekslenebilir sayfalar.

## E-posta
- Konu (karar): TR "US Panel lisans anahtarın"; EN "Your US Panel license key" (BRAND_A'da EN konu yok; adlandırma/ton BRAND_A'ya göre — onay notu). Testle sabitlendi; delivery.py bağlı değil (testle denetleniyor). Yama: `patches/global_i18n_05_email_loader.patch` yeniden üretildi.

## Testler (gerçek)
| Test | Sonuç |
|---|---|
| `python3 validate.py` (14 sayfa / 14 sitemap URL) | PASS |
| validate negatif: hukuki yola link + TASLAK + hukuki dizin enjekte | FAIL beklendiği gibi yakalandı (exit 1) |
| `python3 make_release.py` (build→validate→FILES→SHA256SUMS→sha256sum -c) | PASS, 16 dosya |
| `cd release/public_html && sha256sum -c ../SHA256SUMS.txt` | PASS 16/16 OK |
| Slogan tek anahtar (geçici kopyada test sloganı) | PASS: TR/EN 6'şar yansıma, eski slogan 0 |
| CONFLICT_CHECK (curl, canlı) | 16/16 paket yolu SPA_FALLBACK; 6 hukuki yol SPA_FALLBACK; FARKLI 0 |
| VERIFY_AFTER_UPLOAD.sh, .htaccess benzetimli yerel sunucu | PASS 41/41 |
| VERIFY_AFTER_UPLOAD.sh canlıya (yükleme öncesi) | beklendiği gibi dosya kontrolleri FAIL, hukuki/SPA kontrolleri PASS |
| `test_i18n_mail` (10 test, Go locales senkron dahil) | PASS |
| kanus Python `unittest discover` | PASS (119, 41 skip) |
| `git apply --check` 05 on us-panel-dev/work2 | PASS |
| Go `go vet . ./i18n`, `go test -race . ./i18n`, `node check_ui.mjs` (yamalı r10_base) | PASS |

---
# TUR 4 (06.11.2026 ~10:15 TSİ) — PANEL DEV incelemesi (PANEL_GLOBAL_REVIEW_20261106.md) kapatma

Taban: sabit `/workspace/us-panel-dev/pinned/v10/` (PANEL DEV manifesti 181 dosya, SHA-256 3956798c…af3). Çalışma kopyası `global-work/v10dev`. Pinned ağaca yazılmadı.

## DÜZELTME NOTU (önceki turlardaki yanlış beyanlar)
- Tur 1–3 tabanı olarak anılan `work/api/r10_base` sabit .10 manifestiyle **eş değildir** (PANEL DEV: 33/181 dosya farklı/eksik). Bu turdan itibaren taban `pinned/v10`.
- 02 kapsamı: gui.go'da "+1 satır" değil **+2 satır / 2 hunk** (import + route). Metin sayısı "36" değil **37 data-i18n** (36 mevcut + yeni "Dil" etiketi) + **3 placeholder**.
- Python: "PASS (117/119…)" ve "PASS (119, 41 skip)" yazımları yanıltıcıydı. Gerçek yeniden sayım (bu tur): `unittest discover -v` **Ran 119 = 78 PASS + 41 SKIP + 0 FAIL**; `US_GO_LOCALES` ayarlıyken **79 PASS + 40 SKIP + 0 FAIL** (senkron testi koşar). PANEL DEV'in 79/40 sayısı ikinci koşuya karşılık gelir.
- Tur 1–2'de "licensesrv TestParsePlansBothExportSchemas önceden var olan FAIL" denmişti: kök neden kod hatası değil, testin modül dışı `../../../../reports/US_PLANS_JSON_*.json` fixture yolunu araması. Doğru dizin derinliğinde (fixture'lı) koşuda taban da yamalı ağaç da PASS.

## Bulgular — önce FAIL / sonra PASS
| Bulgu | Regresyon testi | Önce (eski 03) | Sonra (yeni 03) |
|---|---|---|---|
| G-01 FIFO | TestPrefGetFIFODoesNotBlock | FAIL "Get blocked on FIFO" (500 ms) | PASS |
| G-01 symlink dosya | TestPrefGetRejectsSymlinkFile | FAIL "symlink followed: en" | PASS |
| G-01 büyük dosya | TestPrefGetRejectsLargeFileQuickly (1 GiB sparse) | FAIL "read took 1.15s (unbounded read)" | PASS |
| G-02 symlink üst dizin | TestPrefSetRejectsSymlinkParent | FAIL "succeeded" + "write escaped" | PASS (kök dışı dosya yok) |
| G-02 symlink son bileşen | TestPrefSetRejectsSymlinkFile | FAIL "Set over symlink succeeded" | PASS (hedef bozulmadı) |
| G-02 normal yazım kökte kalır | TestPrefSetStaysInsideRoot | (yeni regresyon) | PASS (0600, temp kalmaz) |
| G-03 yabancı Origin | TestLangPostRejectsForeignOrigin | FAIL "status 200" | PASS (403; aynı origin ve Origin'siz 200) |
| G-04 ek veri / ikinci nesne / 2 MB | TestLangPostRejectsTrailingAndOversized | FAIL 3×200 + tercih değişti | PASS (400/400/413) |
Loglar: `r4_before_i18n.log`, `r4_before_gui.log`, `r4_after_i18n.log`, `r4_after_gui.log`.

## Uygulama
- `i18n/pref.go`: Get mutex tutmaz; okuma `readPrefFile` (en fazla 4096 B, LimitReader). Set yalnız yazımda kilitler. PrefHandler: `allow` (Origin) parametresi, `http.MaxBytesReader` 4 KB, tek JSON belgesi + EOF kontrolü, 413/400.
- `i18n/pref_unix.go` (!windows): kökten itibaren her bileşen `openat(O_NOFOLLOW|O_DIRECTORY)` (golang.org/x/sys/unix — modülde zaten v0.30.0, yeni bağımlılık yok), eksik dizin `mkdirat 0700`; okuma `O_NOFOLLOW|O_NONBLOCK` + fstat regular + boyut; yazma dirfd üzerinden `O_EXCL|O_NOFOLLOW` temp + fsync + `renameat` + dir fsync; son bileşen symlink/özel dosya ise ret.
- `i18n/pref_windows.go`: openat yok → Lstat ile bileşen yürüyüşü (symlink/reparse=ModeIrregular ret), Lstat+Open+SameFile, rename öncesi dizin yeniden kontrol. GOOS=windows vet/build PASS; gerçek Windows'ta NOT TESTED.
- `gui_lang.go`: ortak `sameOriginRequest`; `/api/act` ve `/api/lang` POST aynı yardımcıyı kullanır (act'in hata metni/kodu aynı).

## Yama etkisi
- Yeni 03 sha256: `171bf6fc4a45e72fea90ce73ef680d957abf21c3d3af083ae60f9679e091a41a`.
- 04 yeniden üretildi (sha256 `5ffc05358658152587ca484a131befc8e7ca39a69e3966743497cd59b405b65d`): mantık aynı; yalnız 03'ün act bloğunu sadeleştirmesi nedeniyle bir satırın girintisi/bağlamı ve `net/http` importunun 03'e taşınması. Eski 03/04 `global-work/patch_archive_r3/`.
- core (e4f10a33…), 02 (314fab48…), 05 (7c917564…, PANEL DEV'in incelediği bayt) DEĞİŞMEDİ.

## Testler (pinned/v10 + 01→04, fixture'lı düzen /tmp/fx_pat)
| Komut | Sonuç |
|---|---|
| git apply --check && apply 01→04 (sıralı) | PASS 4/4; sonuç ağaç v10dev ile birebir |
| go build ./... | PASS |
| go vet ./... | PASS |
| GOOS=windows go vet ./... / go build ./... | PASS / PASS |
| gofmt -l (yamaya ait dosyalar) | PASS (boş) |
| node check_ui.mjs | PASS |
| go test -race -count=1 -json ./... (yamalı) | PASS: 410 pass / 0 fail / 7 skip, data race 0 |
| Aynı, temiz pinned/v10 | PASS: 386 / 0 / 7 (PANEL DEV ile aynı) |
| Python discover (kanusdev, 05) | 78 PASS + 41 SKIP; US_GO_LOCALES ile 79 + 40; 0 FAIL |

## Kalan risk
- Linux/macOS'ta config yolunun herhangi bir bileşeni symlink ise (ör. symlink'li HOME) tercih kaydedilmez (POST 500, Get → tr). Güvenli varsayılan; kullanıcı etkisi düşük.
- Windows Lstat yöntemi TOCTOU'ya karşı openat kadar güçlü değil; gerçek Windows testi yok.

---
# TUR 5 (06.11.2026 ~10:30 TSİ) — PANEL DEV R2 (R2-01, R2-02)
Ayrıntı ve PANEL DEV'e dönüş: `global-work/GLOBAL_R2_FIX_FOR_PANEL_DEV.md`.
- Yeni 03 sha256 `425a76afd9e07b9bdc711b5496a1539046d975e63b54380d6e4f503f2c2c40a0`; 04 yalnız bağlamdan yeniden üretildi `1c937497ee30ef7de86245b2fa4e2b6668b8660e2db62017bd7a7c2db5864991`; core/02/05 değişmedi.
- Symlink'li HOME/config desteklenir (kök bir kez çözülür, fd/handle ile sabit); kökün altında symlink reddi ve G-01…G-04 korumaları sürer; kök açılamazsa bellek + 503; GUI başarısız kayıtta geri alır ve TR/EN mesaj gösterir.
- Windows: NtCreateFile(RootDirectory)+FILE_OPEN_REPARSE_POINT+GetFileInformationByHandle, NtSetInformationFile(FileRenameInformation) ile handle tabanlı; Windows testleri derlendi, koşulmadı (wine kurulamadı) → NOT TESTED.
- go test -race -p=1 ./...: 415/0/7 PASS; build/vet/windows vet+build/gofmt/check_ui/UI testi PASS.
- Yığın .11→UYUM(v3/v2)→API(r11)→GLOBAL: check+apply 4/4 PASS, build+vet PASS; birleşik suite NOT TESTED.


---
# TUR 6 — 08.11.2026 00:11 TSİ — geri yükleme sonrası doğrulama
Yayın, yükleme, uzak kod aktarımı veya ücretli kaynak oluşturma YOK. Yamalara ve site paketine dokunulmadı.

## Bütünlük
- Beş yamanın SHA-256'sı R3 beklenen değerlerle **5/5 eş** (core e4f10a33, 02 314fab48, 03 425a76af, 04 1c937497, 05 7c917564). Tam değerler: `tur6/patches.json`.
- PANEL DEV manifesti SHA256 `3956798c527456ad7f32aeb5ef32db5db17fe654eeddae0186f3576eb6f81af3`; pinned/v10 **181/181**, eksik=0, ek=0, farklı=0 (`tur6/baseline.json`). Pinned'a yazılmadı.
- 01→04 yeni kopyada check/apply 4/4 PASS. Eski /tmp/fx3'e güvenilmedi; kalıcı yeni fixture: `tur6/fixture/work2/kiyas/term`, rapor JSON'ları `tur6/fixture/reports/`.

## Gerçek test sonuçları
| Komut / kontrol | Sonuç |
|---|---|
| go build ./... / go vet ./... | PASS / PASS |
| GOOS=windows go vet ./... / go build -o /dev/null ./... | PASS / PASS |
| GOOS=windows go test -c ./i18n | PASS (non-race exe); Windows runtime NOT TESTED |
| go test -race -count=1 -p=1 -json ./... | **415 PASS / 0 FAIL / 7 SKIP**, exit 0, DATA RACE 0 |
| Ana paket / hook / i18n / licensesrv dağılımı | 315+6 skip / 5 / 22 / 73+1 skip |
| node check_ui.mjs | PASS |
| node gui_lang_ui_test.mjs | **4 PASS / 0 FAIL** (500,503,ağ hatası geri alır; 200 geçer) |
| Python test_i18n_mail, US_GO_LOCALES=new tree | **10 PASS / 0 FAIL / 0 SKIP** |
| Site release/public_html sha256sum -c | **16/16 OK** |
| Site validate.py (varsayılan site/out) | **FAIL**: geri yüklemede out/sitemap.xml yok. Paket bozuk değil, ara çıktı eksik. |
| Gerçek release paketinin izole kopyasında aynı validate.py | **PASS**: 14 sayfa / 14 sitemap URL; noindex=0; yasak slogan/hukuki sayfa/iç-link testleri aktif |

Eksik `site/out` yeniden build edilmedi (slogan gelmedi, site paketine dokunma sınırı). Paket `tur6/site_validation/out` altına kopyalanıp orada doğrulandı. Hukuki sayfa veya slogan değişikliği yok. Loglar `tur6/static.log`, `race.json`, `race.exit`, `race.stderr`, `python.log`, `site.log`, `site_release_validation.log`.

## PANEL DEV §2 / S2 regresyonu
R2 raporunun **§2'si eski/yeni 04 karşılaştırması**, doğrudan bir regresyon koşucusu değil. O karşılaştırmayı da yineledim: `patch_archive_r4` ile güncel 04'ün değişen +/- satırları birebir aynı (PASS).
Raporda adı geçen **S2 saldırı regresyon koşucusu**, `panel-evidence/review/regression/run_s2_regression.sh` (rapor §5 test tablosu); yanındaki poc testlerini geçici kopyaya ekler. Yeni 03/04 ile tekrar koştum: **36 PASS / 0 FAIL / 0 BLOCKED**, exit 0; kaynak öncesi/sonrası manifest eş (kaynağa dokunmadı). Kanıt: `tur6/s2/table.md`, `raw_test.log`, `src_manifest_before.txt`, `src_manifest_after.txt`.

## Yeni yerel teslimler (hiçbiri yüklenmedi)
- `GLOBAL_R3_REVIEW_GUIDE.md`: dosya/fonksiyon haritası, güvenlik değişmezleri↔test eşlemesi, ilk kullanımda kök çözme/handle yaşam süresi/Windows sınırları.
- `ci/windows-i18n.yml`: manual dispatch, windows-latest, yalnız ./i18n race + dört Windows testi, skip'i eksik kapsam sayar.
- `ci/i18n_windows_amd64.test.exe` (9,163,776 B, SHA256 548944ea038b05c5b7202e5f1e228527bddcc7c2d1bc9e57bcf4db3e5b6b9ac4), `ci/SHA256SUMS.txt`, `ci/run_windows_tests.ps1` (binary veya source/race modu).
- `tur6/WINDOWS_OPTIONS.md`: ayrıntılı kaynak/maliyet/onay karşılaştırması. PowerShell/Actions runtime NOT TESTED (pwsh/actionlint yok); yalnız yerel dosya taslakları.

## Windows seçenekleri ve öneri
| Seçenek | Kanıt | Maliyet / onay |
|---|---|---|
| Mevcut kişisel Windows | Bağlı kayıtlı DESKTOP-48D4KOL (OS bu tur sorgulanmadı); binary+PS1 hazır | Mevcut Windows amd64 ise ek servis ücreti yok; transfer/çalıştırma yerel onayı gerekir. Bilgisayarda hiçbir komut çalıştırılmadı. **İlk öneri** |
| Private repo Actions | Salt okuma aramasında demirtumen08-design/project ve project-2 private; workflow hazır | Hesap kota durumu bilinmiyor; resmi Windows standart oranı $0.010/dk (~$0.03–0.10/koşu, tahmin). Kod yükleme+repo seçimi+olası harcama onayı gerekir |
| Yerel QEMU/KVM eval | /dev/kvm var, açılış EACCES; qemu-system-x86_64 yok. 119 GiB boş disk, 3.3 GiB available RAM | ISO **8,225,329,152 B / 7.66 GiB** (Microsoft HEAD), indirilmedi. 8 GiB RAM + 64–80 GiB dinamik disk pratik öneri; KVM erişimi/kurulum/lisans+indirme onayı gerekir; şu an BLOCKED |
| Mevcut lisanslı başka Windows VM | Varlığı UNKNOWN; aynı binary+PS1 kullanılabilir | Yerel transfer/onay; yeni ücretli kaynak yok |

Microsoft ISO yalnız HEAD ile boyut doğrulandı (`tur6/windows_iso_head.txt`), gövde indirilmedi. 90 günlük evaluation. QEMU/KVM kurulumu ve VM başlatılmadı. Actions çalıştırılmadı, hiçbir repoya commit/push yapılmadı. Gerçek Windows **NOT TESTED** kalır.

---
# TUR 7 — 08.11.2026 ~00:25 TSİ — gerçek Windows sonucu ve DACL düzeltmesi
Yayın, GitHub, harici yükleme YOK. Kanıtlar: `global-work/tur7/`.

## Gerçek Windows sonucu (eski binary)
DESKTOP-48D4KOL, Windows 10 19045, **race'siz** binary sha256 548944ea…9ac4, log `ci/windows-results-DESKTOP-48D4KOL-20261108/binary-tests.log` (UTF-16, PS 5.1 Tee).
- TestWinPref*: **3 PASS / 1 SKIP** (SymlinkFileRejected: "Gereken ayrıcalık istemci tarafından sağlanmıyor" = SeCreateSymbolicLinkPrivilege yok).
- **1 FAIL**: TestPrefStoreRoundTrip `pref_test.go:26: perm -rw-rw-rw-`. Diğer 10 platform-bağımsız test PASS. Toplam: 13 PASS / 1 SKIP / 1 FAIL.
- Betik: PS 5.1'de param varsayılanında `$PSScriptRoot` boş → KAN.US -Binary/-LogDir'i elle verdi.

## Kök neden
Go `os.Stat` Windows'ta (`syscall` / `os/types_windows.go`) mod bitlerini yalnız FILE_ATTRIBUTE_READONLY'den türetir: salt okunur değilse 0666, ise 0444. 0600 Windows'ta asla görünmez; erişimi ACL belirler. Yani FAIL bir test varsayımı hatası. **Ancak** kod incelemesi gerçek bir boşluk da gösterdi: NtCreateFile SecurityDescriptor=nil ile çağrılıyordu → lang.json ve oluşturulan alt dizinler ebeveynin devralınabilir ACE'lerini alır. Profil altında (%APPDATA%) bu yalnız kullanıcı+SYSTEM+Administrators demektir (düşük risk); ama config kökü profil dışında olursa (ör. US_DESK=C:\ShareDir, C:\ kökünde varsayılan Authenticated Users: Modify devralınır) başka yerel kullanıcılar lang.json'u değiştirebilirdi. İçerik yalnız dil kodu (gizlilik düşük), etki: başka kullanıcının arayüz dilini değiştirme (bütünlük, düşük-orta).

## Değişiklik (03 yaması)
- **b) uygulandı** (bağımlılık eklemiyor, x/sys zaten go.mod'da): `privateSD()` SDDL `O:<kullanıcı>D:P(A;;FA;;;<kullanıcı>)(A;;FA;;;SY)(A;;FA;;;BA)`; dizinler için OICI. NtCreateFile'a `OBJECT_ATTRIBUTES.SecurityDescriptor` ile verilir → ACL oluşturma anında atomik uygulanır (yarış penceresi yok), handle tabanlı akış (RootDirectory, FILE_OPEN_REPARSE_POINT, FileRenameInformation) aynen korunur; rename güvenlik tanımlayıcısını korur. Gerekçe: test-only (a) gerçek boşluğu açık bırakırdı; değişiklik küçük ve yalnız oluşturma yolunda.
- Okuma: `NumberOfLinks != 1` dosya reddedilir (yetkisiz hardlink ile kök dışı dosya aliasing'i). Yazma zaten rename ile girdiyi değiştirir, hedef dosyaya yazmaz.
- **a) test ayrımı**: `checkPrivateFile` — `pref_perm_unix_test.go` (0600) / `pref_perm_windows_test.go` (GetNamedSecurityInfo: sahip=geçerli kullanıcı, SE_DACL_PROTECTED, Everyone/Users/Authenticated Users için yazma sınıfı hak yok).
- Yeni Windows testleri: `TestWinPrefDACLSharedRoot` (Everyone/Users/AU FA devralan kök → dosya+alt dizin özel DACL), `TestWinPrefHardlinkAliasRejected` (yetkisiz eşdeğer senaryo). Symlink testi yetki ister; junction (dizin) senaryoları zaten yetkisiz koşuyor. Dosya symlink'inin yetkisiz eşdeğeri yok → SKIP belgeli kalır (Developer Mode veya SeCreateSymbolicLinkPrivilege ile kalkar).
- Kalan risk: config kökünün kendisi bizim oluşturmadığımız paylaşımlı bir dizinse, oradaki kullanıcılar FILE_DELETE_CHILD ile lang.json'u silip yeniden oluşturabilir (sonuç yine yalnız dil tercihi; okuma tarafı doğrulamaları sürer). Kök DACL'ine dokunulmadı (kullanıcının dizini).

## Yamalar
- 03 yeni sha256 `ad63ce899ab221444ccc4b9228374f6cb09d5134f8d028f778fa4e5d67af0ce7` (eski 425a76af… → `patch_archive_r5/`).
- 04 **değişmedi** `1c937497…` (i18n/pref* dosyalarına dokunmuyor; yeni 03 üstüne temiz uygulanıyor). Arşive referans kopyası konuldu.
- core e4f10a33…, 02 314fab48…, 05 7c917564… değişmedi.
- pinned/v10 kopyasında 01→04 sıralı --check+apply **4/4 PASS**; sonuç ağaç v10dev ile birebir (.git hariç).

## Testler (Linux, fixture `tur6/fixture/work2/kiyas/term`)
| Komut | Sonuç |
|---|---|
| go build / go vet ./... | PASS / PASS |
| GOOS=windows go vet ./... / go build ./... | PASS / PASS |
| gofmt -l i18n gui.go gui_lang.go | PASS (boş) |
| node check_ui.mjs / gui_lang_ui_test.mjs | PASS / 4 PASS |
| go test -race -count=1 -p=1 -json ./... | **415 PASS / 0 FAIL / 7 SKIP**, exit 0, DATA RACE 0 (yeni testler yalnız Windows'ta derlenir) |
| S2 regresyonu run_s2_regression.sh | **36 PASS / 0 FAIL / 0 BLOCKED**, exit 0 (`tur7/s2/`) |
| run_windows_tests.ps1 parse (pwsh 7.4.6, /tmp'ye indirildi) | PASS, 0 parse hatası; sahte binary ile: hash uyuşmazlığı → exit 3; eski gerçek log ile özet "TestPrefStoreRoundTrip: pref_test.go:26: perm -rw-rw-rw-", exit 1 |
| PS 5.1 gerçek koşum, yeni Windows binary | NOT TESTED |

## Windows teslimleri
- `ci/i18n_windows_amd64.test.exe` 9,204,736 B, sha256 `ca7a79d56865eb24adfb338d86770c30cfc6e5f813aef599905f7c5d7e4e9a46` (race'siz; 6 TestWinPref + DACL kontrollü RoundTrip).
- `ci/run_windows_tests.ps1` sha256 `c84403f7c1e940c2c8b1d3823570817f4c67bc456281e8b49be7514c87987809`: PS 5.1 yol çözümü ($PSScriptRoot → $MyInvocation.MyCommand.Path → Get-Location), SHA256SUMS karşılaştırması (uyuşmazsa exit 3), symlink yetkisi / Developer Mode raporu, FAIL adı+mesajı özeti. Çıkış: 0 PASS, 1 FAIL, 2 PARTIAL (skip/eksik), 3 hash/kurulum.
- `ci/SHA256SUMS.txt` güncellendi. Eski exe+betik+sums: `ci/archive_548944ea/`.

---
# TUR 7b — 08.11.2026 ~00:30 TSİ — PANEL DEV R3-01 (kök handle önbelleği)
Kaynak: `/workspace/us-panel-dev/PANEL_GLOBAL_REVIEW_R3_20261106.md` R3-01 LOW (64 kökte fd 7→71).
- **Önce FAIL (yeniden üretildi):** eski 03 üzerinde yeni test `TestPrefRootCacheBounded`: `handle growth 7 -> 71 after 64 roots`.
- **Düzeltme (pref.go, pref_unix.go, pref_windows.go):** `maxRoots = 1`; yeni kök çözülünce eski kayıt önbellekten çıkarılır (`evicted`), handle'ı referans sayımıyla (`refs`, PrefStore.mu altında) kullanımda değilse hemen, kullanımdaysa son `release` ile tam bir kez kapanır (`closeRoot`: Unix `unix.Close`, Windows `CloseHandle`; aynı davranış). Get/Set her `root()` sonrası `defer release`; Set'te release, yazım kilidi bırakıldıktan sonra çalışır. `PrefStore.Close()` eklendi.
- **Testler:** `pref_rootcache_test.go` — 64 kök sonrası handle artışı ≤16 (sızıntı +64 olurdu; Linux'ta ölçülen 7→7), önbellek ≤1, Close sonrası artış yok; kullanımda evict edilen kök erken kapanmaz, release'te kapanır, yeniden açılır; 8 goroutine kök değiştirirken Get/Set (race). Handle sayımı: Linux `/proc/self/fd`, Windows `GetProcessHandleCount`, diğerleri Skip.
- 03 yeni sha256 `f66d4416bce2b5d61382fb097705104baf6e9d742e085cce9c0dd77933441a8f`; 04 değişmedi `1c937497…`. Eski 03/04 (ad63ce89…) `patch_archive_r6/`. pinned/v10 üzerinde 01→04 check+apply 4/4 PASS, ağaç v10dev ile eş.
- Fixture'lı `go test -race -count=1 -p=1 ./...`: **418 PASS / 0 FAIL / 7 SKIP**, DATA RACE 0 (`tur7b/`); i18n -race -count=3 PASS. S2 **36/36 PASS**. vet, GOOS=windows vet+build, gofmt: PASS.
- Windows binary yeniden derlendi: `ci/i18n_windows_amd64.test.exe` sha256 `d866b9aac921074e65c874b4435e65561ba0b2030bea2821119c36edad1239a2` (9,231,872 B; 6 TestWinPref + 3 TestPrefRoot*). Betik `ci/run_windows_tests.ps1` sha256 `b5b6c2ce1ba3134c4f7d156c4cf8dc5eb2c77016892ddbd3561faba59e916c44` (yalnız yorum güncellendi). Eski exe/betik/sums `ci/archive_ca7a79d5/`. Windows'ta koşum NOT TESTED.

---
# TUR 8 — 08.11.2026 ~00:45 TSİ — §9: kanus.org/us TR/EN/DE/RU, panel DE/RU, fiyat gösterimi
Yayın, yükleme, DNS, Shopify, GitHub YOK. Kanıtlar `global-work/tur8/`.

## §9'dan GLOBAL'e düşen (KANUS_PROMPT_20261107 §9, harfiyen)
- GLOBAL: "TR/EN/DE/RU seçeneklerinin gerçekten bütün arayüzü çevirdiğini doğrula", "eksik çevirileri tamamla", **çok dilli panel**, **çok dilli hata mesajları**, **çok dilli e-posta**, **uluslararası fiyat gösterimi**.
- API'ye ait (dokunulmadı): global hesap sistemi, API erişimi, OpenAPI şeması, Python SDK, JavaScript SDK, kullanıcı organizasyonları, yetki yönetimi, global performans ölçümü (API/ortak).
- UYUM'a ait: GDPR/KVKK teknik kontrolleri, veri dışa aktarma, veri silme talebi; "hukuki ve vergisel uygunluğu teknik uygulamadan ayrı değerlendir".
- GLOBAL'i bağlayan diğer kurallar: §13 kök neden→test→düzelt→regresyon, go test/-race/vet; §14 üretime dağıtım ve müşteriye e-posta onaysız yok; §15 rapor alanları; §2 sınıflandırma etiketleri.

## Canlı sayfa incelemesi (salt okuma, curl 00:22 TSİ)
- `/us` = SPA kabuğu `index.html` (1600 B, lang="en", last-modified 03.11.2026 23:22 GMT) + `/assets/index-uscommerce.js?v=20261104-usticari` (248,418 B) + `/assets/index-BaRMt_G_.css`. SPA kaynağı box'ta YOK (rg: yalnız canlı kopyalar/raporlar). Canlı dosyalar `tur8/live/` (sha256 `live_sha256.txt`).
- Seçici: `localStorage["kanus.lang"]` + React context; sözlükler `en`(temel) + `tr/de/ru` = `{...en, ...}` (eksik anahtar sessizce EN'e düşer). URL/hreflang değişmez.
- Sözlük paritesi: 4 dilde **126/126 anahtar**; EN ile aynı kalan değerler TR 5, DE 9, RU 3 — hepsi marka/terim veya dilde doğru (US, CVC, Forum, DE "Journal/Name/Prompt"). /us gövdesi ayrı 4 dilli dizide, tam.
- Eksikler (4 dilin tamamında): (1) `<head>` title/description/og yalnız EN, `lang="en"` JS öncesi; (2) günlük/galeri/forum tarihleri ham ISO (`2026-11-06T06:05:00.000Z`); (3) iletişim/forum/yükleme gönderimi hata verince **hiç mesaj yok** (yakalanmamış promise), API hata kodları (`api-html`, `missing`, `credits`) yerelleştirilmemiş; (4) forum yanıt formu hatası da sessiz (paket dışında kaldı); (5) fiyat gösterimi YOK (payments rotası /us'a yönleniyor; ödeme anahtarları kullanılmıyor); yerel yedek ledger stub'ında `currency:"USD"` sabiti var — TRY-only politikasıyla çelişiyor, Ticari/US'e not, dokunulmadı.
- Baseline gerçek tarayıcı testi (canlı kopya): **4 PASS / 8 FAIL** (/us metni 4/4 PASS; tarih 0/4, hata 0/4) — `tur8/ui_live_baseline.log`.

## Yapılanlar
### Panel (yama 06, Go) — `patches/global_i18n_06_de_ru.patch`
- `i18n/locales/de.json`, `ru.json` (93 anahtar, TR/EN ile tam parite; hata mesajları + e-posta konusu/gövdesi dahil). Ton BRAND_A: Kanus/US Panel, DE "du" (canlı site ile aynı), RU "Вы"; yerli/milli vurgu yok, yasak sloganlar yok. `brand.slogan` DE/RU geçici slogan çevirisi (Demir'in sloganı gelince tek anahtardan değişir).
- `_meta.currency` DE/RU = TRY (gösterim varsayılanı; kur kaynağı olmadığından EUR/RUB dönüşümü YOK).
- `i18n.go`: RU sayı biçimi (NBSP binlik, virgül ondalık), RU tarih, RU para "1 234,50 ₺" (tutar+sembol), RUB sembolü. Ödeme/ledger koduna dokunulmadı.
- `gui.html`: hesap geçmişi metinleri 4 yeni anahtara (`gui.history.*`) ve tarih `Intl.DateTimeFormat(bcp47)`'e bağlandı; `i18nT/i18nDate/i18nNumber` yardımcıları. Dil listesi sunucudan geldiği için seçicide DE/RU otomatik görünür.
- Testler: `i18n_de_ru_test.go` (4 dil parite, EN kopyası yok, DE/RU biçim); `TestNegotiate` (de/ru artık destekli, ja→en), `TestPrefHandler` (desteksiz örnek de→ja) güncellendi.
### E-posta (yama 07, Python) — `patches/global_i18n_07_mail_de_ru.patch`
- `kanus/locales/de.json`, `ru.json` (Go ile senkron alt küme); `test_four_languages_parity` + resolve_lang de/ja beklentisi. Gönderime bağlı değil.
### kanus.org/us paketi — `site/release_us/` (YÜKLENMEDİ)
- Canlı bundle'ın yerel kopyasına 6 noktasal değişiklik (`site/us/build_us.py`, her eşleşme tam 1 kez doğrulanır): tarihler `Intl.DateTimeFormat` (tr-TR/en-US/de-DE/ru-RU); iletişim, forum konusu, yükleme gönderim hatalarında 4 dilde mesaj. Yeni asset `assets/index-uscommerce-i18n-75f56119.js` (yeni dosya), `index.html` yalnız script src farkı (**üzerine yazar**), `rollback/index.html` = canlı orijinal.
- `SHA256SUMS`, `CONFLICT_CHECK.md` (canlı sha'lar + ön-yükleme eşleşme koşulu), `VERIFY_AFTER_UPLOAD.sh`, `README_RELEASE.md`.

## Yamalar (sıra: core→02→03→04→06 Go; 05→07 Python)
| Yama | sha256 |
|---|---|
| panel_core | e4f10a339c78818ceddbc49e2f687b4ee80afa86a7fce9d407c22bd8536892c7 |
| 02 | 314fab48b009f06e221ce140758724105b9f5b972c69d0c15c7ad2c9db8d3669 |
| 03 | f66d4416bce2b5d61382fb097705104baf6e9d742e085cce9c0dd77933441a8f |
| 04 | 1c937497ee30ef7de86245b2fa4e2b6668b8660e2db62017bd7a7c2db5864991 |
| 05 | 7c917564df055f151f8bd0a0b334a09e248db1481723123196fbfb0377093f3d |
| 06_de_ru | 83aaca5d4a3aeb8cf7d4ed0e3854d0c3408d1553bb139ddb2eee22f8b5d013c6 |
| 07_mail_de_ru | 70acc9b0789a982c2abfe3b0e6f3febbf62db16cf3957343e7fc916f50efae18 |

## Testler (gerçek)
| Kontrol | Sonuç |
|---|---|
| pinned/v10 + core→02→03→04→06 check+apply | PASS 5/5; ağaç v10dev ile eş |
| boş dizin + 05→07 check+apply | PASS 2/2 (izole dizinde test_delivery_not_wired delivery.py olmadığından ERROR — beklenen; gerçek kanusdev'de PASS) |
| go test -race -count=1 -p=1 ./... (fixture) | **420 PASS / 0 FAIL / 7 SKIP**, DATA RACE 0 |
| Locale parite (Go, 4 dil) TestFourLocaleParity + TestNoMissingKeys | PASS |
| go vet / GOOS=windows vet + build + test -c / gofmt | PASS |
| check_ui.mjs / gui_lang_ui_test.mjs | PASS / PASS (4) |
| S2 regresyonu | **36 PASS / 0 FAIL / 0 BLOCKED** |
| Python discover (kanusdev) | Ran 120: **79 PASS + 41 SKIP**, 0 FAIL; test_i18n_mail US_GO_LOCALES ile 11/11 PASS |
| /us gerçek headless Chrome (CDP, yerel stub sunucu) release_us | **12 PASS / 0 FAIL** (4 dil × /us metni, tarih, hata mesajı) |
| Aynı test canlı kopya (baseline) | 4 PASS / 8 FAIL |
| validate_us.py | 21/21 PASS (node --check, hash adı, tek fark, yasak slogan, sözlük parite 126, SHA256SUMS 6/6) |
| VERIFY_AFTER_UPLOAD.sh yerel simülasyon | 13 PASS / 0 FAIL |
| Önceki statik site validate.py (release kopyası) | PASS 14/14 |
| Canlı yükleme / gerçek CDN / mobil | NOT TESTED (BLOCKED: onay + KAN.US yüklemesi) |
| DE/RU anadil kontrolü | NOT TESTED (makine/ekip çevirisi) |

## Açık sorular / onay
1. DE/RU metinlerinin anadil kontrolü (özellikle hata mesajları, e-posta).
2. Döviz gösterimi isteniyor mu? Kur kaynağı yok; şu an yalnız TRY, DE/RU yerel sayı biçimiyle.
3. SPA kaynağı: yoksa her düzeltme derlenmiş bundle yamasıyla yapılır; `<head>` meta çevirisi ve forum yanıt hatası için kaynak gerekir.
4. release_us yükleme onayı (index.html üzerine yazar); canlı ledger stub'ındaki `currency:"USD"` için Ticari/US kararı.

---
# TUR 9 — 08.11.2026 ~00:45 TSİ — Windows r3 sonucu, R3-01 testi, betik exit kodu
Yayın/GitHub/harici yükleme YOK. Kanıtlar `tur9/`, `tur8/exitcodes.txt`.

## Gerçek Windows r3 sonucu (kayıt)
DESKTOP-48D4KOL, Win10 19045, PS 5.1.19041.6456, race'siz binary d866b9aa (hash doğrulandı), log `ci/windows-results-DESKTOP-48D4KOL-20261108-r3/` (UTF-8 BOM). **18 PASS / 1 FAIL / 1 SKIP** (20 test). PASS: DACLSharedRoot, HardlinkAlias, Junction×2, NormalSave, PrefStoreRoundTrip (DACL), RootCacheEviction/ConcurrentSwitch vb. SKIP: SymlinkFileRejected (SeCreateSymbolicLinkPrivilege absent, Developer Mode off). FAIL: TestPrefRootCacheBounded `pref_rootcache_test.go:51: Close left handles: 146 > 141`. Logda çağrı komut satırı yok (environment.txt'te bu alan yoktu).

## Kök neden: ölçüm gürültüsü (gerçek sızıntı kanıtı yok)
- Aynı testin 64-kök döngüsü sonrası kontrolü (`after-before ≤ 16`) Windows'ta **geçti**; kök başına sızıntı +64 verirdi. Başarısız olan yalnız Close sonrası **sıfır toleranslı** mutlak karşılaştırmaydı (+5). GetProcessHandleCount süreç genelidir (runtime thread/IOCP/event/timer, AV/EDR); +5 bu gürültüyle uyumlu, kök sayısıyla orantılı değil.
- Kod incelemesi (pref_windows.go): openRoot verify hatasında CloseHandle; walkBelow her adımda önceki ara dizini closeBelow ile kapatır, hata yolunda next'i kapatır; readPrefFile handle'ı os.File ile sarar ve defer Close; writePrefFile eski hedef handle'ını hemen kapatır, temp için defer CloseHandle (hata yolunda önce delete-on-close); rename ayrı handle açmaz; GetFileInformationByHandle handle açmaz; kök handle'ı evict/Close'da refcount ile tek kez closeRoot. Açık kalan yol bulunmadı → **kod değişmedi, test düzeltildi**.
- Sınır: Windows'ta yeni testler henüz koşmadı; kesin Windows kanıtı r4 koşumuyla gelir.

## Test düzeltmesi (03)
- Dahili sayaçlar: `rootOpened/rootClosed` (atomic) + test hook `rootCloseHook` (prod'da nil). `TestPrefRootCacheBounded` artık OS sayısı değil **canlı kök handle sayısını** ölçer: 64 kök sonrası ≤1, Close sonrası 0.
- `TestPrefRootHandleScaling`: LockOSThread, 32 kök ısınma, runtime.GC, 64 ve 512 kökte OS handle deltası; `d512-d64 ≤ 32` (sızıntı ≥448 fark verir). Linux: +0 / +0.
- `TestPrefRootClosedHandleInvalid`: kapatılan her kök handle'ı kapanır kapanmaz geçersiz olmalı (Linux fcntl F_GETFD, Windows GetHandleInformation); 16/16.
- EvictionWhileInUse: kullanımdaki handle geçerli; ConcurrentSwitch: sonunda canlı kök 0.
- **Mutasyon kanıtı:** closeRoot çağrısı kaldırılınca Scaling (64→+64, 512→+512) ve ClosedHandleInvalid FAIL; kapatma koşulu devre dışı bırakılınca 5 kök testinin 5'i FAIL. Düzgün kodda tümü PASS (-race -count=3).
- 03 yeni sha256 `ba8845c5f34855d8f2f2fa49a055c4e9d20698c9fbd7572462531c2746f6b58e` (eski f66d4416… → `patch_archive_r7/`). 04 ve 06 değişmedi, core→02→03→04→06 check+apply 5/5 PASS, ağaç v10dev ile eş.

## Betik exit kodu
- Eski betik `exit N`'i fonksiyon içinden (Show-Summary) ve source modunda try/finally içinden çağırıyordu. `-File` altında doğru, ama PS 5.1 `-Command` script'in `exit N`'ini sürecin çıkış koduna yansıtmaz (son komutun $?'ı belirler) → KAN.US'in gördüğü 0. Logda komut satırı olmadığından hangi biçimin kullanıldığı **kanıtlanamadı**; betik artık `cmdline=[...]`'i environment.txt'e yazar.
- Yeni: tek çıkış yolu `Finish N` → `[Environment]::Exit(N)` (+ `trap` → 3). pwsh 7.4.6 ile sahte binary: yeni betik -File / -Command(&) / sh -c+-Command hepsinde **0/1/2/3** doğru; eski betik -Command'da 2 ve 3'ü 1'e düşürdü (`tur8/exitcodes.txt`). PS 5.1 gerçek koşum NOT TESTED.

## Kontroller
| Kontrol | Sonuç |
|---|---|
| go test -race -count=1 -p=1 ./... (fixture) | **422 PASS / 0 FAIL / 7 SKIP**, DATA RACE 0 |
| S2 | 36/36 PASS |
| vet / GOOS=windows vet + build / gofmt | PASS |
| Windows binary (tam ağaç 01→04+06) | `ci/i18n_windows_amd64.test.exe` sha256 `166d869a5dad7e21985a5b910f894c4283af1799244653fce3f0c97104f974fe`, 9,263,104 B, 24 test |
| Betik | `ci/run_windows_tests.ps1` sha256 `193728fabe80135acd0b6bb2bb5496209d86a461a7f95bfe533f155231835278`; eski exe/betik/sums `ci/archive_d866b9aa/` |
| Windows r4 koşumu | NOT TESTED (beklenen: 23 PASS / 0 FAIL / 1 SKIP, exit 2 PARTIAL) |

## Tur 10 — Yaklaşık döviz (TCMB) + SPA kaynağı araması (08.11.2026, TSİ)
Yerel iş. Canlıya/DNS/Shopify/GitHub'a hiçbir şey yüklenmedi; Hostinger yalnız okundu.

### Fiyat nerede görünüyor (tespit)
- **kanus.org/us (SPA):** render edilen fiyat YOK. `/payments` → `/us` yönlendirmesi; `payTitle/rateLine` ("1 USD = 10 kredi", "Tutar (USD)") sözlükte duruyor ama hiçbir rota kullanmıyor (ölü metin). /us sayfası yalnız shop.kanus.org (Shopify) bağlantısı veriyor. Uydurma fiyat eklenmedi → **fiyat verisi yok**; bileşen hazır, `data-kanus-try` taşıyan öğe gelince çalışır.
- **Panel (EXE):** `gui.html` `{{US_FIYAT}}` → `pricing.go:pricingSummary()` (licensesrv.Catalog TL fiyatları, yalnız TR metin, statik `USDTRYRef=50.05`). Panelde fx.json kaynağı yok → `global_i18n_08_fx.patch` yalnız `i18n/fx.go` (+test): `ParseFX` doğrulama + `Localizer.ApproxFX` (4 dil not). pricingSummary'ye bağlanmadı (PARTIAL; licensesrv/ödeme/ledger'a dokunulmadı).

### fx.json (gerçek TCMB, 08.11 00:39 TSİ çekildi)
`{"schema":"kanus.fx/1","tcmb_tarih":"06.11.2026","bulten_no":"2026/209","date":"2026-11-06","base":"TRY","rates":{"USD":50.0724,"EUR":55.9609}}` (ForexSelling). Eski (>5 gün), gelecek, bozuk şema/sayı, eksik kur, çapraz kur dışı → reddedilir, döviz gizlenir.

### Paket `site/release_us_fx/` (YÜKLENMEDİ)
Ön koşul doğrulandı (salt okuma): canlı index.html sha = release_us `a828fa73…` (release_us canlıda).
| Dosya | sha256 |
|---|---|
| public_html/assets/index-uscommerce-i18nfx-5a375f29.js | 5a375f292604694e6f905a9654353700edf9dc4bb595c5df250354048991c5d0 |
| public_html/index.html | 4ab2f9207612f0f64f835f2a5b2b3a6316070894e1f95e8642b99134b69e79b8 |
| public_html/fx.json | 9eac8d08317d5e5414efb3194ba12658498833ca4ca26fd2db9bf890147ae935 |
| rollback/index.html (=release_us) | a828fa73880192c4cf75c2b37cd5c648be8054c30108c8e84a2bb6c5e866a926 |
| rollback/index.live_original.html | 87b980dec6020cb8071b33aeef441fa80b146352a7761612065f393588841e34 |
| VERIFY_AFTER_UPLOAD.sh | 5165d3159afea4f348dd1dc364c3f121ec1d4d6a636f28ed786ee4609b362818 |
| CONFLICT_CHECK.md | 72726a921c3e8b871159075c9964d7558de288ae7a6f52e7c9e08389425d15e6 |
| cron/fx_cron.php (öneri) | a17b603d18f36bc864f6f4b2d2e34138d6d9fae1e8507b2a4914d7300f3bf314 |
| cron/fx_snapshot.py | 561b3488d569dc26ffdba60f256c65d5157b48e91d09391548becc4997169a5c |
Panel: `patches/global_i18n_08_fx.patch` sha256 `6ac3588bd36e76be008ea0e5d7180bbade76b15718cb9b27f592164668e5f363` (yalnız yeni dosyalar; v10dev üzerinde `git apply --check` PASS).

### Testler
| Test | Sonuç |
|---|---|
| fx_ui_test.mjs (gerçek headless Chrome; 4 dil TL+€+$, not+TCMB tarihi, dil değişimi, fx.json eksik/bozuk/eski, TCMB'ye istek yok) | **18 PASS / 0 FAIL** |
| ui_test.mjs regresyon (fx paketi üzerinde) | **12 PASS / 0 FAIL** |
| test_fx_snapshot.py (gerçek today.xml, eski/gelecek/bozuk/entity/eksik/çapraz, CLI eski dosyayı korur) | **12 PASS** |
| validate_us.py release_us_fx (parite 4×126, slogan yok, rollback, fx şema, TCMB yok) | **25 PASS / 0 FAIL**; release_us ALL PASS |
| go test -race ./i18n (TestFX* 3 dahil), go vet, GOOS=windows vet, gofmt | PASS |
| sha256sum -c SHA256SUMS | PASS (10/10) |
| fx_cron.php | NOT TESTED (kutuda php yok) |
| VERIFY_AFTER_UPLOAD.sh | NOT TESTED (yükleme sonrası) |
| İlk denemede 5 FAIL: test beklentisi hatası (337,73→337,74; RU dar boşluk) + RU notunda "г.." çift nokta → bileşen düzeltildi, yeniden koşturuldu. |

### Günlük güncelleme önerisi (uygulanmadı)
Hostinger cron `5 16 * * 1-5 /usr/bin/php ~/domains/kanus.org/fx/fx_cron.php` (script public_html dışında, yalnız public_html/fx.json'u atomik yazar; hatada dokunmaz). Alternatif fx_snapshot.py veya KAN.US'in günlük fx.json yüklemesi. Demir onayı gerekir.

### SPA kaynağı
BULUNAMADI. Bakılanlar: /workspace rg (`uscommerce`, `kanus.lang`, `path:"payments"`, `usTitle`) → yalnız canlıdan çekilmiş kopyalar/raporlar; tüm diskte package.json/vite.config (yalnız hanedan Next.js, tools puppeteer, sistem paketleri). Hostinger (salt okuma): Git repo listesi boş; public_html = api/, assets/ (19 derlenmiş bundle, kaynak yok), downloads/, storage/, deneme.html, us.html, index.html, .htaccess; public_html üstü API ile listelenemiyor (422). `downloads/US-kaynak.zip` (sha 4db233bf…, 51 dosya) = panel Go kaynağı (kiyas/term) + api/license_mint.php; SPA yok. Bundle'da sourceMappingURL yok, .map istekleri SPA fallback. → Demir'e sorulmalı (KAN.US).

## Tur 9 sonucu — gerçek Windows r4: PASS (symlink SKIP belgeli)
Kaynak yalnız log: `ci/windows-results-DESKTOP-48D4KOL-20261108-r4/` (binary-tests.log, environment.txt).
- Ortam: DESKTOP-48D4KOL, Windows NT 10.0.19045, PS 5.1.19041.6456, `cmdline=[powershell -NoProfile -ExecutionPolicy Bypass -File .\run_windows_tests.ps1]`, SeCreateSymbolicLinkPrivilege=absent, DeveloperMode=off/unknown. Binary 166d869a (hash PC'de doğrulandı — KAN.US beyanı; logda hash satırı yok).
- **23 PASS / 0 FAIL / 1 SKIP** (24 test), son satır `PASS`. TestWinPref: 5 PASS + 1 SKIP (SymlinkFileRejected: "Gereken ayrıcalık istemci tarafından sağlanmıyor" — yetki yok; dosya symlink'inin yetkisiz eşdeğeri yok, junction ve hardlink senaryoları PASS). DE/RU parite + biçim, DACL, R3-01 testleri (Bounded, HandleScaling, ClosedHandleInvalid, EvictionWhileInUse, ConcurrentSwitch) PASS.
- HandleScaling logu: **64 kök +6, 512 kök +2** (KAN.US özetinde "+0" yazıyordu; log +2 diyor). Kök sayısıyla orantısız → sızıntı yok, Tur 9 gürültü teşhisi gerçek Windows'ta doğrulandı.
- `RESULT exit=2` / ERRORLEVEL 2 (PARTIAL, beklenen) KAN.US konsol çıktısına dayanır; konsol satırı binary-tests.log'da yer almaz.
- Exit kodu okuma düzeltmesi: `... & echo %ERRORLEVEL%` aynı satırda yanlış (0). Doğru kullanım `ci/README_WINDOWS.md`: ayrı `echo %ERRORLEVEL%`, ya da `cmd /v:on /c "... & echo !ERRORLEVEL!"`, PowerShell'de `$LASTEXITCODE`. Kutudaki belgelerde hatalı örnek bulunmadı (rg); betik baytları değişmedi (sha256 193728fa… aynı, SHA256SUMS değişmedi).

## Tur 9 ek — PANEL DEV R4B kayıt eksikleri (kod değişikliği yok, 03 donmuş)
- `ci/BUILD_INFO_166d869a.md`: v10dev = pinned/v10 + core→02→03(ba8845c5)→04→06; `GOOS=windows GOARCH=amd64 go test -c ./i18n`, go1.27.3, CGO_ENABLED=0, GOAMD64=v1, -trimpath yok, mtime 00:38:34 TSİ. Aynı komutla yeniden derleme **sha256 166d869a… birebir aynı**.
- HandleScaling Windows r4: **64 kök +6, 512 kök +2** (log). Rapordaki "+0" geçen yerler yalnız Linux ölçümüdür (Linux +0/+0, doğru); Windows için +0 kaydı yok.
- r4 gerçek exit: **RESULT exit=2 / ERRORLEVEL 2 — KAN.US beyanı, logda yok.** r4'te hash stdout'u da arşivlenmedi.
- Betik artık `RESULT exit=N`'i `<LogDir>/result.txt`'ye de yazar: yeni sha256 `32c3034e97a619ea1c930e9b14bf6c793ed69e40f8f5bcc193cb044f5fe023bb` (eski 193728fa… → `ci/archive_193728fa/`), SHA256SUMS güncellendi. pwsh 7.4.6 sahte binary: -File ve -Command'da exit 0/1/2/3 ve result.txt içeriği eşleşiyor (`tur9/exitcodes_result_txt.txt`). PS 5.1 NOT TESTED.

## Tur 11 — fx_cron gerçek PHP testi, R4B FX düzeltmeleri, 08 lisans sunucusu kaynağı (08.11.2026 ~01:00–01:25 TSİ)
Yerel iş; canlıya/Hostinger'a/GitHub'a hiçbir şey gönderilmedi. Canlı salt-okuma (curl): index → `i18nfx-5a375f29` (release_us_fx canlıda), `/fx.json` 200 JSON, `/_fx/` henüz yok (SPA fallback 200 text/html).

### 1) fx_cron.php
- PHP: apt erişilemedi (deb.debian.org 500); statik PHP 8.3.31 CLI (static-php-cli, tgz sha256 4f340f03…) `/workspace/tools/php/php`. Hostinger PHP sürümü doğrulanmadı → 8.x varsayıldı; script PHP ≥ 8.0.
- Önceki sürümde bulunan hatalar: kilit yok; sabit `.tmp` adı public_html içinde; 6 gün + 1 gün gelecek kabulü; `source_url` slash kaçışı (Python ile şema farkı); kurulum yolu connector ile yüklenemiyordu.
- Düzeltme / Seçenek A: `public_html/_fx/fx_cron.php` + `_fx/.htaccess` (Require all denied + Deny from all); `PHP_SAPI !== 'cli'` → 403, hiçbir şey yazmaz/kilitlemez; yalnız `_fx/` içinden çalışır; temp `_fx/` içinde → `public_html/fx.json`'a rename; flock `_fx/fx_cron.lock` (meşgulse çıkış 3); zaman aşımı ≤20 sn; tarih politikası panel ile aynı.
- test_fx_cron.py **11/11 PASS**: canlı TCMB (06.11.2026 USD 50.0724 EUR 55.9609) yazıldı, yalnız fx.json değişti, 0644, temp kalmadı; Python ile alan alan aynı şema; ağ hatası, bozuk XML (4 tür), eksik kur (2), tatil/eski tarih, kilit, zaman aşımı (<8 sn), eşzamanlı okuyucu ile atomiklik (15 tur, yarım dosya 0), `php -S` HTTP → 403 + yazma yok, yanlış dizin reddi; çıktıda sır kalıbı yok.
- .htaccess sözdizimi yerel Apache ile: **NOT TESTED** (kutuda Apache yok). php-cgi: NOT TESTED (yerine php -S cli-server SAPI).
- Kurulum: `cron/INSTALL_CRON.md` — A önerilir (connector ile yüklenip doğrulanabilir, çift koruma); B (public_html dışı) connector ile doğrulanamaz.

### 2) PANEL DEV R4B — 08 FAIL bulguları
- FX01 (MEDIUM) düzeltildi: TCMB tarihi Türkiye takvim günü (sabit UTC+3, Windows'ta tzdata gerekmez); bugün−5..bugün dahil; gelecek ve bugün−6 red; saat enjekte (`ParseFX(raw, now)`, panelde `fxNow`). `i18n.FXFresh` tek kural.
- FX02 (LOW) düzeltildi: kuruş tam sayı, `big.Rat` ile kesin bölme, yarım sentte sıfırdan uzağa; NaN/Inf/aralık dışı kur ve >10^15 kuruş → gizle (taşma yok).
- PANEL DEV'in 39 vakalık probu (`fx_review_r4b_test.go`, aynen) 08'e eklendi: **39/39 PASS**.
- Aynı kural site/cron/snapshot'a uygulandı; ortak vektör testi fx_policy_test.py (Python 21 + JS 26 + PHP 12): **59/59 PASS**.
- Canlı etkisi: canlı `i18nfx-5a375f29` eski kurala sahip (1 gün gelecek/6. gün kabulü, float yuvarlama). Sitede `data-kanus-try` öğesi olmadığından görünür etki YOK. Düzeltilmiş paket `site/release_us_fx2/` (YÜKLENMEDİ).

### 3) 08 — lisans sunucusundan kur
- Okunan: pinned/v10 licensesrv, US-kaynak.zip, API r11 + api-0001..0012 → fiyat/kur uç noktası yok; yalnız sabit `USDTRYRef=50.05`. İmza: HMAC-SHA256 + US_LICENSE_SECRET (yeni sır üretilmedi).
- Sözleşme: `FX_CONTRACT_FOR_US_API.md` — `GET /v1/fx`, kanus.fx/1 gövde, `X-Kanus-FX-Signature: v1=b64url(HMAC(secret,"kanus.fx/1\n"||body))`, `Cache-Control: public, max-age=3600`, geçersiz/eski → 503 no-store.
- Panel (08): HTTPS zorunlu, yönlendirme yok, 5 sn, ≤4 KB, imza+şema+tazelik; son geçerli yanıt `fx.json` (desk.json yanında, PrefStore ile aynı güvenli yazma: `ReadRaw/WriteRaw`), okunurken yeniden doğrulanır; sayfa ağı beklemez; başarısızlıkta yalnız TL; 50,05'e dönüş YOK. `pricingSummary()` ve {{US_FIYAT}} paragrafı 4 dilde i18n anahtarları (`price.*`); TL tutarları Catalog'dan; `/api/pricing?lang=` (GET) ile dil değişiminde güncellenir; sunucu çiziminde TR varsayılanı korunur. licensesrv/ödeme/ledger'a dokunulmadı.
- Sunucu önerisi: `global_fx_licensesrv_PROPOSAL.patch` (licensesrv/fx.go + test, UYGULANMAYACAK; ayrı kopyada vet + test PASS).

### Paket / yama sha256
| Dosya | sha256 |
|---|---|
| patches/global_i18n_08_fx.patch | 9cd86d6bb6aa2a852361ac28e7d745b304647e10bc5cd71680709fca1781d04b |
| patches/global_fx_licensesrv_PROPOSAL.patch | 25e20f9d28822f14d60c24eda52b430c9d7c74840edf44a2599bbd538c629171 |
| patches/global_i18n_03_pref.patch (değişmedi) | ba8845c5f34855d8f2f2fa49a055c4e9d20698c9fbd7572462531c2746f6b58e |
| release_us_fx/SHA256SUMS (cron/ güncellendi; Tur 11 ilk ara sürüm) | 6ae3708a8be1cf86c19653549504390efd7e30831968fcd5b539252c997d9b1f |
| release_us_fx2/SHA256SUMS | d713ebf473d98adac7797d0648840e5584ea4dd216d4154b105680de7373d7f6 |
| release_us_fx2 asset index-uscommerce-i18nfx2-dc1e3536.js | dc1e3536a4b96edf00a2443208efdee6ffbab6f02a5ee299076426c585e43f6d |
| release_us_fx2 index.html | bf768146cfc406eb654f13bfe9b5455b05a88be158183dbe9702c6f6dccf2163 |
| release_us_fx2 cron/upload_public_html/_fx/fx_cron.php | 3ff10ac51f88c35ef2a35f909553a0f156794ba92716d7ecff521db23b46724e |
| release_us_fx2 cron/upload_public_html/_fx/.htaccess | 6559fbacdc833562010b31370bc9d2aa39c1b28c8fadb7c08ada5e3fc916caf1 |

### Testler (pinned/v10 + core→02→03(ba8845c5)→04→06→08)
| Test | Sonuç |
|---|---|
| git apply --check + apply, 6 yama sırayla | PASS 6/6 |
| go test -race -count=1 -p=1 ./... | **473 PASS / 0 FAIL / 7 SKIP**, exit 0, DATA RACE 0 |
| (içinde) R4B FX probu | 39/39 PASS |
| (içinde) yeni panel FX testleri (TestFX*, TestPricing*) | PASS |
| §2 S2 regresyonu | **36/36 PASS** |
| go vet / GOOS=windows vet / GOOS=windows build | PASS / PASS / PASS |
| gofmt (08'in Go dosyaları) | PASS (boş) |
| check_ui.mjs / gui_lang_ui_test.mjs | PASS / PASS |
| fx_ui_test.mjs (headless Chrome, fx2; +gelecek tarih vakası) | **19/19 PASS** |
| ui_test.mjs regresyon (fx2) | 12/12 PASS |
| validate_us.py release_us / _fx / _fx2 | 21/0, 24/0, 24/0 |
| test_fx_snapshot.py / test_fx_cron.py / fx_policy_test.py | 12/12, 11/11, 59/59 PASS |
| .htaccess (Apache) / gerçek Windows koşusu / gerçek /v1/fx | NOT TESTED |
Not: ilk koşu yanlış tabanda (global-work/base) yapıldı: §2 7/29 ve 1 fixture FAIL; pinned/v10'da tekrarlandı (yukarıdaki sonuçlar). Ara düzeltmeler: DE/RU `price.try` "TRY" → "TL" (parite testi), sunucu çizimi TR varsayılanı (TestPricingTextFromCatalog). Loglar: `global-work/tur11/`.
