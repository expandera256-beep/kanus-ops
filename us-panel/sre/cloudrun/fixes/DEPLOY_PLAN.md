# Ayrı servis build / deploy planı — YALNIZ TASLAK, ÇALIŞTIRILMADI

Hazırlama: 08.11.2026 00:41:13 TSİ. Bu belge onay değildir. Build, Artifact Registry push, Cloud Run update/traffic değişiklikleri **bu görevde çalıştırılmadı**. GPU/VM/VPS/usta-wake çağrısı ve Scheduler job run/create yok.

## Mevcut durum ve ayrı yayın hedefleri

| Servis | Şu an revizyon / trafik | Mevcut image | Yerel patch / hedef taban | Geri dönüş |
|---|---|---|---|---|
| us-watchdog | us-watchdog-00007-fnc, %100 | gateway:2026-11-08-readywindow; `europe-west4-docker.pkg.dev/project-27082a40-6083-4952-9db/us-runtime/gateway@sha256:f8a85dbcdee57c0d21639feed0b722beac1ab4b9d94781a06a9d1fb85011e1a6` | `us-watchdog/local-acceptance.patch`, temiz base-00007 | us-watchdog-00007-fnc |
| us-commerce | us-commerce-00002-pr2, %100 | `europe-west4-docker.pkg.dev/project-27082a40-6083-4952-9db/us-runtime/gateway@sha256:3e001dbec0a7d033ac105366bf282a59e4babff84e60413a33dbc2a7c474098a` | `us-commerce/local-acceptance.patch`, **release-20261108 / readywindow tabanına geçiş** | us-commerce-00002-pr2 |

Mevcut durum salt-okuma describe kanıtı `reports/r20261108/raw_cloudrun/deploy_plan_inventory.json`. İşleme geçmeden yeniden okuyun; arada başka revizyon varsa rollback hedefini ve image tabanını güncelleyin.

**Önemli:** KAN'ın READY sonrası lease penceresi düzeltmesi `inference_lifecycle.py`, `managed_inference.py` ve `commerce_http.py` içindedir. us-commerce şu an eski image'dadır; bu düzeltme ona ancak bu servis readywindow image tabanına geçtiğinde ulaşır. Buradaki commerce patch eski image üzerine körlemesine uygulanmamalı; temiz `base-00007` üzerine uygulanır. KAN us-commerce'i deploy etmeyecek; o servis için ayrıca açık yayın onayı gerekir.

## Watchdog zorunlu ön koşulları / STOP kapısı

1. KAN'ın lease/READY penceresi revizyonu Ready olmalı ve kaynak/image/hash/test doğrulaması geçmeli. Yerel rebase 00007 tabanı üzerinde yapıldı; KAN'ın 114 testi ile yeni testler birlikte koşuldu. Ready=True tek başına yeterli değildir.
2. **STOPPING kuyruğu ve VM içi idle.py 600 sn/discard davranışı KAN tarafından doğrulanmış olmalı.** Bu executor VM/VPS/usta-wake'e erişmedi. KAN_FIXES_20261108.md idle.py.fixed için "ds-a100'e uygulanmadı" diyor; bu ön koşul doğrulanmadan watchdog için yeni build/deploy'a geçmeyin.
3. Yeni KAN revizyonu çıkarsa önce o image kaynağını/hash'ini doğrulayın, ayrı yerel kopyaya rebase edin; lease/managed-inference dosyalarının davranışını ve suite'i tekrar test edin. KAN release dizinini değiştirmeyin.
4. Cloud Run revision, Scheduler job/SA/audience, Secret Manager latest sürümü, env US_ALLOW_DISCARD_LOCAL_SSD=1 ve IAM yalnız salt okuma ile tekrar kontrol edilmeli. Politika/source pin tutarlılığı kaydedilmeli.
5. Yerel limit önerileri (health2/sn burst10, read5/sn burst30, genelAPI10/sn burst60) işletim açısından onaylanmalı. Instance-local olması/cold-start ve scale-out reseti kabul edilmeli. Webhook/login/model yolları genel API bucket'ına dahildir. Watchdog IAM-private Scheduler reconcile muafiyeti korunmalı.

## Build komutları — ÇALIŞTIRMAYIN; ayrı onaydan sonra

