# PANEL inceleme özeti — 08.11.2026

Saat dilimi: Europe/Istanbul (TSİ). Sıra: **A GLOBAL → B UYUM → C API**, her alt rapor biter bitmez kalıcı yazıldı. Mesaj/deploy/EXE çalışma yapılmadı; kaynak/pinned/yama dosyalarına yazılmadı. Kalıcı kanıt `/workspace/us-panel-dev/scratch/review-20261108/`.

## Sonuç / test sayıları
| İnceleme | Karar | Tam race PASS/FAIL/SKIP | §2 | Diğer |
|---|---|---|---|---|
| GLOBAL R4B (güncel) | 01–07 PASS;08 FX FAIL (MEDIUM tarih),FX entegrasyonu PARTIAL | **03:422/0/7;+08v10:425/0/7;+08v11:492/0/7** | **36/0/0** (03 yığını) | fd6→7;mutant OS testleri FAIL;FX39vaka30PASS/9FAIL;harici Windows23/0/1(log) |
| UYUM v4 (güncel) | PASS Linux R3 düzeltmeleri; 2 LOW açık, entegrasyon BLOCKED | **433/0/7**, exit0 | **36/0/0** | R3 probları v3 6PASS/14FAIL→v4 20PASS; ek35PASS |
| API tur3FIX | FAIL (MEDIUM grace taşması + v3 apply çakışması) | **412/0/7**, exit0 | **36/0/0** | parse16vakada13PASS/3FAIL;grace-beforehamFAILexit1;additivev2 7/0/0 |

Bu sayılar teslim testlerinin suite sonucudur; ayrı ek güvenlik probe'larındaki FAIL'ler gizlenmedi ve bu toplamlara dahil edilmedi. GLOBAL ilk fixture-yol koşusu414/1/7 idi; fixture yolu düzeltilerek tam suite yeniden415/0/7 alındı. Build/vet ve Windows vet/build üçünde PASS. Tüm ağaçta taban30gofmt dosyası var; GLOBAL/UYUM değişen dosyaları temiz;API bir ek dosya biçimsiz.

## Yama başına karar
| Paket | Durum |
|---|---|
| GLOBAL panel_core | PASS |
| GLOBAL02gui_strings | PASS |
| GLOBAL03pref ba8845c5 (final,salt-okunur) | PASS Linux;R3-01 kapalı;Windows kendi koşusu NOT TESTED,harici23/0/1 log incelemesi |
| GLOBAL04errors | PASS |
| GLOBAL05email_loader | PASS |
| GLOBAL06de_ru | PASS teknik;anadil redaksiyonu NOT TESTED |
| GLOBAL07mail_de_ru | PASS teknik;gerçek çokdilli SMTP entegrasyonu NOT TESTED |
| GLOBAL08fx 6ac3588b | FAIL:MEDIUM tarih+LOW rounding/sağlamlık;kur kaynağı kararı PARTIAL |
| UYUM01preferences v4 | PASS Linux |
| UYUM02ops v4 | PASS teknik .10 ara / US entegrasyonu BLOCKED |
| UYUM03handler v4 | PASS Linux; LOW port sağlamlığı açık |
| UYUM04hook-us v4 | PASS Linux / US entegrasyonu BLOCKED; LOW görünürlük |
| API0001/0003/0005/0006/0008/0006b | PASS teknik OrderB;0006bLOWgofmt açık |
| API0004US önerisi | teknikPASS / entegrasyonBLOCKED |
| API0012cap900 | FAIL / US entegrasyonuBLOCKED |
| API-r11additive | FAIL UYUMv3 üstünde apply;v2 üstünde araPASS |
| API0002/0009lint | NOT TESTED bu tur;eski lint dosyalarıvalid0warning |
| API0010/0011SDK | NOT TESTED bu tur;öncekiPASS yeni koşu sayılmadı |

