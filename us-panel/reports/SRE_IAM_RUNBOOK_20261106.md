# SRE IAM Runbook: ds-a100 watchdog için en az yetki (06.11.2026, TSİ)

**Uygulayıcı:** KAN.US, ancak KAN'ın ds-a100 yaşam döngüsü ve local-SSD testleri bittikten sonra. **SRE bu runbook'taki hiçbir komutu çalıştırmadı.** Yazma komutlarının tamamı öneri.

**Kaynak (salt okunur, önceden kaydedilmiş):**
- `sre/iam.json`: proje politikası
- `sre/role_ustaWake.json`
- `sre/instiam_ds.json`: ds-a100 instance politikası, etag `BwZfjVoIPvw=`
- `sre/sa.txt`
- `sre/desc_*.json`
- `gcloud iam roles describe usModelStart/usModelStop` (06.11 08:15 TSİ)

## 0. Mevcut durum (kayıtlardan)
| Öğe | Değer |
|---|---|
| Proje | `project-27082a40-6083-4952-9db` |
| Proje düzeyi bağlama | `roles/ustaWake` (get/start/stop/suspend/resume, stage ALPHA) → `serviceAccount:<USTA_SA_EMAIL>` (kayıtta [REDACTED], büyük olasılıkla compute varsayılan SA'sı, **doğrulanmadı**) |
| ds-a100 instance düzeyi | `usModelStart` (get, start) → us-gateway SA; `usModelStop` (get, stop) → us-watchdog SA |
| ds-a100 / usta-wake / instance-2026… VM SA | `<USTA_SA_EMAIL>` ([REDACTED]), scope `cloud-platform` |
| Mevcut SA'lar | us-watchdog, us-builder, us-scheduler, us-gateway, `<USTA_SA_EMAIL>` |

Değişkenler (yer tutucuları KAN.US dolduracak):
```bash
P=project-27082a40-6083-4952-9db; Z=europe-west4-b; VM=ds-a100
WD_SA=ds-watchdog@$P.iam.gserviceaccount.com
USTA_SA='<USTA_SA_EMAIL>'          # [REDACTED] — iam.json'dan gerçek değer okunmalı
OLD_RUN_SA=us-watchdog@$P.iam.gserviceaccount.com
```

## Uygulama sırası (h)
**Önce ekle ve doğrula, en son daralt:**
A (SA) → B (rol) → C (bağlama) → D (çalışma kimliği) → canlı doğrulama → E (ustaWake daraltma) → F (VM SA/scope, KAN'ın penceresinde).

Her adım ayrı uygulanır. Doğrulama geçmeden bir sonraki adıma geçilmez.

Genel ön kontrol, her adımdan önce yedek alınır:
```bash
gcloud projects get-iam-policy $P --format=json > iam_before_$(date +%H%M).json
gcloud compute instances get-iam-policy $VM --zone $Z --project $P --format=json > instiam_before_$(date +%H%M).json
```

## A. Watchdog servis hesabını oluşturma (a)
- **Ön kontrol:**
  ```bash
  gcloud iam service-accounts describe $WD_SA --project $P
  ```
  NOT_FOUND bekleniyor.
- **Uygula:**
  ```bash
  gcloud iam service-accounts create ds-watchdog --project $P --display-name="ds-a100 SRE watchdog" --description="get/start/stop only on ds-a100"
  ```
- **Doğrula:**
  ```bash
  gcloud iam service-accounts describe $WD_SA --project $P --format="value(email,disabled)"
  ```
  `False` dönmeli. Ayrıca **anahtar oluşturulmamalı**:
  ```bash
  gcloud iam service-accounts keys list --iam-account $WD_SA --managed-by=user
  ```
  Liste boş olmalı.
- **Geri dönüş:**
  ```bash
  gcloud iam service-accounts delete $WD_SA --project $P
  ```

## B. Özel rol (b)
- **İzinler:** `compute.instances.get`, `compute.instances.start`, `compute.instances.stop`, `compute.instances.setMetadata`.
- **`compute.zoneOperations.get` gerekli mi?** **Hayır.**
  - Watchdog, start/stop'tan dönen operation'ı izlemiyor. Sonucu `instances.get` ile `status` okuyarak doğruluyor (watchdog.py `_wait_until`).
  - Ayrıca zoneOperations izni instance düzeyinde bağlanamaz, proje veya zone düzeyi gerekir ve bu da kapsamı genişletir.
  - Operation hata ayrıntısı istenirse **ayrı** salt okunur bir rol kullanılmalı: `compute.zoneOperations.get` + `compute.zoneOperations.list`, proje düzeyinde. Bu zone'daki tüm operasyonları okumaya izin verir ama değişiklik yetkisi vermez. Varsayılan öneri: **ekleme**.
- **⚠ `setMetadata` riski:**
  - Bu izinle startup-script yazılabilir. VM'in SA'sıyla kod çalıştırma yoluna eşdeğer.
  - Watchdog yalnızca `ds-controller` bayrağını **okuyor**, bunun için `get` yeterli. Bayrağı KAN elle ya da kendi kimliğiyle ayarlıyor.
  - **SRE önerisi:** rolü önce setMetadata olmadan oluşturmak (B-1). setMetadata yalnızca Demir ve KAN watchdog'un bayrağı kendisinin yazmasına karar verirse eklenmeli (B-2).
- **Ön kontrol:**
  ```bash
  gcloud iam roles describe dsWatchdog --project $P
  ```
  NOT_FOUND bekleniyor.
- **Uygula (B-1, önerilen):**
  ```bash
  gcloud iam roles create dsWatchdog --project $P --title="ds-a100 watchdog" --stage=GA \
    --permissions=compute.instances.get,compute.instances.start,compute.instances.stop
  ```
- **B-2 (onay olursa):**
  - **KARAR (06.11.2026 08:19 TSİ, Demir, SRE sohbetinde doğrudan):** setMetadata eklensin. B-1 uygulandıktan hemen sonra B-2 de uygulanacak. Geri dönüş: `gcloud iam roles update dsWatchdog --project $P --remove-permissions=compute.instances.setMetadata`. Bağlama yalnızca ds-a100 instance düzeyinde kalmalı.
  ```bash
  gcloud iam roles update dsWatchdog --project $P --add-permissions=compute.instances.setMetadata
  ```
- **Doğrula:**
  ```bash
  gcloud iam roles describe dsWatchdog --project $P --format="value(includedPermissions)"
  ```
- **Geri dönüş:**
  ```bash
  gcloud iam roles delete dsWatchdog --project $P
  ```
  Silinen rol 7 gün içinde `undelete` ile geri alınabilir.

## C. Rolü yalnızca ds-a100'e, instance düzeyinde bağlama (c)
- **Ön kontrol:** instance politikası yedeklenir (yukarıda). Etag'in değişip değişmediğine bakılır; kayıtlı değer `BwZfjVoIPvw=`.
- **Uygula:**
  ```bash
  gcloud compute instances add-iam-policy-binding $VM --zone $Z --project $P \
    --member="serviceAccount:$WD_SA" --role="projects/$P/roles/dsWatchdog"
  ```
- **Doğrula:**
  ```bash
  gcloud compute instances get-iam-policy $VM --zone $Z --project $P --format=json
  ```
  usModelStart, usModelStop ve dsWatchdog bağlamalarının üçü de görünmeli.
  ```bash
  gcloud projects get-iam-policy $P --flatten=bindings --filter="bindings.members:$WD_SA" --format="value(bindings.role)"
  ```
  **Boş** dönmeli (proje düzeyinde bağlama yok).
- **İşlevsel doğrulama (salt okuma, impersonation ile):**
  ```bash
  gcloud compute instances describe $VM --zone $Z --project $P --impersonate-service-account=$WD_SA --format="value(status)"
  ```
  Başarılı olmalı.
  ```bash
  gcloud compute instances describe usta-wake --zone us-central1-c --project $P --impersonate-service-account=$WD_SA
  ```
  **403** dönmeli.
  Impersonation için uygulayıcıda `roles/iam.serviceAccountTokenCreator` gerekir; bu yetki geçici verilir ve test sonrası kaldırılır.
- **Geri dönüş:**
  ```bash
  gcloud compute instances remove-iam-policy-binding $VM --zone $Z --project $P \
    --member="serviceAccount:$WD_SA" --role="projects/$P/roles/dsWatchdog"
  ```

## D. Watchdog'un çalıştığı yerde kimliği değiştirme (d)
Hedef çalışma yeri henüz kesinleşmedi; Cloud Run `us-watchdog` ya da VPS. Aşağıdaki komutlar Cloud Run için.
- **Ön kontrol:**
  ```bash
  gcloud run services describe us-watchdog --region europe-west4 --project $P --format="value(spec.template.spec.serviceAccountName,status.latestReadyRevisionName)"
  ```
  Beklenen değer `us-watchdog@…`. Revision adı not edilir.
- **Gerekli ek bağlama:** deploy eden kimliğe yalnızca bu SA üzerinde `roles/iam.serviceAccountUser`:
  ```bash
  gcloud iam service-accounts add-iam-policy-binding $WD_SA --project $P --member="user:<DEPLOYER>" --role=roles/iam.serviceAccountUser
  ```
- **Cloud SQL:** us-watchdog lease store için Cloud SQL kullanıyorsa yeni SA'ya da `roles/cloudsql.client` gerekir. Şu an `us-watchdog` SA'da var (iam.json). Gerekirse proje düzeyinde eklenir. Bu, compute yetkisi vermez.
- **Uygula:**
  ```bash
  gcloud run services update us-watchdog --region europe-west4 --project $P --service-account=$WD_SA
  ```
- **Doğrula:**
  - Yeni revision Ready olmalı.
  - Watchdog **observe modunda** `/health` çağrısı `status` döndürmeli ve `degraded=false` olmalı.
  - Cloud Run loglarında 403 olmamalı.
  - Scheduler'ın OIDC kimliği değişmez: us-scheduler, invoker bağlaması aynı kalır.
- **Geri dönüş:**
  ```bash
  gcloud run services update-traffic us-watchdog --region europe-west4 --project $P --to-revisions=<ÖNCEKİ_REVISION>=100
  gcloud run services update us-watchdog --region europe-west4 --project $P --service-account=$OLD_RUN_SA
  ```
- **Not:** watchdog aktif moda geçip yeni SA ile stop yetkisi aldıktan sonra, `us-watchdog` SA'nın ds-a100 üzerindeki `usModelStop` bağlaması E adımında kaldırılabilir. Bu, eski reconcile'ın da stop yapmasını engeller ve çift denetleyiciyi önler.

## E. Eski yetkileri daraltma (e). Ancak A–D doğrulandıktan ve idle.py devri (tasarım §6) tamamlandıktan sonra
- **Ön kontrol:** son 24 saatte usta-wake'in hâlâ start/stop yapıp yapmadığına bakılır:
  ```bash
  gcloud logging read 'protoPayload.methodName=~"instances.(start|stop)" AND protoPayload.authenticationInfo.principalEmail="'$USTA_SA'"' --project $P --freshness=1d --format="value(timestamp,protoPayload.resourceName)"
  ```
  Kayıt varsa **dur**: eski sistem hâlâ aktif.
- **Seçenek 1 (önerilen ara adım, daraltma):** usta-wake geçiş süresince yalnızca ds-a100'ü yönetsin.
  ```bash
  gcloud compute instances add-iam-policy-binding $VM --zone $Z --project $P --member="serviceAccount:$USTA_SA" --role="projects/$P/roles/ustaWake"
  # doğrula, sonra:
  gcloud projects remove-iam-policy-binding $P --member="serviceAccount:$USTA_SA" --role="projects/$P/roles/ustaWake"
  ```
- **Seçenek 2 (kaldırma, usta-wake devreden çıktıktan sonra, §4.13):**
  ```bash
  gcloud projects remove-iam-policy-binding $P --member="serviceAccount:$USTA_SA" --role="projects/$P/roles/ustaWake"
  gcloud compute instances remove-iam-policy-binding $VM --zone $Z --project $P --member="serviceAccount:$USTA_SA" --role="projects/$P/roles/ustaWake"   # Seçenek 1 uygulandıysa
  ```
  Rol artık kullanılmıyorsa:
  ```bash
  gcloud iam roles delete ustaWake --project $P
  ```
- **Doğrula:**
  ```bash
  gcloud projects get-iam-policy $P --flatten=bindings --filter="bindings.role:roles/ustaWake" --format="value(bindings.members)"
  ```
  Boş dönmeli.
- **Geri dönüş:**
  ```bash
  gcloud projects add-iam-policy-binding $P --member="serviceAccount:$USTA_SA" --role="projects/$P/roles/ustaWake"
  ```
  Rol silindiyse önce:
  ```bash
  gcloud iam roles undelete ustaWake --project $P
  ```
- **⚠ idle.py etkisi:** ds-a100'ün içindeki idle.py, VM'in kendi SA'sıyla (`<USTA_SA_EMAIL>`) stop çağırıyor (KANUS_GPU_REPORT.md:32). Proje bağlaması idle.py devrinden **önce** kaldırılırsa idle.py stop yapamaz ve GPU açık kalır. Bu yüzden E, yalnızca tasarım §6'daki devirden sonra uygulanır.

## F. ds-a100 ve usta-wake VM SA ve scope değişikliği (f). **VM durdurulmalı; yalnızca KAN'ın penceresinde, KAN tarafından**
- **Neden:** üç VM aynı SA'yı `cloud-platform` scope ile kullanıyor; herhangi biri ele geçirilirse proje genelinde yetki elde edilir.
- **Hedef:**
  - **ds-a100:** ayrı bir SA, örneğin `ds-a100-vm@`. Yetkileri: model bucket'ında `storage.objectViewer` ve `logging.logWriter`. Scope: `storage-ro,logging-write,monitoring-write`. idle.py devredildiyse compute yetkisi verilmez.
  - **usta-wake:** devreden çıkana kadar mevcut SA kalır. Ancak scope `compute` ile sınırlandırılır.
- **Ön kontrol:**
  ```bash
  gcloud compute instances describe $VM --zone $Z --project $P --format="json(status,serviceAccounts,scheduling)"
  ```
  Mevcut SA ve scope'lar yedeklenir.
- **Uygula (ds-a100, KAN penceresinde):**
  - Önce durdurma. ⚠ Spot ve 4×375 GB local SSD: verinin korunup korunmayacağı KAN'ın ölçümüne bağlı. `--discard-local-ssd` değeri KAN tarafından seçilir.
    ```bash
    gcloud compute instances stop $VM --zone $Z --project $P --discard-local-ssd=<true|false>
    ```
  - Ardından:
    ```bash
    gcloud compute instances set-service-account $VM --zone $Z --project $P --service-account=ds-a100-vm@$P.iam.gserviceaccount.com --scopes=storage-ro,logging-write,monitoring-write
    ```
  - Başlatma, watchdog ya da KAN tarafından yapılır.
- **usta-wake:**
  ```bash
  gcloud compute instances stop usta-wake --zone us-central1-c --project $P
  gcloud compute instances set-service-account usta-wake --zone us-central1-c --project $P --service-account=$USTA_SA --scopes=compute
  gcloud compute instances start usta-wake --zone us-central1-c --project $P
  ```
  Bu sırada uyandırma kapısı kapalı kalır (kesinti).
- **Doğrula:**
  - `describe` ile `serviceAccounts` kontrol edilir.
  - ds-a100 içinde `ds-boot` modeli bucket'tan okuyabilmeli.
  - `ds-serve` hazır olmalı: `/v1/models` 200 dönmeli. Bu kontrol KAN'ın testi, SRE istek atmaz.
- **Geri dönüş:** VM durdurulur, `set-service-account` ile yedeklenen SA ve `--scopes=cloud-platform` geri yazılır, VM başlatılır.

## Sonuç doğrulama listesi
- [ ] `$WD_SA` yalnızca ds-a100 üzerinde dsWatchdog rolüne sahip; proje düzeyinde bağlama yok; anahtar yok.
- [ ] Watchdog observe modunda 403 almadan `status` okuyor.
- [ ] idle.py devri tamamlandı (`ds-controller=sre-watchdog`).
- [ ] Proje düzeyinde `ustaWake` bağlaması yok.
- [ ] ds-a100 VM SA'sında compute yetkisi yok, scope dar.
- [ ] Her adımın `*_before_*.json` yedeği saklanıyor.
