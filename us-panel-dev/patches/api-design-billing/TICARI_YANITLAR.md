# Ticari'nin 8 sorusuna yanıtlar (API botu)
Kök: `/workspace/us-panel-dev/work/api/r10_base/`. Ayrıntı: `BILLING_LEDGER_DESIGN.md` (aynı dizin).
Biçim: **(a) OLGU** (dosya:satır) · **(b) ÖNERİ** · gerekiyorsa **KARAR GEREKİR (Demir)**.

**1. Go hook'ta subscription nesnesi var mı?**
(a) Yok. Yalnız tek seferlik grant: her satın alınan adet bir `License` (`licensesrv/store.go:29-42`, üretim `licensesrv/webhook.go:176-196`). `Plan.Kind "sub"` abonelik değil, istem-kredisi türü (`licensesrv/store.go:22`); US-MONTHLY-200 de 30 günlük tek seferlik grant (`licensesrv/pricing.go:42`). Webhook yalnız `orders/paid`, `orders/cancelled`, `refunds/create` işliyor (`licensesrv/webhook.go:93-124`); yenileme olayı yok. Paddle/LemonSqueezy kodu yok.
(b) `subscription` varlığı + dönem başına `credit_grant` (tasarım §2–4).

**2. US_PLANS_JSON yenilemeyi nasıl temsil edecek?**
(a) Bugün temsil etmiyor: biçim `product_id → {kind, days, prompts, seconds, credits}` (`licensesrv/plans.go:11-60`, `licensesrv/store.go:21-27`); `US_PLANS_FILE` önceliklidir (`cmd/us-license-hook/main.go:37-47`); boşsa `LivePlansJSON()` (`licensesrv/pricing.go:55-65`). Recurring/interval alanı yok.
(b) Geriye uyumlu, isteğe bağlı alanlar: `"recurring": true, "interval": "month", "renewal_mode": "replace|stack", "rollover_cap": N, "max_concurrent": N`, ve sağlayıcı anahtarı için `"provider_refs": {"paddle":"pri_…","lemonsqueezy":"…"}`. Eski dosyalar aynen geçerli kalır. `renewal_mode`/`rollover_cap` değerleri **KARAR GEREKİR (Demir)**.

**3. Shopify + Paddle müşterisi hangi anahtarla tek account_id'de birleşir?**
(a) Bugün Shopify müşteri kimliği saklanmıyor; lisansa sipariş e-postası yazılıyor (`licensesrv/webhook.go:164-173,192`). Hesaba bağlama yalnız lisans anahtarının `/account/redeem` ile girilmesiyle (`licensesrv/accounts.go:201-214,531-545`); otomatik e-posta eşleşmesi yok. Hesap e-postası normalize ediliyor (`licensesrv/accounts.go:44`).
(b) Birincil anahtar iç `account_id`; `account_provider_link(provider, provider_customer)` tablosu. Otomatik bağlama yalnız **doğrulanmış** hesap e-postası = sağlayıcının e-postası olduğunda; aksi halde redeem/checkout'a `account_id`'yi custom_data/note attribute olarak geçirme. Düz e-posta eşleşmesiyle doğrulanmamış hesaba kredi bağlanmaz (hesap ele geçirme riski).

**4. Ledger prod VPS'te PostgreSQL'de mi, JSON store'da mı? Geçiş planı?**
(a) JSON store: tek dosya, flock + atomik rename (`licensesrv/store.go:101-186`), yol `US_STORE_PATH` (`cmd/us-license-hook/main.go:50`; deploy örneği `/opt/us-license-hook/licenses.json`, `us-panel/deploy/us-license-hook/us-license-hook.env.example:5`). "PostgreSQL/SQLite aktif dağıtım değildir" (`us-panel/reports/LISANS_URETICI_PLAN_TR.md:7`). Go'da PostgreSQL kodu yok.
(b) Önce JSON'a ekleme-only `Ledger` gölge yazımı + mutabakat testi; PG'ye geçiş: DDL (tasarım §2.1) → çift yazım → mutabakat → okuma geçişi → JSON salt-okunur arşiv. **KARAR GEREKİR (Demir):** PG'ye geçilecek mi, ne zaman (VPS'te PG kurulumu/yedek sorumluluğu dahil).

