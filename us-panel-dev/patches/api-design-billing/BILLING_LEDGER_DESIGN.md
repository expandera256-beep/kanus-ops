# KAN.US — Abonelik / Kredi Defteri Tasarımı (TASLAK)

Durum: **yalnız tasarım**. Kod değişikliği, migration, deploy yok. Kaynak taban: `/workspace/us-panel-dev/work/api/r10_base/` (aşağıda yollar bu köke göredir). Girdi: `us-panel/reports/KANUS_PRICING_GLOBAL.md` §5, `GLOBAL_PROMPT_20261106.txt` §7.4–7.6.
Etiketler: **[OLGU]** koddan/belgeden doğrulandı · **[ÖNERİ]** tasarım önerisi · **[KARAR GEREKİR (Demir)]** iş/politika kararı.

## 0. Mevcut durum (doğrulanan olgular)
| Konu | Olgu | Kaynak |
|---|---|---|
| Depolama | Tek JSON dosyası, flock + temp-file + rename ile atomik yazım; `US_STORE_PATH` (varsayılan `us-license-store.json`; deploy örneği `/opt/us-license-hook/licenses.json`) | `licensesrv/store.go:101-186`, `cmd/us-license-hook/main.go:50`, `us-panel/deploy/us-license-hook/us-license-hook.env.example:5` |
| PostgreSQL | Go hook'ta **yok**. "PostgreSQL/SQLite aktif dağıtım değildir" | `us-panel/reports/LISANS_URETICI_PLAN_TR.md:7` |
| Plan modeli | `Plan{Kind sub|use, Days, Prompts, Seconds, Credits}`; yineleme (recurring) alanı **yok** | `licensesrv/store.go:21-27` |
| Hak birimi | Satın alınan her adet = bir `License` (kalan `Credits`, `Reserved`, `ExpiresAt`, `Revoked`) | `licensesrv/store.go:29-42`, `licensesrv/webhook.go:176-196` |
| Subscription nesnesi | **Yok.** "sub" adı yalnız istem-kredisi türü; US-MONTHLY-200 de 30 günlük tek seferlik grant | `licensesrv/pricing.go:35-44` |
| Reserve | Atomik; `Credits-Reserved >= amount`, TTL'li `Reservation` | `licensesrv/store.go:232-253` |
| Settle | İdempotent (aynı commit değeri tekrar = no-op; çelişen = hata); süresi geçmiş rezervasyon commit edilemez | `licensesrv/store.go:261-300` |
| Release/expire | Süresi dolan rezervasyon `Settlement{Commit:false}` olarak kapanır | `licensesrv/store.go:302-312` |
| TTL | HTTP kredi API: 45 dk sabit; model proxy: `budget(US_MODEL_TIMEOUT_SECONDS, vars. 2700 s)+5 dk` | `licensesrv/credits.go:19`, `licensesrv/model_meter.go:56`, `cmd/us-license-hook/main.go:114` |
| Tüketim sırası | Hesapta: önce `sub` (istem), sonra `use` (saat); aynı türde en erken bitiş önce | `licensesrv/accounts.go:117-137` |
| Webhook idempotency | `Webhooks[X-Shopify-Webhook-Id]`; ayrıca `Orders[order_id]` ikinci grant'i engeller; `Refunds[refund_id]`; `Cancelled[order_id]` tombstone geç `orders/paid`'i engeller | `licensesrv/webhook.go:135-160`, `licensesrv/reversals.go:17-59` |
| İade | Adet iadesi o kadar lisansı **revoke** eder (harcanmış olanlar dahil), kredi negatife inmez; yalnız tutar iadesi krediye çevrilmez | `licensesrv/reversals.go:59-78`, `licensesrv/store.go:314-327` |
| İmza | Tek `SHOPIFY_WEBHOOK_SECRET`, HMAC-SHA256; çoklu secret/rotasyon **yok** | `licensesrv/webhook.go:17-46` |
| Hesap birleşimi | Account = e-posta+parola; lisans `/account/redeem` ile anahtarla bağlanır (`LicenseOwners`); sipariş e-postasıyla otomatik eşleşme **yok** | `licensesrv/accounts.go:24-35,201-214,531-545` |
| Vergi | Fiyatlar "KDV dahil TL"; brüt/net/vergi alanı **yok** | `licensesrv/pricing.go:9-11` |
| Paddle/LemonSqueezy | Kodda **yok** | grep (licensesrv, cmd) |
| E-posta | Outbox + SMTP; şablonlar kod içinde sabit metin (sipariş, kayıt, doğrulama, sıfırlama); yenileme/kart reddi şablonu **yok** | `licensesrv/accounts.go:384,448,452`, `licensesrv/reversals.go:102-107`, `licensesrv/mail.go` |
| Admin grant API | **Yok** | route listesi `licensesrv/accounts.go:304`, `cmd/us-license-hook/main.go:71-105` |

