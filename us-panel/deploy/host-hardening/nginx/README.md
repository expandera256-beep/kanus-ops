# nginx host-hardening drafts (api-test.kanus.org)

**Status:** PREPARATION ONLY — not applied. LIVE already has per-IP webhook/credits rate limits, SSE 2700s paths, daily backup, fail2ban nginx jails, journald 500M.

## Diff summary vs LIVE (read-only snapshot 2026-11-06)

| Item | LIVE | This draft |
|------|------|------------|
| Webhook rate key | `$binary_remote_addr` 10r/s burst20 | `$server_name` **global** 40r/s burst80 (Shopify many IPs share one bucket) |
| /v1 rate | `uslic_credits` 5r/s | `uslic_v1` 10r/s burst20 (chat burst10) |
| `/account/` | **404** | Proxy to `127.0.0.1:8787`; login/register/reset/verify → strict `uslic_account_auth` 1r/s |
| Security headers | none on api-test | HSTS, nosniff, DENY/frame-ancestors, Referrer-Policy |
| SSE `/v1/chat*` `/v1/models` `/v1/credits/status` | buffering off, 2700s | **preserved** |

Binary/README confirms `/account/*` routes exist on Go; LIVE deliberately 404s them — enabling account is intentional product change.

## Apply (operator)

```bash
# 1) backup live
ts=$(date +%Y%m%d%H%M%S)
cp -a /etc/nginx/conf.d/us-license-hook-ratelimit.conf /root/nginx-bak-$ts-ratelimit.conf
cp -a /etc/nginx/snippets/us-license-hook-locations.conf /root/nginx-bak-$ts-snippet.conf
cp -a /etc/nginx/sites-available/api-test.kanus.org /root/nginx-bak-$ts-site.conf

# 2) install drafts
cp us-license-hook-ratelimit.conf /etc/nginx/conf.d/us-license-hook-ratelimit.conf
cp us-license-hook-locations.conf /etc/nginx/snippets/us-license-hook-locations.conf
cp api-test.kanus.org /etc/nginx/sites-available/api-test.kanus.org

# 3) test + reload (zone *rate* changes may need nginx restart once; prefer reload first)
nginx -t && systemctl reload nginx
# If limit_req_zone rate change does not take effect: systemctl restart nginx
# (does NOT restart us-license-hook)

# 4) smoke
curl -sI https://api-test.kanus.org/healthz
curl -s -o /dev/null -w '%{http_code}\n' -X POST https://api-test.kanus.org/webhooks/shopify/orders-paid
```

**Note:** Changing existing `limit_req_zone` parameters often requires a full `nginx` restart for the zone to recreate. Reload alone may keep the old zone definition in memory.

## Rollback

```bash
cp -a /root/nginx-bak-$ts-ratelimit.conf /etc/nginx/conf.d/us-license-hook-ratelimit.conf
cp -a /root/nginx-bak-$ts-snippet.conf /etc/nginx/snippets/us-license-hook-locations.conf
cp -a /root/nginx-bak-$ts-site.conf /etc/nginx/sites-available/api-test.kanus.org
nginx -t && systemctl reload nginx
# or restart nginx if zones stuck
```

Do not edit aradinazun / turkdini / turkiyegundemi sites.
