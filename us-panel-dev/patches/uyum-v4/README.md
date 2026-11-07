> **v4.1 yerini aldı**: PANEL DEV R4 iki LOW düzeltmesi `/workspace/us-panel-dev/patches/uyum-v4.1/` altında; v4 yamaları silinmedi. (08.11.2026 Europe/Istanbul)

# UYUM v4 — PANEL DEV R3 düzeltmeleri

Tarih: 08.11.2026 (Europe/Istanbul). **Yerel öneri; canlıya uygulanmadı. Hukuki uyum tamamlandı anlamına gelmez.**

- Sabit taban: `/workspace/us-panel-dev/pinned/v10` — değiştirilmeyecek.
- Çalışma: `/workspace/us-panel-dev/work/uyum/v10_privacy_v4`.
- Teslimin temiz uygulanmış kopyası: `work/uyum/v10_privacy_v4_apply`.
- Kanıtlar: `/workspace/us-panel-dev/scratch/uyum-v4-evidence/`.
- v3 korunur; bu sürüm v3'ün yerini alır. Yamalar **v3 üstüne değil temiz v10/.11 tabanına** uygulanır.

## Sıra ve sorumluluk

| Sıra | Klasör | İçerik / önkoşul |
|---|---|---|
| 1 | 01-preferences | iface, Linux descriptor IO, diğer platformlarda açık fail-closed sınırlaması, tercihler ve testler |
| 2 | 02-ops | **US alanı**: privacy_ops + minimal Store.Close temizliği; 01'den bağımsız derlenebilir |
| 3 | 03-handler | HTTP; 02'nin ConfirmEraseWithCleanup metodunu kullanır; prefs implementasyonu zorunlu değil |
| 4 | 04-hook-us | **US alanı**: küçük main.go mount/config kancası ve test; 01/02/03 ardından |

```bash
cd /path/to/clean-v10-copy
for dir in /workspace/us-panel-dev/patches/uyum-v4/0[1-4]-*; do
  for p in "$dir"/*.patch; do patch -p1 --batch --fuzz=0 < "$p" || exit 1; done
done
```

US alanına hiçbir merge yapılmadı. 02 ve 04, US'e öneridir. API botunun kaynak/yama dosyaları değiştirilmedi. Sabit nihai .11 + API birleşimi ayrı US/PANEL DEV incelemesi gerektirir; `r11_base --check` bunun yerine geçmez.

## R3 bulgu → düzeltme → test

