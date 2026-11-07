# FINDING-06 — Secret/IAM lifecycle ve en-az-yetki açıkları

Tarih: 08.11.2026 01:15 TSİ (Europe/Istanbul). **Owner: SRE. Genel önem: MEDIUM; alt maddeler aşağıda. Durum: OPEN / FAIL (§12 Secret Manager).** Credential compromise/izinsiz erişim gösterilmedi.

## Kanıt / provenance
`reports/r20261108/SRE_SECRETS_IAM_METADATA.md:4–46,61–87` ve `raw_secrets_iam/` SRE tarafından teslim edilen redacted, salt-okuma metadata kanıtıdır; QA bağımsız GCP ölçümü yapmadı. QA `review_sre_metadata.py` ile JSON tutarlılığını yerelde doğruladı; bulguların doğrulanması GCPye bağlanmayı gerektirmedi. Kaynak latest bindingleri `work2/kanus/deploy/gateway.service.yaml:30–47`, `watchdog.service.yaml:29–40` ile de doğrulandı. Repo templateı canlı bindinglerin bütünüyle aynı diye sunulmaz.

| Alt madde | Önem / durum | Beklenen → gerçek | Kanıt |
|---|---|---|---|
|Credential rotation yönetişimi|MEDIUM / OPEN|Tanımlı sahip/takvim/rotation kanıtı → 6 secretta rotation/nextRotationTime metadata yok; signing/model-token/postgres-dsn tek ENABLED v1. Manual rotation hiç olmadı diye kanıtlanmış değildir.|SRE:5–17,68,81; secrets_list.json + ver_*.json|
|Sürüm pinleme/eski policy sürümleri|LOW / OPEN|Kontrollü version rollout → Cloud Run env/volume latest; worker-policy v5 yanında v1/v3/v4 ENABLED, v2 DISABLED. Policy credential değil; bu durum credential sızıntısı sayılmadı.|SRE:15–17,35–36,73; run_services.json, ver_us-worker-policy.json|
|VM kimliği/scope sertleştirme|MEDIUM / OPEN|İş yüküne özel SA ve dar scope → ds-a100/usta-wake ortak default compute SA, cloud-platform scope. Effective IAM dar custom ustaWake (get/start/stop/suspend/resume); scope tek başına secretAccessor/Owner yetkisi vermez.|SRE:30,38–40,74; instances.json, role_ustaWake.json|
|Kullanılmayan enabled SA|LOW / OPEN|Envanterde gerekçeli aktif kimlik → sre-oidc-negtest enabled, proje rolü yok; kullanım yokluğu SRE envanter iddiası. us-builderın kullanılmadığı **NOT VERIFIED**, onu unused diye sınıflamadık.|SRE:45–46,75; sas.json, project_iam.json|
|Box eski token artığı|MEDIUM / OPEN, token validity NOT VERIFIED|Gerekçeli retention/decommission → **/workspace/secrets/usta-wake/a100.token.old** mevcut, regular file,0600,uid/gid1000. Yalnız lstat; içerik/hash/token değeri okunmadı. Geçerli olduğu veya sızdığı gösterilmedi.|SRE:61–62,86; QA old-token-stat.json|

Tek sürümlü secretların version create zamanları **05.11.2026 23:36–23:39 TSİ**; policy v5 **08.11.2026 00:21:40 TSİ**. CreateTime rotation yapıldığı ispatı değildir. Public .pub0644 veya dizin0755 tek başına private-token disclosure olarak sınıflanmadı.

## Repro — yalnız teslim edilmiş metadata + stat, sıfır network
```bash
python3 /workspace/us-panel/qa/audit12/review_sre_metadata.py \
 --evidence /workspace/us-panel/reports/r20261108/raw_secrets_iam \
 --output /workspace/us-panel/reports/r20261108/qa-raw/sre-recheck-$(date +%Y%m%dT%H%M%S%N).json
```
Gerçek koşum: `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/sre-metadata-verification.json` + .log: metadata assertions PASS, **security policy assessment FAIL**. Test PASS, configuration güvenliği PASS demek değildir.

## NOT VERIFIED / owner ayrımı
- **HOST:** canlı `/etc/us-license-hook.env`, `/etc/us-license-hook-test.env`, `/etc/us-model/wake.token`/wake.pin mode/UID/GID ve servis erişimi. Kaynaklar `deploy/us-license-hook/us-license-hook.service:7`, `deploy/hook-move/us-license-hook-test.service:8`, main.go:30,57,104,132,156; docs600/640 ve wake owner çelişkisi var. Canlı dosyalar açılmadı; çelişki ölçülmüş kötü izin değildir. HOST metadata-only stat + unit User/Group + okunabilirlik teyidi sağlamalı.
- **SRE:** us-watchdog DB-secret ihtiyacının kaldırılabilirliği ve us-builder kullanım durumu NOT VERIFIED. IAM mevcut bağları yalnız gateway/watchdog secret-level accessor olarak gösteriyor; public/project-wide accessor yokluğu olumlu kanıt olarak korundu.
- SRE aktarımı5P/5F/2NV denmişti; raporun açık tablosu: **6PASS/4FAIL/2NOT VERIFIED**. Eski token ayrı risk, tablo satırı değil; alt kontrolleri14-madde QA sayısıyla karıştırmadık.

## Önerilen fix / §13
SRE credential rotation/runbook/alarmlarını tanımlasın; sürüm pinleme/controlled latest kararını ve policy eski version retentionını gerekçelendirsin; özel VM SA/scope tasarlasın; gereksiz SA ve eski token için owner decommission kararı alsın. Secret disable/delete/rotate, IAM/SA değişikliği veya dosya silme **bu görevde yapılmadı**; gerekli açık onay olmadan uygulanmaz. HOST dosya izinlerini yalnız gerçek unit erişim gereksinimiyle doğrulasın. Fix sonrası aynı metadata testi ve owner runtime attestation tekrar edilmeli; secret değerleri istenmemeli.

### 08.11.2026 01:19 TSİ (Europe/Istanbul) — Negatif safety regression eklendi
Metadata consistency PASS ile policy kabulünü ayırmak için `review_sre_metadata.py --assert-safe` eklendi. Aynı metadata üzerinde5 bulgu (rotation/pinning/VM identity/unused test SA/old token) nedeniyle **FAIL/exit1**, xfail yok. Repro yukarıdaki komuta `--assert-safe` eklenmesidir; yeni output path kullanılır, var olan kanıt üzerine yazılmaz. `reports/r20261108/qa-raw/teammates-20261108T011233+0300-qnuld4zp/sre-policy-negative-final.json` + .log. Script hiçbir secret içeriği veya live-cloud API okumaz. Owner fix sonrası güncel metadata ve izinli decommission kanıtıyla aynı safety testinde PASS beklenir; runtime rotation başarısı ayrıca owner attestation gerektirir.
