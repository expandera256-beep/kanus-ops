# SSH flaky / timeout investigation (READ-ONLY) — 2026-11-06

Host: srv1662550 / 31.220.63.191. No config changes made in this preparation pass.

## Executive cause

**Primary cause: host CPU + memory pressure (not fail2ban, not conntrack, not MaxStartups).**

Snapshot (~05:11 UTC / 08:11 Europe/Istanbul):
- **Load average: 42.69 / 32.45 / 29.43** (multi-core saturated; `vmstat` showed **st ~65–96%** steal time — classic VPS oversubscription under load)
- **RAM:** 7941 MiB total, ~5804 used, **~152 MiB free**, ~2137 available; **Swap 1.7G / 4G used** with ongoing swap-in/out (`si/so` in vmstat)
- **OOM history:** kernel OOM killed tradebot python workers (Nov 4 tradebot3-worker; Nov 6 tradebot-okx-worker ~02:27 UTC)
- Heavy processes: tradebot3 (~2GB RSS), turkdini node (~1.3GB), tradebot live-24h (~0.75GB), tradebot-okx, growth-os operators, disk `grep -rl` scans

Under this pressure SSH handshake/auth can stall past client `ConnectTimeout` even though sshd eventually accepts keys when the machine catches up.

## Ruled out / low likelihood

| Check | Finding |
|-------|---------|
| fail2ban sshd bans | **Currently banned: 0**; total banned 0; not blocking our box |
| MaxStartups | `10:30:100` (default); no evidence of “exceeded MaxStartups” spam in sampled logs |
| ClientAlive | interval 0 (defaults); not causing connect timeouts |
| PermitRootLogin | `without-password` (keys OK) |
| conntrack | count **161** / max **262144** — not exhausted |
| Docker | **no containers** |
| Disk | root ~51% — OK |
| sshd logs | Many **Accepted publickey** from automation IPs; occasional “Connection closed by …” (scanner/client abort), not mass refusals |

## Contributing factors

1. **CPU steal + run-queue length** — SSH competes with tradebots/turkdini/growth-os.
2. **Memory reclaim / swap** — latency spikes during si/so and OOM killer activity.
3. **Concurrent automation** — many short-lived root SSH sessions from several cloud IPs (agents), increasing contention with MaxSessions 10 (secondary).

## Safe recommendations (do NOT apply without approval; none stop production blindly)

1. **Short term (ops):** raise client `ConnectTimeout` / retry with backoff (already used); prefer multiplexing (`ControlMaster`) for agent SSH.
2. **Memory:** reduce tradebot/turkdini footprint or move one heavy stack (tradebot3 or turkdini) off this VPS; add alert on swap >50% and load >2×nCPU.
3. **sshd soft knobs (reload only):** `ClientAliveInterval 30` / `ClientAliveCountMax 3` for dead session cleanup; optionally `MaxStartups 20:50:100` if many parallel agent connects — low risk.
4. **Avoid** disabling fail2ban or opening UFW further.
5. **Prod api.kanus.org:** do **not** colocate another heavy hook+model stack here without reclaiming RAM (see `../PROD_PLAN.md`).

## Evidence pointers (on host)

- `uptime`, `free -m`, `vmstat 1 3`, `ps aux --sort=-%mem`
- `journalctl -k | grep -i oom`
- `fail2ban-client status sshd`
- `sshd -T | grep maxstartups`
