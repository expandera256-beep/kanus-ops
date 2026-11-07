# Cloud Run yerel düzeltmeleri — 00007 tabanı

**Deploy/GCP yazması/GPU/VM/VPS/usta-wake çağrısı yapılmadı.** Bu dizin source-control repository değildir; talep edilen yerel, geri alınabilir image snapshot yamalarıdır. KAN'ın release dizini değiştirilmedi.

## Taban ve dosyalar
- base-00007/: KAN release-20261108 kopyası; tarball SHA256 80d16639d3c3440518e4db519a98feb0a7f44f31cb6bf8aaea7ace03aaba5584.
- app-00007/: iki servis için birleşik test edilen yerel sonuç.
- us-watchdog/local-acceptance.patch: yalnız watchdog factory/ready + yardımcı modüller; gateway factory/commerce route'u değişmez.
- us-commerce/local-acceptance.patch: yalnız gateway factory/commerce ready + yardımcı modüller; watchdog factory değişmez.
- Bu iki patch birbirinin üzerine uygulanmaz. Her biri temiz base-00007 kopyasına `patch --batch -p1 -i ABSOLUTE_PATCH_PATH` ile uygulanır.
- readiness.patch → rate-limit.patch → structured-logging.patch sıralı birleşik alternatif; combined.patch tek birleşik alternatif. Hizmetlerin ayrı build/deploy planı için DEPLOY_PLAN.md kullanın.
- build_patches.py çıktıların temiz ve tekrar üretilebilir olmasını sağlar; patch_apply_check.json uygulanan dosyaların beklenen ağaçla eşleşmesini kanıtlar.
- tests/baseline/: 114 KAN testi (10 ready-window testi dahil), tests/new/: yeni testler. Gerçek PostgreSQL/DDL testlerini açan US_POSTGRES_TEST_DSN unset tutulmalıdır.

## Davranış
- /ready health-method gate öncesinde config + yalnız SELECT 1; GPU/provider, lease/heartbeat/reconcile veya mail yok. /health aynı; /healthz platform rezervi nedeniyle önerilmez. /readyz alias eklenmedi.
- Instance başına, thread-safe, sabit 3 token bucket: health=2 req/sn burst10; read=5/sn burst30; API=10/sn burst60. 429 ve Retry-After. Değerler yerel öneridir; işletim etkisi deploy onayında değerlendirilmeli. Dağıtık kullanıcı kotası değil; cold-start/scale-out sıfırlar. Genel API grubu webhook/login/model yollarını da kapsar.
- Yalnız IAM-private watchdog POST /internal/reconcile limiter dışında; caller header'ı bu ayrıcalığı sağlamaz. OIDC güvenlik sınırı Cloud Run IAM'dir. Commerce aynı path için bypass sağlamaz.
- logging_app(rate_limit_app(app)): 429 dahil JSON stdout; severity/message/httpRequest/server-generated request_id, policies/env'den project ve doğrulanmış trace context ile logging.googleapis.com/trace. Caller request-id, Authorization, header dump, body, query string ve hata metni loglanmaz. Bilinmeyen URL path'i [unmatched].
- Yakalanan app/iterator istisnası generic 500 veya başlamış response'da sanitize edilmiş stream sonu + ERROR kaydı üretir; exception/DSN/token metni Gunicorn'a tekrar fırlatılmaz. Streaming başarısızlığında mevcut close/cleanup çağrısı korunur. Tam stack trace yerine güvenli generic hata kullanılır.
- KAN'ın READY'den başlayan lease/stream penceresi, managed_inference.py, inference_lifecycle.py ve gce_provider.py değişmeden korunur. Yeni AST/byte-eşitlik testleri ve 10 davranış testi bunu kontrol eder.

## Test komutu
```sh
PYTHONPATH=/workspace/us-panel/sre/cloudrun/fixes/app-00007 env -u US_POSTGRES_TEST_DSN python3 -m pytest -q -rs /workspace/us-panel/sre/cloudrun/fixes/tests
```
37 skip: 36 opt-in PostgreSQL contract + 1 PostgreSQL/Gunicorn runtime. Bunlar izinli canlı SELECT testinin yerine geçmez; salt-okuma dışı DDL içerdiğinden açılmadı.

Canlı eksikler yerel test başarısıyla kapanmış sayılmaz. DEPLOY_PLAN.md bir yürütme izni değildir.

## Doğrulanmış son sonuçlar
Birleşik 00007: 142 PASS /37 SKIP. Ayrı watchdog ve commerce patch ağaçlarının her biri 138 PASS /37 SKIP /4 diğer-servis yeni testi deselected. KAN114 + yeni28 toplamı; ready-window/lease10 test dahil. Gerçek SQL read-only kabulü raporda ayrıca PASS; migration sürüm tablosu yok (FAIL).