## 1. Tasarım ilkeleri [ÖNERİ]
1. **Geriye uyum:** `License` = `CreditGrant`'in özel hali. Mevcut `Reserve/settleFor/expireLocked/revokeLocked` imza ve semantiği değişmez; ledger satırları bu fonksiyonların **yan ürünü** olarak eklenir (gölge yazım), bakiye hâlâ `License.Credits/Reserved`'dan okunur. Ledger bakiyesi ile `License` bakiyesi tutarlılık testi geçmeden kaynak değiştirilmez.
2. **Append-only ledger:** Satır güncellenmez/silinmez; düzeltme = yeni `adjust` satırı.
3. **Tek yazar:** JSON store'da zaten flock ile tek yazar; PG'de `SERIALIZABLE` veya grant satırında `SELECT … FOR UPDATE`.
4. **Sağlayıcıdan bağımsız iç model:** Provider durumları iç duruma tablo ile eşlenir (§3.2).
5. Ürün kimliğine göre plan (fiyat/başlığa göre değil) — mevcut kural korunur (`licensesrv/store.go:20`).

## 2. Şema

### 2.1 PostgreSQL DDL taslağı [ÖNERİ — uygulanmadı]
```sql
CREATE TYPE sub_provider AS ENUM ('shopify','paddle','lemonsqueezy','manual');
CREATE TYPE sub_status   AS ENUM ('trialing','active','past_due','paused','canceled','expired');
CREATE TYPE ledger_kind  AS ENUM ('grant','reserve','commit','release','refund','expire','adjust');
CREATE TYPE credit_unit  AS ENUM ('prompts','seconds');

CREATE TABLE account (
  id              text PRIMARY KEY,                 -- mevcut Account.ID
  email_norm      text NOT NULL UNIQUE,             -- normalizedEmail()
  email_verified  boolean NOT NULL DEFAULT false,
  created_at      timestamptz NOT NULL DEFAULT now()
  -- parola/oturum alanları mevcut accounts modunda kalır
);

CREATE TABLE account_provider_link (               -- çoklu sağlayıcı birleşimi
  provider          sub_provider NOT NULL,
  provider_customer text NOT NULL,                  -- Shopify customer.id / Paddle ctm_… / LS customer_id
  account_id        text NOT NULL REFERENCES account(id),
  linked_by         text NOT NULL,                  -- 'verified_email' | 'redeem' | 'admin'
  linked_at         timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (provider, provider_customer)
);

CREATE TABLE plan (                                 -- US_PLANS_JSON karşılığı
  id            text PRIMARY KEY,                   -- SKU, ör. US-MONTHLY-200
  kind          text NOT NULL CHECK (kind IN ('sub','use')),  -- mevcut Plan.Kind
  unit          credit_unit NOT NULL,
  amount        integer NOT NULL CHECK (amount > 0),
  valid_days    integer NOT NULL CHECK (valid_days BETWEEN 1 AND 3650),
  recurring     boolean NOT NULL DEFAULT false,
  interval      text CHECK (interval IN ('month','year')),
  renewal_mode  text CHECK (renewal_mode IN ('replace','stack')),   -- KARAR GEREKİR
  rollover_cap  integer,                                             -- KARAR GEREKİR
  max_concurrent_reservations integer NOT NULL DEFAULT 1
);
CREATE TABLE plan_provider_product (
  provider sub_provider NOT NULL, product_ref text NOT NULL, plan_id text NOT NULL REFERENCES plan(id),
  PRIMARY KEY (provider, product_ref)
);

CREATE TABLE subscription (
  id                    uuid PRIMARY KEY,
  account_id            text NOT NULL REFERENCES account(id),
  provider              sub_provider NOT NULL,
  provider_sub_id       text NOT NULL,
  plan_id               text NOT NULL REFERENCES plan(id),
  status                sub_status NOT NULL,
  current_period_start  timestamptz NOT NULL,
  current_period_end    timestamptz NOT NULL,
  cancel_at_period_end  boolean NOT NULL DEFAULT false,
  trial_end             timestamptz,
  status_changed_at     timestamptz NOT NULL,
  last_event_at         timestamptz NOT NULL,       -- geç/sırasız event koruması
  UNIQUE (provider, provider_sub_id)
);

CREATE TABLE credit_grant (                        -- mevcut License
  id               text PRIMARY KEY,               -- mevcut License.Key (yeni grant'lerde uuid)
  account_id       text REFERENCES account(id),    -- NULL = henüz redeem edilmemiş lisans
  unit             credit_unit NOT NULL,
  amount           integer NOT NULL CHECK (amount >= 0),
  source           text NOT NULL CHECK (source IN ('order','subscription_period','admin','rollover','migration')),
  source_ref       text NOT NULL,                  -- order_id:line_id:i | sub_id:period_start | admin_req_id
  subscription_id  uuid REFERENCES subscription(id),
  plan_id          text REFERENCES plan(id),
  priority         smallint NOT NULL,              -- tüketim sırası (§6)
  starts_at        timestamptz NOT NULL,
  expires_at       timestamptz NOT NULL,
  revoked_at       timestamptz,
  gross_minor      bigint, net_minor bigint, tax_minor bigint,   -- §9
  currency         char(3), tax_collector text CHECK (tax_collector IN ('mor','merchant','none')),
  UNIQUE (source, source_ref)                      -- çift grant'e karşı ikinci kilit
);

CREATE TABLE ledger_entry (                         -- append-only
  seq            bigserial PRIMARY KEY,
  grant_id       text NOT NULL REFERENCES credit_grant(id),
  account_id     text,
  kind           ledger_kind NOT NULL,
  delta_avail    integer NOT NULL,                 -- kullanılabilir bakiye değişimi
  delta_reserved integer NOT NULL,                 -- rezerve değişimi
  reservation_id text,                             -- mevcut Reservation id
  idem_key       text NOT NULL UNIQUE,             -- ör. 'reserve:'||res_id, 'commit:'||res_id
  cause_event    bigint REFERENCES webhook_event(id),
  actor          text NOT NULL,                    -- 'system'|'webhook'|'admin:<id>'
  reason         text,
  created_at     timestamptz NOT NULL DEFAULT now()
);
-- UPDATE/DELETE yasak:
REVOKE UPDATE, DELETE ON ledger_entry FROM PUBLIC;
CREATE RULE ledger_no_update AS ON UPDATE TO ledger_entry DO INSTEAD NOTHING;
CREATE RULE ledger_no_delete AS ON DELETE TO ledger_entry DO INSTEAD NOTHING;
CREATE VIEW grant_balance AS
  SELECT grant_id, SUM(delta_avail) AS available, SUM(delta_reserved) AS reserved
  FROM ledger_entry GROUP BY grant_id;

CREATE TABLE reservation (                         -- açık rezervasyonlar (mevcut Reservations)
  id text PRIMARY KEY, grant_id text NOT NULL REFERENCES credit_grant(id),
  amount integer NOT NULL, created_at timestamptz NOT NULL, expires_at timestamptz NOT NULL,
  closed_kind ledger_kind, closed_at timestamptz   -- commit|release|expire
);

CREATE TABLE webhook_event (
  id           bigserial PRIMARY KEY,
  provider     sub_provider NOT NULL,
  event_id     text NOT NULL,                      -- Shopify X-Shopify-Webhook-Id / Paddle event_id / LS meta.event_id
  topic        text NOT NULL,
  payment_ref  text,                               -- tekilleştirme anahtarı (§7.1)
  occurred_at  timestamptz,                        -- sağlayıcının olay zamanı
  received_at  timestamptz NOT NULL DEFAULT now(),
  payload_sha256 bytea NOT NULL,
  secret_slot  text,                               -- hangi imza secret'ı doğruladı (current|previous)
  processed_at timestamptz, result text,           -- applied|duplicate|ignored_late|error
  UNIQUE (provider, event_id)
);
CREATE UNIQUE INDEX webhook_payment_once ON webhook_event(payment_ref) WHERE result = 'applied' AND payment_ref IS NOT NULL;

CREATE TABLE admin_audit (                          -- §7.6 denetim kaydı, admin grant için zorunlu
  id bigserial PRIMARY KEY, actor text NOT NULL, action text NOT NULL,
  target text NOT NULL, request_id text NOT NULL UNIQUE, detail jsonb, created_at timestamptz NOT NULL DEFAULT now()
);
```