| Bulgu | Düzeltme | Teslim testi / durum |
|---|---|---|
| R3-U01 MEDIUM | Linux: tüm parent zinciri fd üzerinden openat; leaf O_NOFOLLOW + O_NONBLOCK; açılmış fd Fstat ile regular/1MiB kontrolü. Lstat→open penceresi yok. | TestV4PrefsFIFOOpenedFDRejectedWithoutBlocking; mevcut leaf symlink/FIFO testleri; TestV4PrefsLeafSwapRaceNeverReadsOutside |
| R3-U02 MEDIUM | / kök fd'sinden her bileşende O_DIRECTORY/O_NOFOLLOW ile ilerle; mkdirat, temp openat, renameat ve unlinkat yalnız pinned parent fd üzerinden. Yol üstündeki symlink reddedilir. | TestV4PrefsParentSymlinkReadAndWriteRejected; TestV4PrefsPinnedDirectorySurvivesSwap |
| R3-U03 MEDIUM | Her pre-rename hata yolunda defer unlinkat; dosya fsync, Close hatası kontrolü, renameat ve dizin fsync. Yeni dizinin parent'ı da fsync edilir. | TestV4PrefsFaultsCleanTempsAndRollback (write/sync/rename enjeksiyonu + gerçek rename hatası); TestV4PrefsDirectorySyncFailureIsExplicitCommittedState |
| R3-U04 MEDIUM | Kod/parola/süre/HMAC/session doğrulaması ve cleanup sırası aynı Store.mu altında. Geçersiz doğrulamada cleanup hiç çağrılmaz. | TestV4InvalidConfirmationPreservesAccountAndPreferences (HTTP400 + disk/bellek değişmez); TestV4CleanupRunsOnlyAfterValidationAndRetries |
| R3-U05 MEDIUM | Ortak config helper: 1..3650 gün; boş env/KeepDays=0 varsayılan90. Geçersiz/negatif/taşan env/int girdisi90 + sabit uyarı (ham değer loglanmaz). Süre çarpımından önce sınır/overflow kontrolü. | TestV4RetentionBounds; TestV4MountRetentionBounds; mevcut TestPrivacyRetentionDaysMounted |
| R3-U06 LOW | Önce origin yapısı, scheme ve localhost/IP doğrulaması, sonra allowlist. HTTP yalnız tam localhost veya loopback IP ve açık allowlist kaydıyla. HTTPS aynı Host veya allowlist. URL userinfo/path/query/fragment reddi. | TestV4OriginAllowlistCannotBypassPolicy; mevcut TestOriginSchemeHTTPSOnly |
| LOW alıcı normalizasyonu | net/mail ParseAddress/List, trim ve kesin case-insensitive mailbox eşleşmesi. Display name kimlik değildir, substring yok. Çoklu To'da yalnız eşleşen adreslerin adı/adresi anonimleşir; diğer alıcıların mailbox/display name bilgisi korunur. | TestV4MailNormalizationAndSharedRecipients; mevcut unrelated/OrderID testi |
| LOW challenge bellek tutma | Close, Store.mu altında eraseBooks kaydını bırakır. Kapalı Store erase isteğini reddeder; map yeniden oluşmaz. | TestV4CloseDropsChallengeBookAndRefusesRecreation; mevcut multi-Store race |
| U-07 / 02+04 entegrasyonu | Klasör ayrımı, minimal Close/main değişiklikleri, US'e öneri. | **BLOCKED: US entegrasyonu / sabit nihai .11 beklenir.** |

### Erase tasarımı ve tekrar deneme

`ConfirmEraseWithCleanup` kilidi alır; session, challenge hash/süre, parola ve HMAC yapılandırması doğrulanmadan **hiçbir tercih silinmez**. Callback yalnız accountID alır; Store'a geri çağrı yapmamalıdır. Doğrulama başarılıysa `Prefs.Delete` (idempotent) çalışır; ardından lisans/account işlemi mevcut `transactionLocked` ile kalıcı yazılır. Kod ancak bu kalıcı kayıt başarılı olduktan sonra tüketilir.

- Yanlış kod/parola/süresi geçmiş kod: HTTP400, hesap ve tercihler (disk+bellek) değişmez.
- Tercih temizliği hatası: HTTP500, account/challenge korunur; aynı onayla yeniden denenir.
- Tercih silindi ama ledger kaydı başarısız: ledger geri alınır, challenge korunur; yeniden deneme idempotent tercihler temizliğini tekrar çalıştırır. **İki ayrı dosya tek transaction değildir; bu hata halinde tercihler önceden silinmiş olabilir.**
- Doğrulama ile cleanup arasında ayrı Verify çağrısının getireceği TOCTOU yok; Store kilidi doğrulamadan kalıcı kayda kadar tutulur.
- Başarılı erase eski session'ı geçersiz kılar; aynı HTTP isteğinin tekrar200 dönmesi şeklinde bir idempotency protokolü vaat edilmez.
- RequestErase challenge'ı ancak outbox kalıcı kaydı başarılı olduktan sonra belleğe koyar. Challenge'lar restart'ta kaybolur; yeni onay isteği gerekir.

### Dosya güvenliği ve dayanıklılık sınırları