**5. Rezervasyon TTL ile period_end nasıl etkileşir?**
(a) TTL: HTTP kredi API 45 dk (`licensesrv/credits.go:19`), model proxy `US_MODEL_TIMEOUT_SECONDS` (vars. 2700 s) + 5 dk (`licensesrv/model_meter.go:56`, `cmd/us-license-hook/main.go:114`). Reserve ve commit ikisi de `now < License.ExpiresAt` istiyor (`licensesrv/store.go:240,281`); süresi dolmuş rezervasyon commit edilemez (`licensesrv/store.go:276-278`) ve release olarak kapanır (`licensesrv/store.go:302-312`). Sonuç: bitişe yakın başlayan uzun çağrı bitişten sonra commit edilemez → kredi düşmez.
(b) Commit'i rezervasyonun açıldığı andaki geçerliliğe bağla; yeni reserve için `expires_at - now >= TTL` şartı, yetmiyorsa sıradaki grant. Davranış değişikliği → önce regresyon testi, ayrı yama.

**6. Yenileme başarısız / kart reddi e-posta şablonlarını kim hazırlayacak?**
(a) Kodda bu şablonlar yok. Mevcut şablonlar kod içinde sabit metin: kayıt, sıfırlama, doğrulama (`licensesrv/accounts.go:384,448,452`), sipariş (`licensesrv/reversals.go:102-107`); gönderim Outbox + SMTP (`licensesrv/mail.go`).
(b) MoR (Paddle/LS) kendi dunning e-postalarını gönderebiliyorsa onları kullanmak (çift e-postayı önlemek için bizim taraf göndermez); Shopify/manual için Outbox'a iki yeni şablon (metni Ticari yazar, API botu koda bağlar). **KARAR GEREKİR (Demir):** metinlerin sahibi ve MoR e-postalarının açık olup olmayacağı.

**7. Faturasız kurumsal kredi için admin-only API gerekli mi?**
(a) Yok; mevcut uçlar yalnız webhook, `/v1/credits/*`, `/account/*`, model proxy (`cmd/us-license-hook/main.go:71-105`, `licensesrv/accounts.go:304`). Bugün tek kredi kaynağı Shopify `orders/paid`.
(b) Enterprise (`KANUS_PRICING_GLOBAL.md` §4.4) satılacaksa gerekli: `POST /v1/admin/credit-grants` (loopback/mTLS + admin anahtarı, `Idempotency-Key` zorunlu, üst sınır, `admin_audit` kaydı) — tasarım §11. **KARAR GEREKİR (Demir):** gerekli mi.

**8. KDV: Shopify'da kim hesaplıyor? Ledger'da brüt/net olacak mı?**
(a) Kodda vergi alanı yok; katalog "Fiyatlar KDV dahil TL" (`licensesrv/pricing.go:9-11`). Webhook payload'ından tutar/vergi okunmuyor (`licensesrv/webhook.go:48-63` order yapısı).
(b) Grant'e `gross_minor, tax_minor, net_minor, currency, tax_collector(mor|merchant|none)`; değerler sağlayıcı payload'ından okunur, bizde hesaplanmaz; kredi bakiyesi parasal değil. **KARAR GEREKİR (Demir / muhasebe):** Shopify TR satışlarında KDV hesap/beyan sorumluluğu ve e-Fatura.

**Ek düzeltme (§5.6 "şu an abonelik önce — teyit"):** (a) Doğrulanamadı. Mevcut sıra: önce istem kredisi (`Kind "sub"`: paketler + aylık birlikte), sonra saat (`use`); aynı türde en erken biten önce (`licensesrv/accounts.go:117-137`). Abonelik ile paket arasında ayrı öncelik yok. (b) Öncelik alanı: abonelik dönemi → rollover → paket → admin → saat. **KARAR GEREKİR (Demir)** teyit.
