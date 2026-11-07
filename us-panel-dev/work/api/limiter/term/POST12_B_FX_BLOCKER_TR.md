# B: Ed25519 FX anahtar önkoşulu

Durum: NOT DONE (yalnız TL-only güvenli fallback DONE).

.12 doğrulandı; tüm kaynakta crypto/ed25519, Ed25519 private key yükleme veya gömülü public key yok. Lisans üretimi/parseLicense ve licensesrv webhook mint'i HMAC-SHA256 kullanıyor. Config.LicenseSecret string'i bir Ed25519 anahtarı değildir. Gömülü embeddedLicenseSecret public key değil, legacy HMAC secret alanıdır; ticari hook modunda boştur. Sır dosyası okunmadı.

İstenen EXISTING Ed25519 anahtarını ve EXE'de zaten gömülü public key'i yeniden kullanmak bu kaynak sürümünde doğrulanamıyor. HMAC secret'tan Ed25519 seed türetmek, ayrı key yaratmak, algoritmayı HMAC'a çevirmek veya EXE'ye HMAC secret dağıtmak yapılmadı. Sahte key slotu/çalışmayan doğrulayıcı eklenmedi.

Devam için mevcut Ed25519 lisans imzalayıcısının ve EXE public-key çözümleyicisinin bulunduğu doğru kaynak dosya/sürüm paylaşılmalı (key değerini değil dosya/simge konumunu). O sürüm yoksa yeni anahtar/algoritma migrasyonu tasarım kararı gerekiyor.

GLOBAL global_i18n_08_fx.patch incelendi. Tam uygulanmadı: .12'de i18n, Localizer, PrefStore, langPref ve önceki 01..07 yamalarının altyapısı yok. Patch08 /v1/fx, HMAC, EUR+USD, kanus.fx/1 şeması bekliyor; yeni /v1/fx/usd-try + Ed25519 ile mantıksal çakışmalı. GLOBAL'in Türkçe price.summary TL-only fallback metni yeniden kullanıldı. Yeni bir paralel i18n/formatlama katmanı yaratılmadı. FXMaxAge ve approx-notes mantığı doğru kaynak/key geldiğinde aynı i18n/fx.go içinde birleştirilmeli; 5 günlük sınırda patch08'in +24h toleransı yeni şartla aynen alınmamalı.

Sabit kur kaldırıldı: katalog PriceText TL-only; lisans mint'i artık doğrulanmamış USD tutarı üretmez (USD JSON alanı geriye uyum için 0 olarak kalır); eski USD-only lisans fiyatı gösterilmez; pricing-gen JSON/csv ödemeleri TRY kalır ve usdtry_ref/hesaplanmış USD kolonları kaldırıldı. Ledger, lisans HMAC algoritması, anahtarlar ve hesap protokolü değiştirilmedi.

Endpoint/fetcher/cache/signature ve EXE doğrulama bu yama içinde uygulanmadı. 503, imza, bayat kur, TCMB fixture ve rate-limit testleri NOT TESTED. Bunlar tamamlanmış özellik olarak sunulmamalı.
