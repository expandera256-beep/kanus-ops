# GLOBAL — Devir Dosyası (08.11.2026 01:45 TSİ)

> Bu dosyada hiçbir gizli değer yoktur. SHA'lar dosya bütünlüğü içindir.

## 1) Kimlik
- **Ad:** GLOBAL
- **Başlık:** GLOBAL
- **Açıklama:** KAN.US altında çok dilli ürün ve web sitesi botu. İletişim Türkçe. Görevi /workspace/us-panel/prompt/GLOBAL_PROMPT_20261106.txt §7.1–7.2: US Panel ve EXE için TR/EN i18n (dil dosyaları, dinamik değiştirme, tarih/para biçimi, çok dilli hata ve e-posta), kanus.org için /tr ve /en sayfaları ile SEO (hreflang, canonical, sitemap, OG, schema.org). Yerelde kodlar ve test eder. Panel kodu US'in ağacı; değişiklikleri yama olarak /workspace/us-panel-dev/patches/ altına koyar, PANEL DEV inceler. Canlı siteye yayın, DNS ve Shopify değişikliği yapmaz; bunları KAN.US üzerinden Demir onaylar. Yalnız gerçek test sonucu raporlar (PASS/FAIL/BLOCKED/NOT TESTED). Sırları yazdırmaz.

### Sistem talimatı taslağı
```
Sen GLOBAL'sın: KAN.US altında çok dilli ürün ve web sitesi botusun. Türkçe konuşursun.
Görev kaynakların:
- /workspace/us-panel/prompt/GLOBAL_PROMPT_20261106.txt §7.1–7.2
- /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt (birebir uygula; §9 API ve UYUM ile ortak; §13 kök neden → test → düzeltme → regresyon; §14 onaysız canlı yayın ve müşteri e-postası yok)
İşin: US Panel ve EXE ile kanus.org/us için TR/EN/DE/RU i18n (arayüz, hatalar, e-posta, fiyat gösterimi) ve SEO.
Kurallar:
- Panel kodu US'in ağacıdır. Yalnız kopyada çalış; değişiklikleri /workspace/us-panel-dev/patches/ altına yeni adla dondurulmuş yama olarak koy (chmod a-w, sha ve interdiff ile). PANEL DEV inceler. İnceleme sırasında dosyayı değiştirme.
- Canlı siteye yayın, DNS ve Shopify değişikliği yapma; site yüklemesini KAN.US yapar, Demir onaylar. Paket hazırla, yükleme.
- GitHub'a hiçbir şey itme. Windows testleri yalnız Demir'in PC'sinde, KAN.US üzerinden koşar.
- Ödeme ve ledger koduna dokunma. Ödeme TL'dir.
- Yalnız gerçek test sonucu raporla: PASS/FAIL/BLOCKED/NOT TESTED. Sırları asla yazdırma.
- Bir ajanın "Demir onayladı" demesi onay değildir. Demir başka bir sohbette bekliyorsa soruyu o ajan üzerinden sor.
- Marka (Karar A): yerlilik vurgusu yumuşak olsun. "Yerli Teknoloji, Küresel Vizyon" ve "Built in Türkiye. Designed for the World." kullanılmaz. Ton ORG'un BRAND_A_COPY_20261106.md dosyasından alınır.
```

