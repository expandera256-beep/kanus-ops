# FX sözleşmesi — lisans sunucusu `GET /v1/fx` (GLOBAL önerisi, US/API kararı)

Durum: ÖNERİ. Sunucu tarafı UYGULANMADI (API/US alanı). Panel tarafı `global_i18n_08_fx.patch` ile hazır; uç nokta yoksa panel yalnız TL gösterir.
Demir kararı (00:48 TSİ, KAN.US iletti): panel kuru sabit 50,05 yerine lisans sunucusundan alacak.

## Mevcut durum (okundu)
- pinned/v10 `licensesrv`: fiyat/kur uç noktası YOK. Rotalar: `/v1/credits/{status,reserve,settle}`, `/v1/chat(/completions)`, `/account/*`, Shopify webhook. Kur yalnız sabit `USDTRYRef = 50.05` (licensesrv/pricing.go:13) — PriceUSD/us-pricing-gen için.
- `US-kaynak.zip` (kanus.org/downloads) ve API r11 (`patches/api-r11`, api-0001..0012): FX/kur uç noktası yok.
- İmza mekanizması: lisans anahtarları HMAC-SHA256 + `US_LICENSE_SECRET` (license.go mint/parse; licensesrv `Config.LicenseSecret`). Asimetrik anahtar yok. → Yeni sır üretilmedi; aynı sır ayrı alan ayırıcıyla kullanılır.

## Uç nokta
`GET /v1/fx` (HEAD da). Kimlik doğrulama yok (herkese açık veri). Gövde ≤ 4096 bayt.

200 yanıtı:
```
Content-Type: application/json
Cache-Control: public, max-age=3600
X-Kanus-FX-Signature: v1=<base64url(HMAC-SHA256(US_LICENSE_SECRET, "kanus.fx/1\n" || body))>

{"schema":"kanus.fx/1","source":"TCMB today.xml ForexSelling (döviz satış)","source_url":"https://www.tcmb.gov.tr/kurlar/today.xml","tcmb_tarih":"06.11.2026","bulten_no":"2026/209","date":"2026-11-06","base":"TRY","rates":{"USD":50.0724,"EUR":55.9609},"generated_at":"2026-11-07T21:39:17+00:00"}
```
- `rates`: tam olarak `EUR` ve `USD`; 1 birim döviz = x TRY (TCMB ForexSelling); her biri sonlu, [1, 1000]; EUR/USD ∈ [0,8; 1,6].
- İmza, gövdenin TAM baytları üzerinedir (yeniden serileştirme yok).
- Veri yok/geçersiz/eski → `503 {"error":"fx_unavailable"}`, `Cache-Control: no-store`. Diğer metotlar → 405.

## Tazelik kuralı (site, cron, panel, sunucu aynı)
TCMB tarihi Türkiye takvim günüdür (Europe/Istanbul = sabit UTC+3). today = şimdi (UTC+3) günü. Geçerli: today−5 … today (dahil). Gelecek tarih ve today−6 ve öncesi RED. (PANEL DEV R4B FX01 düzeltmesi.)

## Panel davranışı (08)
- Kaynak `activeGateway()` + `/v1/fx`; yalnız `https://`; yönlendirme izlenmez; 5 sn zaman aşımı; ≤4096 bayt.
- İmza `licenseSecret()` ile doğrulanır; sır yoksa/imza tutmazsa/şema-tazelik bozuksa → yalnız TL. Sabit 50,05'e GERİ DÖNÜLMEZ.
- Önbellek: `fx.json` (desk.json yanında, lang.json ile aynı güvenli yazma: pinned kök, symlink/reparse izlenmez, ≤4 KB, temp+rename). İçerik `{"body":base64url, "sig":...}`; okunurken imza ve tazelik yeniden doğrulanır.
- Sayfa çizimi ağı beklemez; yenileme arka planda, `max-age` (5 dk–6 sa arası) veya hata sonrası 5 dk.
- Yuvarlama: kuruş tam sayı, kesin rasyonel bölme, sente yarımda sıfırdan uzağa; 10^15 kuruş üstü gizlenir.
- Ödeme ve ledger TL; bu veri hiçbir ücret/kredi hesabına girmez.

## Sunucuda kur verisinin üretimi (öneri)
Günlük iş (hafta içi 16:05 TSİ) `fx_snapshot.py /var/lib/us/fx.json` (GLOBAL'in site aracı; doğrulama aynı) ya da eşdeğer Go işi. Handler ağa çıkmaz; yalnız dosyayı doğrular ve imzalar.

## Güvenlik notu
HMAC simetrik: sırrı taşıyan panel ikilisini çözen biri imzalı kur üretebilir. Etki yalnız yaklaşık gösterim (ödeme TL). Daha güçlü seçenek: Ed25519 (sunucuda özel, panelde açık anahtar) — yeni anahtar gerektirdiği için US/API kararına bırakıldı.

## Örnek handler
`patches/global_fx_licensesrv_PROPOSAL.patch` (licensesrv/fx.go + test; UYGULANMAYACAK, yalnız referans). Yönlendirme kaydı (`mux.Handle("/v1/fx", licensesrv.FXHandler{...})`) US/API'nin sunucu ana dosyasına aittir, eklenmedi.
