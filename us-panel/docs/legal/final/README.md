# Yasal metinler — yayın için final (ORG taslakları korundu)

**Üretim:** 2026-11-06 09:15 Europe/Istanbul  
**Kaynak taslaklar:** `/workspace/us-panel/docs/legal/*.md` (üzerine yazılmadı)  
**Bu klasör:** `/workspace/us-panel/docs/legal/final/`  
**Uyarı:** Hukukçu incelemesi önerilir. Şirket bilgisi uydurulmamıştır; eksikler `<DOLDUR>` ile işaretlidir. Shopify’a henüz yapıştırılmamıştır.

## Mağaza olguları (metinlere işlendi)
- Mağaza: https://shop.kanus.org  
- Para birimi: TRY; fiyatlar KDV dahil  
- Dijital lisans/kredi/süre; fiziksel kargo yok  
- Destek: support@kanus.org  

## Dosya → Shopify hedefi

| Dosya | Shopify hedefi |
|---|---|
| `CAYMA_VE_IADE_POLITIKASI.md` | **Ayarlar → Politikalar → İade politikası** (Refund policy) |
| `KULLANIM_SARTLARI.md` | **Ayarlar → Politikalar → Hizmet şartları** (Terms of service) |
| `GIZLILIK_POLITIKASI.md` | **Ayarlar → Politikalar → Gizlilik politikası** (mevcut şablonun yerine/güncelleme) |
| `KVKK_AYDINLATMA_METNI.md` | **Online Mağaza → Sayfalar** — örn. handle `kvkk` + footer link |
| `MESAFELI_SATIS_SOZLESMESI.md` | **Sayfalar** — örn. `mesafeli-satis` + checkout ek sözleşme alanı |
| `ON_BILGILENDIRME_FORMU.md` | Checkout öncesi bilgilendirme / ek uygulama veya Sayfa + ödeme adımı referansı |
| `ILETISIM.md` | **Sayfalar → İletişim** (`contact` — şu an boş) ve/veya Politikalar → İletişim bilgileri |

Markdown’u Admin’e yapıştırırken gerekirse HTML’e çevirin; başlıkları koruyun.

## `<DOLDUR>` alanları (dosya bazında)

### `CAYMA_VE_IADE_POLITIKASI.md`
- - Unvan: `<DOLDUR>`
- - Açık adres: `<DOLDUR>`
- - VKN / MERSİS: `<DOLDUR>`
- - Telefon: `<DOLDUR>`
- - KEP (varsa): `<DOLDUR>`
- - Yasal bildirim / cayma bildirimi kanalı: `<DOLDUR>` (support@kanus.org yetkili ise bunu açıkça belirtin)
- Örnek onay metni (checkout’ta gösterilecek sürüm: `<DOLDUR>`):
- Sistemdeki rezervasyon çözümü / kredi iadesi **kart-banka para iadesi değildir**. Para iadesi, ödeme aracına uygun şekilde ve satıcının mali sürecine göre yapılır. Süreç: `<DOLDUR>`.

### `KULLANIM_SARTLARI.md`
- **Sağlayıcı:** Unvan `<DOLDUR>` · Adres `<DOLDUR>` · VKN/MERSİS `<DOLDUR>` · Telefon `<DOLDUR>` · KEP `<DOLDUR>` · Destek support@kanus.org · Yasal bildirim `<DOLDUR>`.
- Yürürlük / sürüm tarihi: `<DOLDUR>`.
- Lisansın cihaz / kullanıcı / devir kuralları: `<DOLDUR>`. Hesap kısıtlama ve itiraz süreci: `<DOLDUR>`.
- GPU/model hazırlık veya kuyruk mesajı, hizmetin hazır olduğu veya sonucun doğru olduğu garantisi değildir. Kullanılabilirlik / SLA: `<DOLDUR>`.
- İçeriğin model eğitiminde kullanımı ve saklama: `<DOLDUR>`.
- Cayma/iade politikası ve mesafeli satış sözleşmesi uygulanır. Emredici tüketici hakları kaldırılamaz. Kanunen sınırlandırılamayan sorumluluklar saklıdır; ek sorumluluk sınırları: `<DOLDUR>`.
- Şart güncellemeleri: `<DOLDUR>`. Tüketici uyuşmazlıklarında 6502 ve ilgili mevzuat uygulanır.

