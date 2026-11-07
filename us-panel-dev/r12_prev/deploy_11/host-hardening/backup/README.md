# Backup drafts

Install (when approved):
```bash
install -m 755 backup-us-license-hook-hourly.sh /usr/local/sbin/
install -m 644 us-license-hook-hourly.cron /etc/cron.d/us-license-hook-hourly
# optional: install restore helper
install -m 755 restore-us-license-hook-store.sh /usr/local/sbin/
```
No service restart required. Do not scp env backups to the box.
