# SRE standart sertifika izlemesi

## Kapsam ve karar
Pinleme **iptal edildi, kullanılmayacak**. Sistem CA deposuyla zincir ve hostname/SAN doğrulaması kullanılacak. SPKI gözlemleri pin önerisi veya onaylı yapılandırma değildir. Daha önce oluşturulan `pin_check.py` çalıştırıldığında UNKNOWN/exit 3 ile durur; `pins.example.json` ve `CANCELLED_PINNING.txt` açıkça iptal olarak işaretlendi. Dosyalar tarihsel yerel taslak olarak tutuldu.

VPS SSH, özel anahtar okuma/yazdırma, GCP yazma/deploy, DNS değişikliği ve HTTP isteği yapılmadı. Hiçbir bota mesaj gönderilmedi. Yerel testlerde yalnız bellekte geçici anahtar üretilip sentetik public sertifikalar imzalandı; anahtar dosyası/çıktısı yok.

## Host keşfi
- **PASS:** `.12` teslim raporu `/workspace/us-panel/dist/2026.11.06.12/TESLIM_RAPORU_TR.md:34` embeddedLicenseURL ve embeddedGatewayURL değerlerini `https://api-test.kanus.org` gösteriyor.
- **PASS (kaynak taslağı):** `/workspace/us-panel-dev/patches/global_i18n_08_fx.patch:417,482-489,569-571`: panel kur isteği `activeGateway()` tabanının `/v1/fx` yoluna gidiyor. Ayrı bir kur hostname'i tanımlı değil. .12 varsayılan gateway esas alınırsa lisans ve kur aynı `api-test.kanus.org` hostunda.
- **NOT VERIFIED:** Kur yamasının panelin dağıtılmış sürümünde bulunduğu ve kullanıcıda URL override olmadığı bu metadata incelemesiyle doğrulanmadı.
- **PASS:** Web sitesi kuru `kanus.org/fx.json` üzerinde; `GLOBAL_FX_UPLOAD_20261108.md:12-15,26-30`. Bu, masaüstü panelin `/v1/fx` sözleşmesinden farklıdır.
- **PASS:** `api.kanus.org` production lisans adayı; host-hardening `PROD_PLAN.md` ve .12'deki staging varsayılanı birbirinden ayrıldı.

## Public TLS gözlemi
8 Kasım 2026 01:22:20–01:22:21 TSİ aralığında **host başına yalnız bir** `openssl s_client` bağlantısı. `-showcerts -verify_return_error -verify_hostname` kullanıldı. Hiç HTTP verisi gönderilmedi. Aynı public handshake kanıtları expiry ve yerel zincir kontrollerinde tekrar kullanıldı; ek bağlantı yok.

| Host | Leaf issuer | notAfter | Kalan gün | Zincir/SAN + süre |
|---|---|---|---:|---|
| api-test.kanus.org | CN=YE1,O=Let's Encrypt,C=US | 2027-02-04 03:48:11 TSİ | 88.10 | PASS |
| api.kanus.org | CN=YE1,O=Let's Encrypt,C=US | 2027-02-04 08:13:20 TSİ | 88.28 | PASS |
| kanus.org | CN=YE1,O=Let's Encrypt,C=US | 2027-01-31 18:28:13 TSİ | 84.71 | PASS |

Her üç hostta leaf ve ara anahtarlar EC. SAN api-test ve api için birebir hostname; kanus.org için `kanus.org`, `www.kanus.org`.

Sunucuların yayınladığı zincir: **leaf → YE1 → Root YE (X2 imzalı) → ISRG Root X2 (X1 çapraz imzalı)**. Sistem deposundaki self-signed ISRG Root X1 güven çıpası; sunucunun kök sertifikayı göndermemesi normaldir. Zincirde görünen çapraz imzalı X2, self-signed X2 ile aynı anahtara sahip olsa da farklı sertifikadır. Eksik ara sertifika ile kökün gönderilmemesi birbirine karıştırılmaz.

Ek sertifika metadata (üç host için ortak; tüm zamanlar TSİ):
| Subject | Issuer | notAfter |
|---|---|---|
| YE1 | Root YE | 2028-09-03 02:59:59 TSİ |
| Root YE | ISRG Root X2 | 2032-09-03 02:59:59 TSİ |
| ISRG Root X2 (çapraz imzalı) | ISRG Root X1 | 2032-09-03 02:59:59 TSİ |
| Sistem deposu ISRG Root X1 | ISRG Root X1 | 2035-06-04 14:04:38 TSİ |
| Sistem deposu ISRG Root X2 | ISRG Root X2 | 2040-09-17 19:00:00 TSİ |