- Go modülü **1.22**; os.Root kullanılmadı. Zaten go.mod'da bulunan `golang.org/x/sys v0.30.0/unix` kullanıldı; yeni bağımlılık/vendor değişikliği yok.
- Linux üzerinde gerçek descriptor tabanlı persistence vardır. Tüm parent bileşenlerinde symlink reddedildiğinden yapılandırılmış yol gerçek, symlinksiz bir yol olmalıdır (örn. `/var/lib/kanus/privacy/prefs.json`). Yapılandırma HTTP'den alınmaz. Hardlink / yetkili dosya sistemi yöneticisine karşı ilave izolasyon iddia edilmez.
- **Windows ve diğer non-Linux platformlarda kalıcı prefs açık hata ile kapalıdır.** Plain os.Open/reparse-follow fallback yoktur. Native Windows handle-relative/reparse-safe persistence henüz uygulanmadı; Windows çalışır persistence PASS iddiası yoktur. Çapraz vet/build başka, runtime güvenliği başkadır. Boş yol yalnız bellek içi boş store açabilir; diske Put/Delete desteklenmez.
- Configured prefs açılamazsa 04 kancası erase rotalarını kapalı tutar; mevcut sidecar'ı atlayarak hesap silmez. Preferences503; export prefs olmadan çalışabilir. Prefs yolu hiç yapılandırılmamışsa opsiyonel prefs olmadan erase hâlâ mümkündür.
- Rename öncesi hata: temp temizlenir ve bellek geri alınır. Rename sonrası **directory fsync** hata verirse commit görünür ama kalıcılığı belirsizdir: açık hata döner, görünür dosyayla bellek tutarlı bırakılır; idempotent retry mümkündür. Güç kesintisinde iki dosya atomikliği vaat edilmez.
- Swap testi hem gerçek leaf yarışını hem parent fd alındıktan sonra parent yolunun symlink ile değiştirilmesini kapsar. Sonuç pinned dizine yazılır, symlink hedefi değişmez.

### Çoklu alıcı ve PII kapsamı

Mail ilişkisi yalnız normalize mailbox eşleşmesi veya hesaba ait OrderID ile kurulabilir; Body∋OrderName heuristiği yoktur. İlişkisiz kayıt byte/anlam olarak korunur. Paylaşılan bir mailin gövdesi kişi bazında güvenle ayrılamadığından ilişkili kaydın Subject/Body/OrderName/LastErr kişisel içeriği tümden temizlenir ve **tüm kayıt gönderime kapatılır (`Skipped=anonymized`)**; unrelated alıcı adresleri değişmez. Bu, diğer alıcıya teslimatı koruduğumuz anlamına gelmez: shared kayıt tekrar gönderilmez. Bozuk/parçalanamayan To tek başına ilişki ispatı değildir; OrderID ilişkisi varsa bozuk To boşaltılır ve gönderilmez. Her keyfi string'den bütün PII'nin çıkarıldığı iddia edilmez.

### Retention ve önceki kararlar

1..3650 gün (yaklaşık10 yıl), süre taşmasını ve yanlış operasyon ayarlarını engelleyen **teknik sınırdır; kesin hukuki saklama süresi değildir**. Geçersiz değer90 güne döner ve log uyarısı üretir; mali kayıtlar bu sweep ile silinmez. Mali saklama/sürelerin uygunluğu **HUKUKİ İNCELEME** gerektirir.

Aktif lisanslar mevcut `revokeLocked` ile iptal edilir; rezervasyonlar release edilir; License.Email anonimleşir ve LicenseOwners bağı kaldırılır. Orders/önceki Settlements/Refunds korunur. Otomatik para iadesi **yok** (`refundTriggered:false`). Kullanılmamış süre iadesi, mesafeli sözleşme ve ticari sonuçları hukukçu/ürün kararıdır.

`US_PRIVACY_HMAC_KEY`: HOST tarafından VPS env dosyasında mod0600 tutulacak. Üretim anahtarı üretilmedi/okunmadı/yazdırılmadı. Testlerde yalnız sabit sentetik fixture değerleri vardır. Gerçek e-posta/model/canlı sistem kullanılmadı.