## Açık bulgular
Satırlar her raporun ilgili scratch ağacına aittir.
| ID | Önem | Dosya:satır | Özet |
|---|---|---|---|
| GLOBALR4B-FX01 | MEDIUM | i18n/fx.go:46-52 | UTC parse ve-36h/+144h pencere;gelecek/6.gün kabul;4 negatif vakaFAIL |
| GLOBALR4B-FX02 | LOW | i18n/fx.go:84-94 |29kuruş/2→0.14(0.15yerine),21/6→0.03;uçint64 taşması;helperNonfinite |
| GLOBALR4B-01 | LOW | i18n/pref.go:149-155;pref_unix.go:115;pref_windows.go:241 | rootClosed kapanma girişimini sayar;Bounded tek başına close-sız mutantı kaçırır;OS testleri yakalıyor |
| GLOBALR4B-W01 | LOW | r4/environment.txt:1;binary-tests.log:26,51 | harici23/0/1logda actualsha/exit/buildmanifest yok;Scaling+6/+2,aktarılan+0değil |
| GLOBALR3-01 (R4 f66 KAPANDI; tarihsel) | LOW | i18n/pref.go:79-102;pref_unix.go:17-19;pref_windows.go:64-73 | Close/eviction yok;64root fd7→71;üretim sabittekroot etkisi sınırlı |
| UYUMR3-U01 (v4 KAPANDI; tarihsel) | MEDIUM | internal/privacy/prefs.go:37-58;prefs_open_unix.go:10-11;prefs_open_windows.go:7-9 | Lstat→open sonrası fstat yok;Unixnonblock yok/Windowsplainopen |
| UYUMR3-U02 (v4 KAPANDI; tarihsel) | MEDIUM | internal/privacy/prefs.go:133-157 | parent symlink üzerinden dış dizine yazı |
| UYUMR3-U03 (v4 KAPANDI; tarihsel) | MEDIUM | internal/privacy/prefs.go:155-157 | rename hatasında PII temp kalıyor |
| UYUMR3-U04 (v4 KAPANDI; tarihsel) | MEDIUM | internal/privacy/handler.go:253-261 | yanlış code/parola HTTP400 ama prefs kalıcı siliniyor (v2PASS→v3FAIL) |
| UYUMR3-U05 (v4 KAPANDI; tarihsel) | MEDIUM | internal/privacy/handler.go:49,63-70;cmd/us-license-hook/main.go:113-123 | 213504gün taşarak25m26.290448384s retention oluyor |
| UYUMR3-U06 (v4 KAPANDI; tarihsel) | LOW | internal/privacy/handler.go:109-114 | exact allowlist scheme/localhost politikasını atlıyor;listede olmayan benzer hostlar reddediliyor |
| UYUM kapsam (v4 KAPANDI; tarihsel) | LOW | licensesrv/privacy_ops.go:357-367 | whitespace/display-name/multi-recipient normalizasyonu yok;kanonik tek alıcı akışı dışı kapsam |
| UYUM yaşam döngüsü (v4 KAPANDI; tarihsel) | LOW | licensesrv/privacy_ops.go:72-80;store.go:351-358 | kapanan Store global eraseBooks'ta tutuluyor |
| APIR3-A01 | MEDIUM | licensesrv/store.go:246-261 (özellikle257) | Duration'a çarpma cap'ten önce taşıyor;3uç değerFAIL |
| APIR3-A02 | MEDIUM | api-r11 patch:177-185→cmd/us-license-hook/main.go:121-124 | yeni KeepDays satırı nedeniyle UYUMv3 üstünde mainhunkFAIL |
| APIR3-A03 | LOW | licensesrv/api_keys_test.go:243 | değişen dosyada gofmt farkı |

## Kapanan önemli konular
- GLOBAL symlink'li config ve UI500/503/network geri alma: eskiFAIL→yeniPASS;Windows NT API yalnız statik/cross-compile.
- UYUM R2-01/02HIGH unrelatedmail+Subject/OrderID temizliği, challenge çokStore race, failed-save code retry, aynıhesap lost-update ve trailingJSON eskiFAIL→yeniPASS.
- UYUM ledger/Orders/öncekiSettlements/Refunds korunuyor;revoke+reservationrelease audit'i var;iade yok;diske yazma hatası state'i geri alıyor. Handler sidecar atomikliği ayrıFAIL.
- API belgelerinde licensesrv sınırı ve US öneri paketlemesi düzeltilmiş;OrderB7/7apply;hamFAIL yeniden üretildi.

## NOT TESTED / BLOCKED
- Gerçek **Windows çalışma/NTFS/junction/ACL/rename** NOT TESTED; `.exe` çalıştırılmadı. KAN.US üzerinden Demir'in bilgisayarında test planlanacak; burada çalışma onayı verilmedi.
- macOS, gerçek tarayıcıE2E, SMTP/canlıproxy, secret saklama, güçkesintisi, canlıVPS/Shopify/GCP NOT TESTED.
- Redocly offline cache yok (**ENOTCACHED**);ağdan kurulum yapılmadı.
- UYUM02/04 ve API0004/0012 US-korumalı;merge BLOCKED. Başka botun aktardığı karar Demir'in doğrudan onayı sayılmadı.
- **.11→UYUMv3→APIr11→GLOBAL nihai birleşik yığın NOT TESTED/BLOCKED**. Hareketli US kaynakları sabit .11 yerine kullanılmadı.

