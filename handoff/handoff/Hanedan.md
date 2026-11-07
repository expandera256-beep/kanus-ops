# Hanedan — Devir Dosyası

Yazıldı: 8 Kasım 2026, 01:45 TSİ (Avrupa/İstanbul, UTC+3)

## 1) Kimlik
- **Ad:** Hanedan
- **Başlık:** (boş)
- **Açıklama:** (boş)
- **Rol:** Demir'in kod ve GitHub/Cursor Origin işleri asistanı. Öncelikle Demir'in tarihsel strateji oyunu "Hanedan"ı (Next.js + TypeScript) geliştiriyor, planlıyor ve değişiklikleri Origin deposuna PR olarak alıyor.
- **Sistem talimatı taslağı:**
  > Sen Hanedan'sın. Demir'in kod ve GitHub işleri asistanısın. Her zaman Türkçe yaz; kısa, sıcak ve net ol. Ana işin Cursor Origin'deki `viaveritasvita/hanedan` deposundaki "Hanedan" oyununu geliştirmek. Yol haritası Canvas sayfasında duruyor (Faz 1–10, 41 madde). Depo işlerini bulut ajanı (CloudAgent) ile yap. main'e doğrudan push etme, force push yapma. Değişiklikleri dal + PR olarak, PR açıklamasını Türkçe yazarak gönder. Kimlik bilgisi, token ya da çerez arama, kopyalama ya da atlatma girişiminde bulunma. Tarihsel içeriği "taslak, doğrulanmamış" diye işaretle. Simülasyon deterministik (tohumlu `roll`, `commandLog`) kalmalı. Her değişiklikten sonra lint, typecheck, test, `npm run content:check`, `npm run pack:check`, build ve `npm run gate` çalışmalı. Demir'in adına mesaj gönderme.

## 2) Hafıza
Profile:
- (2026-10-24) Demir'in Cursor Origin hesabı `viaveritasvita`. Repoları: `hanedan` (özel, varsayılan dal main) ve `turkdini-alfa`.
- (2026-10-24) Bağlı GitHub hesabı `demirtumen08-design`. Repoları: New-project-2, project-2, project.
- (2026-10-24) Demir en çok kod ve GitHub işlerinde yardım istiyor.
- (2026-10-24) Demir Türkçe konuşuyor ve Türkçe yanıt istiyor.

Journal: Kayıtlı journal girdisi yok. Ortak (shared) hafıza da boş.

Çalışma geçmişinden olgular:
- 2026-10-24 → 11-06: Hanedan oyunu için 10 fazlık yol haritası hazırlandı ve tüm fazlar uygulandı. Toplam 44 commit, kutuda yerel olarak duruyor.
- Origin MCP'de yalnızca okuma araçları var. Push yapılamıyor. `git clone` kimlik bilgisi istediği için başarısız oldu, ağaç MCP okumalarıyla yeniden kuruldu.
- 2026-11-08: Bulut ajanı başlatıldı, yarıda hata verdi. Devam ettirme denemesi "Cursor kullanımı doldu" hatasına takıldı.

## 3) Rutinler
Rutin yok.

## 4) Beceriler, bağlayıcılar, servisler
- Beceriler: code-changes, source-control, canvas, unblock-research
- Bağlayıcılar: cursor-origin (yalnızca okuma), cursor-github
- Servisler: Cursor CloudAgent (bulut ajanı), Canvas/WriteCanvas

## 5) Açık işler ve kalınan yer
- **Ana açık iş:** Faz 1–10 değişikliklerini `viaveritasvita/hanedan` deposuna yeni bir dal ve Türkçe PR olarak almak. Henüz Origin'de değil.
  - Son bulut ajanı: https://cursor.com/agents/bc-50461edc-8816-5be9-8542-f8f44907cf4a. Yamayı 9fc9df6 üzerinde uygulayıp güncel main'e uyarlıyordu, hata verip durdu ve PR açmadı.
  - Engel: Cursor kullanım hakkı dolu. Devam için on-demand kullanımın açılması gerekiyor.
  - Alternatif: Demir yamayı kendisi uygulayabilir: `git checkout -b hanedan-fazlar && git am hanedan-tum-fazlar.patch && git push`
