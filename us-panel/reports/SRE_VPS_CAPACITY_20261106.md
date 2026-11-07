# SRE VPS kapasite analizi — 06.11.2026

Ölçüm tamamlandı: **09:41 TSİ** (Europe/Istanbul). Kaynak: KAN.US salt-okuma dosyası `sre/vps/srv1662550_capacity_20261106.txt` (UTC 2026-11-06T06:41:02+00:00). Box'tan VPS'e SSH yok. Hiçbir seçenek uygulanmadı; süreç/servis/config değiştirilmedi.

Özet sonuç: **FAIL** (test VPS aşırı yük + yüksek CPU steal). Prod VPS headroom **PASS**. Hook alarm eşikleri ölçüye uygun; config değişmedi.

## 1. Kanıt kaynakları

| Dosya | Rol |
|---|---|
| `sre/vps/srv1662550_capacity_20261106.txt` | Test VPS (31.220.63.191 / srv1662550) tam OS örneklemesi — **birincil** |
| `sre/vps/srv2133566_capacity.txt` | Prod VPS (82.29.184.225 / srv2133566) headroom — 09:18 TSİ |
| `sre/vps/srv1662550_capacity_20261106_0942.txt` | SSH timeout metni; ölçüm değil (yok sayıldı) |
| `sre/vps/inventory.txt`, `vps_catalog.txt` | Hostinger envanter / fiyat kataloğu |
| `reports/HOOK_VPS_STATUS_20261106.md`, `LOGIN_PERF_ONERI_TR.md` | Önceki bağlam; yeni ölçümle çelişmiyor |

Box'tan SSH denemesi yapılmadı (fail2ban ban riski, Demir kararı 06.11.2026).

## 2. Ölçüm matrisi (test VPS srv1662550, 2 vCPU / 8 GiB)

PASS = değer salt-okuma dosyasından okundu; sağlık garantisi değildir.

| Ölçüm | Durum | Bulgu / kanıt |
|---|---|---|
| Tahsis CPU/RAM (Hostinger + nproc) | PASS | nproc=2; inventory KVM 2, memory=8192. free -m Mem total 7941 MiB |
| load1/5/15 | **FAIL** | `32.39, 31.00, 32.10` → load1/nCPU ≈ **16.2** (kritik eşik 4) |
| CPU steal (vmstat `st`) | **FAIL** | Örnek satırlar: st=**65, 95, 95, 95, 96** %. Misafir neredeyse CPU alamıyor |
| CPU iowait (`wa`) | PASS (düşük) | wa=3,0,0,0,0 %. Darboğaz IO değil |
| CPU user/system | PASS (düşük pay) | us≈4–5 %, sy≈1 %; idle 0–5 % — kalan süre steal |
| CPU PSI | **FAIL** | some avg10=**96.75** avg60=**97.44** avg300=**95.57**; full=0 |
| Memory PSI | PASS (uyarı altı full) | some avg10=1.71 avg60=0.95 avg300=9.38; full avg60=**0.00** avg300=0.20 |
| IO PSI | PASS | some avg60=3.90; full avg10/60/300≈0 |
| Bellek available/cache | PASS (sıkışık) | used 6493 / free 456 / buff+cache 1287 / **available 1448** MiB |
| Swap doluluk | **FAIL** (yüksek kullanım) | Swap 4095 MiB, **used 1927** MiB (`/swapfile`) |
| Swap in/out (anlık) | PASS (örnekte sakin) | vmstat si/so sonraki 4 satırda 0; ilk satır boot-ort. (si 959 so 1264) yok sayılır |
| Top CPU süreçler | PASS | tradebot python %45.9 RSS 773 MiB; growthos python %22/%21; tradebot python %18.9/%18.1; turkdini node %16.2; **us-license-hook %9.5** |
| Top RSS | PASS | tradebot python **~2005 MiB (24.6 %)**; turkdini node **~1867 MiB (22.9 %)**; tradebot ~773 MiB |
| cgroup / systemd-cgtop | **NOT TESTED** | Capture'da cgroup çıktısı yok |
| us-license-hook CPUWeight/Nice/MemoryMax | **NOT TESTED** | `systemctl show` satırı capture'da yok |
| Login p50/p95 | **NOT TESTED** | Journal/access login süreleri bu dosyada yok. Önceki rapor ~20–28 sn ile aşırı steal tutarlı; yeni percentile yok |
| bcrypt cost (yerel kaynak) | PASS | accounts.go cost 12 (önceki yerel bulgu; bu turda yeniden doğrulanmadı) |
| Prod VPS load/RAM/steal | PASS | srv2133566: load 0.14/0.04/0.01; st=0; idle≈100 %; Mem available 7454 MiB; Swap 0 |

## 3. Yükü kim üretiyor?

**Doğrulandı (09:41 TSİ örneklemesi):**

