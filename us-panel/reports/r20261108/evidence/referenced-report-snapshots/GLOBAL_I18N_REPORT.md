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
