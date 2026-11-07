# TLS / UFW / fail2ban (read-only findings 2026-11-06 Europe/Istanbul ~08:11)

## TLS (certbot certificates — read-only)
| Name | Expiry (UTC) | Notes |
|------|--------------|-------|
| api-test.kanus.org | 2027-02-04 (~89d) | webroot renew; earlier dry-run **success** |
| aradinazun.com | 2027-01-25 | OK |
| turkdini.com | 2027-01-10 | OK |
| turkiyegundemi.com | 2026-12-08 (~31d) | renew soonest |

Renewal conf for api-test: `authenticator=webroot`, `webroot_path=/var/www/html`. Port 80 ACME location present on api-test site.

**Recommendation:** keep certbot timer; alert at <21 days on turkiyegundemi; optional `certbot renew --dry-run` weekly via cron (read-only-ish).

## UFW (active)
- Default: deny in / allow out / deny routed
- ALLOW: OpenSSH (22), Nginx Full (80/443), **8090–8095 from Anywhere** (TradeBot UIs)
- Recommendation (safe, reversible): restrict 8090–8095 to admin VPN/IP allowlist; leave 22/80/443 as-is or prefer key-only (already `PermitRootLogin without-password`).

## fail2ban (already applied today)
- Jails: `sshd`, `nginx-http-auth`, `nginx-limit-req`
- Currently banned: **0** on sshd and nginx-limit-req
- Recommendation: after global webhook rate change, nginx-limit-req may fire less on Shopify; keep jail. Add `ignoreip` for office/box egress if not already (LIVE jail.local includes box IP from earlier apply).

## journald
- Drop-in already: SystemMaxUse=500M, RuntimeMaxUse=200M; usage ~431M.
