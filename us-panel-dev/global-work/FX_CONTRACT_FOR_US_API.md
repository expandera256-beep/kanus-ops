# FX sözleşmesi — `GET /v1/fx` (ortak belge, GLOBAL; sürüm 08.11.2026 ~01:40 TSİ)

Tek ortak sözleşme budur. API incelemesi (`patches/api-0016-fx-contract/FX_CONTRACT_UNIFIED_TR.md`) işlendi; uyum tablosu §9. Sunucu ucu US'in alanı: referans `patches/global_fx_licensesrv_PROPOSAL.patch` (UYGULANMAYACAK). Panel referans uygulaması: `patches/global_i18n_08b_fx_unsigned.patch` (US panel tarafını bu sözleşmeye göre kendisi yazabilir).

## 0. Kesin kararlar
- **İmza yok.** HMAC, `X-Kanus-FX-Signature`, `sig`, `kid` yok; `US_LICENSE_SECRET` panelde hiçbir şekilde kullanılmaz. (Demir 01:13 TSİ)
- **Pin yok.** Yalnız standart TLS doğrulaması: sistem kök sertifikaları, hostname kontrolü, yalnız https, sabit host, yönlendirme yok, en düşük TLS 1.2; `InsecureSkipVerify` veya özel doğrulama geri çağrısı yok. **Karar: Demir 01:22 TSİ (API aktarımı).** HOST'un SPKI değerleri kullanılmaz.
- **503 yalnız hiç geçerli kur yoksa**; bayat ama geçerli kur **200** ile sunulur. Tazelik kararı panelde.
- Kaynak: TCMB `today.xml` **ForexSelling**. Kapsam: **EUR + USD**.
- Host (HOST bildirdi): canlı `api.kanus.org`, test `api-test.kanus.org` (ikisinin zincirinde Let's Encrypt YE1).

## 1. Uç nokta
- `GET /v1/fx` ve `HEAD /v1/fx`; kimlik doğrulama yok. Rota tam yol `"/v1/fx"` (sonda `/` yok, `/v1/fx/` alt ağacı yok); UYUM/`MountAccountAlias`'tan sonra, `/healthz`'den önce.
- Diğer metotlar → 405 + `Allow: GET, HEAD` + hata zarfı. IP başına sınır aşımı → 429 + `Retry-After` + zarf.

## 2. 200 gövdesi `kanus.fx/1` (≤ 4096 bayt)
```json
{"schema":"kanus.fx/1","source":"TCMB today.xml ForexSelling (döviz satış)",
 "source_url":"https://www.tcmb.gov.tr/kurlar/today.xml","rate_type":"forex_selling",
 "tcmb_tarih":"06.11.2026","bulten_no":"2026/209","date":"2026-11-06","base":"TRY",
 "rates":{"EUR":55.9609,"USD":50.0724},"generated_at":"2026-11-06T16:05:12+03:00"}
```
- Zorunlu: `schema`, `date` (TCMB yayın tarihi, TSİ, YYYY-MM-DD), `base:"TRY"`, `rates.EUR`, `rates.USD`.
- `rates` değerleri **JSON number**, TCMB ondalığı aynen (yeniden yuvarlama yok; ör. `json.Number`). String reddedilir.
- `rate_type` isteğe bağlı; varsa `"forex_selling"` olmalı, başka değer reddedilir. Bilinmeyen alanlar yok sayılır.
- Aralık: her kur sonlu ve [1, 1000]; çapraz EUR/USD [0,8; 1,6].
- Başlıklar: `Content-Type: application/json`, `Cache-Control: public, max-age=3600`.

## 3. Hata
- 503 (yalnız hiç geçerli kur yok: ilk açılış ve hiç başarılı çekim yok ya da kalıcı dosya bozuk): api-0003 zarfı `{"error":"fx_unavailable","code":"fx_unavailable","message":"FX rate unavailable","request_id":"…"}`, `X-Request-Id`, `Cache-Control: no-store`. FX'te `error` alanı bilinçli olarak kodu taşır.
- Panel: 503+`fx_unavailable` ve 200 dışı her durum → son geçerli önbellek (tazelik içinde) ya da yalnız TL.

## 4. Tazelik (panel uygular)
- Geçerli: `today−5 ≤ date ≤ today`, takvim günü **Europe/Istanbul (UTC+3)**. Gelecek tarih ve `today−6` ve öncesi reddedilir. Site (`fx_component.js`, `fx_cron.php`, `fx_snapshot.py`) ve panel (`i18n.FXFresh`) aynı kural.
- Sunucu bayat kuru reddetmez.

## 5. Panel davranışı (08b referansı)
- Host sabit: `fxBase()` yalnız derleme sabitlerinden (`embeddedLicenseURL`/`embeddedGatewayURL`, `-ldflags`) `https://api-test.kanus.org` seçebilir; aksi her durumda `https://api.kanus.org`. `US_GATEWAY_URL`, `US_LICENSE_URL` ve GUI'deki bağlantı adresi alanı **yok sayılır** — kullanıcı FX host'unu değiştiremez.
- https zorunlu; http taban → istek yok, yalnız TL. Yönlendirme izlenmez; son URL host'u istenen host'a eşit olmalı. 5 sn zaman aşımı; gövde > 4096 → ret.
- Önbellek: `fx.json` (desk.json yanında; lang.json ile aynı güvenli yazma: sabit kök fd, symlink/reparse izlenmez, temp+rename). İçerik = sunucu gövdesinin ham baytları (≤ 4096, zarf yok); okunurken şema+aralık+tazelik yeniden doğrulanır. Kayıt hatası loglanır.
- Yenileme: `max-age` 5 dk–6 sa aralığına kırpılır; hata sonrası 5 dk. Ağ çağrısı UI'yi bloklamaz.
- Log yalnız sabit kategori (`fx-unavailable`, `tls-certificate`, `fetch-or-validate`); URL, gövde, sertifika içeriği yazılmaz.
- Gösterim: `≈ € · ≈ $`, `big.Rat` ile kesin bölme, kuruşa yarım-sıfırdan-uzağa yuvarlama, `Intl`/yerel biçim, 4 dil; "yaklaşık, ödeme TL ile" notu ve kur tarihi. **Sabit 50,05'e geri dönüş yok.** Ödeme ve ledger TL; dokunulmaz.

## 6. Sunucu veri üretimi (API önerisi, KABUL)
- Sunucu TCMB today.xml'i **kendisi** çeker: https, 10 sn, ≤ 64 KB, yalnız aynı host'a yönlendirme. Açılışta bir kez; hafta içi ~16:05 TSİ; başarısızlıkta sınırlı yeniden deneme (ör. +15/+30/+60 dk, gün sonuna kadar).
- Parse hatası, eksik EUR/USD, aralık/çapraz dışı, gelecek tarih veya son geçerliden eski tarih → **son geçerli korunur**, yalnız neden loglanır.
- Son geçerli gövde atomik kalıcı (temp+rename, ör. `/var/lib/us/fx.json`), açılışta yüklenir.
- `kanus.org/fx.json`'a bağımlılık yok. İleri adım (KAN.US/GLOBAL kararı): site `fx_snapshot.py`/`fx_cron.php` TCMB yerine `https://api.kanus.org/v1/fx`'ten beslenir.

## 7. Kalan risk
Sunucu ya da TLS (CA/host) ele geçirilirse yanlış kur gösterilebilir. Etki yalnız yaklaşık gösterimdir; panel aralık/çapraz/tazelik kontrolleriyle sapmayı sınırlar; ödeme TL ile alınır. Pin bilinçli olarak yok (sertifika yenilemede kırılganlık; Demir 01:22).

## 8. Açık noktalar
- US sunucu ucunu uygulayacak; API/US yanıtına göre belge güncellenir.
- Üretim derlemesinin `US_BUILD_GATEWAY=https://api.kanus.org` ile yapılması gerekir: `US_DERLE.bat:25` varsayılanı `https://api-test.kanus.org` (aşağı bkz. §10).
- OpenAPI: `patches/api-0016-fx-contract/openapi-fx.yaml` — lint NOT TESTED (Redocly yok).

## 9. API inceleme uyumu
| # | API kararı | Durum |
|---|---|---|
| M1 GET/HEAD, auth yok, ≤4096 | KABUL | §1–2 |
| M2 İmza başlığı kaldır | DEĞİŞİKLİK | Uygulandı (§0) |
| M3 `rate_type` ekle, number kalsın | DEĞİŞİKLİK | Uygulandı (§2; panel `rate_type` ≠ forex_selling'i reddeder) |
| M4 EUR+USD, aralık/çapraz | KABUL | §2 |
| M5 503 yalnız hiç kur yoksa, api-0003 zarfı | DEĞİŞİKLİK | Uygulandı (§3); panel zarfı tanır |
| M6 Tazelik panelde | KABUL | §4 |
| M7 https, yönlendirme yok, host eşitliği | KABUL + ek | Host eşitliği uygulandı; host artık sabit (§5) |
| M8 Önbellekte imza yok | DEĞİŞİKLİK | Uygulandı: ham gövde, yeniden doğrulama |
| M9 TTL/yuvarlama/TL | KABUL | §5 |
| M10 Sunucu TCMB'yi kendisi çeker | DEĞİŞİKLİK | Uygulandı (§6, öneri yaması) |
| M11 Güvenlik notu | DEĞİŞİKLİK | §7 |
| M12 Öneri yaması: HMAC yok, dosya okuma yok, bayatta 503 yok, zarf | DEĞİŞİKLİK | Uygulandı (yeni PROPOSAL) |
| S7 Pin | KARAR | Pin yok (Demir 01:22) |

## 10. Mevcut kodda host kanıtı (pinned/v10)
Bkz. Tur 12 raporu: `US_DERLE.bat:25` (varsayılan test host), `US_DERLE.bat` ldflags `embeddedGatewayURL`/`embeddedLicenseURL`; `gateway.go:24` ve `gui.go:509` (kullanıcı/ortam gateway'i değiştirebilir — bu yüzden FX `activeGateway()` kullanmaz).
