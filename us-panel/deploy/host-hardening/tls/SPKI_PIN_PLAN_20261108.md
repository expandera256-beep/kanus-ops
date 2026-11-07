İPTAL EDİLDİ: Demir 01:22 TSİ — "Pin olmasın, yalnız standart sertifika doğrulaması kullanılsın" (API üzerinden aktarıldı). Yedek anahtar ve reuse_key işi yapılmayacak; sunucuda değişiklik yok.

# SPKI pin planı — 8 Kasım 2026

Durum: PLAN / ONAY BEKLİYOR. İnceleme 01:21 TSİ (Europe/Istanbul). VPS üzerinde yapılandırma değişikliği, anahtar üretimi ve restart yapılmadı. Özel anahtar yazdırılmadı veya kopyalanmadı.

## Ölçüm
82.29.184.225 üzerinde /etc/letsencrypt/live/*/fullchain.pem okundu. Yalnızca aşağıdaki iki lineage mevcut. SPKI = base64(SHA256(DER SubjectPublicKeyInfo)); sertifika parmak izi değildir. Her iki domain için 127.0.0.1:443 + SNI ile sunulan leaf ve zincir dosyadaki değerlerle aynı. Bu, yerel nginx kontrolüdür; dış DNS/uçtan uca satış kabulü değildir.

### api-test.kanus.org

**Leaf**
- SPKI: `/mDQtXQR7xQxXMCtk3hPpVy9EvYNgIYODuGXjxl1cjc=`
- Anahtar: Public Key Algorithm: id-ecPublicKey, Public-Key: (256 bit), ASN1 OID: prime256v1
```text
subject=CN=api-test.kanus.org
issuer=C=US, O=Let's Encrypt, CN=YE1
notBefore=2026-11-06 03:48:12 TSİ
notAfter=2027-02-04 03:48:11 TSİ
X509v3 Subject Alternative Name: 
    DNS:api-test.kanus.org
```

**Zincir sertifikası 1**
- SPKI: `brzvtCELCIZUo4sD/qPX0ccRtPsd3DY6RfmxpOU9oB4=`
- Anahtar: Public Key Algorithm: id-ecPublicKey, Public-Key: (384 bit), ASN1 OID: secp384r1
```text
subject=C=US, O=Let's Encrypt, CN=YE1
issuer=C=US, O=ISRG, CN=Root YE
notBefore=2025-09-03 03:00:00 TSİ
notAfter=2028-09-03 02:59:59 TSİ
```

**Zincir sertifikası 2**
- SPKI: `sCkq5UWXjg+7mKu9lMhhYF5bGLsy7VI/UNW3tccdR7w=`
- Anahtar: Public Key Algorithm: id-ecPublicKey, Public-Key: (384 bit), ASN1 OID: secp384r1
```text
subject=C=US, O=ISRG, CN=Root YE
issuer=C=US, O=Internet Security Research Group, CN=ISRG Root X2
notBefore=2026-05-13 03:00:00 TSİ
notAfter=2032-09-03 02:59:59 TSİ
```

**Zincir sertifikası 3**
- SPKI: `diGVwiVYbubAI3RW4hB9xU8e/CH2GnkuvVFZE8zmgzI=`
- Anahtar: Public Key Algorithm: id-ecPublicKey, Public-Key: (384 bit), ASN1 OID: secp384r1
```text
subject=C=US, O=Internet Security Research Group, CN=ISRG Root X2
issuer=C=US, O=Internet Security Research Group, CN=ISRG Root X1
notBefore=2026-05-13 03:00:00 TSİ
notAfter=2032-09-03 02:59:59 TSİ
```

### api.kanus.org

**Leaf**
- SPKI: `RMWkVD8VDAnv0bthc/pfT0BDB54THxSNo6UcGicoRTI=`
- Anahtar: Public Key Algorithm: id-ecPublicKey, Public-Key: (256 bit), ASN1 OID: prime256v1
```text
subject=CN=api.kanus.org
issuer=C=US, O=Let's Encrypt, CN=YE1
notBefore=2026-11-06 08:13:21 TSİ
notAfter=2027-02-04 08:13:20 TSİ
X509v3 Subject Alternative Name: 
    DNS:api.kanus.org
```

**Zincir sertifikası 1**
- SPKI: `brzvtCELCIZUo4sD/qPX0ccRtPsd3DY6RfmxpOU9oB4=`
- Anahtar: Public Key Algorithm: id-ecPublicKey, Public-Key: (384 bit), ASN1 OID: secp384r1
```text
subject=C=US, O=Let's Encrypt, CN=YE1
issuer=C=US, O=ISRG, CN=Root YE
notBefore=2025-09-03 03:00:00 TSİ
notAfter=2028-09-03 02:59:59 TSİ
```

**Zincir sertifikası 2**
- SPKI: `sCkq5UWXjg+7mKu9lMhhYF5bGLsy7VI/UNW3tccdR7w=`
- Anahtar: Public Key Algorithm: id-ecPublicKey, Public-Key: (384 bit), ASN1 OID: secp384r1
```text
subject=C=US, O=ISRG, CN=Root YE
issuer=C=US, O=Internet Security Research Group, CN=ISRG Root X2
notBefore=2026-05-13 03:00:00 TSİ
notAfter=2032-09-03 02:59:59 TSİ
```

**Zincir sertifikası 3**
- SPKI: `diGVwiVYbubAI3RW4hB9xU8e/CH2GnkuvVFZE8zmgzI=`
- Anahtar: Public Key Algorithm: id-ecPublicKey, Public-Key: (384 bit), ASN1 OID: secp384r1
```text
subject=C=US, O=Internet Security Research Group, CN=ISRG Root X2
issuer=C=US, O=Internet Security Research Group, CN=ISRG Root X1
notBefore=2026-05-13 03:00:00 TSİ
notAfter=2032-09-03 02:59:59 TSİ
```

Root YE ve Root X2 burada çapraz imzalı kök sertifikalarıdır; YE1 ile aynı türden issuing intermediate diye değerlendirilmemelidir.

## Wake/model :8443
- Uç nokta: 34.41.186.170:8443; VPS üzerinden TLS bağlantısı erişilebilir.
- Yerel public sertifika: /etc/us-model/wake.crt. Sunucunun sunduğu sertifikayla SPKI ve metadata eşleşiyor.
- SPKI: `3edOilLr+zTl2YGigD8AqZ3m53J7xTktnZGV3v7ADxk=`
- Subject ve issuer: CN=34.41.186.170, O=usta-wake, C=TR (self-signed).
- SAN: IP Address:34.41.186.170; ECDSA prime256v1 / 256 bit.
- Geçerlilik: 2026-11-06 09:46:48 TSİ — 2028-11-05 09:46:48 TSİ.
- Bu bağlantı yalnızca public sertifika topladı; wake HTTP çağrısı yapmadı. Self-signed sertifikanın güven kökü/pin ile ayrıca doğrulanması gerekir. API leaf pinleri burada kullanılamaz.

## Certbot / mevcut yedek durumu
- /etc/letsencrypt/renewal/api.kanus.org.conf: key_type = ecdsa; reuse_key satırı YOK.
- /etc/letsencrypt/renewal/api-test.kanus.org.conf: key_type = ecdsa; reuse_key satırı YOK.
- authenticator = webroot; production ACME v2. Global cli.ini kontrolünde reuse_key veya key_type override çıkmadı. Etkin anahtarın korunması garanti değil; reuse_key henüz açık olarak ayarlanmamış.
- certbot.timer active; OnCalendar=*-*-* 00,12:00:00 (VPS UTC; TSİ karşılığı 03:00/15:00), RandomizedDelaySec=43200, Persistent=true.
- Son timer çalışması 2026-11-08 01:15:03 TSİ; bir sonraki gözlenen 2026-11-08 13:25:27 TSİ. Bu, başarılı renewal kanıtı değildir.
- /root/tls-backup-keys mevcut DEĞİL. /etc/letsencrypt içinde maxdepth 2 backup-adlı dosya varlık taraması sonuç vermedi. Başka konumda offline yedek yoktur denemez. Archive içindeki eski lineage anahtarları bağımsız önceden pinlenmiş offline backup sayılmaz. Özel anahtar içeriği incelenmedi/kopyalanmadı.

## Önerilen pin kümesi (en az iki yedek)
Her hostname için ayrı politika:
1. Aktif: yukarıdaki ilgili leaf SPKI.
2. Yedek A: **TO BE GENERATED / ONAY GEREKİYOR** bağımsız offline backup anahtarı. Henüz pin değeri yok; uydurulamaz. Önerilen saklama /root/tls-backup-keys altında root-only dizin (0700), dosya 0600 + güvenli offline kopya. Offline kopyalama ve anahtar üretimi ayrı onay gerektirir. VPS kopyası tek başına offline değildir.
3. Yedek B (daha geniş güven, yalnızca açık risk kabulüyle): gözlenen Let's Encrypt YE1 SPKI `brzvtCELCIZUo4sD/qPX0ccRtPsd3DY6RfmxpOU9oB4=`. Normal PKI zinciri, SAN/hostname ve süre kontrolleri daima korunmalı; pin yalnızca doğrulanmış zincirde eşleşmeli. Tek başına sunucunun gönderdiği herhangi bir sertifikayı eşleştirmek yeterli değildir.

Bu üçlü bugün tamamlanmış DEĞİL: backup A üretilmedi; dolayısıyla iki yedek hazır denemez. Geniş CA fallback istenmezse ikinci bağımsız offline leaf anahtarı da önceden üretilip dağıtılmalı.

### Güncel Let's Encrypt hiyerarşisi
Resmî kaynak 8 Kasım 2026 incelemesinde aktif issuance CA'larını **YE1, YE2, YR1, YR2**, standby CA'ları YE3/YR3 olarak listeliyor. E5/E6/E7/E8/E9 ve R10/R11/R12/R13/R14 retired. Örnekteki E5/E6/R10/R11 güncel fallback diye seçilmemeli. Sahadaki iki zincir YE1 kullanıyor. YE2 gibi başka issuing CA'ya geçiş YE1-only fallback'i bozabilir. Root YE veya X2'yi fallback seçmek daha geniş güven alanı doğurur; otomatik önerilmiyor.
Kaynak: https://letsencrypt.org/certificates/ (sayfa güncellemesi 2026-07-08; bu incelemede açıldı).

## Rotasyon / geri dönüş
1. Onaydan sonra her renewal config için `reuse_key = True` veya kontrollü Certbot `--reuse-key` kullanımı; sürüm destek/etki kontrolü, kontrollü dry-run ve sonrasında public SPKI eşleşme kontrolü. Şimdi uygulanmadı. Aynı anahtarın sonsuza kadar tutulması önerilmez; ihlalde acil rotasyon gerekir.
2. Yeni anahtar/backup'ın yalnız public SPKI değerini önce panellerin hostname-scoped pin kümesine ekle. Dağıtım ve eski istemci kapsaması kanıtlanmadan TLS anahtarını değiştirme.
3. Yeni anahtar ile sertifika issuance ve nginx geçişini ayrıca onayla; mevcut public sertifika/konfigürasyon referanslarını koru. Yenileme yöntemi backup anahtarını gerçekten kullanacak şekilde önceden doğrulanmalı.
4. Geçiş sonrası normal zincir, SAN, süre, leaf SPKI, API erişimi ve negatif fail-closed testlerini doğrula. Eski leaf pinini istemci güncelleme süresi boyunca tut; sonra ayrı onayla kaldır.
5. Sorunda eski sertifika/anahtar referanslarına kontrollü geri dön; özel anahtarı taşıma/yazdırma. Eski anahtar compromised ise geri dönüş yapılmaz; güvenli backup ve acil panel güncellemesi kullanılır.

## Riskler / .12 uyumsuzluğu
- Yanlış/eksik pin panelleri fail-closed bırakır; lisans/kur çekimi kesilebilir. Salt doğru SPKI, SAN/zincir doğrulamasının yerine geçmez.
- Intermediate pin hostname doğrulamasını kaldırmaz ama aynı CA'nın tüm geçerli issuance alanına güveni genişletir; intermediate rotasyonu fallback'i bozabilir. CA pinleri dar leaf pininin güvenlik garantisini zayıflatır.
- .12 wake pin okuyucusu yalnız PEM sertifikası kabul ediyor; base64 SPKI değeri buraya doğrudan yazılamaz. Panelin API/kur SPKI pinleme talebi ile hook'un wake PEM politikası ayrı protokoller/ayarlar. Uyum için PANEL DEV tasarım/test/PASS ve ayrı kurulum onayı gerekir. .12 mevcut FAIL nedeniyle kurulmayacak.

## Ayrı onay gereken adımlar
- Demir'den bağımsız backup anahtarı üretimi, root-only saklama ve güvenli offline kopya için açık onay.
- Her iki certbot lineage'da reuse_key değişikliği ve kontrollü renewal testi için açık onay.
- CA fallback risk kabulü; pin kümelerinin GLOBAL/US/panellere dağıtımı ve uygulama değişiklikleri için onay.
- PEM/SPKI uyum değişikliği, yeni sürüm PASS sonrası deploy için yeni onay.
- Sertifika/anahtar rotasyonu, nginx reload ve olası rollback için ayrı onay. Bu rapor bunların hiçbirini uygulamaz.
