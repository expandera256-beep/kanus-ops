# UYUM — .14 görsel/video üretimi için içerik güvenliği ve veri koruma görüşü
_2026-11-08 01:25 (UTC+3) · Hazırlayan: UYUM · Talep: US (.14 planı) · Durum: TASLAK, hukuki tavsiye değildir_

Etiketler: **[Z]** zorunlu (yayın engeli) · **[Ö]** önerilen · **HUKUKİ** hukukçu doğrulaması gerekir.
Bu belge hukuki uyumu tamamlanmış saymaz; teknik asgari çizgiyi tanımlar.

## 1. Prompt filtreleme
- **[Z]** Filtre sunucu tarafında, GPU kuyruğuna girmeden önce ve kredi rezervasyonundan önce çalışmalı. İstemci tarafı filtre yalnızca UX içindir, güvenlik sayılmaz.
- **[Z]** Filtre, modele giden **son birleşik metne** uygulanmalı: kullanıcı promptu + negatif prompt + kullanıcı sistem talimatı + şablon/stil ön ekleri + (varsa) yüklenen görselden çıkarılan metin/altyazı.
- **[Z]** Engellenecek kategoriler (sıfır tolerans, istisna yok): reşit olmayanların cinselleştirilmesi (CSAM) ve buna yaklaşan her şey; gerçek ve tanınabilir kişilerin rızasız cinsel/çıplak görüntüsü ("soy", "nudify", "undress" türü dönüşümler dahil).
- **[Z]** Engellenecek kategoriler (politika): terör/şiddet propagandası, korunan gruplara yönelik nefret içeriği, ağır şiddet/vahşet yüceltme, kendine zarar teşviki, silah/patlayıcı yapımına yönelik görsel talimat, aldatıcı siyasi/seçim deepfake'i, belge/kimlik/para sahteciliği.
- **[Ö]** Ürün kararı yoksa yetişkin (NSFW) içerik tamamen kapalı olmalı. Açılacaksa yaş doğrulama, ayrı sözleşme ve ayrı hukuki değerlendirme gerekir (**HUKUKİ**, Türkiye'de ayrıca 5651 ve genel ahlak hükümleri).
- **[Ö]** Çok dilli normalizasyon (TR/EN + diğer arayüz dilleri), leetspeak/ayrık yazım, Unicode homoglif normalizasyonu; anahtar kelime listesi + sınıflandırıcı birlikte.
- **[Ö]** Engellenen istekte kredi düşülmemeli ya da rezervasyon serbest bırakılmalı (mevcut reservation/release akışıyla); kullanıcıya kategori adı verilip ayrıntı verilmemeli.

## 2. Çıktı denetimi
- **[Z]** Her çıktıda sunucu tarafı görsel sınıflandırıcı: NSFW + yaş/reşit olmayan sinyali. Video için kare örnekleme (ör. her saniye en az 1 kare + ilk/son kare) ve küçük resim.
- **[Z]** Sıfır tolerans kuralı: "reşit olmayan sinyali" ile "cinsel/çıplaklık sinyali" birlikteyse eşik ne olursa olsun engelle. Prompt geçmiş olsa bile çıktı engellenir, kullanıcıya hiç gösterilmez, depolanmaz, yeniden deneme hakkı verilmez.
- **[Z]** Böyle bir olayda hesap otomatik askıya alınır ve insan incelemesine düşer. Kanıt saklama ve yetkili mercilere bildirim yükümlülüğü (TCK 226 kapsamı, bildirim usulü, kanıtın kim tarafından ne süre tutulacağı) **HUKUKİ** — hukukçu kararı gelmeden içerik geliştirici makinelerine, loglara veya yedeklere kopyalanmamalı.
- **[Ö]** Yüklenen görsellerde bilinen istismar içeriği için hash eşleştirme (ör. sektörel hash listeleri); erişim koşulları **HUKUKİ**.
- **[Ö]** Sınıflandırıcı modelleri de kendi GPU'muzda (ds-a100) çalışırsa üçüncü tarafa aktarım olmaz; harici API kullanılırsa m.9 aktarım envanterine eklenmeli.

## 3. Gerçek kişi benzerliği ve deepfake
- **[Z]** Politika: gerçek ve tanınabilir kişilerin cinsel/çıplak, aşağılayıcı veya aldatıcı (yaşanmamış olayı gerçekmiş gibi gösteren) içeriği yasak. Kamusal figürlerin aldatıcı siyasi içeriği yasak.
- **[Z]** img2img / i2v'de kişi fotoğrafı yüklenirken kullanıcı beyanı: "Bu kişinin rızası var / görsel üzerinde hakkım var." Beyan kaydı (zaman, hesap takma adı, sürüm) tutulur.
- **[Z]** Yüklenen görselde yüz tespit edilirse cinsel/çıplaklık/soyma yönlü prompt doğrudan engellenir. Yüklenen görselde reşit olmayan sinyali varsa vücut/kıyafet değiştiren dönüşümler engellenir.
- **[Z]** Yüz tespiti yalnızca anlık (transient) yapılmalı: yüz gömmesi (embedding) veya biyometrik şablon **saklanmamalı**, kimlik eşleştirmesi yapılmamalı. Aksi halde KVKK m.6 / GDPR m.9 özel nitelikli veri (biyometrik) rejimi devreye girer (**HUKUKİ**).
- **[Ö]** Yüklenen kişi fotoğrafları kısa saklanmalı (ör. iş bitince veya en geç 24 saat; kullanıcı kütüphanesine açıkça kaydetmediyse).

## 4. Filigran ve C2PA
- **[Z]** AB Yapay Zekâ Yasası m.50(2): sentetik görsel/video üreten sistem sağlayıcısı çıktıları makinece okunabilir ve tespit edilebilir biçimde işaretlemeli. Bu yükümlülüğün uygulama tarihi 2 Ağustos 2026'dır; AB'ye hizmet verilecekse yükümlülük **şu an yürürlükte** kabul edilmeli. Güncel erteleme/geçiş düzenlemeleri (Digital Omnibus vb.) ve KANUS'un "sağlayıcı" sayılıp sayılmadığı **HUKUKİ** doğrulama.
- **[Z]** Uygulama: C2PA manifesti (üretici: KANUS, "AI generated" iddiası, model adı/sürümü, zaman) + metadata silinse bile kalan görünmez filigran. Video için her iki katman.
- **[Z]** Manifest ve filigran içine **kişisel veri konmamalı** (hesap kimliği, e-posta, IP yok). Gerekirse yalnız rastgele üretim kimliği; bu kimlikle hesap arasındaki eşleme sunucuda tutulur ve anonimleştirmede kırılır.
- **[Ö]** m.50(4) deepfake ifşa yükümlülüğü dağıtıcıya (kullanıcıya) düşer; sözleşmede kullanıcıya aktarılmalı. İsteğe bağlı görünür "AI" etiketi seçeneği.

## 5. Prompt ve çıktı loglama, saklama, anonimleştirme
- **[Z]** Hesaba bağlı prompt, çıktı, yüklenen görsel ve kullanıcı sistem talimatı kişisel veridir. Hepsi mevcut veri dışa aktarma (export) çıktısına ve silme/anonimleştirme (erase) akışına eklenmeli; yani uyum yamalarındaki `privacy_ops` kapsamı .14'te genişletilmeli.
- **[Z]** Ayrı tutulacak iki log: (a) **ürün geçmişi** (kullanıcının kendi galerisi: prompt + çıktı) — kullanıcı silene veya hesap silinene kadar; (b) **güvenlik/moderasyon logu** — karar, kategori, skor, içerik hash'i, zaman, hesap HMAC takma adı (mevcut `US_PRIVACY_HMAC_KEY` ile). Moderasyon logunda ham prompt metni yalnız engellenen/işaretlenen kayıtlarda tutulmalı.
- **[Ö]** Saklama önerisi: moderasyon logu 90 gün (mevcut retention mekanizması ve 1–3650 gün sınırı içinde, yapılandırılabilir); işaretlenmiş kötüye kullanım kayıtları hukuki saklama bayrağıyla (legal hold) istisna; ham GPU iş logları / geçici dosyalar 7 gün; yüklemeler §3. Kesin süreler **HUKUKİ**.
- **[Z]** Erase akışı: galeri, çıktı dosyaları (disk + nesne deposu + önbellek/thumbnail), yüklemeler ve kullanıcı sistem talimatı silinir; moderasyon logu takma adlıysa kalır, eşleme kırılır; legal hold kayıtları istisna olarak belgelenir. Kredi defteri (generation charge kayıtları) mevcut kurala göre silinmez, kişisel alanlar anonimleşir. Yedeklerdeki kalıntı mevcut rapordaki yedek imha sorusuna dahil.
- **[Z]** Promptlar/çıktılar açık rıza olmadan model eğitimi veya ince ayar için kullanılmamalı; kullanılacaksa ayrı tercih anahtarı (uyum-preferences) ve aydınlatma.
- **[Ö]** ds-a100'ün fiziksel bölgesi/ülkesi aktarım envanterine yazılmalı (açık modeller yerelde çalıştığı için model sağlayıcıya aktarım yok — bu, mevcut DeepSeek aktarımından farklı ve olumlu; metinlerde ayrıştırılmalı).

## 6. Kullanıcı sözleşmesi ve kabul edilebilir kullanım politikası (AUP)
- **[Z]** AUP'ye §1–3'teki yasak kategoriler, gerçek kişi rızası beyanı, filigran/manifest kaldırma yasağı, deepfake ifşa yükümlülüğü, engellenen istek ve hesap askıya alma süreci, şikâyet/bildirim kanalı (içerik raporla) eklenmeli.
- **[Z]** Model lisansları: OpenRAIL-M ailesi (ör. CreativeML OpenRAIL-M) kullanım kısıtlamalarının (Attachment A) son kullanıcı sözleşmesine **en az aynı sıkılıkta aktarılmasını** şart koşar; bu modeller kullanılacaksa aktarım zorunlu.
- **[Z]** Her model için lisans envanteri çıkarılmadan yayın yapılmamalı: ticari kullanım izni, gelir/kullanıcı eşikleri, bölge kısıtları, çıktı hakları, atıf. Örnek risk (her biri **doğrulanmalı**): bazı popüler modeller yalnız ticari olmayan kullanıma izin verir (ör. FLUX.1 [dev]); bazıları gelir eşiğine bağlıdır (Stability Community License); bazı video model lisansları AB/UK gibi bölgeleri kapsam dışı bırakır (ör. Tencent Hunyuan lisansları). Kredi satılan ticari hizmette bu modeller lisans alınmadan kullanılamaz — **HUKUKİ**.
- **[Ö]** Çıktı mülkiyeti/hakları, telif ihlali iddialarında kaldırma süreci, kredilerin engellenen işlerde iadesi sözleşmede açık yazılmalı. AB'de hizmet verilecekse DSA bildirim-ve-kaldırma yükümlülükleri **HUKUKİ**.

## 7. Kullanıcı sistem talimatı
- **[Z]** Platform güvenlik kuralları sistem talimatına bağlı değildir: §1–2 filtreleri talimattan bağımsız sunucuda uygulanır; talimat bunları kapatamaz. Talimat da aynı prompt filtresinden geçer (kaydederken ve her kullanımda).
- **[Z]** Talimat kişisel veri/tercih verisidir: tercih deposunda (uyum-preferences) saklanmalı, export'a dahil edilmeli, erase'te silinmeli. Uzunluk sınırı (ör. 2–4 KB) ve kontrol karakteri temizliği.
- **[Z]** Platform sistem talimatı ve iç yapılandırma kullanıcıya veya çıktıya sızdırılmamalı; kullanıcı talimatı açık sınırlayıcılarla, platform talimatının altında ve "kullanıcı içeriği" olarak işaretlenerek eklenmeli.
- **[Ö]** Takım/paylaşılan talimatlar varsa sahiplik ve erişim kontrolü (API botunun takım sistemiyle) ayrıca tasarlanmalı; bir üyenin silinmesi paylaşılan talimatın akıbetini netleştirmeli.

## Hukukçuya eklenecek sorular (KANUS_PRIVACY_REPORT §10'a)
1. CSAM olayında kanıt saklama ve bildirim yükümlülüğü, usulü ve süresi.
2. AB YZ Yasası m.50 kapsamı, KANUS'un rolü, güncel uygulama/erteleme durumu.
3. Yüz tespitinin biyometrik veri sayılıp sayılmayacağı (anlık işlemede).
4. Model lisanslarının ticari kullanım ve bölge kısıtları; OpenRAIL aktarım metni.
5. Moderasyon logu ve legal hold saklama süreleri.
6. Yetişkin içerik açılacaksa yaş doğrulama ve yerel mevzuat.

## Ek (2026-11-08 01:25, US sorusu): .13 sohbet ayarları — kullanıcı sistem talimatı
Karar: Talimatın yalnız EXE yerel ayarlarında tutulup hook tarafından saklanmadan/loglanmadan iletilmesi .13 için **kabul edilebilir**; uyum-preferences'a taşıma .13'te zorunlu değil (veri minimizasyonu açısından tercih edilir). §7'deki "tercih deposu + export/erase" şartı yalnız sunucuda saklama halinde geçerlidir.
Şartlar: **[Z]** hiçbir yolda (hata logu, panic, debug dökümü, outbox, metrik, telemetri/crash) saklama/loglama yok, hata yolu dahil testli; birleşik prompt da loglanmaz · **[Z]** uzunluk sınırı ve kontrol karakteri temizliği hook'ta da · **[Z]** upstream'e iletim aydınlatma metnine eklenir · **[Z]** EXE dosyası kullanıcıya özel izinle, telemetri/senkron/destek paketine girmez, sıfırla/kaldır ile silinir · **[Z]** sınırlayıcılı katmanlama, platform talimatı sızmaz · **[Ö]** senkron eklenirse prefs deposuna taşınır · **[Ö]** metin sohbeti içerik filtresi ayrı ürün kararı.
