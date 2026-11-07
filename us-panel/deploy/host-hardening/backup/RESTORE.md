# Store / env restore (DRAFT — not applied)

## What is backed up
| Path | Backup name | Mode on backup |
|------|-------------|----------------|
| `/opt/us-license-hook/licenses.json` | `licenses.json.<UTC>` | 640 |
| `/opt/us-license-hook/plans.json` | `plans.json.<UTC>` | 640 |
| `/etc/us-license-hook.env` | `us-license-hook.env.<UTC>` | **600 root-only** |

**Never** copy env backup files off the VPS or into this git/box tree.

## Hourly vs existing daily
- LIVE already has daily cron `15 3 * * *` → `backup-us-license-hook-store.sh` (plans + licenses, 14d).
- This draft adds **hourly** flock-based backup, **7-day** retention, and **env** (600). Keep both or replace daily — operator choice.

## Dry restore test (no live overwrite)
```bash
# On VPS only
newest=$(ls -1t /var/backups/us-license-hook/licenses.json.* | head -1)
tmpdir=$(mktemp -d)
cp -a "$newest" "$tmpdir/licenses.json.restored"
sha256sum /opt/us-license-hook/licenses.json "$tmpdir/licenses.json.restored"
# expect match if store unchanged since backup
```

## Live restore (maintenance)
1. Announce short webhook downtime.
2. `systemctl stop us-license-hook`
3. Snapshot current store: `cp -a licenses.json licenses.json.pre-restore.<stamp>`
4. `install -m 640 -o uslic -g uslic <backup> /opt/us-license-hook/licenses.json`
5. `systemctl start us-license-hook` → `curl -s http://127.0.0.1:8787/healthz`
6. Rollback: restore `pre-restore` file if needed.

Env restore: same stop/start; `install -m 600 -o root -g uslic` (or root:root). Never paste secrets into chat/logs.
