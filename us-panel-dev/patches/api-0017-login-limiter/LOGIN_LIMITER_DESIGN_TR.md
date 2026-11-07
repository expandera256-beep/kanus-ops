# Giriş sınırlayıcı — US için uygulama önerisi

08.11.2026, Europe/Istanbul (TSİ). Deploy/merge yapılmadı. Demir'in bu sohbetteki 01:28 TSİ kararı:
> E-posta+IP ve artan bekleme olsun, ayrıca bilinen cihaz/oturum kilitten muaf olsun

## Kapsam ve taban

`/workspace/us-panel-dev/post12/term` kopyası seçildi. `work2/kiyas/term` eski, account limiter içermiyor. Seçilen ağacın accounts.go ve login_limiter.go dosyaları `pinned/v12` ile SHA-256 olarak aynı. Taban manifesti ayrıca verilir. Bu **saf .12 değil, .12-tabanlı post12 çalışma kopyasıdır**: pinned/v12 karşısındaki ek/farklı dosyalar `logs/base-vs-pinned12.txt` içinde tam listelenir (pricing/gateway/license ve retry/FX testleri dahil). Bu yamada onlara dokunulmadı; saf release ve .13 için ayrıca apply/entegrasyon kabulü gerekir. Yalnız kopyada licensesrv değişti; US korumasına karşı bu teslim öneri diff'idir. main.go, model_proxy.go, ledger, UYUM kablolaması değişmedi. .13 ve API additive stack üzerinde rebase/bağımsız kabul US/PANEL DEV'e ait.

## Anahtarlar ve kaynak sınırı

- E-posta: mevcut `normalizedEmail` (trim+lower, exact mail.ParseAddress); Gmail dot/+ özel eşitlemesi yok.
- Kaynak: mevcut proxy güven sınırı korunur. RemoteAddr loopback ise yalnız geçerli X-Real-IP alınır; diğer uzaktan gelen X-Real-IP/XFF yok sayılır. HOST reverse proxy istemci başlığını **overwrite** etmeli, arka uç internete doğrudan açık olmamalı. Loopback'e erişebilen başka süreç güven sınırı içindedir.
- IPv4 canonical /32, mapped IPv6 IPv4'e çevrilir; IPv6 /64 gruplanır. Privacy-address döndürmesiyle aşmayı azaltır; ortak /64/NAT kullanıcılarını birlikte yavaşlatma riski vardır. E-posta da anahtarın parçası olduğundan normal farklı hesaplar pair bütçesini paylaşmaz.
- Pair anahtarı `unknown:` veya `trusted:` + SHA256(normalized_email + NUL + source). Sabit boy; ham e-posta/IP map içinde tutulmaz. Global anahtar SHA256(email). Bu hash anonimleştirme değildir: tahmin edilebilir adresler bulunabilir; loglanmaz.
- `/account/login` ile `/v1/account/login` aynı handler durumunu kullanır. Handler aliası tanır; main.go değişmez. Saf .12 sunucuda /v1 mux kaydı yoksa mevcut API-0005 mux yaması yine gerekir; doğrudan handler testi ağda mux kaydı olduğunu kanıtlamaz.

## Eşikler, bekleme ve başarı

| Katman | Varsayılan | Amaç |
|---|---|---|
| Bilinmeyen cihaz pair (email+source) | İlk 5 doğrulanmış yanlış parola; 5'inci yanlış 401, sonraki istek 1 sn 429 | Yazım hatası toleransı, tek kaynaktan brute force |
| Bilinen cihaz yanlış parola pair | Ayrı `trusted:` ad alanında aynı 5 eşiği | Muafiyet yanlış parola denemelerini serbest bırakmaz |
| E-posta global | 100 yanlış parola; sonra 1 sn 429 | Kaynak döndüren saldırı; 10'luk genel kilide göre daha az kolay DoS |
| Login coarse kaynak | 100 istek / 15 dakika | Bir kaynaktan çok e-postaya spray/CPU koruması |
| bcrypt semaphore | Mevcut US ayarları, varsayılan 2 slot | Kaynak tükenmesine karşı korunur |

