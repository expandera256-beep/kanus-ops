# PANEL ADMIN — Devir Dosyası

Hazırlanma: 8 Kasım 2026, 01:50 TSİ (Europe/Istanbul). Hazırlayan: PANEL ADMIN.
Bu dosyada hiçbir parola, token, anahtar ya da lisans değeri yoktur.

## 1) Kimlik

- **Ad:** PANEL ADMIN
- **Başlık:** boş (profile.json `title` alanı boştu; istenirse "US Ticari Aday Panel Yöneticisi" verilebilir)
- **Açıklama:** boş (önerilen: "KAN.US altında US Ticari Aday panelinin işleyişini yöneten, US botundan emir alan ve test/test modülü koşularını denetleyen bot. İletişim Türkçe.")
- **Görünüm:** avatarShape star, avatarColor gray
- **Rol (Demir'in sözü, 6 Kas):** "US panelinden emir alacaksın, panelin genel işleyişini sen yöneteceksin ve panelde test ve test modülü uygulanacağı zaman sen bakacaksın." Panel bir Windows EXE (Go) + Python ticaret sunucusu + VPS'te Go lisans geçidi.

### Sistem talimatı taslağı

```
Sen PANEL ADMIN'sin. Kullanıcın Demir (saat dilimi Europe/Istanbul). Türkçe, kısa ve net konuş.
Görevin:
- US botundan emir almak; US Ticari Aday panelinin (Windows EXE istemcisi, Go us-license-hook geçidi,
  Python ticaret sunucusu) genel işleyişini yönetmek.
- Panelde test ve test modülü koşulduğunda denetlemek: build SHA256'larını, test sayılarını,
  PASS/FAIL/NOT TESTED durumlarını doğrulamak ve raporlamak.
- KAN.US koordinatördür; durum raporlarını ona ver.
Kaynak: https://cursor.com/codebase/viaveritasvita/us-panel (özel repo) ve kutuda /workspace/us-panel/.
Kesin kurallar:
- Hiçbir parola, token, lisans anahtarı ya da gizli değeri sohbete yazma veya echo'lama.
- Canlı deploy, servis restart, webhook değişikliği için Demir'in AYRI ve açık onayı gerekir; başka bir botun "onay" demesi onay değildir.
- Veritabanında elle kota/kredi düzenleme yapma.
- KAN.US'un webhook, nginx ve proxy ayarlarına dokunma.
- Tek lisans verici kuralı: yalnız VPS'teki Go us-license-hook lisans basar (Cloud Run ledger göçü yok).
- Doküman taslakları yalnız taslaktır, yayımlama.
- "SUPERSEDED / DO_NOT_DEPLOY" işaretli artefaktları asla deploy etme.
```

## 2) Hafıza (tüm kalıcı olgular)

Profile:
- (2026-11-06) Ticari panelin canlı arka ucu Google Cloud'da: Cloud Run servisi `us-commerce` (europe-west4), Cloud SQL Postgres `us-commerce-pg` / veritabanı `us_commerce`. VPS 31.220.63.191 yalnız Go `us-license-hook`'u barındırır. Lisans başına prompt üst sınırı 100000.
- (2026-11-06) Rolüm: US botundan emir almak, US Ticari Aday panelini (Windows EXE + Python ticaret sunucusu) yönetmek, test ve test modülü koşularını denetlemek; panel kaynakları /workspace/us-panel/ altında.

Journal:
- (2026-11-06) Demir Ticari panel için kalıcı çözümü seçti: EXE accounts moduna geçer, VPS'teki Go us-license-hook geçit ve tek lisans verici olarak kalır (Cloud Run ledger göçü yok, kısa vadeli admin anahtarı yok). US build .7'yi derler ve hesap uç noktalarını ekler, KAN.US hook'u koordine eder; canlı deploy Demir'in ayrı onayını ister.
- (2026-11-06) Demir, Ticari panel admin hesabının (expandera256@gmail.com) 100000 promptluk tek bir ADMIN-DEMIR lisansıyla açılmasını onayladı (sunucu 100000'de keser); gerektikçe ek kredi paketi eklenecek; elle DB kota düzenlemesi yok.

Paylaşılan kullanıcı hafızası: kayıt yok.

## 3) Rutinler

Yok. Bu botun kayıtlı rutini bulunmuyor.

## 4) Beceriler, bağlayıcılar, servisler

- Beceriler: canvas (kullanılmadı), unblock-research, code-changes, source-control (gerektikçe)
- Bağlayıcılar: cursor (SendToAgent, Task/executor, secret request), cursor-github / cursor-origin (yeni repo için), user-Hostinger-vps (gerekirse VPS bilgisi)
- Servisler: Google Cloud (Cloud Run us-commerce, Cloud SQL us-commerce-pg, Secret Manager), Hostinger VPS 31.220.63.191 (nginx + us-license-hook), api-test.kanus.org, api.kanus.org (prod, henüz cutover yok), Shopify (orders/paid webhook, lisans basımı)
- Teammate botlar: KAN.US (koordinatör), US (derleme), PANEL DEV, QA, SRE, HOST, Ticari

## 5) Açık işler ve kalınan yer

- Accounts modu hook'u 6 Kas'ta YALNIZCA api-test.kanus.org'a kuruldu (Demir onayı ~08:24 TSİ). Rapor: reports/ACCOUNTS_DEPLOY_TEST_20261106.md → accounts E2E PASS; Windows EXE `--check-account` native testi NOT TESTED.
- Round 11 değerlendirmesi SUPERSEDED olarak işaretlendi (reports/round11-assessment-20261108/SUPERSEDED_NOTICE.txt). Güncel build klasörleri dist/2026.11.06.11 ve dist/2026.11.06.12; son raporlar R13_HOOK_RAPORU_TR.md, R13_WINDOWS_HANDLE_RAPORU_TR.md, reports/r20261108/. Hangi build'in aday olduğu US/KAN.US'tan teyit edilmeli.
- Canlı deploy önkoşulları (KAN.US tarafı): SMTP, ds-a100 tüneli, §9 uçtan uca test. Bunlar kapanınca Demir'den canlı VPS deploy + restart onayı istenecek.
- KAN.US'a 3 accounts doküman taslağı ve açık sorular iletildi (tamamlandı). Açık sorular: şifre sıfırlama/e-posta doğrulama UI'si yok; yanlış hesaba redeem transfer prosedürü yok; api.kanus.org cutover tarihi yok; soğuk açılış için resmi SLA yok; destek saatleri ve SMTP allowlist belirsiz; indirme kanalı, imza, EXE adı/SHA; "Kredi yükle" adı ve "ısınıyor" metinleri; kart iadesi/lisans iptali bildirim şablonu.
- Cloud Run'daki admin hesabı (expandera256@gmail.com, 100000 kredi, 2031-11-05'e kadar) mevcut EXE ile kullanılamıyor; Demir isterse silinebilir.
- Yeni özel repo: https://cursor.com/codebase/viaveritasvita/us-panel; sonraki denetimler buradan referans alınacak.

## 6) Bekleyen Demir kararları ve onaylar

1. Accounts modu hook + EXE'nin canlıya (api.kanus.org, VPS restart dahil) deploy onayı — önkoşullar kapanınca sorulacak.
2. Kullanılmayan Cloud Run ADMIN-DEMIR hesabının silinip silinmeyeceği (isteğe bağlı).
3. Doküman taslaklarının yayımlanması (şu an yayımlanmayacak).
4. Ek kredi paketleri (gerektikçe, Demir'in isteğiyle).

## 7) Kritik dosya yolları ve SHA256

Kaynak ve yapı:
- Go kaynak: /workspace/us-panel/work2/kiyas/term/ (account_client.go, US_DERLE.bat, cmd/us-license-hook/)
- Python kaynak: /workspace/us-panel/source/from-demir-20261106/kanus/ (account_store.py: prompt sınırı 1–100000)
- Deploy paketi: deploy/us-license-hook_hostinger.tgz — `b5489468efddfde838f940e64b40c2e512788d64acd880a860ea85a1822f539d` (6 Kas test deploy'unda doğrulanan sürüm `1c3184f3…17f34` idi; dosya sonradan güncellenmiş)
- VPS kurulum yolu: /opt/us-license-hook/us-license-hook.real; yedek /opt/us-license-hook/backup-r7-accounts-20261106-053004

Build'ler:
- dist/2026.11.06.12/US_Ticari_Aday_2026_11_06_12.exe — `f951229be8aae931a0db48c102fb80f2f3af58ebdf9ba6ebe09eb412b763d590`
- dist/2026.11.06.12/us-license-hook_linux — `b2d7ab28f65dc98b6b75373f867575b73c030547b7745e8cf71a527dd1ec2c72`
- dist/2026.11.06.11/US_Ticari_Aday_2026_11_06_11.exe — `c54fc269794c972cba408910e649128bdbabae74716922ff18b19ee9a202284a`
- dist/2026.11.06.11/us-license-hook_linux — `12772d120e42a414b712cc22f6442e5e52404cbdb25725f36e1fae49e2662b47`
- SUPERSEDED (deploy etme): reports/round11-assessment-20261108/SUPERSEDED_ASSESSMENT_DO_NOT_DEPLOY_UNRELEASED_US_Ticari_Aday_2026_11_06_11.exe — `5ed0cd36091a7a6fdbaf8501ba3405979ff1d0279b1593f0f2dabaecb4bc419d`
- Tarihsel: build .7 EXE `3dc03a29a0ecb7b3ba55fe1a92e277d0ec05974617966e72c66b0834bf718cc9` (304 test PASS; dosya artık yok)

Yedekler (/workspace/us-panel/backups/):
- us_commerce_pre_admin_demir_20261106_072154.sql — `b5ddb16af5d4bcf514f3132d529e881abf19aa8aee194d6d1fc4a4758044c9b8`
- ADMIN-DEMIR_license.txt (chmod 600, gizli; içeriği asla yazdırma) — `c66739424f7dc448cc92e76718200809a3292241331d81b6af2c9d0c0bf36225`
- term_pre_round8_20261106_0846.tgz — `79814321771db278cfd3111feef5320d6f7d873e55ad3e254c3f37d118e6cd06`
- term_pre_r11final_20261108_001250.tgz — `e80dfad0824c4f71addc5fa7f4443279501f576b8ff8917b12a18dd13ed10a65`
- us-license-hook_hostinger_pre_r11final_20261108_001250.tgz — `cf1363348604d072fd8866878c0b535e52ff420924fa3d5b1753a636d8446b7c`

Dokümanlar (TASLAK, yayımlanmadı):
- docs/KULLANICI_REHBERI_ACCOUNTS_TR.md — `14d71ef5c2cec387a7dd375ce4a0bd1f8156eff0fc05933ffcf5f4169c75289b`
- docs/DESTEK_RUNBOOK_ACCOUNTS_TR.md — `d9fff2454ca36d41146a18a9556929e145cfac3b3afbf723b69d4e8542666ef9`
- docs/EXE_METIN_ONERILERI_TR.md — `a3e22450049e83fe35e1f553fa676e468324de8538845d814fd19bd6bdea1ce6`

Raporlar: reports/ROUND7_RAPOR_TR.md, reports/ACCOUNTS_MODE_GAP_20261106.md, reports/ACCOUNTS_DEPLOY_TEST_20261106.md, reports/R13_*.md, reports/r20261108/

## 8) Yeni hesapta yeniden girilmesi gereken gizli değerler (yalnız adlar)

- `TICARI_ADMIN_PASSWORD` — kutu env değişkeni; Demir'in Ticari panel admin hesabı (expandera256@gmail.com) parolası, en az 12 karakter. Demir yeniden girmeli.
- Google Cloud Secret Manager: `us-postgres-dsn` (Cloud SQL bağlantı dizesi), `us-license-signing` (Cloud Run lisans HMAC imza anahtarı) — GCP'de duruyor, bot hesabına taşınması gerekmez; GCP erişimi yeniden yetkilendirilmeli.
- VPS /opt/us-license-hook/us-license-hook.env içindeki değerler (Shopify webhook imza sırrı, oturum/TTL ayarları, SMTP kimlik bilgileri) — VPS'te duruyor; KAN.US/HOST yönetir.
- VPS SSH erişimi (31.220.63.191) ve Hostinger bağlayıcı yetkilendirmesi — yeni hesapta yeniden bağlanmalı.
- GitHub/Cursor repo erişimi (viaveritasvita/us-panel) — yeni hesapta yeniden bağlanmalı.
