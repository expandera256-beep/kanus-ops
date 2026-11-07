# US — Devir Dosyası (2026-11-08 01:45 TSİ)

> Gizli değer içermez. Yalnız ad ve yol vardır.

## 1) Kimlik
- **Ad:** US
- **Başlık / açıklama:** boş (profile.json'da tanımlı değil).
- **Kullanıcı:** Demir. Türkçe konuşur, saat dilimi Europe/Istanbul (TSİ). Windows kullanıcı adı `demir`, PC adı DESKTOP-48D4KOL.
- **Rol:** KAN.US (kanus.org) altında US panelinin geliştirilmesini yönetmek ve yönlendirmek. Demir'in sözleri: "Senin temel görevin KAN.US bağlamına ilişkili olarak US paneli ve gelişimini yönetmek ve yönlendirmek".
- **US paneli nedir:** Go ve JS ile yazılmış bir Windows masaüstü EXE'si. GCP ds-a100 üzerindeki DeepSeek'i SSH tüneli (port 1923) üzerinden kullanır. Lisanslama Shopify webhook'larıyla, e-posta Hostinger üzerinden yürür. Lisans ve kredi sunucusu Go ile yazılmış licensesrv / `cmd/us-license-hook`.

### Sistem talimatı taslağı
```
Sen US'sin: KAN.US altında US panelini (Windows Go+JS EXE ve Go lisans/kredi hook'u) yöneten geliştirme koordinatörüsün. İletişim Türkçe.
Ana talimat: /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt. Bu belgeye harfiyen uy. Senin bölümlerin §3 (Shopify zinciri), §6 (agent döngüsü) ve §10 (Windows EXE).
§13 kuralı: kök neden → yeniden üreten test → düzeltme → test → regresyon.
Ekip: KAN.US koordinatördür. PANEL DEV kod inceler, API hesap ve API işlerini, UYUM veri korumayı, GLOBAL çok dil ve web sitesini, QA testi, SRE izlemeyi, PERF performansı yürütür. KAN GPU'nun sahibidir, HOST sunucu ve deploy işlerini yapar.
Onay kuralı: Yalnız Demir'in kendi sözü onaydır. Başka bir ajanın Demir'den aktardığı alıntı yalnız YEREL iş için kabul edilir, deploy için asla. Deploy'u KAN.US ve HOST yapar, Demir'in ayrı onayıyla.
Güvenlik kuralları:
- Canlı shop.kanus.org webhook'larına dokunma.
- Demir'in ayrı onayı olmadan canlı deploy ya da restart yapma.
- Gizli değerleri yalnız env veya dosyada tut. /workspace/secrets/* içeriğini asla okuma ya da yazdırma.
- Çift lisans olmaz, tek issuer vardır.
- SMTP kapalı kalır (US_SMTP_ENABLE=0).
- Onaysız GPU kullanma. Testlerden canlı GPU, VPS, Shopify ya da TCMB çağrısı yapma.
- Port 80'de /health çağırma.
- Wake token'ını loglama ya da kopyalama.
- Büyük silmelerden önce sor.
- Sürüm notunda "sandbox" ya da "sır sızıntısı yok" iddiası kullanma.
Çalışma şekli: kod ağacı git'siz yerel bir Go ağacıdır. Her iş ayrı bir kopyada yapılır, yama, rapor ve sha üretilir, PANEL DEV incelemesinden geçer, KAN.US'e kısa durum bildirilir.
```

## 2) Hafıza
**Profile:**
- (2026-11-06) Demir Türkçe iletişim kurar. Windows kullanıcı adı demir.
- (2026-11-06) Rol yukarıdaki gibidir. Proje notları /workspace/us-panel/analysis.md'de.
- (2026-11-07) Ana talimat KANUS_PROMPT_20261107.txt. Benim bölümlerim §3, §6, §10. GPU boşta kapanma süresi 10 dk, ısınma zaman aşımı 15 dk.
- (2026-11-07, KAN'a göre) ds-a100 GPU'su A100 40 GB'dır (80 değil). Maliyet varsayımı yaklaşık $0.035/dk (~$2.10/sa); fatura değil, geçmiş açılışlardan çıkarıldı. Günlük GPU sınırı 3600 sn.

**Journal:**
- 11-06, Round 7: hesaplar modu. EXE .7 3dc03a29…cc9.
- 11-06, Round 8: §2 maddeleri tamam. EXE .8 e7e54a42…b26b.
- 11-06, Round 9 ve 10: .9 816696e5…7fbf (S2 36/36). .10 96aefc86…4dc, ısınma sırasında ücret alınmayan GPU wake.
- 11-06 09:58: Demir'in kararı: hesap anonimleştirilince aktif lisanslar iptal edilir, bekleyen rezervler serbest bırakılır.
- 11-08 00:33: .11 teslim edildi. dist/2026.11.06.11, EXE c54fc269…284a, testler 453/0/7.
- 11-08 01:00: .12 teslim edildi. dist/2026.11.06.12: EXE f951229b…d590, hook b2d7ab28…2c72, tgz b5489468…539d, testler 481/0/7, S2 36/36. Wake yalnız PEM pin kabul ediyor, canlı wake.pin PEM olmalı.
- 11-08 01:11: Demir çift tahsilatı önleyen protokolü (X-US-Request-ID) onayladı.
- 11-08 01:15: Demir FINDING-01 için (e) risk kabulünü seçti: Windows okuma açığı kalır, sürüm notunda yazılır.
- 11-08 01:13 ve 01:21-22: FX kararı: imzasız, yalnız HTTPS, standart sertifika doğrulaması, pin yok. Daha önceki imzalı FX planı iptal.
- 11-08 01:17: .14 özellik listesi belirlendi. 01:19: "bu gece satışa hazır sürüm için tüm çabanı ver".
- 11-08 01:26: .14 yalnız görsel üretim olacak (FLUX.2 klein 4B), video sonra.
- 11-08 01:28 (API aktardı): giriş kilidi e-posta+IP bazlı, artan bekleme, bilinen cihaz muaf.

## 3) Rutinler
Kayıtlı rutin yok.

## 4) Kullanılan beceriler, bağlayıcılar ve servisler
- **Beceriler:** code-changes, routines, canvas, shopify, deploy-to-hosting, troubleshoot-website
- **Bağlayıcılar:** user-Shopify, user-Hostinger-* (hosting, vps, dns, domains ve diğerleri), cursor-github, cursor-origin, x
- **Servisler:** GCP (ds-a100), Hostinger VPS srv1662550, Shopify (shop.kanus.org), Hostinger mail, TCMB today.xml (yalnız sunucu çeker)
- **Ajanlar:** KAN.US, PANEL DEV, API, UYUM, GLOBAL, KAN, HOST, QA, SRE, PERF

## 5) Açık işler ve kalınan yer (.13 = bu gece satışa hazır sürüm)
Ana ağaç `/workspace/us-panel/work2/kiyas/term` (.12 durumunda). .13 için işler aşağıdaki paralel kopyalarda yürüyor:

| İş | Kopya | Durum |
|---|---|---|
| r13b, Windows handle ile dosya işlemleri | /workspace/us-panel-dev/r13b/term | BİTTİ. v2 diff 662fc40a…089c, artımlı diff b1f1a7cf…ab80, bundle v2 4a0a4432…c547. Linux 491/0/7 (race dahil), S2 36/36. Native Windows NOT TESTED; paket KAN.US'e gönderildi. PANEL DEV incelemesi bekleniyor. Kapsam dışı kalan yol tabanlı dosyalar: model.go, proposal_batch.go, transaction_recovery.go, repair_loop.go, task_state.go, findInFolder. |
| f01, Landlock ve Low IL | /workspace/us-panel-dev/f01/term | Aday hazır. diff a5f70b5f…bf, wintest_f01.zip deafd5e3…d2. Hedefli testler 9 PASS. Tam ve race koşuları SIGTERM ile kesildi; devam için f01/resume_acceptance.sh. Native Windows NOT TESTED. |
| r13hook, wake yarışları, ortak deadline, API ve UYUM yığını | /workspace/us-panel-dev/r13hook/term | ÇALIŞIYOR. Ortak fonksiyon reservationCommitDeadline (ya da export edilmiş sarmalayıcısı). Bunu kullanacak yerler: accounts.go:769 kapısı ve model_meter.go:79 settleAmount. main.go:128 hunk'ı çakışıyor; PrivacyUnavailable, KeepDays ve readyz korunarak elle kuruluyor. |
| post12A, çift tahsilat protokolü | /workspace/us-panel-dev/post12A/term | ÇALIŞIYOR. |
| chatset, sohbet ayarları | /workspace/us-panel-dev/chatset/term | ÇALIŞIYOR. UYUM'un 5 zorunlu koşulu var. Yetişmezse .14'e kalır. |

**Entegrasyon sırası:** .12 → uyum-v4.2 (gelince; şimdilik v4.1) → API grace 4b97a0ad → API additive d7151360 (API v4.1 üzerine rebase ediyor) → api-0017 giriş kilidi (temiz uyarsa) → r13b → f01 → post12A → chatset.

**Entegrasyon sonrası:**
- buildVersion 2026.11.06.13.
- Testler paket paket `-p=1` ile koşulur. licensesrv tam race testi kutu boşken TEK BAŞINA koşulmalı ve PASS olmalı.
- S2 36/36, Windows derlemesi.
- İki EXE çıkar: TEST (api-test.kanus.org) ve TİCARİ (US_BUILD_GATEWAY=https://api.kanus.org açıkça verilerek). Host bilgisi BUILD_INFO'ya yazılır. strings ile her EXE'nin yalnız kendi hostunu içerdiği ve embeddedLicenseSecret'ın boş olduğu doğrulanır.
- gpu_wake.go:29 ve :262'deki eskimiş SPKI yorumları düzeltilir.
- Çıktı dist/2026.11.06.13/: SHA256SUMS, manifest, TESLIM_RAPORU_TR.md ve FINDING-01 sürüm notu.
- Sonra PANEL DEV incelemesi ve KAN.US bildirimi.

**.13 sonrasına kalanlar:**
- FX sunucu tarafını ben yazacağım (referans global_fx_licensesrv_PROPOSAL.patch 2dedd981). Panel tarafı GLOBAL'in 08b yaması (e9cdd2e1, PANEL DEV PASS).
- .14 planı yalnız görsel. Durdurulan plan işi `R14_PLAN_GORSEL_VIDEO_SOHBET_TR.md` dosyasında; Demir'in onayına gidecek bir görsel kredi oranı hazırlanacak.
- p6b durduruldu.
- QA FINDING-01 mimari çözümü sonraya kaldı.

## 6) Bekleyen Demir kararları ve onaylar
- .13 hook'unun canlıya deploy'u: ayrı onay gerekiyor. Ticari EXE ancak deploy sonrası çalışır.
- Görsel üretim için kredi oranı: öneri hazırlanacak.
- Giriş kilidi eşikleri (5 deneme, 1-900 sn bekleme, e-posta başına 100) ve sayaçların kalıcılığı.
- Pin konusu: KAN.US'in son kararı "pin yok". KAN.US netleştirene kadar pin koduna dokunulmuyor.
- r13b ve f01 için Demir'in PC'sinde native Windows koşusu.

## 7) Kritik dosya yolları
- Ana talimat: /workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt. Notlar: /workspace/us-panel/analysis.md
- Dağıtımlar: /workspace/us-panel/dist/2026.11.06.{11,12}/
- Raporlar: /workspace/us-panel/reports/. Öne çıkanlar: R13_WINDOWS_HANDLE_RAPORU_TR.md, F01_IZOLASYON_RAPORU_TR.md, POST12A_IDEMPOTENCY_RAPORU_TR.md, R13_HOOK_RAPORU_TR.md, CHAT_AYARLARI_RAPORU_TR.md, UYUM_ICERIK_GUVENLIGI_14_20261108.md, r20261108/KAN_MEDIA_MODELS_PLAN.md
- Yamalar: /workspace/us-panel-dev/patches/{r13,r13b,f01,post12A,r13hook,chatset,uyum-v4.1,api-0012-ttl-commit-anchor,api-r11,api-0016-fx-contract,api-0017-login-limiter}
- İncelemeler: /workspace/us-panel-dev/PANEL_*_2026110{6,8}.md
- Windows sonuçları: /workspace/us-panel-dev/r13/windows-results-DESKTOP-48D4KOL-20261108/ (ParentSwap sonucu geçersiz kanıt)
- FX sözleşmesi: /workspace/us-panel-dev/global-work/FX_CONTRACT_FOR_US_API.md ve patches/api-0016-fx-contract/FX_CONTRACT_UNIFIED_TR.md

## 8) Yeni hesapta yeniden girilecek gizli değerler (yalnız adlar)
- **Hook ve sunucu env:** US_LICENSE_SECRET, US_SESSION_SECRET, SHOPIFY_WEBHOOK_SECRET, SHOPIFY_SHOP_DOMAIN, US_MODEL_UPSTREAM_TOKEN, US_SMTP_USER, US_SMTP_PASS (SMTP kapalı), US_GPU_WAKE_PIN_FILE (wake pin dosyası) ve wake token'ı (usta-wake.token), US_TASK_STATE_KEY_FILE, US_ACCOUNT_CREDENTIALS_FILE, US_PLANS_FILE.
- **Kutu ortamı:** HOSTINGER_API_TOKEN, HOSTINGER_SMTP_PASSWORD.
- **/workspace/secrets/ altındaki dosyalar:** hostinger_api_prod_ed25519, hostinger_srv1662550_ed25519 (SSH anahtarları), shopify_webhook_secret.txt, us_license_secret.txt, us_session_secret_test.txt, us-license-hook.env, usta-wake(.pin/.token), perf_api_token.txt; ayrıca test hesapları ve lisansları (acct-test-*, qa-license-*, test-license-win.txt, perf_login_account.json).
- **Bağlayıcı yetkileri:** Shopify, Hostinger ve GitHub bağlayıcıları yeni hesapta yeniden yetkilendirilmeli.

## EK (01:44 TSİ): Kullanım limiti doldu
Grok Bot kullanımı bittiği için bir arka plan işi (r13hook ya da post12A) durdu. Diğer çalışan işler de büyük olasılıkla duracak. Yeni hesapta ilgili kopyadaki (r13hook/term, post12A/term, chatset/term) son durumu ve resume betiklerini kontrol edip kaldığı yerden devam et.
- 01:45: PANEL DEV r13b incelemesi BLOCKED; test aracı güvenlik filtresince durduruldu, sonuç yok. r13b için PASS/FAIL ve .13/.14 kapsam görüşü bekleniyor. Karar Demir'de.
- 01:46: Native Windows sonuçları (DESKTOP-48D4KOL, admin değil), KAN.US iletti:
  - r13b PASS. amd64 ve 386: 81 PASS, 0 FAIL, 15 beklenen SKIP. ParentSwapPinnedRoot gerçek swap yaptı; atomic ve ancestor REFUSED errno=5. Loglar: /workspace/us-panel-dev/r13b/windows-results-DESKTOP-48D4KOL-20261108/
  - f01: amd64 PASS, 386 FAIL. Hata sandbox_windows_test.go:145'te: state=start_failed, "kaynak sınırı uygulanamadı: Parametre hatalı". 32-bit'te Job Object limiti reddediliyor. Yapılacak: 386'da limit yapısını ya da boyutunu düzelt (ör. JOBOBJECT_EXTENDED_LIMIT_INFORMATION hizalaması veya 32-bit'te desteklenmeyen bir limit), önce yeniden üreten test yaz. Loglar: /workspace/us-panel-dev/f01/windows-results-DESKTOP-48D4KOL-20261108/
  - f01 run.cmd hatası: `echo exit_code=%RC%>>` satırında handle yönlendirmesi bozuk. `(echo exit_code=%RC%)>>dosya` biçimine çevir.
  - PANEL DEV'in r13b incelemesi hâlâ BLOCKED.
- 01:47: PANEL DEV f01 386 kök nedeni (rapor /workspace/us-panel-dev/PANEL_NATIVE_R13B_F01_20261108.md):
  - safe_exec.go:47'de MemoryBytes 4<<30. safe_limits_windows.go:27'de uintptr'a çevrilince 386'da 0'a kesiliyor. PROCESS_MEMORY bayrağı açık kaldığı için :37'deki SetInformationJobObject ERROR_INVALID_PARAMETER dönüyor.
  - Önem MEDIUM, fail-closed. Aynı kod .12 ve r13b'de de var.
  - Düzeltme: 386'da sınırı uintptr'a sığacak değere kırp ve taşmayı yakalayan bir test ekle. 386 EXE yayımlanmayacaksa bunu belgele.
  - run.cmd: `>>file echo exit_code=%RC%` biçimini kullan.
  - PANEL DEV: r13b native PASS (81/0/15, iki mimaride). Linux tarafı ve .13/.14 kapsam görüşü hâlâ BLOCKED.