## Uygulama ve eşikler
Dizin: `/workspace/us-panel/sre/pin_monitor/` (isim tarihsel; etkin özellik pinleme değil).
- `cert_expiry.py`: JSON; exit 0 OK/INFO, 1 WARN, 2 CRITICAL, 3 UNKNOWN. Kalan süre **≤21 gün WARN**, **≤7 gün CRITICAL**. **<25 gün** ve leaf notBefore önceki gözlemden yeni değilse WARN. Önceki gözlem yoksa yenileme kanıtı olmadığı açıkça bildirilir ve <25 gün için ihtiyatlı WARN verilir. Kanıtlanmış yenileme 21/7 eşiklerini bastırmaz.
- `cert_hosts.json`: sadece host adları, pin içermiyor. Yenileme karşılaştırması için isteğe bağlı `previous_not_before` ISO tarihi; ilk çalıştırmada geçmiş yok. Otomatik geçmiş deposu/cron kurulmadı.
- `common.observe`: host başına tek TLS bağlantısı, `subprocess` timeout **5 saniye**, sistem güven deposu ve hostname doğrulaması. Bağlantı hataları UNKNOWN; retry yok. Sonraki yerel OpenSSL doğrulama işlemi ayrıca 5 saniye timeout taşır; toplam işlem bütçesi ağ timeoutundan ayrıdır.
- `chain_check.py`: sistem `/etc/ssl/certs/ca-certificates.crt`, SSL-server purpose, zorunlu DNS SAN, OpenSSL hostname eşleşmesi; alınan aralar yalnız `-untrusted` olarak kullanılır, sunucunun gönderdiği kök güven çıpası yapılmaz. Alternatif CApath/CAstore kapalı. AIA üzerinden ara indirme yok. SAN uyuşmazlığı, eksik ara veya güvenilmeyen kök CRITICAL/exit2.
- Sistem güven deposunda alternatif geçerli yol mevcutsa bazı sunucu ara eksikleri yine doğrulanabilir; politika zincirin her sertifikasının gönderilmesi değil, standart güvenilir yol kurulmasıdır.
- Kriptografik zincir kontrolü CRL/OCSP iptal kontrolü anlamına gelmez; bu kapsamda revocation **NOT TESTED**.

## Test ve sonuç
- **PASS: 15/15 yerel birim testi.** Eşik sınırları 26/25/24.9/21/7/negatif gün, yenileme var/yok, doğru SAN, yanlış SAN, SAN yok, eksik ara, güvenilmeyen kök, tek bağlantı ve 5 saniye timeout, HTTP verisi gönderilmemesi.
- **PASS:** Gerçek public handshake kanıtlarına expiry kontrolü üç hostta OK/exit0.
- **PASS:** Aynı gerçek zincirler sistem deposuyla üç hostta OK/exit0; hostname/SAN eşleşmesi geçiyor.
- **NOT TESTED:** Sunucunun renewal otomasyonu/certbot ayarları ve gelecekte yenileme başarıları. Public gözlem mevcut sertifikayı kanıtlar; sunucu işini kanıtlamaz.
- **NOT TESTED:** Zamanlanmış izleme, bildirim kanalı ve panelin dağıtılmış binary'sinde sertifika doğrulaması. Hiç cron/job/routine kurulmadı.

Kanıtlar: `evidence/*.handshake.txt`, `observations.json`, `roots.json`, `unit-tests.txt`, `cert-expiry.json`, `chain-check.json`. JSON metadata kaynak ISO offsetlerini taşır; bu rapordaki tarihler Europe/Istanbul/TSİ'ye çevrilmiştir.

## Çalıştırma
```sh
cd /workspace/us-panel/sre/pin_monitor
python3 -m unittest -v test_cert_monitor
# Canlı: host başına tek handshake; yalnız TLS, HTTP yok
python3 cert_expiry.py --config cert_hosts.json
# Bu denetimde bağlantıyı tekrar açmadan aynı canlı kanıt kullanıldı:
python3 cert_expiry.py --config cert_hosts.json --snapshot evidence/observations.json
python3 chain_check.py --config cert_hosts.json --evidence-directory evidence
```
Snapshot seçeneği çevrimdışı kanıt tekrar değerlendirmesidir; sürekli canlı izlemenin yerine geçmez. Bağımlılık: Python `cryptography` (box'ta 43.0.0), OpenSSL CLI, sistem CA deposu.

## Açık sorular
1. Panelin production lisans/kur adresi ne zaman api-test'ten api'ye geçecek? `activeGateway` kullanıcı override'ları var mı?
2. GLOBAL kur yaması hangi dağıtılmış panel sürümünde etkin? Web sitesi fx.json ile panel /v1/fx ayrımı korunmalı.
3. HOST standart ACME yenilemesini ve hata bildirimini nerede çalıştırıyor? Yenileme başarı geçmişi nerede?
4. Zamanlanmış izlemenin periyodu, çalışacağı makine ve alarm alıcısı ayrıca kararlaştırılmalı; bu görevde hiçbir otomasyon etkinleştirilmedi.
