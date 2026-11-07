# KAN §4 — Canlı GPU test planı (ÖNERİ, çalıştırılmadı)

Bütçe: Demir'in 6 Kasım 8 testlik onayından kalan **4 açılış**, ~0,40 $/açılış, soğuk ~7 dk. Her açılış sonunda VM'in diskleri silerek TERMINATED olduğu `gcloud compute instances describe` ile doğrulanır. Tümü usta-wake :8443 veya mevcut tünel yoluyla; hook/firewall/DNS'e dokunulmaz.
Not: watchdog stop düzeltmesi deploy edilmeden watchdog'un VM'i durdurması canlıda test edilemez (W19/W20). Bu plan düzeltme olmadan da çalışır; düzeltme gelirse Boot 3 ona ayrılır.

## Boot 1 — idle 600 uygulama + 10 dk idle + otomatik kapanış (~25 dk VM, ~0,45 $)
1. Soğuk açılış (/a100/start), hazır olunca: `sudo cp -p /opt/ds/idle.py /opt/ds/idle.py.bak-20261108-idle600`; box `idle.py.600` kopyala; `python3 -m py_compile`; `cmp`/sha256 (d036c8c3…ae78).
2. Tek chat isteği (ilk yanıt ölçümü), sonra hiç istek yok.
3. Beklenen: son istekten 10:00–11:00 dk sonra idle.py stop (discardLocalSsd=true, Python-urllib, Audit Log) → TERMINATED. PASS ölçütü: 10–11 dk; 9:59 öncesi FAIL.
Kapsar: 10 dk etkinliksizlik, otomatik kapanış, API doğrulama, soğuk başlangıç/ilk yanıt tekrar.

## Boot 2 — eşzamanlı istek + iptal + tekrar açılış (~22 dk, ~0,40 $)
1. TERMINATED'dan 5 eşzamanlı istek (istemci tarafında paralel `/a100/start` + chat) → Audit Log'da **tek** start.
2. Hazır olunca 1 uzun istek başlat, ortada iptal et (istemci kesme) → nginx/Strata'da isteğin kesildiğini ve idle sayacının iptalden itibaren saydığını doğrula.
3. idle stop sonrası (Boot 1'den sonra bu "tekrar açılış"ın ikinci kanıtıdır).

## Boot 3 — beklenmeyen kesinti / Spot preemption + kurtarma (~12 dk, ~0,25 $)
1. Hazır olunca `gcloud compute instances simulate-maintenance-event ds-a100 --zone europe-west4-b` (Spot'ta preemption davranışı beklenir; yerel SSD silinir). Alternatif: `stop --discard-local-ssd=true` ile ani kapatma.
2. Doğrula: TERMINATED, watchdog `phase=stopped` (olay tablosu/log), usta-wake `/a100/status` doğru durum; sonra yeni istek yeniden açmayı tetikler mi — tetiklerse bu açılış Boot 4 sayılır, bu yüzden adım 2'de yalnız status okunur.
- Watchdog düzeltmesi deploy edildiyse bu boot yerine: idle.py timer'ı durdurup watchdog'un 600 sn sonra discard ile durdurduğunu doğrula (watchdog kurtarma/stop).

## Boot 4 — yedek
Başarısız/şüpheli adımın tekrarı veya watchdog düzeltmesi sonrası watchdog stop testi. Kullanılmazsa iade.

Toplam tahmini: ~1,1–1,5 $ (4 boot). Önceki harcama ~0,78 $.
Ön koşullar: Demir'in her boot için GO'su (§14), bitişte ds-a100 TERMINATED + disk savedState kontrolü, sonuçlar `r20261108/` altına.
Kapsanamayanlar: gerçek (Google kaynaklı) Spot preemption zamanlaması; 70 dk health-fail yolu (maliyet nedeniyle önerilmez, SİM I10 yeterli sayılır).
