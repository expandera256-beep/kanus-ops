# usta-wake Güvenli Devreden Çıkarma Planı (TASLAK) — 06.11.2026
Yalnız öneri. Silme/kapatma yok. Ön koşul: yeni istem-başı sistem (Go proxy → us-gateway start / us-watchdog stop) canlı kabul testlerini geçmiş olmalı. Her adım Demir onayı + KAN koordinasyonu ister.

Mevcut: usta-wake (e2-micro, us-central1-c, statik IP usta-ip 34.41.186.170, tag kalfa-http, SA cloud-platform), proje düzeyi custom rol `ustaWake` (get/start/stop/suspend/resume), ds-a100 içinde idle.py (15 dk) + boot.sh, istemciler `X-Usta-Worker: a100` ile usta-wake'e gidiyor.

## Faz 0 — Envanter (salt-okuma, şimdi yapılabilir)
1. usta-wake'e kimlerin gittiğini bul: VPS/Go proxy US_MODEL_UPSTREAM, EXE, snippets (`snippets/gcloud_ssh_tunnel.*`), usta-wake nginx/access log (KAN'dan).
2. ds-a100 içi idle.py ve usta-wake'in kendi idle/70 dk stop mantığı belgelensin (KAN).
3. Kabul kriteri tanımla: yeni yolla ≥3 soğuk açılış, eşzamanlı istekte tek start, aktif lease varken stop yok, 15 dk idle stop, health'in GPU açmaması — audit log principal'ı us-gateway/us-watchdog olmalı.
## Faz 1 — Paralel çalışma (gölge)
4. us-watchdog `adopt_running=true` ile ds-a100'ü yönetilir hale getir; usta-wake açık kalır.
5. İki stop kaynağı yarışmasın: önce VM içi idle.py süresi > watchdog süresi (örn. 30 dk) yapılır, watchdog 15 dk ile birincil olur. 48 saat audit log izlenir.
## Faz 2 — Trafik geçişi
6. Tüm istemcilerin upstream'i yeni yola çevrilir; usta-wake'e gelen /v1 isteği sıfırlanana kadar (≥72 saat) log izlenir.
7. Geri dönüş: upstream'i usta-wake'e geri al (IP ve VM henüz duruyor).
## Faz 3 — Pasifleştirme (geri alınabilir)
8. usta-wake'te wake endpoint'i devre dışı (servis stop), VM durdurulur (STOP, silme değil). Statik IP rezervde kalır.
9. Proje düzeyi `ustaWake` rol bağlaması kaldırılır; ds-a100'e özel SA (bucket okuma + log yazma) atanır, scope daraltılır (VM stop gerektirir → KAN penceresi).
10. VM içi idle.py kapatılır ya da yalnız "watchdog ulaşılamazsa emniyet" (örn. 60 dk) olarak bırakılır.
## Faz 4 — Kaldırma (≥14 gün sorunsuz sonra, Demir yazılı onayı)
11. usta-wake disk snapshot → VM silme; usta-ip statik IP serbest bırakma; `kalfa-http` firewall kuralı (yalnız başka kullanan yoksa); custom rol `ustaWake` silme; dokümanlardan (GPU_SOGUK_ACILIS_TR.md) usta-wake referansları kaldırma.
## Ek Faz — us-commerce (Cloud Run) allUsers erişiminin yeniden değerlendirilmesi
- Durum (06.11.2026): KAN.US kararıyla allUsers erişimi şimdilik **korunuyor**.
- Kanıt: SRE_AUDIT_20261106.md "us-commerce trafik" bölümü. Son 7 günde 45 istek: 10 probe, Shopify user-agent'lı 0, 35 probe dışı (curl/route-audit 31, Python-urllib register/login 2, Chrome favicon 2). Kaynak IP sahipleri doğrulanmadı.
- Ön koşul: accounts modu VPS'teki Go hook'a taşınmış ve EXE .7 kabul testi PASS olmuş olmalı.
- Adımlar (ön koşuldan sonra):
  1. Son 7 günün Cloud Run request loglarını yeniden incele; probe dışı gerçek trafik olmadığını doğrula.
  2. allUsers run.invoker bağlamasını kaldırma önerisini KAN.US'a sun. Uygulayan KAN.US; IAM rolünü kaldırmak için Demir onayı gerekli.
  3. Geri dönüş: bağlamayı yeniden ekle (`gcloud run services add-iam-policy-binding us-commerce --member=allUsers --role=roles/run.invoker`).
- Durum: NOT TESTED.

## Riskler
- Spot kesintisi sırasında geçiş; iki denetleyicinin aynı anda stop/start'ı (audit'te 2'şerli start/stop çiftleri zaten görülüyor: audit.txt); usta-ip'ye sabitlenmiş eski EXE sürümleri.