### 2.2 Mevcut JSON store eşlemesi [OLGU alanlar `licensesrv/store.go:79-90`; eşleme ÖNERİ]
| Mevcut `state` alanı | Yeni varlık | Not |
|---|---|---|
| `Accounts` | `account` | aynen |
| `LicenseOwners[key]=accountID` | `credit_grant.account_id` | |
| `Licenses[key]` | `credit_grant` (id=key) + bir `grant` ledger satırı (`delta_avail=Plan.Credits` veya `Seconds`) | `Plan.Kind` → `unit` (`sub`→prompts, `use`→seconds) |
| `Reservations[id]` | `reservation` + `reserve` ledger satırı (`delta_avail=-a, delta_reserved=+a`) | |
| `Settlements[id]` `Commit:true` | `commit` satırı (`delta_reserved=-a`) | |
| `Settlements[id]` `Commit:false` (settle) | `release` (`delta_avail=+a, delta_reserved=-a`) | |
| `Settlements` (expireLocked) | `expire`… **dikkat:** mevcut kod rezervasyon süresi dolmasını da `Commit:false` yazar; ayırt etmek için yeni alan gerekir | ayrım için `Settlement.Reason` (opsiyonel, omitempty) önerilir — geriye uyumlu |
| `revokeLocked` | `refund` (iade kaynaklıysa) veya `adjust` + `credit_grant.revoked_at` | |
| `Webhooks[id]` | `webhook_event(provider='shopify', event_id=id)` | |
| `Orders[oid]` | `credit_grant.source_ref` önekleri | |
| `Cancelled[oid]` | `webhook_event` + order tombstone | |
| `Refunds[fid]` | `webhook_event` + `refund` satırları | |
| `Outbox` | ayrı `mail_outbox` tablosu (bu belgenin kapsamı dışı) | |