## R3 kanıt yolları

Raporun üst düzey `scratch/review-20261108/v2-probe/` ve `v3-probe/` yolları yok. Gerçek dosyalar `scratch/review-20261108/B/v2-probe/` ve `B/v3-probe/` altında. `B/fifo_helper.log`, `B/atomic.log`, `B/v3_negative.log` ve probe testleri okundu. V4'ün yeni kanıtları ayrı dizindedir, eski kanıtlar değiştirilmedi.

## Doğrulama

Aşağıdaki sonuçlar ham loglardan alınmıştır (`summary.json`). Race komutu etkilenen tüm paketler `./licensesrv ./internal/privacy ./cmd/us-license-hook` üzerindedir; vet ve Windows çapraz derleme ayrıca temiz uygulanmış ağacın tamamında çalıştı.

| Komut / kontrol | Sonuç | Kanıt |
|---|---|---|
| `go vet ./...` | PASS / exit0 | `vet-all.log`, `vet-all.exit` |
| `GOOS=windows go vet ./...` | PASS / exit0 | `windows-vet-all.log` |
| `GOOS=windows go build ./...` | PASS / exit0 | `windows-build-all.log`, `windows-all.exit` |
| `go test -race -count=1 -p=1 -json ./licensesrv ./internal/privacy ./cmd/us-license-hook` | **125 PASS / 0 FAIL / 1 SKIP**, exit0 | `race.jsonl`, `race.exit` |
| R3 bağımsız probe'ları: `go test -race -count=1 -json -run '^TestReview' ./internal/privacy ./licensesrv` | **18 PASS / 0 FAIL**, exit0 (alt testler dahil) | `r3-probes.jsonl`, `r3-probes.exit` |
| Temiz pinned/v10 kopyasında 16 patch, 01→04, `patch -p1 --batch --fuzz=0` | PASS; her klasör sonrası üç paket `go test -run '^$'` derlemesi PASS | `cumulative-apply.log`, `stages.json` |
| 02-ops tek başına temiz taban + `go vet ./licensesrv ./cmd/us-license-hook` | PASS | `ops-standalone.log` |
| Her patch için `/workspace/us-panel-dev/r11_base` üzerinde `git apply --check` (salt okunur) | PASS; ayrıca ayrı .11 kopyasına birikimli check+apply PASS | `r11-check.log`, `r11-check.exit` |
| Değişen Go dosyalarında gofmt | PASS / boş çıktı | `gofmt.log` |
| pinned/v10 öncesi/sonrası SHA256 | **181/181 değişmedi** | `pinned-before.json`, `pinned-after.json` |
| Çalışma ağacı / temiz uygulanmış teslim dosyaları | Birebir aynı | `summary.json` |

Ortak ortam: GOPROXY=off, GOSUMDB=off; gerçek model/network test bayrakları0. R3 probe'ları `work/uyum/v10_privacy_v4_review` adlı ayrı kopyada çalıştırıldı; teslim yamalarına eski probe dosyaları eklenmedi. Yeni V4 regresyon testleri teslim yamalarında vardır. Race sayıları alt testleri de kapsar. Tek SKIP, mevcut `TestStoreWriterProcess` alt-süreç yardımcı testi (doğrudan çalıştırıldığında helper koşulu yok); gerçek lock testleri suite içindedir, ham kayıtta gerekçesi vardır. §2'nin36 senaryosu bu tur ayrıca çalıştırılmadı; eski R3 sonucunu V4 sonucu gibi sunmuyoruz.

**Açık kalan / test edilmemiş:** US'in 02/04 kabulü ve nihai sabit .11/API yığını; native Windows/non-Linux kalıcı prefs implementasyonu ve gerçek çalışma; canlı CORS/proxy/SMTP, güç kesintisi. Linux R3 teknik regresyonları yeni testler ve bağımsız R3 probe'larıyla geçti; bağımsız PANEL DEV onayı henüz alınmadı.