### `KVKK_AYDINLATMA_METNI.md`
- - Unvan: `<DOLDUR>`
- - Adres: `<DOLDUR>`
- - VKN / MERSİS: `<DOLDUR>`
- - Telefon: `<DOLDUR>`
- - KEP / KVKK başvuru kanalı: `<DOLDUR>`
- - Genel destek: support@kanus.org — KVKK başvurusu için yetkili olup olmadığı: `<DOLDUR>`
- | E-posta, parola özeti, oturum | Hesap / erişim | Sözleşmenin ifası (m.5/2-c); güvenlik için meşru menfaat değerlendirmesi `<DOLDUR>` |
- | Sipariş kimliği/adı, e-posta, ürün/plan, lisans ilişkisi (Shopify sipariş bildirimleri) | Dijital hakkın teslimi, iptal/iade mutabakatı | m.5/2-c; mali kayıt dayanağı `<DOLDUR>` |
- | Lisans/kredi rezervasyon ve kullanım kayıtları | Bakiye, uyuşmazlık | m.5/2-c; m.5/2-e `<DOLDUR>` |
- | IP / hız sınırlama kayıtları | Kötüye kullanım önleme | m.5/2-f denge testi `<DOLDUR>` |
- | Modele gönderilen istem / proje içeriği | İstenen işlemi yürütme | Kapsam ve yetki `<DOLDUR>` |
- | Destek yazışmaları | Talep yönetimi | `<DOLDUR>` |
- Özel nitelikli veri göndermeyin. İstemlerin eğitimde kullanılıp kullanılmadığı: `<DOLDUR>`.
- - E-ticaret altyapısı (Shopify) — rol/ülke/sözleşme: `<DOLDUR>`
- - Ödeme kuruluşu: `<DOLDUR>`
- - Barındırma / e-posta (ör. Hostinger SMTP kaydı yerel belgelerde geçer) — tüzel kişi/rol: `<DOLDUR>`
- - Model / bulut altyapısı — ülke ve aktarım mekanizması: `<DOLDUR>`
- **Yurt dışı aktarım:** Fiilî envanter ve m.9 mekanizması `<DOLDUR>`. «Aktarım yoktur» denilemez; model/barındırma konumu doğrulanmalıdır.
- Amaç ve mevzuat için gerekli süre: hesap `<DOLDUR>`, sipariş/mali `<DOLDUR>`, log `<DOLDUR>`, destek `<DOLDUR>`. Güvenlik önlemleri: `<DOLDUR>`.
- m.11 hakları (öğrenme, düzeltme, silme, itiraz vb.). Başvuru: `<DOLDUR>`. En geç 30 gün içinde yanıt; Kurula şikâyet hakkı saklıdır.

### `MESAFELI_SATIS_SOZLESMESI.md`
- **Satıcı/sağlayıcı:** Unvan `<DOLDUR>` · Adres `<DOLDUR>` · VKN/MERSİS `<DOLDUR>` · Telefon `<DOLDUR>` · Sözleşme/şikâyet e-postası `<DOLDUR>` · Destek support@kanus.org · KEP `<DOLDUR>`.
- | Ödeme yöntemi | Checkout’ta seçilen · sağlayıcı `<DOLDUR>` |
- | EXE / sistem gereksinimi | `<DOLDUR>` |
- Lisans anahtarı üretimi / e-posta iletimi ifanın parçasıdır; her paket «tüm hizmet bitmiştir» anlamına gelmez. Kredi ve saniye ayrıdır. Teknik rezervasyon çözümü para iadesi değildir. Ölçüm özeti ürün sayfasında / ön bilgilendirmede: `<DOLDUR>`.
- Şikâyet: support@kanus.org ve `<DOLDUR>`. Tüketici hakem heyeti / mahkeme yolları saklıdır.
- KVKK aydınlatma metni ayrıca sunulur; bu sözleşme genel açık rıza değildir. Saklama: `<DOLDUR>`.

