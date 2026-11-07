# Test VPS SSH timeout diagnosis — 2026-11-06 (Europe/Istanbul)

**Scope:** READ-ONLY. No firewall/sshd/fail2ban/service changes, no unban, no restarts.  
**Target:** srv1662550 / `31.220.63.191` (test). Compare: srv2133566 / `82.29.184.225` (prod).  
**Collected:** ~09:31–09:35 Europe/Istanbul (server TZ `Etc/UTC`).

## Summary (root cause)

**Primary (yüksek güven):** Test VPS aşırı yük altında (load ~31–33, CPU PSI ~92–95%, RAM baskısı, ~1.9 GiB swap). SSH handshake gecikiyor veya TCP `:22` ara sıra 10 s timeout veriyor; HTTP `:80`/`:443` aynı anda milisaniye cinsinden açık. Bu, “fail2ban ban” veya Hostinger cloud firewall ile uyumlu değil.

**Değil (kanıtla elendi):**
- fail2ban sshd: **Currently banned = 0**, Total banned = 0 (HOST’un 09:19 gözlemiyle uyumlu).
- Hostinger VPS firewall: hesapta **firewall yok**; VM `firewall_group_id = null`.
- ufw: `22/tcp ALLOW Anywhere` (aktif).
- sshd journal (son 1–6 sa): **MaxStartups / drop connection / throttling yok**.

**İkincil / latent risk (şu an timeout’un nedeni değil):** Box egress `140.248.50.x` içinde **dönüyor** (aynı dakikada birden fazla IP). fail2ban `ignoreip` yalnızca `140.248.50.124`. Ban oluşursa çoğu box IP korunmaz — ama bugün ban listesi boş.

**Belirsizlik:** Prod `:22` TCP de ara sıra 10 s timeout verdi (box/path etkisi). Ancak test SSH ~24–89 s sürerken prod SSH ~1.8 s tamamlandı; asıl fark test host overload.

## Evidence

### 1) Box egress (api.ipify.org ×5, ~09:31 Istanbul)

| # | IPv4 |
|---|---|
| 1 | 140.248.50.39 |
| 2 | 140.248.50.13 |
| 3 | 140.248.50.103 |
| 4 | 140.248.50.24 |
| 5 | 140.248.50.96 |

IPv6 (ifconfig.me / icanhazip): `2a04:4e41:80:…` (egress de değişiyor).  
HOST ~09:19: `140.248.50.13`. fail2ban ignoreip (test): **yalnız** `140.248.50.124`.

### 2) TCP reachability (bash `/dev/tcp`, timeout 10 s)

**Test 31.220.63.191**

| Port | try1 | try2 | try3 |
|---|---|---|---|
| 22 | OK 63 ms | **TIMEOUT 10004 ms** | OK 94 ms |
| 80 | OK ~4 ms | OK ~4 ms | — |
| 443 | OK ~3 ms | OK ~3 ms | — |

**Prod 82.29.184.225**

| Port | try1 | try2 |
|---|---|---|
| 22 | OK 143 ms | **TIMEOUT 10003 ms** |
| 80/443 | OK ~3–6 ms | OK ~3 ms |

### 3) SSH attempts (az sayıda)

| Hedef | ConnectTimeout | Sonuç | Süre |
|---|---|---|---|
| test | 45 s | **OK** (`SSH_OK` srv1662550) | ~24216 ms |
| test (diag session) | 60 s | **OK** | ~89324 ms |
| prod | 20 s | **OK** (`PROD_SSH_OK`) | ~1857 ms |

### 4) On-host (test) — salt okuma

- **uptime:** load average **31.46, 32.94, 33.00** (19 gün up).
- **PSI:** CPU some avg10=**91.88**; memory some avg10=**31.83**, full avg10=**8.63**; IO some avg10≈12.
- **memory:** 7941 MiB total, used ~6673, available ~1268; **swap 4095 / used ~1937 MiB**.
- **Top CPU/RSS:** tradebot live-24h (~46% CPU / 818 MiB), tradebot3 (~18% / **~2.0 GiB**), turkdini node (~16% / **~1.8 GiB**), tradebot-okx, growth-os operator/publisher, unattended-upgrade, Hostinger usage-telemetry, multiple gunicorn :8090–8095, us-license-hook.
- **OOM:** dmesg **02:26 UTC** OOM kill tradebot-okx python; journal ~06:03 gunicorn worker SIGKILL “Perhaps out of memory?”. Son 1 saatte yeni kernel OOM yok; bellek baskısı sürüyor.
- **sshd -T:** `maxstartups 10:30:100`, `maxsessions 10`, `logingracetime 120`, `usedns no`. Journal’da MaxStartups throttling **yok**.
- **ss :** SYN-RECV :22 ≈ 0–1 satır; established :22 az (diag anında 1–2). Listen :22 sshd aktif.
- **fail2ban:** jails `nginx-http-auth`, `nginx-limit-req`, `sshd` — hepsi **0 banned**. sshd total failed=15, currently failed=1.
- **ignoreip:** `127.0.0.1/8 ::1 140.248.50.124` (`/etc/fail2ban/jail.local`).
- **ufw:** active; OpenSSH/22 + 80/443 + 8090–8095 ALLOW.
- **iptables/nft:** UFW policy DROP INPUT; 22 user-allow zincirinde (ban kuralı görünmedi; fail2ban banned=0 ile tutarlı).

### 5) Hostinger API (MCP timeout → curl Bearer token; token raporlanmadı)

- MCP `user-Hostinger-vps` execute: **timeout** (−32001).
- API `GET /api/vps/v1/virtual-machines`: srv1662550 **running**, ipv4 31.220.63.191, plan KVM 2, Ubuntu 24.04; srv2133566 running 82.29.184.225.
- `GET /api/vps/v1/firewall`: **data=[]** (hiç firewall yok).
- VM detail: `firewall_group_id = null`.
- Metrics (panel, ~son 2 sa): cpu_usage noktaları **4.99%** (şüpheli tavan/örnekleme; host load 32 ile çelişiyor — panel metriğine tek başına güvenme), ram ~5.7–6.4 GiB, uptime ~471 h.

## Fix proposal (uygulanmadı — yalnızca öneri)

1. **Asıl çözüm:** Test VPS yükünü düşür — tradebot/turkdini/growth-os fleeti ayır veya prod/ayrı VPS’e taşı; RAM/swap baskısını azalt. SSH flakiness’in kökü bu.
2. **Operasyonel SSH:** `ConnectTimeout 45–60`, birkaç retry; “timeout = ban” varsayma.
3. **fail2ban (latent):** `ignoreip` içine `140.248.50.0/24` (veya bilinen box egress CIDR) ekle — **yalnızca onay sonrası**; bugünkü timeout’u çözmez ama yanlış ban riskini azaltır.
4. **sshd (opsiyonel, onaylı):** aşırı yükde `LoginGraceTime` / `MaxStartups` gözden geçir; şu an MaxStartups kanıtı yok — öncelik değil.
5. **Hostinger cloud firewall:** şu an yok; eklerken 22’yi box CIDR ile kısıtla, aksi halde kendi SSH’ini kesersin.
6. **İzleme:** panel CPU %5 vs load 32 sapmasını not et; host `uptime`/PSI esas alınsın.

## Değişiklik yapılmadı

Firewall, fail2ban, sshd, unban, restart, service flip — hiçbiri.
