# /v1/image yerel API paketi — 8 Kasım 2026 ~04:50 TSİ

**Durum:** Odaklı, ayrı derlenebilir `imageapi` paketi yazıldı ve test edildi. Bu, licensesrv’ye bağlanmış canlı entegrasyon değildir. Müşteri Windows EXE’si yoktur. Siteye yüklenmedi. GPU/VM/BOOT6 açılmadı.

## Ne eklendi

Yol: `/workspace/us-panel-dev/dist/image-api-20261108/imageapi/`

Sözleşme r2 (`US_V1_IMAGE_SOZLESMESI_20261108.md`, SHA256 `e4ee10ce930bd00b2da7e62de5c57d8844f732314ca7ecf93b7538b21b65e8ce`) doğrultusunda:

- `POST /v1/image`
- `GET /v1/image/jobs/{id}`
- `GET /v1/image/outputs/{ref}`

Paket `licensesrv` içe aktarmaz; `Store`, `Mail`, `Plan` tanımlanmaz.

Üretim varsayılanı: tarife yok, ready işareti yok, üretici (generator) yok. Bu yüzden gerçek GPU çağrısı ve canlı kredi kesimi yoktur.

## Testler (gerçek)

Ortam: kutu, Linux, `go test -count=1 ./...` (paket: `imageapi`).

| Sonuç | Ayrıntı |
|---|---|
| PASS | 12 test, 0 FAIL, 0 SKIP, ~0.004s |
| Log | `imageapi-go-test.log` |

Kapsanan davranış: Idempotency-Key yokluğu; geçersiz model/size/n; unknown_field; unsupported_parameter (`fallback`, `allow_fallback`, `amount`, `ttl`, gövde `idempotency_key`, `steps`); 401; Kind `use` → 403 `plan_kind_not_supported`; tarifesiz istekte 503 `reservation_failed` / `rate_table_pending` (kredi kesilmez); tarife fixture varken ready yoksa 503 `model_not_staged`; IMAGE_READY + test üreticisi ile 200, 24s `authorized_24h`, rezervasyon id sızmaz, yetkili indirme, başka lisans 401; idempotency replay ve 409; cold 202 + boot **niyeti** (GPU açılmaz) ve poll’un niyeti tekrarlamaması; 405; 413; 24 saat sonra erişim reddi ama otomatik silme yok.

Testlerde kullanılan `rate` sayıları yalnız HTTP yollarını işletmek için fixture’dır; ürün tarifesi değildir ve canlıya işlenmedi.

Supervisor 70 CPU testi (ayrı, daha önce yeniden koşuldu): 70 OK, FAIL/SKIP yok. Log: `/workspace/us-panel-dev/dist/image-integration-20261108-r13h-local.work/evidence/cpu-tests-rerun.log` (SHA256 `4b4a155c34d4a2699e8d179f886a19cb2d47216cfd8a0aab21cf2fc548e78cb6`). Bu, panel/API veya gerçek GPU testi değildir.

## Hâlâ BLOCKED / açık

- Müşteri Windows panel EXE’si: yok. `licensesrv` içinde `Store`/`Mail`/`Plan` üretim tanımları mevcut ağaçlarda bulunamadı; tip uydurulmadı.
- Tam paket `go test` / `GOOS=windows` panel derlemesi: BLOCKED (önceki arama sonucu).
- Bu paket licensesrv’ye bağlanmadı (aynı eksik tipler).
- Native Windows / Windows GUI / satış / madde 33: NOT TESTED.
- FINDING-01 (seçenek e) açık.
- f01 386 FAIL açık.
- Sayısal kredi tarifesi PENDING.
- Gerçek FLUX üretimi / GPU: yapılmadı (200 yolundaki PNG, test fixture 1×1 görselidir).

Sürüm dizgesi değiştirilmedi (`2026.11.06.12`). Sözleşme dosyası değiştirilmedi. Donmuş r13b–h ağaçlarına yazılmadı. Site, hook, SMTP, canlı fiyat yok.