### `ON_BILGILENDIRME_FORMU.md`
- - Unvan: `<DOLDUR>`
- - Adres: `<DOLDUR>`
- - VKN / MERSİS: `<DOLDUR>`
- - Telefon: `<DOLDUR>`
- - E-posta: support@kanus.org / yasal: `<DOLDUR>`
- - KEP: `<DOLDUR>`
- Toplam: *[sipariş tutarı]* **TRY** (**KDV dahil**). Ek ücret yoksa «yok» yazılır; varsa: `<DOLDUR>`.
- Ödeme: checkout’taki yöntem · sağlayıcı `<DOLDUR>`.
- İfa: ödeme onayından sonra lisans anahtarı e-posta ile ve/veya US uygulamasında Hesap ekranına bağlama. Taahhüt edilen azami süre: `<DOLDUR>`.
- **Dijital anında teslim / ifa:** Yönetmelik m.15 istisnaları gündeme gelebilir. Erken ifa için **ayrı açık onay** alınır; onay verilmezse ifa ertelenir / alternatif: `<DOLDUR>`.
- Desteklenen işletim sistemi / EXE sürümü: `<DOLDUR>`. İnternet bağlantısı gerekebilir.
- support@kanus.org · `<DOLDUR>`.

### `ILETISIM.md`
- **Satıcı unvanı:** `<DOLDUR>`
- **Açık adres:** `<DOLDUR>`
- **VKN / MERSİS:** `<DOLDUR>`
- **Telefon:** `<DOLDUR>`
- **KEP:** `<DOLDUR>`
- **Yasal bildirim / resmi yazışma:** `<DOLDUR>`
- Destek saatleri: `<DOLDUR>`.

### `GIZLILIK_POLITIKASI.md`
- Unvan `<DOLDUR>` · Adres `<DOLDUR>` · VKN/MERSİS `<DOLDUR>` · Telefon `<DOLDUR>` · KEP `<DOLDUR>` · Destek support@kanus.org.
- Shopify, ödeme kuruluşu `<DOLDUR>`, barındırma/e-posta `<DOLDUR>`, model altyapısı `<DOLDUR>` ve zorunlu merciler. Yurt dışı aktarım: `<DOLDUR>`.
- KVKK m.11 hakları. Başvuru: `<DOLDUR>`. Ayrıntı: KVKK aydınlatma metni sayfası.

## Ortak satıcı kimliği (çoğu dosyada tekrarlar)
Tüm metinlerde doldurulması gereken ortak küme:
- `<DOLDUR>` Unvan (şirket / gerçek kişi)
- `<DOLDUR>` Açık adres
- `<DOLDUR>` VKN / MERSİS
- `<DOLDUR>` Telefon
- `<DOLDUR>` KEP (varsa)
- `<DOLDUR>` Yasal bildirim / KVKK başvuru / cayma bildirimi kanalı (support@kanus.org yetkisi netleştirilsin)
- `<DOLDUR>` Ödeme kuruluşu tüzel kişisi
- `<DOLDUR>` Barındırma / model altyapısı aktarım envanteri (KVKK)
- `<DOLDUR>` Checkout açık onay metni sürümü ve kayıt yöntemi (dijital anında ifa)

## Sonraki adım
Demir + hukukçu doldurur → Demir «canlıya yapıştır» onayı → SHOP politikaları/sayfaları günceller (ayrı görev).