**JSON'da ara aşama [ÖNERİ]:** `state`'e yalnız eklemeli alanlar: `Subscriptions map[string]*Subscription`, `Ledger []LedgerEntry` (append), `Events map[string]WebhookEvent` (anahtar `provider+":"+event_id`). Mevcut alanlar silinmez; eski binary yeni dosyayı okuyabilir (bilinmeyen JSON alanları yok sayılır). Dosya büyümesi için ledger'ın PG'ye taşınması ayrı karar (Q4).

## 3. Abonelik durum makinesi

### 3.1 Geçişler [ÖNERİ]
| Kaynak → Hedef | Tetik | Kredi etkisi |
|---|---|---|
| (yok) → trialing | sağlayıcı trial başladı | trial grant (miktar KARAR GEREKİR; varsayılan 0) |
| (yok)/trialing → active | ilk ödeme başarılı | dönem grant'i |
| active → active | yenileme ödemesi başarılı (yeni period) | yeni dönem grant'i; eski dönem grant'i replace/stack kuralına göre (§4) |
| active → past_due | yenileme ödemesi başarısız | bkz. erişim kuralı |
| past_due → active | tahsilat sonradan başarılı | dönem grant'i (bir kez, payment_ref ile) |
| past_due → canceled/expired | sağlayıcı dunning bitti | kalan abonelik grant'leri dönem sonunda `expire` |
| active → paused / paused → active | sağlayıcı pause/resume | pause'da yeni grant yok; mevcut grant donmaz (ÖNERİ) |
| active → canceled (`cancel_at_period_end=true`) | kullanıcı iptal | erişim period_end'e kadar sürer |
| canceled → expired | `now >= current_period_end` | kalan grant `expire` satırı |
| her durum → expired | iade/chargeback ile tam geri alma | §5 |
Geçersiz geçiş (ör. expired → active) yeni abonelik kaydı gerektirir; olay `result=ignored_invalid_transition` ile kaydedilir.