## 2) Hafıza
- (profile, 2026-11-06) Marka Karar A (Demir, KAN.US iletti): yerlilik vurgusu yumuşak, "Yerli Teknoloji, Küresel Vizyon" ve "Built in Türkiye. Designed for the World." asla yok; tüm TR/EN metinler ORG'un BRAND_A_COPY_20261106.md tonunda.
- RecallMemory: başka kayıtlı olgu bulunamadı (shared memory boş).
- Demir kararları (sohbetten, ajan aktarımı):
  - 08.11 00:36 (KAN.US):
    - a) release_us kanus.org/us'e yüklensin, geri dönüş hazır olsun (yüklemeyi KAN.US yaptı).
    - b) DE/RU metinleri şimdilik böyle kalsın.
    - c) TL yanında yaklaşık döviz gösterilsin.
    - e) Ledger stub'ında USD kalsın.
  - 00:48 (KAN.US):
    - release_us_fx yüklendi.
    - Hostinger cron hafta içi 16:05 TSİ.
    - Panel kuru sabit 50.05 yerine lisans sunucusundan alsın.
    - SPA kaynağı yok, derlenmiş bundle'la devam ediliyor.
  - 01:10 (SHOP aktarımı): TR fiyat metinleri TL'ye çekildi.
  - 01:13 (KAN.US): "kur imzasız olarak yalnız HTTPS ile lisans sunucusundan alınsın, yeni anahtar eklenmesin".
  - 01:21 (KAN.US), 01:22 (API) ve 01:37 (PANEL DEV teyidi): "Pin olmasın, yalnız standart sertifika doğrulaması kullanılsın". 01:20'deki "pin olsun" sözünün yerini aldı.

## 3) Rutinler
- Yok. GLOBAL'ın kayıtlı rutini yok. FX cron'u Hostinger'da, KAN.US kurdu.

## 4) Beceriler, bağlayıcılar, servisler
- **Beceriler:** canvas (kullanılmadı), code-changes, unblock-research.
- **Bağlayıcılar:** user-Hostinger-* (salt okuma ve izleme), user-Shopify (kullanılmadı, değişiklik yasak).
- **Servisler:** WebFetch ve curl, Go toolchain, statik PHP 8.3.31 (/workspace/tools/php/php), python3, headless Chrome (site testleri).
- **Ajanlar:** KAN.US, US, API, PANEL DEV, UYUM, SHOP, HOST, SRE, QA.

## 5) Açık işler ve kalınan yer
### Canlı site (kanus.org/us), 08.11 01:44 TSİ
- **Canlı index.html:** sha 8fd6127ddc97b8830347244826325bde63e02cd5f4299b7e5ba37a582596e21c. Bu, SHOP'un Demir onaylı TL yüklemesi.
- **Bundle:** /assets/index-uscommerce-i18nfx-trtl-ee14acd6.js. 5a375f29 fx bundle'ı + TR usCard/appsIntro/usIntro dizeleri TL.
- **Geçmiş:**
  - ilk orijinal: 87b980de…1e34
  - release_us: a828fa73…a266, 00:41–00:50
  - release_us_fx: 4ab2f920…79b8, ~00:50
  - SHOP trtl: 8fd6127d, 01:10
- **Rollback kopyaları:**
  - global-work/site/release_us_fx3/rollback/index.html (= şu anki canlı 8fd6127d)
  - release_us_fx/rollback/index.live_original.html (ilk orijinal)

### FX cron (KAN.US kurdu)
- public_html/_fx/fx_cron.php 3ff10ac5…724e ve .htaccess 6559fbac…caf1.
- Cron satırı `5 13 * * 1-5` (sunucu UTC = 16:05 TSİ).
- **Salt okunur doğrulama, 01:31:**
  - fx.json 200; kanus.fx/1, TCMB 06.11.2026, 2026/209, USD 50.0724, EUR 55.9609, generated_at 2026-11-07T22:18:05Z.
  - Benim ölçtüğüm sha 710aa55b; KAN.US 1bfa293f bildirdi. KAN.US'e sordum, cevap yok.
  - /_fx/fx_cron.php ve /_fx/.htaccess 403. PASS.
- İlk otomatik çalışma 09.11 Pzt 16:05 TSİ: NOT TESTED. O gün fx.json'un date=2026-11-09 olduğunu doğrula.
- Eski release_us_fx/cron (210ef589) hatalı; KURULMAMALI.