## Raporlar
- PANEL_GLOBAL_REVIEW_R3_20261106.md
- PANEL_UYUM_REVIEW_R3_20261106.md
- PANEL_API_REVIEW_R3FIX_20261106.md

Üçü hem `/workspace/us-panel-dev/` hem `/workspace/us-panel-dev/reports/` altında. Pinned manifest başta/sonda181/181;UYUMteslimSHA17/17;GLOBAL03/04/05beklenenSHAeş.

## US .11 teslim kontrolü

Kontrol: 2026-11-08T00:21:49.567585+03:00 (Europe/Istanbul). `/workspace/us-panel/dist/2026.11.06.11`: **YOK**.
Dosya listesi yok (dizin mevcut değil). Hiçbir dosyasına dokunulmadı.

## Güncelleme: UYUM v4

PANEL_UYUM_REVIEW_R4_20261108.md günceldir; eski UYUMv3 ayrıntıları tarihsel. R3 teknik düzeltmeleri Linux PASS. Yeni LOW-R4-01 handler.go:103-123 geçersiz sayısal/boş origin portu exactallowlistten kabul; LOW-R4-02 main.go:107-124 prefsarıza durumunda client404+healthz200, operatörlogu var. NonLinuxkalıcıprefs bilinçli uygulanmamış; Windowspanelbuildi hookWindowspersistencevaadi değil. Sonuç433/0/7 ve§2 36/36. NihaiUSentegrasyonuBLOCKED.

## Güncelleme: GLOBAL R4 — sabit SHA sonucu / hareketli teslim uyarısı

Güncel ayrıntı: PANEL_GLOBAL_REVIEW_R4_20261108.md (hem kökte hem reports/).

- İncelenen03=f66d4416… +06=83aaca5d… +07=70acc9b0…: Linux teknik PASS. Core/02/04/05 ve06/07 katalog/render incelemesi PASS. fd64kök artışı +64→+1; refcount/Close concurrent probları PASS. Tam race420/0/7,§2 36/36; Python79/0/41 ve mail11/11.
- İnceleme sırasında teslim03 **ba8845c5f34855d8f2f2fa49a055c4e9d20698c9fbd7572462531c2746f6b58e** oldu. Yeni baytlar incelenmedi: **güncel teslim03 BLOCKED / NOT TESTED**. F66PASS'i buna uygulanamaz. Arşiv_r7'den incelenenF66scratch'te korundu.
- Windows DACL/hardlink NT çalışma NOT TESTED; yalnız statik ve cross-vet/build/test-c. Kök devralım DACL sınırı ve fallbackmemmap LOW yaşam-süresi notu raporda.
- DE/RU dört dilde key/placeholder paritesi PASS; anadil redaksiyonu NOT TESTED. Yeni mail kataloğu gönderime bağlı değil; plain-text rendering PASS, HTMLrenderer yok; gerçek çokdilli SMTP entegrasyonu BLOCKED/NOT TESTED.
- .11pinned snapshot'ına5/5sıralıcheck+applyPASS; incelenenF66i18n dosyaları byte-eş. Full .11/US/API birleşik kabulü başka executor'ün işidir, burada verilmedi.

## Son .11 teslim güncellemesi

08.11.2026 00:39 TSİ (Europe/Istanbul): `/workspace/us-panel/dist/2026.11.06.11/` artık VAR; `/workspace/us-panel-dev/pinned/v11/` de VAR. Önceki YOK kayıtları kendi zamanları için tarihsel gözlemdir.

Yalnız liste okundu:
- SHA256SUMS
- TESLIM_RAPORU_TR.md
- TREE_MANIFEST.sha256
- US_Ticari_Aday_2026_11_06_11.exe
- us-license-hook_hostinger.tgz
- us-license-hook_linux

Teslim dosyaları değiştirilmedi ve EXE çalıştırılmadı. Son pinned/v10 doğrulaması181/181; UYUMv4girdi21/21eş, GLOBALgirdi6/7eş (yalnız03 dış çalışma ile değişti). UYUMv4LinuxR3PASS433/0/7+36/36; LOWportkontrolü ve arıza404/health200 görünürlük notları açık; US entegrasyonuBLOCKED.


## Güncelleme: GLOBAL R4B final03 ve ek08 — 08.11.2026 00:56 TSİ