Her izin verilmiş ek yanlış parola sonrası bekleme 1,2,4,8,...512,900,900 sn. Bloklu **probe süreyi uzatmaz**, 429 denemesi yeni parola hatası sayılmaz. Geçmiş son sayılan hatadan 30 dakika sessizlikte silinir; sabit takvim penceresi yoktur. Böylece pencere kenarındaki çift bütçe önlenir ve 15 dk maksimum gecikme erken sıfırlanmaz. Retry-After saniyesi yukarı yuvarlanır. Paralel admission+compare atomik değil: mevcut bcrypt slot sayısı kadar uçuşta deneme eşiği aşabilir; semaphore sınırı kalır.

Global 100 eşiği ürün kararı olarak önerilen başlangıç parametresidir, yük/saldırı verisiyle ayarlanmalı. Dağıtık saldırı hâlâ bilinmeyen cihazı geciktirebilir; bu kasıtlı kabul edilen trade-off'tur. Bilinen doğru parolalı kullanıcı global/pair/coarse kapasite kilidini aşar. Başarı yalnız ilgili pair'i sıfırlar; global saldırı sayacı ve diğer kaynakların geçmişi sıfırlanmaz.

Yeni env: `US_LOGIN_PAIR_THRESHOLD=5` (3–20), `US_LOGIN_GLOBAL_THRESHOLD=100` (20–1000), `US_LOGIN_FAILURE_QUIET_SECONDS=1800` (1800–86400). Geçersiz değerler varsayılana döner. Eski `US_LOGIN_ACCOUNT_LIMIT/WINDOW_SECONDS` artık kullanılmaz; US release notunda operatör migrasyonu belirtmeli, sessiz davranış beklentisi taşınmamalı. IP login bütçesi eski 10/900 yerine 100/900; yeni pair ilk 5'ten sonra daha erken yavaşlatır, büyüyen beklemeyi eski sabit IP kilidi gölgelemez. Diğer auth gruplarının mevcut sayısal sınırları değişmedi; kaynak canonicalization hepsine uygulanır.

## Bilinen cihaz/oturum tanımı

Bu öneri ayrı kalıcı cihaz kimliği üretmez; **geçerli mevcut usa1 oturumu** seçeneğini uygular. HMAC, rastgele nonce, sub ve sessionVersion mevcut mekanizmadır; token çözüldüğünde bulunan hesabın e-postası login hedefiyle eşleşmelidir. IP'ye bağlanmaz (kullanıcı ağ değiştirebilir). Sahte/süresi bitmiş/başka hesaba ait proof muafiyet sağlamaz. İstemcinin serbestçe verdiği device-id, email headerı veya User-Agent asla güven sayılmaz.

- Başarılı parola girişinde mevcut JSON `session` korunur; ek `__Host-kanus_login` çerezi aynı usa1 tokenını taşır. Path=/, Domain yok, HttpOnly, Secure, SameSite=Strict, Max-Age ve Expires mevcut TTL kadar.
- Süre mevcut TTL: kurucu 1 dakika–7 gün aralığını zaten zorunlu tutar. Mevcut EXE oturum TTL'si neyse muafiyet de o kadar; süresiz hatırlanan cihaz yok.
- Parola reset/değişiminde mevcut sessionVersion artar; hem header hem cookie hemen geçersizleşir. Yeni sessionSecret tümünü iptal eder. Tek cihaz iptali/logout sunucu denylisti bu yamada yok; tüm oturum iptali version artışıyla yapılır. İstemci logout'ta saklanan tokenı/çerezi silmeli.
- Masaüstü/EXE cookie taşımıyorsa son geçerli tokenı güvenli OS credential store'da saklar; sonraki parola logininde `Authorization: Bearer <usa1>` kullanır. JSON yanıtı geriye uyumludur; istemcide bu taşımanın fiilen uygulanması ayrı EXE değişikliğidir, bu patch istemciyi değiştirmez. Token loglanmaz, URL'ye konmaz. Süresi bitince yeni cihaz gibi davranır.
- Muafiyet auth değildir: parola yine bcrypt ile doğrulanır, snapshot/version yeniden kontrol edilir. Yanlış parola 200 alamaz; trusted yanlış pair bütçesinden sayılır, ayrıca global bütçeyi besler. Önceden trusted pair blokluysa yanlış parola ancak doğrulamadan sonra 429 alır; doğru parola bloklu olsa da 200 alır. Böylece çalınmış geçerli token ile bcrypt CPU tüketimi tamamen önlenemez; mevcut concurrency/queue 503 koruması sürer. Yalnız doğru parolayı muaf tutarken parolayı hesaplamadan bloklamak mümkün değildir.
- Secure cookie HTTP'de geri taşınmaz. Çerez için HTTPS gerekir; dış URL ve standard sertifika/hostname doğrulaması operatör sorumluluğu. TLS pin yok. Cookie uç noktalarında CSRF/origin ve logout UX bu dar limiter yamasından bağımsız değerlendirilmelidir.