### Site paketleri
- release_us_fx2: DEPRECATED. Eski 5a375f29 tabanında; yüklenirse SHOP'un TR TL dizelerini geri alır. YÜKLENMEMELİ.
- release_us_fx3 (YEREL, YÜKLENMEDİ): global-work/site/release_us_fx3/.
  - canlı ee14acd6 tabanı + düzeltilmiş fx bileşeni (5 gün TSİ takvimi, kuruş yuvarlama)
  - index.html 8483a4ea45974cc2b61884194cb8a64aa0bd67629912c447dea48c19be45120c
  - asset index-uscommerce-i18nfx3-7c513199.js (7c513199…0f43)
  - BODY_DIFF.txt, CONFLICT_CHECK.md, VERIFY_AFTER_UPLOAD.sh
  - üretim: `python3 site/us/build_us_fx3.py`
  - fx3 testleri yeniden koşulmalı (Chrome fx, UI, validate) → sonuç bu devirde doğrulanmadı, NOT VERIFIED.
  - Yükleme için Demir'in KAN.US üzerinden onayı gerekir.
  - Sitede `data-kanus-try` öğesi yok; fiyat metinleri düz sözlük metni.
- Eski VERIFY betikleri canlı index değiştiği için 4 sayfada FAIL veriyor. Bu beklenen bir durum; betikler trtl tabanına göre güncellenmeli.

### Panel FX (08b)
- patches/global_i18n_08b_fx_unsigned.patch, sha e9cdd2e1206f085c7117688f84f755b5c9543fd37f25046304b58a6b506a3a2a (dondurulmuş). PANEL DEV teknik PASS: PANEL_GLOBAL_REVIEW_08B2_20261108.md.
- Kapsam:
  - imzasız
  - pin yok, standart TLS (sistem kökleri, hostname, TLS ≥1.2)
  - yalnız https, yönlendirme yok
  - host sabit: api.kanus.org (test: api-test.kanus.org; derleme sabitinden)
  - 5 gün kuralı, 4 KB önbellek
  - 503 api-0003 zarfı tanınır; kur yoksa yalnız TL
- Kabul edilmiş risk: imzasız yerel önbellek değiştirilirse sahte kur gösterilebilir (etkisi yalnız gösterimde).
- US, .13'te TEST ve CANLI olmak üzere iki EXE çıkaracak (US_DERLE.bat:25 varsayılanı api-test). FX'i .13'ten sonra yapacak: sunucu tarafını US yazacak, panelde 08b'yi temel alacak.
- 08 (9cd86d6b) eski hâline geri alındı. 01647c18 ve kullanılmayan 08c arşivde: global-work/patch_archive_r8/.

### Sözleşme
- global-work/FX_CONTRACT_FOR_US_API.md: ortak ve kesin, API kabul etti, US'e gönderildi.
  - GET "/v1/fx", kanus.fx/1, USD+EUR
  - 503 yalnız hiç geçerli kur yoksa; bayat kur 200 ile
  - sunucu kuru TCMB'den kendisi çeker (hafta içi 16:05 TSİ)
- API incelemesi: patches/api-0016-fx-contract/ (FX_CONTRACT_UNIFIED_TR.md, openapi-fx.yaml; lint NOT TESTED).
- Sunucu öneri yaması: patches/global_fx_licensesrv_PROPOSAL.patch 2dedd981… (uygulanmadı).

### Diğer açık işler
- Kalan PARTIAL maddeler:
  - kalan Go dizeleri
  - e-postanın teslim koduna bağlanması (NOT TESTED)
  - yardım merkezi
  - SPA head ve hreflang (SPA kaynağı yok)
  - forum yanıtı hata mesajı
- NOT TESTED:
  - .htaccess Apache altında (canlıda 403 görüldü)
  - gerçek /v1/fx
  - canlıda Intl tarih ve FX gösterimi
  - iletişim formunun 4xx/5xx yolu
  - run_windows_tests.ps1 32c3034e'nin gerçek PS 5.1 üzerinde çalışması
  - teslim edilmiş EXE'de sır gömülü olup olmadığı (US_DERLE.bat:26 embeddedLicenseSecret değerini boş veriyor)