1. **Hypervisor steal (~95 %)** — birincil duvar. 2 vCPU'nun fiili kapasitesi neredeyse yok.
2. **tradebot** kullanıcı süreçleri (birden fazla python + gunicorn) — en yüksek runnable CPU ve RSS.
3. **growthos** python süreçleri — kısa ömürlü ama yüksek %CPU.
4. **turkdini** node — yüksek RSS (~1.9 GiB) + anlamlı CPU.
5. **us-license-hook** — yalnızca ~%9.5 CPU; yükün ana kaynağı değil, kurbanı.

Runnable kuyruk (vmstat `r` 16–43) steal ile birleşince login bcrypt'inin 20–28 sn sürmesi beklenen sonuçtur; bcrypt kendi başına ~0.2 sn (önceki box ölçümü).

## 4. Seçenekler — hiçbiri uygulanmadı

Fiyat: Hostinger `vps_catalog.txt` (cent → USD). Mevcut abonelik upgrade teklifi / prorata **NOT VERIFIED**.

| Seçenek | Ölçüme göre yeniden değerlendirme | Öneri |
|---|---|---|
| **A: Hook'u prod 82.29.184.225'e taşı** | Prod idle (load≈0, st=0, available≈7.5 GiB). İzolasyon hook login SLO'sunu steal'den kurtarır. Kapasite artışı değil; bot yükü test VPS'te kalır. | **Öncelikli öneri.** DNS/TLS/HMAC/store/GPU tüneli kesim planı + Demir onayı şart. Uygulanmadı. |
| **B: CPUWeight/Nice + bot CPUQuota** | Hook zaten düşük CPU; steal %95 iken weight misafir içi paylaşımı düzeltir, hypervisor steal'i çözmez. Bot quota trade emir gecikmesi riski. CPUWeight/Nice canlı değeri **NOT TESTED**. | Steal varken **ikincil / geçici**. Önce A. B ancak A gecikir ve bot sahibi ticari riski kabul ederse; unit adları doğrulanmadan komut yok. |
| **C: KVM4/8 yükseltme** | Daha fazla vCPU runnable kuyruğu emebilir; steal host oversubscribe ise plan büyütmek host'u değiştirmeyebilir. KVM4 liste 42.99 USD/ay (ilk 25.99); KVM8 73.99 (ilk 50.99). KVM2 liste 24.99/ay. | **A sonrası** veya botlar aynı hostta kalmak zorundaysa Hostinger'a steal sorulup teklif alınarak. Ölçüm olmadan / steal kanıtı yokken yükseltme önerilmezdi; artık steal kanıtlı ama önce ucuz izolasyon (A). Uygulanmadı. |

**Net öneri (uygulama yok):** 1) Hook'u prod'a taşı (A). 2) Test VPS'te tradebot/growthos/turkdini sahipleriyle CPU/RSS bütçesi veya ayrı host. 3) Steal sürerse Hostinger destek + gerekirse KVM4 teklifi (C). 4) B yalnız geçici band-aid. Yükseltmeyi tek başına ilk adım yapma.

### B komut taslağı — ÇALIŞTIRILMADI

Önceki rapordaki taslak geçerli; canlı CPUWeight **NOT TESTED** olduğu için hâlâ uygulanmaz. `systemctl show` capture'a eklenmeden set-property yok.

## 5. Hook alarm eşik gözden geçirme

Dosya: `sre/hook_watchdog/config.example.json`.

| Eşik | Config | Ölçülen | Karar |
|---|---|---|---|
| load_warning/critical per CPU | 2 / 4 | ~16.2 | Doğru kritik; **değişmedi** |
| cpu_psi warning/critical % | 20 / 50 | avg60 97.44 | Doğru kritik; **değişmedi** |
| memory_psi full warning % | 1 | full avg60 0.00 | Uygun; **değişmedi** |
| io_psi full warning % | 10 | ~0 | Uygun; **değişmedi** |
| login p95 warn/crit sn | 2 / 5 | Bu capture'da yok | Korundu; collector bağlanınca doğrulanır |
| pressure_hold_seconds | 300 | CPU PSI avg300 95.57 — uzun süreli | 300 sn tutmak false-positive azaltır; **değişmedi** |

**Sonuç:** Config değişmedi → pytest yeniden koşturulmadı. Önceki kanıt: `TEST_OUTPUT_20261106.txt` 59 passed. Steal için ayrı metrik yok; load+CPU PSI zaten ateşler. İleride vmstat `st` collector isteğe bağlı (kod değişikliği; bu turda yok).

## 6. Billing / watchdog (önceki tur, hâlâ geçerli)

BigQuery dry-run ve watchdog pytest önceki bölümde PASS; bu turda dokunulmadı.

## 7. Bu turda yapılan / yapılmayan

- Yapıldı: salt-okuma dosyasını okuma; raporu ölçümle tamamlama; eşik gözden geçirme; routine silme talebi.
- Yapılmadı: SSH, yükseltme, taşıma, CPUWeight/Quota, config değişikliği, login sentezi, kaynak silme.

Rapor durumu: ölçüm matrisi tamamlandı (eksik kalanlar yalnızca capture kapsamı dışı: CPUWeight/Nice, cgroup, login p50/p95 → **NOT TESTED**).
