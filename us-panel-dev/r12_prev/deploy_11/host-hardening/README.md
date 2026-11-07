# Host hardening — PREPARATION ONLY (2026-11-06)

Nothing in this tree was applied to the VPS. VPS access for this pack was **read-only**.

Already LIVE earlier today (do not duplicate blindly): per-IP rate limits, daily backup cron, fail2ban nginx jails, journald 500M. Drafts below **diff** against that baseline.

## Layout
- `nginx/` — api-test site + conf.d zones + snippet (global webhook rate, account proxy, headers, SSE preserved)
- `backup/` — hourly flock backup (store+env 600), restore helper, RESTORE.md
- `security/` — TLS/UFW/fail2ban notes + **SSH_FLAKY_REPORT.md**
- `PROD_PLAN.md` — api.kanus.org same-VPS vs new-VPS recommendation

## Key finding
SSH timeouts ↔ **CPU steal + memory/swap pressure + OOM**, not fail2ban/conntrack.
