# Test VPS CPU steal — 2026-11-06 (Europe/Istanbul)

**READ-ONLY.** No buys, tickets, migrations, or host changes performed.  
**Artifacts:** `/workspace/us-panel/sre/vps/srv1662550_capacity_20261106.txt`  
**MCP:** `user-Hostinger-vps` execute timed out → Hostinger public API + one SSH (ConnectTimeout 60; first try timed out, second succeeded ~39s).

## 1) Findings

### Plan / capacity

| | srv1662550 (test) | srv2133566 (prod) |
|---|---|---|
| Plan | **KVM 2** | **KVM 2** |
| vCPU | 2 | 2 |
| RAM | 8192 MiB | 8192 MiB |
| Disk | 102400 MiB | 102400 MiB |
| OS | Ubuntu 24.04 | Ubuntu 26.04 |
| State | running | running (idle) |

Source: `GET /api/vps/v1/virtual-machines/{id}`.

### Metrics (API, ~24h)

- **Test `cpu_usage`:** flat **~4.98–5.00%** (48 points). This tracks **guest run time**, not steal.
- **Test RAM:** ~4.5–7.6 GiB used (tight on 8 GiB).
- **Test disk used:** ~48–49 GiB / 100 GiB.
- **Prod:** ~0.5% CPU, ~0.47 GiB RAM — essentially idle.

### Steal (guest, 09:45 Europe/Istanbul / 06:45 UTC)

| Source | Steal | Guest busy (usr+sys) | Idle |
|---|---|---|---|
| vmstat 1s samples | **95%** | ~5% | 0% |
| mpstat avg | **95.32%** | ~4.5% | 0% |
| sar today 03:40–06:40 UTC | **~95% every 10 min** | ~3% user | ~0% |
| /proc/stat 5s | **95.0%** | ~5% | 0% |

Also: load ~30–32 on 2 vCPU; `vmstat r` = 20–33 runnable; CPU PSI some ~95%.

## 2) Plan throttle vs noisy neighbour?

**Strong evidence of hypervisor CPU contention (stolen time), not “apps using 95% CPU”.**

- Guest only executes **~3–5%** of wall time (`%usr`/`%sys`), while **steal ≈ 95%** for hours (`sar`).
- Runnable queue stays high → workloads *want* CPU but the host does not schedule the vCPU.
- Panel API `cpu_usage≈5%` **agrees** with guest busy fraction and **hides** the steal problem.

**Not fully separable from guest alone:**

- Classic **noisy neighbour** and **intentional host oversubscription / soft cap** both present as high `%steal` with low guest `%usr`.
- No host-side metrics or neighbour visibility → **cannot prove** a single noisy VM vs systematic oversell vs a Hostinger scheduler policy.
- Label: **host-side CPU contention / oversubscription (steal-dominated)**. “Plan CPU limit at 5%” is a possible mechanism but **unproven**; what is proven is the guest is starved via steal.

## 3) Options (not executed)

### (a) Hostinger support / host migration

- Ask support to migrate the VM off the contended hypervisor; attach steal/`sar`/`mpstat` evidence.
- **Cost:** usually $0 if they migrate; time cost + possible short downtime.
- **Risk:** low data risk if they live-migrate; verify IP/hostname unchanged. No guarantee of lasting relief if fleet is generally oversubscribed.

### (b) Plan upgrade (catalog prices)

Source: `GET /api/billing/v1/catalog?category=VPS` (USD **cents**).  

| Plan | vCPU / RAM / disk | Monthly renew | Promo first month | Yearly renew |
|---|---|---|---|---|
| KVM 2 (current) | 2 / 8 GB / 100 GB | **$24.49** (2449) | $13.99 (1399) | $203.88 |
| KVM 4 | 4 / 16 GB / 200 GB | **$42.99** (4299) | $25.99 (2599) | $371.88 |
| KVM 8 | 8 / 32 GB / 400 GB | **$73.99** (7399) | $50.99 (5099) | $647.88 |

- Delta KVM2→KVM4 renew: **+$18.50/mo** (before tax / existing subscription proration — **proration not verified**).
- **Risk:** more vCPU may still steal if host is bad; upgrade helps capacity for tradebot/turkdini load even if steal drops only partially. Existing subscription upgrade path/price in hPanel **not verified** via API this pass.
- TRY public pricing: `hostinger.com.tr` returned **403** from this box; use USD catalog above. Dynamic HTML pricing page fetched but not parsed as authoritative vs API.

### (c) Temporarily park **api-test hook only** on idle prod srv2133566

Prod is KVM2 and idle (~0.5% CPU). Technically possible to run a **second**, isolated hook instance — **not done**.

**Risks / requirements if ever approved:**

| Topic | Requirement |
|---|---|
| Secrets | **Never** reuse live `SHOPIFY_WEBHOOK_SECRET` / `US_LICENSE_SECRET`. Separate env file; test store secret only. |
| Process isolation | Separate user (e.g. `uslic-test`), unit name, **different listen port** (not 8787 if prod uses it), own store dir. |
| nginx / DNS | Separate site `api-test.kanus.org` → this IP; do **not** point live `api.kanus.org` at test hook. ACME for api-test on prod IP. |
| Shopify | Keep webhooks on **kanus-test** only; live shop stays on `api.kanus.org` when cut over. |
| Blast radius | Test traffic / bad deploys / rate-limit experiments can starve or break **prod** on shared 2 vCPU if load grows; steal on test was host-level — moving hook alone does **not** fix tradebot fleets left on srv1662550. |
| Rollback | Keep binary+env on test VPS; DNS TTL low for api-test A record flip. |

**Cost:** $0 infra; operational risk medium–high for mixing environments.

## 4) Recommendation (advisory)

1. Short term: support ticket with steal evidence **or** isolate heavy workloads; do not trust panel 5% CPU.  
2. Capacity: KVM4 if workloads must stay co-located.  
3. Avoid parking api-test on prod except as emergency with strict isolation checklist above.

## Unverified

- Exact upgrade proration / invoice for existing `subscription_id`.
- TRY list prices.
- Whether Hostinger applies a hard CPU quota labeled as steal.
- Neighbour identity on the physical host.