`base-00007/` orijinal Dockerfile, requirements.txt ve deploy/cloudbuild.yaml içerir. Her servis **ayrı temiz context** kullanır; per-service patch'ler üst üste uygulanmaz. Context kopyalama/patch/test yereldir; aşağıdaki Cloud Build komutları ücret/yazma doğurabileceği için mevcut görevde yasaktır.

### us-watchdog
```sh
PROJECT=project-27082a40-6083-4952-9db
REGION=europe-west4
FIXES=/workspace/us-panel/sre/cloudrun/fixes
CONTEXT="$FIXES/build-context-us-watchdog"   # yeni, temiz bir kopya; mevcut context üstüne yazmayın
test ! -e "$CONTEXT" || exit 1
cp -a "$FIXES/base-00007" "$CONTEXT"
(cd "$CONTEXT" && patch --batch -p1 -i "$FIXES/us-watchdog/local-acceptance.patch")
# Yerel testler PASS ve bütün STOP kapıları kapandıktan, ayrıca build onayı geldikten sonra:
IMAGE="europe-west4-docker.pkg.dev/$PROJECT/us-runtime/gateway:sre-watchdog-approved-tag"
gcloud builds submit "$CONTEXT" --project="$PROJECT" \
  --config="$CONTEXT/deploy/cloudbuild.yaml" --substitutions="_IMAGE=$IMAGE" \
  --service-account="projects/$PROJECT/serviceAccounts/us-builder@$PROJECT.iam.gserviceaccount.com"
IMAGE_DIGEST="$(gcloud artifacts docker images describe "$IMAGE" --project="$PROJECT" --format='value(image_summary.fully_qualified_digest)')"
```

### us-commerce
```sh
PROJECT=project-27082a40-6083-4952-9db
REGION=europe-west4
FIXES=/workspace/us-panel/sre/cloudrun/fixes
CONTEXT="$FIXES/build-context-us-commerce"   # hazırda yoksa oluşturulacak temiz kopya
test ! -e "$CONTEXT" || exit 1
cp -a "$FIXES/base-00007" "$CONTEXT"
(cd "$CONTEXT" && patch --batch -p1 -i "$FIXES/us-commerce/local-acceptance.patch")
# KAN lease testleri + commerce regresyonu PASS, ayrıca ayrı build onayı şart:
IMAGE="europe-west4-docker.pkg.dev/$PROJECT/us-runtime/gateway:sre-commerce-approved-tag"
gcloud builds submit "$CONTEXT" --project="$PROJECT" \
  --config="$CONTEXT/deploy/cloudbuild.yaml" --substitutions="_IMAGE=$IMAGE" \
  --service-account="projects/$PROJECT/serviceAccounts/us-builder@$PROJECT.iam.gserviceaccount.com"
IMAGE_DIGEST="$(gcloud artifacts docker images describe "$IMAGE" --project="$PROJECT" --format='value(image_summary.fully_qualified_digest)')"
```

İmaj içindeki modülleri hash ile doğrulayın; test dosyaları dışında yalnız amaçlanan değişiklikleri taşıyın. Yeni SA/rol/secret veya authorizedNetwork gerekmez; eksik yetkide otomatik genişletme yapmayın, durun.

## Deploy komutları — ÇALIŞTIRILMADI; her servis için ayrı açık onay

Mevcut env, IAM, serviceAccount, secret mounts, VPC ve Cloud SQL anotasyonlarını koruyan image update kullanılır. Aşağıdaki değişkenler **ilgili servisin** yukarıdaki build'ine aittir; watchdog ve commerce IMAGE_DIGEST karıştırılmamalı.

### us-watchdog: önce trafiğe çıkmayan aday
```sh
gcloud run services update us-watchdog --project="$PROJECT" --region="$REGION" \
  --image="$IMAGE_DIGEST" --revision-suffix=sre-ready-r1 --no-traffic --tag=sre-candidate-watchdog
# Candidate revizyonunu describe ile doğrula, sağlık/IAM/JSON log kabulünü yap.
NEW_WATCHDOG_REVISION='<describe-ile-alinan-aday-revizyon>'
# Yalnız bütün kontroller ve trafik onayı sonrası:
gcloud run services update-traffic us-watchdog --project="$PROJECT" --region="$REGION" \
  --to-revisions="$NEW_WATCHDOG_REVISION=100"
```
Scheduler hedefi/audience/SA değişmez; job'u elle çalıştırmayın. Doğal dakikalık çağrıyı bekleyin.