- **Kısmi kalanlar (sonraki iş adayları):**
  - Web Worker, IndexedDB ve harita bölme yapılmadı.
  - i18n yalnızca arayüz kabuğunu kapsıyor.
  - Tuna Bulgarları, Bolor ve Sind eklenmedi; haritada eyaletleri yok.
  - 519 eyaletin çoğunda not yok.
  - `/editor` sayfası ham JSON üzerinde çalışıyor.
  - Sesler yer tutucu.
  - Arayüz tarayıcıda elle test edilmedi.
- Tarihsel içeriğin tamamı taslak ve kaynakla doğrulanmadı: 751 Talas dönemi, Maniheizm/Nasturilik, biyografiler, yeni taçlar (Zabulistan, Şiwey, Keş ve Nahşeb) ve eyalet notları.

## 6) Bekleyen Demir kararları
- On-demand Cursor kullanımı açılsın mı? Açılırsa bulut ajanı tekrar başlatılıp PR açılacak. Açılmazsa Demir yamayı kendisi uygulayabilir.
- PR açıldıktan sonra inceleme ve merge onayı Demir'de.

## 7) Kritik dosya yolları ve sha'lar
- Origin deposu: https://cursor.com/codebase/viaveritasvita/hanedan (clone: https://origin.cursor.com/viaveritasvita/hanedan.git). main'in özgün HEAD'i: `9fc9df6774332fd5b389b3d647c71e10b98bcb80`
- Yerel depo: `/workspace/hanedan`, dal `faz-1-ci-kayit-test`
  - Taban: `bee92b85e81d1a411c6aef9ac16a10d815985f4c` (9fc9df6 ağacının yeniden kurulmuş hâli)
  - HEAD: `db193710098ebfcc308c65c7dbd4f33db7b72dff` (Faz 10 / madde 41)
- Tüm yama: `/workspace/hanedan-tum-fazlar.patch` (`bee92b8..HEAD`, 44 commit, yaklaşık 1,7 MB). sha256: `59354d0c825df51b44cc97f2c57b4e76af5c08fdc97ae01f66485c511335fac4`
- Commit başına yamalar: `/workspace/hanedan-patches/` (44 dosya)
- Commit listesi: `/workspace/hanedan-commit-listesi.txt`
- İlerleme günlüğü: `/workspace/hanedan-ilerleme.md`
- Gate kaydı: `/workspace/gate-faz10.log` (GATE OK, 42 ölçüt)
- Yol haritası Canvas'ı "Hanedan — Gelişim Yol Haritası": https://cursor.com/canvas/store-925da0d8-f24e-410a-a360-82cde9895f0b/197acbfe-4038-4c42-95fc-78c31598bb18
- Teknik notlar:
  - SAVE_VERSION 33; kayıt anahtarı `hanedan-save-v25`, yedek anahtarı `hanedan-save-v25-yedek`.
  - Gate betiği `scripts/autoplay-gate.ts`. `GATE_TICK_SCALE` varsayılanı 1, CI'da 2. Bu makinede özgün main'de bile Göktürk ortalaması 50 ms hedefini aşıyor.

## 8) Yeniden girilmesi gereken gizli değerler
- Bu bot kendi kullandığı bir gizli değer tutmuyor.
- Yeni hesapta yeniden bağlanması gerekenler:
  - Cursor Origin bağlantısı (viaveritasvita hesabı): depo okuma için.
  - GitHub bağlantısı (demirtumen08-design): GitHub işleri için.
  - Cursor kullanımı / on-demand: bulut ajanı başlatmak için.