### 3.2 Sağlayıcı eşlemesi [ÖNERİ; sağlayıcı alan adları doğrulanmadı — entegrasyon öncesi sandbox'ta teyit]
| İç | Paddle | Lemon Squeezy | Shopify | manual |
|---|---|---|---|---|
| trialing | trialing | on_trial | (Shopify Subscriptions TR'de NOT KNOWN) | admin |
| active | active | active | ACTIVE | admin |
| past_due | past_due | past_due / unpaid | FAILED/dunning | — |
| paused | paused | paused | PAUSED | admin |
| canceled | canceled (scheduled) | cancelled | CANCELLED | admin |
| expired | canceled + period bitti | expired | EXPIRED | admin |
Eşleme tablosunun sahibi: API/ledger kodu (`provider_status_map`), testle kilitli.

### 3.3 past_due erişimi — **KARAR GEREKİR (Demir)**
Seçenekler: (A) hard-block: abonelik grant'lerinden reserve reddi; paket kredileri kullanılabilir. (B) soft-block: N gün grace (ör. 3), sonra hard. (C) hiç blok yok, dunning bitene kadar. Öneri: **A** veya kısa grace'li **B**; çünkü model çağrısı GPU maliyeti doğurur. Uygulama noktası: `modelLicenseLocked` uygunluk filtresi (`licensesrv/accounts.go:111-115`) — grant'in bağlı aboneliği past_due ise elenir.

## 4. Yenileme, iptal, kota
- **Yenileme grant'i:** `source='subscription_period'`, `source_ref = sub_id:period_start`; `UNIQUE(source,source_ref)` aynı dönemin iki kez verilmesini engeller.
- **replace vs stack — KARAR GEREKİR (Demir).** replace: yeni dönemde eski dönem grant'inin kalanı `expire` satırıyla sıfırlanır (açık rezervasyon varsa §8). stack: eski grant kendi `expires_at`'ine kadar yaşar. Mevcut davranış: her grant bağımsız ve kendi `Days` süresiyle yaşıyor (`licensesrv/webhook.go:177`) → bugünkü US-MONTHLY-200 fiilen **stack** gibi davranıyor [OLGU].
- **Carry-over sınırı — KARAR GEREKİR (Demir).** Uygulama: yenilemede eski kalan `min(kalan, rollover_cap)` kadar yeni `rollover` grant'ine aktarılır (`expire` + `grant` çifti, aynı idem_key öneki).
- **Dönem ortası değişiklik — KARAR GEREKİR (Demir).** Öneri: MoR'da proration'ı **sağlayıcı** hesaplasın; biz yalnız sağlayıcının yeni plan + yeni period bilgisine göre grant açarız. Upgrade: anında yeni plan grant'i (fark = yeni plan − eski kalan, ya da tam yeni plan; politika). Downgrade: dönem sonunda.
- **İptal:** `cancel_at_period_end=true`; dönem sonuna kadar kullanım; dönem sonunda kalan `expire`.
- **Kota:** Abonelik katmanı kotası = dönem grant'i miktarı (Starter 50 / Pro 200 / Business 500, `KANUS_PRICING_GLOBAL.md` §4.3 öneri). `max_concurrent_reservations` plan alanı; mevcut kodda katmana bağlı eşzamanlı rezervasyon sınırı **yok** [OLGU], tek sınır bakiyedir (`licensesrv/store.go:242`) ve proxy kapasitesi (`cmd/us-license-hook/main.go:113-126`). Öneri: Starter 1, Pro 2, Business 4 — KARAR GEREKİR.

## 5. İade ve chargeback
- **Mevcut [OLGU]:** adet iadesi → o kadar lisans revoke (harcanmış dahil), negatif bakiye yok; tutar-only iade → krediye dokunulmaz (`licensesrv/reversals.go:59-61`).
- **Öneri:** Tüm sağlayıcı iadeleri tek bir `applyReversal(provider, refund_id, payment_ref, scope)` yolundan geçer; idempotency `(provider, refund_id)`.
  - Kredi paketi, tam iade: grant'in kalanı `refund` satırıyla sıfırlanır, `revoked_at` set.
  - Kısmi tutar iadesi: oran = `refund_gross / grant_gross`; geri alınacak = `floor(amount * oran)`, kalandan düşülür. **KARAR GEREKİR** (mevcut kod oran uygulamıyor).
  - Abonelik iadesi: ilgili dönem grant'i (`source_ref=sub_id:period_start`).
  - Chargeback: tam iade gibi + hesap `billing_hold` bayrağı (yeni reserve reddi) — öneri.
  - Kullanılmış kredi: kalan < geri alınacak ise fark **borç** olarak `adjust` negatif mi yazılır, yoksa sıfırda mı durur? **KARAR GEREKİR (Demir).** Öneri: sıfırda dur (mevcut davranışla uyumlu), chargeback'te hesap hold.
  - Açık rezervasyonu olan grant iade edilirse: rezervasyon `release` ile kapatılır, sonra `refund` (mevcut `revokeLocked` aynı şeyi yapıyor, `licensesrv/store.go:321-326`).

## 6. Kredi tüketim sırası
- **Mevcut [OLGU]** (`licensesrv/accounts.go:117-137`): önce `Kind=="sub"` (istem kredisi — paketler ve aylık AYNI türde), sonra `use` (saat). Aynı türde en erken `ExpiresAt` önce. Yani bugün "abonelik önce, paket sonra" diye bir ayrım **yok**; US-MONTHLY-200 ve kredi paketleri yalnızca bitiş tarihine göre sıralanıyor. Ticari'nin "şu an abonelik önce" ifadesi bu anlamda **doğrulanamadı**; doğru ifade "istem kredisi saatten önce, sonra en erken biten önce".
- **Öneri:** `priority`: 10 = abonelik dönem grant'i (en kısa ömür, kullanılmazsa yanar), 20 = rollover, 30 = kredi paketi, 40 = admin/kurumsal, 90 = `use` (saniye). İçeride `expires_at ASC, id ASC`. Abonelik-önce kuralı ile en-erken-biten-önce çoğu durumda aynı sonucu verir. Teyit **KARAR GEREKİR** (iş kuralı).

## 7. Idempotency
### 7.1 orders/paid ↔ subscription_payment_succeeded
- Katman 1: `webhook_event UNIQUE(provider,event_id)` → yeniden teslim no-op (mevcut `Webhooks` haritasının genellemesi).
- Katman 2: `payment_ref` = sağlayıcıdaki **ödeme/fatura kimliği** (Shopify: `order.id`; Paddle: `transaction_id`; LS: `order_id`/`subscription_invoice_id`). `webhook_payment_once` kısmi unique indeksi aynı ödemenin iki olaydan iki kez uygulanmasını engeller.
- Katman 3: `credit_grant UNIQUE(source,source_ref)`.
- Shopify abonelik siparişi hem `orders/paid` hem abonelik olayı üretirse: **grant yalnız `orders/paid`'den** verilir (mevcut yol), abonelik olayı yalnız `subscription` durumunu günceller ve `payment_ref` eşleşirse grant'i atlar. Böylece mevcut `Orders[oid]` kontrolü (`licensesrv/webhook.go:148-153`) korunur.
### 7.2 Geç / sırasız event
- `subscription.last_event_at`: `occurred_at < last_event_at` olan durum değişikliği uygulanmaz (`ignored_stale`), ama kaydedilir.
- Ödeme olayı `current_period_end` geçtikten sonra gelirse: grant `starts_at=olay dönemi başı`, `expires_at=olay dönemi sonu` ile hesaplanır; dönem zaten bittiyse grant **verilmez**, `ignored_late` + admin uyarısı (manuel karar). Alternatif (geç geleni kısa grace ile vermek) KARAR GEREKİR.
- İptal/iade önce gelir, ödeme sonra gelirse: mevcut tombstone deseni (`Cancelled`, `licensesrv/reversals.go:22-23`; `applyRefundsLocked` ödeme sonrası yeniden uygulanır `licensesrv/webhook.go:203`) tüm sağlayıcılara genellenir.
### 7.3 İmza secret rotasyonu [ÖNERİ; mevcut kod tek secret, OLGU `licensesrv/webhook.go:34-46`]
1. Yeni env adı: `SHOPIFY_WEBHOOK_SECRET_PREVIOUS` (ve sağlayıcı başına `PADDLE_WEBHOOK_SECRET`, `PADDLE_WEBHOOK_SECRET_PREVIOUS`, `LEMONSQUEEZY_WEBHOOK_SECRET`, `…_PREVIOUS`). Doğrulama: önce current, olmazsa previous; `secret_slot` kaydedilir.
2. Prosedür: sağlayıcıda yeni secret üret → sunucuda current=yeni, previous=eski → reload → sağlayıcıda geçiş → `secret_slot='previous'` olayı 24–72 saat görülmeyince previous boşaltılır.
3. Değerler yalnız sunucu env/secret dosyasında; log'a hiçbir zaman yazılmaz. Paddle/LS imzası zaman damgalı ise ±5 dk tolerans + replay penceresi.

## 8. Rezervasyon TTL ↔ period_end
- **Mevcut [OLGU]:** reserve anında `now < ExpiresAt` şartı (`licensesrv/store.go:240`); commit anında da `now < l.ExpiresAt` şartı (`licensesrv/store.go:281`). Model proxy rezervasyonu ~50 dk (2700 s + 5 dk) sürebilir. Sonuç: grant'in bitişine yakın başlayan uzun bir çağrı, grant bittikten sonra **commit edilemez** ve kredi düşmez (kullanıcı lehine sızıntı); model_meter yolunda `settleAmount` davranışı ayrıca test edilmeli.
- **Öneri:** commit, rezervasyonun **açıldığı andaki** geçerliliğe bağlansın (rezervasyon açıksa grant'in bitişi commit'i engellemez); yeni reserve ise `expires_at - now >= reservation_ttl` şartı arar, yetmiyorsa sıradaki grant'e geçer. replace yenilemesinde eski grant'in açık rezervasyonları yeni dönemde de commit edilebilir; `expire` satırı yalnız rezerve olmayan kalanı yakar. Bu bir **davranış değişikliğidir** → önce regresyon testi, sonra ayrı yama, PANEL DEV onayı.

## 9. US-HOUR-1 çakışması
- **Mevcut [OLGU]:** `US-HOUR-1` = `Kind:"use"`, 3600 s, 30 gün (`licensesrv/pricing.go:43`); saniye birimiyle ayrı tutulur (`MeterUnit="seconds"`, `licensesrv/webhook.go:186-190`); hesapta istem kredisi varken **önce istem kredisi** harcanır (`licensesrv/accounts.go:119-121`).
- **Öneri:** iki birim asla birbirine çevrilmez. Varsayılan sıra korunur (istem önce). Kullanıcıya istemcide "saat paketini kullan" tercihi isteğe bağlı başlık (`X-US-Meter: seconds`) ile sunulabilir — KARAR GEREKİR. Abonelik past_due/hard-block durumunda saat paketi yine kullanılabilir (bağımsız satın alma).

## 10. KDV brüt/net
- **Mevcut [OLGU]:** "Fiyatlar KDV dahil TL" (`licensesrv/pricing.go:11`); ledger/grant'te para alanı yok.
- **Öneri:** grant üzerinde `gross_minor, tax_minor, net_minor, currency, tax_collector ('mor'|'merchant'|'none')`, değerler **sağlayıcı payload'ından** okunur, hesaplanmaz. Kredi bakiyesi parasal değildir; para alanları yalnız raporlama ve kısmi iade oranı içindir.
- Shopify (TR) satışlarında KDV'yi kim hesaplar/beyan eder (Shopify vergi ayarı mı, muhasebe mi, e-Fatura entegrasyonu mu): **KARAR GEREKİR (Demir / muhasebe-hukuk).**

## 11. Admin-only kurumsal kredi API'si [ÖNERİ; bugün yok]
```
POST /v1/admin/credit-grants        (yalnız loopback veya mTLS + admin API anahtarı, scope=billing:grant)
Idempotency-Key: <zorunlu>
{ "account_id": "...", "unit": "prompts", "amount": 1000, "expires_at": "...", "reason": "PO-2026-014", "contract_ref": "..." }
→ 201 { "grant_id": "...", "ledger_seq": 123 }
POST /v1/admin/credit-grants/{id}/revoke   { "reason": "..." }
GET  /v1/admin/accounts/{id}/ledger?cursor=
```
Kurallar: üst sınır (ör. tek istekte ≤ 100 000) aşılırsa ikinci admin onayı; her çağrı `admin_audit`; `source='admin'`, `tax_collector='none'`; kamu internetine açılmaz. API anahtarı altyapısı `api-0006-api-keys` yamasına bağlı (başka çalışan; bu belge ona dokunmaz). Gerekli mi: **KARAR GEREKİR (Demir)** — enterprise satış yapılacaksa evet.

## 12. Uygulama aşamaları [ÖNERİ]
1. Regresyon testleri: mevcut reserve/settle/expire/revoke/refund/sıra davranışını kilitle (bir kısmı `api-0001-regression`'da var).
2. JSON store'a ekleme-only `Ledger` gölge yazımı + bakiye mutabakat testi (`sum(ledger)==License.Credits/Reserved`).
3. `Settlement.Reason` (release|expire|revoke) — omitempty, geriye uyumlu.
4. Çoklu secret doğrulama (current/previous).
5. Subscription varlığı + Shopify/manual; ardından Paddle/LS adaptörleri (sandbox).
6. Admin grant API (karar sonrası).
7. PostgreSQL (karar sonrası): çift yazım → mutabakat → okuma geçişi → JSON salt-okunur arşiv.
Her aşama: önce test, `go vet`, `go test -race`, `patches/api-billing-*` altında yama, PANEL DEV incelemesi, deploy yok.

## 13. Açık kararlar (Demir)
past_due soft/hard · replace/stack · carry-over sınırı · proration sahibi · kısmi iade oranı · kullanılmış kredi iadesinde borç · geç ödeme olayında grant · tüketim sırası teyidi · katmana göre eşzamanlı rezervasyon · saat paketi tercihi · Shopify KDV sorumluluğu · e-posta şablonlarının sahibi · admin grant API gerekliliği · PostgreSQL'e geçiş ve zamanlaması · trial kredisi.