### us-commerce: bağımsız aday ve bağımsız trafik onayı
```sh
gcloud run services update us-commerce --project="$PROJECT" --region="$REGION" \
  --image="$IMAGE_DIGEST" --revision-suffix=sre-ready-r1 --no-traffic --tag=sre-candidate-commerce
NEW_COMMERCE_REVISION='<describe-ile-alinan-aday-revizyon>'
gcloud run services update-traffic us-commerce --project="$PROJECT" --region="$REGION" \
  --to-revisions="$NEW_COMMERCE_REVISION=100"
```
`--allow-unauthenticated`, IAM/env/secret değişikliği veya yeni Scheduler job eklemeyin.

## Canlı doğrulama — yalnız GPU açmayan yollar

Her servis ayrı kayıtla:
1. Service/revision describe: exact immutable digest, %100 doğru trafik, SA, env/secret binding ve Cloud SQL bağlantısı. /ready testi before/after yapılmalı; Ready=True yeterli değil.
2. **us-commerce:** GET /health →200; GET /ready →200 ve checks.store/config true. /healthz GFE rezervi nedeniyle kullanılmaz. Müşteri/model/ödeme/POST endpoint'i çağırmayın.
3. **us-watchdog:** kimliksiz GET /health →403. Mevcut izinli kullanıcı (Owner) token'ı yalnız bellekte kullanılarak GET /health ve /ready →200. SA impersonation veya /internal/reconcile manuel istek yok. Rolsüz SA testi zaten ayrı onaylı dış kanıtta PASS; tekrar token üretmek bu planın parçası değil.
4. **429:** yalnız /health üzerinde, en fazla **60 GET / yaklaşık2.5 saniye**, 429 görüldüğünde dur; Retry-After pozitif tamsayı ve yanıt JSON olmalı. Watchdog için aynı izinli kullanıcı token'ı; token dosyaya/log'a yazılmaz. Birkaç instance'a dağıtım yüzünden eşik doğrulanamazsa PASS yazmayın; daha büyük burst üretmeyin. Bu adım için canlı test kapsamı/onayı ayrıca korunmalı.
5. **JSON log:** ilgili candidate/live revision Cloud Logging stdout kayıtlarında severity, message, httpRequest ve logging.googleapis.com/trace, generated request_id bulunmalı. Authorization/DSN/token/body/query değerleri yok. 429 severity WARNING; readiness503/uygulama500 ERROR. Hata üretmek için gerçek secret/DSN göndermeyin; yerel redaction testleri bunu kapsar.
6. Watchdog için doğal Scheduler POST reconcile 200 ve OIDC audience/SA değişmemiş olmalı; 5xx/LifecycleError/refusing yok. **GPU start/stop senaryosu çalıştırmayın.** Model istemi, /v1/*, /a100/*, wake/start/inference çağrısı yok.
7. Commerce readywindow geçişini canlı GPU isteğiyle doğrulamayın; yerelde 10 lease/READY penceresi davranış testi ve AST/byte-preservation testleri kanıtı kullanılır. Gerçek model uçtan uca kabulü ayrı izinli iş olarak açık kalır.

Örnek salt-okuma log sorgusu (NEW_* değerini ilgili servisin gerçek revision adıyla değiştirin):
```sh
gcloud logging read 'resource.type="cloud_run_revision" AND resource.labels.revision_name="<aday-veya-yeni-revizyon>"' \
  --project=project-27082a40-6083-4952-9db --freshness=15m --order=desc --limit=200 --format=json
```

## Geri dönüş — rollback de ayrı değişiklik/onaydır; burada çalıştırılmadı

| Servis | Komut |
|---|---|
| us-watchdog | `gcloud run services update-traffic us-watchdog --project=project-27082a40-6083-4952-9db --region=europe-west4 --to-revisions=us-watchdog-00007-fnc=100` |
| us-commerce | `gcloud run services update-traffic us-commerce --project=project-27082a40-6083-4952-9db --region=europe-west4 --to-revisions=us-commerce-00002-pr2=100` |

Rollback secret latest/env içeriğini otomatik geri almaz. İlgili eski revizyonun varlığını ve kapasitesini önce describe ile kontrol edin; artifact/image silmeyin. Commerce rollback lease/READY düzeltmesini yeniden kaldırır; etkisi bilinerek karar verilmeli.