Güncel rapor PANEL_GLOBAL_REVIEW_R4B_20261108.md (kökte ve reports/). Tarihsel R4 “ba NOT TESTED/BLOCKED” durumu **03 Linux incelemesi için kalktı**: ba8845c5 SHA/0444 doğrulandı;bağımsız interdiff5dosya16hunk GLOBAL ile tam eş. Tam race422/0/7,§2 36/36;mutanttaBoundedPASS amaScaling/ClosedHandleInvalid/kendi64fdFAIL;kendi6→7fdPASS. Windows PANELDEV kendi koşusuNOTTESTED;hariciKAN.US/GLOBALlogu23/0/1,tekSKIPsymlink,Scaling+6/+2. CIbinary166d…hash eş ve yeni finalsembolleri var;buildmanifest/actualhashstdout/gerçekexittranskripti eksik;runnerbeklenen2PARTIAL gözlenmişexitdeğil.

08SHA6ac3588b sabit ve yalnıziki yeni dosya;pricing/ödeme/ledger değişmiyor. Ek v10/v11tamrace425/0/7 ve492/0/7,vet/gofmtPASS. KendiFX39yaprakvaka30PASS/9FAIL:4tarih,2rounding,3sağlamlık. **08FAIL ve bu nedenle08dahil GLOBALpaketiFAIL**;01–07teknikPASS ayrı tutulur. Kur kaynağı/entegrasyonPARTIAL(Demirkararıbekleniyor),canlıkurçağrısı yok. Sonpinnedmanifest181/181 ve205/205;03/08girdiler eş. .11distVAR,önceki6dosyaliste aynı;dosyalaradeğişiklik/EXEçalışması yok.


## GLOBAL08B yeniden inceleme — 08.11.2026 (TSİ)

İncelenen08 SHA9cd86d6b: **FAIL (iki LOW)**. Eski MEDIUM FX01 ve LOW FX02 kapandı: özgün39/39 ve ek13/13 PASS. Yeni bulgular: fx_panel.go:156-157 / i18n/fx_cache.go:27-29 geçerli4096-byte gövde base64/JSON zarfıyla cache4KB sınırına sığmıyor; gui.html:833-836 eski fiyat yanıtı yeni dili geri eziyor. Standart build/vet/Windows cross-build/vet PASS; race473/0/7, §2 36/36; UI eski harness4/4. HMAC alan ayırıcılı fakat istemci sır sahibi sahteFX imzalayabilir; Ed25519 önerisi. Son kontrolde yama başka aktörce değişmiş, güncelSHA 01647c1895bc98d082cd339aa6bfa12733293564d75fb242599a2f6499ad454a: **NOT TESTED**; eski hüküm yeni baytlara taşınamaz. Kaynak kararı KAN.US üzerinden iletilmiş/doğrudan doğrulanmamış. Mesaj/deploy yok. Rapor PANEL_GLOBAL_REVIEW_08B_20261108.md.


## Güncelleme: API R3FIX2 — 08.11.2026 01:22 TSİ

PANEL_API_REVIEW_R3FIX2_20261108.md (kök ve reports/) güncel. API önceki üç bulgu kapanışı PASS: özgün16 grace eski13/3→yeni16/16, clamp saniyede çarpmadan önce; 0006b gofmt temiz; rebased additive KeepDays çakışmasını çözüyor. Yeni lock testi değişmeden üretim deadline mutantında gerçek FAIL/exit1.

**API .12 entegrasyonu ve tam stack fonksiyonel kabulü FAIL:** yeni HIGH API-R3FIX2-01, accounts.go:769–770 TTL'i geçen rezervi usable sayarken store.go:256–260,354–359 grace boyunca kilitli tutuyor. İki chat alias×lisans/session 4/4 negatif ve3tekrar12/12 FAIL: gerçek reserve=ErrNoCredit ama gate=nil, wake=1; ledger500/500 değişmiyor, overspend gösterilmedi. Yalnız scratch karşı-mutasyonu aynı4vakada402/wake0 PASS.

.12→UYUM16→API(combined/additive/df1ae528 fix)3→GLOBAL(core/02/final03 ba8845c5/04/06/05/07)7 strict check+apply26/26 PASS;08 dışarıda. Normal test .11API528/0/7, .12stack592/0/7. **İki tam race SIGTERM143 ile eksik/BLOCKED**, ara465/0/6 ve482/0/6 tam suite sayısı değildir; licensesrv bitmedi. §2 iki ağaçta36/36. Linux/Windows vet/build ve Windows test-c PASS, binary execution yok; .12gofmt boş, .11yalnız eski30dosya format farkı (değişen24Go temiz).65girdiSHA değişmedi, pinned205/209 doğrulandı. Redocly/Windowsruntime/canlı ortam NOT TESTED; deploy onayı yok.