## Kayıtsız e-posta ve hatalar

Kayıtsız adres de aynı pair/global eşiğe girer, aynı dummy bcrypt yolu korunur. Geçerli parola uzunluğu 12–72; kısa/uzun parola veya geçersiz e-posta 400, **failure** sayılmaz. JSON bozuksa mevcut 400 korunur. 401 ve 429 hesabın varlığını açıklamaz. Üretimde bcrypt cost=12; test helper kayıtlı hesaplar için cost4 kullanır, dolayısıyla test süreleri enumeration-timing kanıtı değildir.

429 gövdesi API-0003 ile aynı additive sözleşme:
`{"error":"account request rejected","code":"rate_limited","message":"account request rejected","request_id":"<rastgele>"}`
Ayrıca X-Request-Id, Cache-Control:no-store, Content-Type:application/json ve Retry-After. JSON error string korunur; request_id yüzünden kayıtlı/kayıtsız gövdeler byte-byte eşitlenmez. Bu taban api_error.go içermediğinden yerel loginRateError helperı kullanılır; US isterse API-0003 WriteAccountError'a birebir yönlendirebilir. bcrypt doluluğunda mevcut 503 ve Retry-After korunur, yeni 429 diye sayılmaz.

Coarse kaynak limitine ait 429 mevcut backpressure polling davranışını (5 sn) kullanır; bu 900 sn pencerenin kesin bitişi değildir. Pair/global Retry-After gerçek kalan süredir. Kapasite doluyken de 5 sn tekrar-dene kullanılır; her 5 sn'de kapasite açılacağı vaat edilmez. Bu sınır dokümante edilmiş kalan operasyonel iyileştirmedir.

## Bellek, spray, restart ve çok örnek

İki ayrı limiter map'i (pair ad alanları ortak 10000 + global 10000), mevcut IP map'i 4096, sabit boy hash anahtarları. Yeni canlı anahtar ekleneceği zaman kapasite doluysa **canlı giriş asla eviction ile silinmez**, yeni anahtar fail-closed 429 alır. Süresi bitmiş kayıtlar temizlenir; en fazla dakikada bir failure yolu sweep yapar, kapasite durumunda da sweep olur. Pair+global toplam en çok 20000; bellek kesin byte ölçümü yapılmadı. Spray hedefin geçmişini temizleyemez ama kapasite doldurup yeni cihazları geciktirebilir. Bilinen doğru parola bu kapasite kilidinden muaf; tamamen yeni cihaz kullanıcısına availability riski kalır.

Sayaçlar süreç içidir; restart'ta sıfırlanır, iki handler/instance paylaşmaz. Güvenilen sessionVersion hesap store'unda kalıcıdır; limiter geçmişi kalıcı değildir. Tek süreçte mevcut IP/nginx katmanları ve artan bekleme ilk iyileştirmedir, fakat restart/çok-replica brute-force bütçesi garantisi değildir. **Çok replica veya güçlü restart dayanıklılığı gerektiren prod kabul için ortak atomik TTL counter (örn. Redis/Lua) gereklidir.** Ledger JSON transaction'ına her yanlış parolayı yazmak disk DoS yaratacağından önerilmez. Yeni altyapı bu patch'e sessizce eklenmedi. Hangi ortak depolama/erişim politikası kullanılacağı US/SRE ile ayrıca kararlaştırılmalı; o zamana kadar bu açık risk release notunda kalmalı. Kullanıcı kararı diye uydurulmuş kalıcılık onayı yoktur.