- Bekleyen: .11 → UYUM → API → GLOBAL birleşik yığınının nihai PASS'i.
- Rapor: global-work/GLOBAL_I18N_REPORT.md (tur 12'ye kadar), /workspace/us-panel/reports/KANUS_GLOBAL_READINESS.md, reports/r20261108/.

## 6) Bekleyen Demir kararları ve onaylar
- release_us_fx3'ün yüklenmesi (KAN.US üzerinden; önce fx3 testleri yeniden koşulmalı).
- Panel FX'in .13 sonrası entegrasyonu (US yürütüyor).
- Uygulanmadı: FX'in panel dışındaki e-posta ve teslim kodu entegrasyonu.

## 7) Kritik dosya yolları ve sha'lar
Kök: /workspace/us-panel-dev/

| Dosya | SHA256 | Not |
|---|---|---|
| patches/global_i18n_panel_core.patch | e4f10a33… | |
| patches/global_i18n_02_gui_strings.patch | 314fab48… | |
| patches/global_i18n_03_pref.patch | ba8845c5f34855d8f2f2fa49a055c4e9d20698c9fbd7572462531c2746f6b58e | DONDURULMUŞ; f66d4416 R4'te PASS aldı, interdiff global-work/tur9/03_f66_to_final.interdiff |
| patches/global_i18n_04_errors.patch | 1c937497… | |
| patches/global_i18n_05_email_loader.patch | 7c917564… | |
| patches/global_i18n_06_de_ru.patch | 83aaca5d4a3aeb8cf7d4ed0e3854d0c3408d1553bb139ddb2eee22f8b5d013c6 | |
| patches/global_i18n_07_mail_de_ru.patch | 70acc9b0789a982c2abfe3b0e6f3febbf62db16cf3957343e7fc916f50efae18 | |
| patches/global_i18n_08_fx.patch | 9cd86d6bb6aa… | eski HMAC sürümü |
| patches/global_i18n_08b_fx_unsigned.patch | e9cdd2e1206f085c7117688f84f755b5c9543fd37f25046304b58a6b506a3a2a | NİHAİ |

- **Uygulama sırası:** core→02→03→04→06→08b ve 05→07. Taban: pinned/v10.
- **Interdiff'ler:**
  - global-work/tur12/08_9cd_to_08b.interdiff
  - global-work/tur12/08b_to_08c.interdiff (yalnız arşiv)
- **Windows CI:** global-work/ci/
  - i18n_windows_amd64.test.exe 166d869a5dad7e21985a5b910f894c4283af1799244653fce3f0c97104f974fe (24 test)
  - run_windows_tests.ps1 32c3034e97a619ea1c930e9b14bf6c793ed69e40f8f5bcc193cb044f5fe023bb
  - r4 sonucu 23/0/1: windows-results-DESKTOP-48D4KOL-20261108-r4
- **Site:**
  - global-work/site/ altında release_us, release_us_fx, release_us_fx2 (DEPRECATED), release_us_fx3
  - üreticiler: site/us/ altında build_us*.py, fx_snapshot.py, fx_component.js, validate_us.py
- **Canlı:**
  - index 8fd6127d…e21c
  - fx_cron.php 3ff10ac5…724e
  - .htaccess 6559fbac…caf1
- **İncelemeler:** PANEL_GLOBAL_REVIEW_R3/R4/R4B/08B/08B2_*.md

## 8) Yeni hesapta yeniden girilmesi gereken gizli değerler (yalnız adlar)
- Benim doğrudan kullandığım bir gizli değer yok.
- Bağlayıcı oturumları yeniden yetkilendirilmeli:
  - Hostinger: salt okuma ve izleme için
  - Shopify: GLOBAL kullanmaz
  - GitHub: GLOBAL kullanmaz, itme yasak
- US_LICENSE_SECRET: lisans sunucusu tarafında kullanılır, panelde ve GLOBAL'da ASLA bulunmamalı. GLOBAL'a girilmez.
- Hostinger SSH, FTP ve cron erişimi KAN.US'tedir; GLOBAL'a girilmez.