## Redeem önerisi (bu patch uygulamaz)

F07'nin authenticated invalid-license redeem denemeleri limitsiz bulgusu ayrı kalır. Öneri: doğrulanmış account_id+source için 10 yanlış/15dk, ayrıca hesap toplamı 30 yanlış/15dk ve source CPU kapısı; yanlış redeem'ler tekdüze hata, 429 zarfı+Retry-After. Bilinen cihaz yanlış lisansa muaf tutulmamalı; başarılı idempotent aynı hesaba bağlama etkilenmemeli. Sayılar ve mevcut istemci davranışı QA/US ile doğrulanmadan buraya davranış eklenmedi.

## QA yeni beklenti tablosu

| Senaryo | Eski F04/F07 beklentisi | Yeni kabul / kanıt |
|---|---|---|
| Aynı email+IP 5 yanlış | İlk 10 yanlış 401 | 5×401, hemen sonraki 429 Retry-After:1; süreyi bekledikten sonraki yanlışla 2,4,...900 |
| Aynı email farklı IPv4 | 10 toplamdan sonra herkes 429 | İlk 100 ayrı-source yanlış 401; yeni kaynak bilinmeyen cihaz 429; doğru parola + geçerli aynı hesap session 200 |
| Saldırı 10 farklı kaynak, yeni meşru IP | Eski F07 güvenlik testi 200 beklerken 429 | Global eşik altı: doğru parola 200; değişmemiş F07 AccountLockDoS bu nedenle PASS |
| Pair saldırısı, aynı IP bilinen session/cookie | Kilitli doğru parola da 429 | Doğru parola 200; yanlış 401 veya trusted-pair 429, asla 200 |
| Cookie/header sahte/süresi geçmiş/başka hesap | Muafiyet yoktu | Kilit atlanamaz; 429 |
| Parola reset sonrası eski proof | Yok | Eski cookie/header geçersiz; bilinmeyen cihaz limiti uygulanır |
| Başarı | Tüm email sayacı reset | Yalnız başarılı pair reset; global ve diğer pairler korunur |
| Registered/unknown, normalize email | Aynı 10/900 | Aynı 5-pair/100-global davranışı, generic zarf; random request_id dışında anlamlı fark yok |
| Kısa parola | 400, failure değil | 400 ve ilk geçerli yanlış deneme 401 |
| IPv6 adres döndürme | Exact-IP | Aynı /64 aynı pair; farklı /64 ayrı kaynak |
| X-Real-IP remote spoof | Yalnız loopback güvenilir | Aynı sınır korunur, mapped IPv4 aynı kaynak |
| 10000 anahtar spray | Eski risk repro evicted=true PASS diyordu | Live target eviction **olmamalı**; eski risk gözlem testi FAIL olması düzeltmenin beklenen etkisi |
| Bellek limiti dolu | Oldest eviction | Bilinmeyen anahtar 429; eski kayıt korunur; trusted doğru parola yine 200 |
| Yeniden başlatma | Sıfırlanıyor | Hâlâ sıfırlanır: açık risk, tam düzeltildi sayılmaz |
| Redeem | Limitsiz risk gözlemi PASS | Hâlâ limitsiz; ayrı öneri, kapandı denmez |
| Gerçek nginx/canlı dağıtım | SOURCE MITIGATED / LIVE NOT VERIFIED | Aynı: hiçbir canlı istek/deploy yok |

F04 tarihsel IP-only DESIGN GAP statüsü geriye dönük değiştirilmez. F07 account-lock DoS yeni senaryolarla yeniden kabul edilmelidir; unknown yeni cihazın eşik üstünde gecikmesi kasıtlıdır. Eski 10/900 testleri aynen korunup yeni politika testleriyle birlikte PASS diye sunulamaz. Eski repo testlerinden yalnız bu politika ile çelişen beklentiler açıkça güncellendi; RED ana politika testleri değişmeden GREEN'de tekrar kullanıldı. Sonradan eklenen clock/concurrency testlerinin eski kodda RED çalıştırıldığı iddia edilmez.
