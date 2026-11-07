# SMTP Hostinger — 2026-11-06

Güncelleme: **06.11.2026 07:58 TSİ (Europe/Istanbul)**. Etiketler: PASS / FAIL / BLOCKED / NOT TESTED.

## Özet

| Madde | Durum | Kanıt |
|---|---|---|
| Mailbox | **PASS** | `support@kanus.org` (Demir); şifre `/workspace/secrets/hostinger_smtp_password.txt` mode 600 |
| Box SMTP auth (login-only) | **FAIL** | 465 SSLEOFError; 587 timeout — box egress; şifre denemesi çoğaltılmadı |
| VPS SMTP auth 465 (login-only) | **PASS** | root + `uslic` `SMTP_SSL` login OK |
| Outbox neutralize (SMTP öncesi) | **PASS** | servis stop → backup → 5/5 entry `sentAt`+`dead` (still_due=0); lisans/order korundu |
| VPS env SMTP | **PASS** | `US_SMTP_ENABLE=1`, HOST=smtp.hostinger.com, PORT=465, USER/FROM=support@kanus.org, PASS set (redacted); env 600 |
| healthz | **PASS** | 200; `us-license-hook` active |
| Test mail → demirtumen08@gmail.com | **PASS** | signed paid order **9000000000202** HTTP 200; outbox `sentAt=2026-11-06T04:56:06Z` (UTC); ilk deneme `context deadline exceeded`, retry sonrası gönderildi |
| Revoke cancelled | **PASS** | signed cancelled HTTP 200; `cancelled[9000000000202]=true`; lisans `revoked=true`, credits=0 |
| Diğer adreslere sızıntı | **PASS** | due leak yok; eski outbox neutralize edildi |

## Auth

- **Box:** FAIL (ağ); Go gönderimi VPS’te olduğu için VPS auth esas alındı.
- **VPS 465 implicit TLS:** PASS (Go binary yalnız `tls.Dial` / 465 kullanır; 587 STARTTLS desteklenmez).
- `US_SMTP_FROM=support@kanus.org` (çıplak adres). Kod `From: KAN.US <%s>` ve `MAIL FROM` için display-name sarmalamaz.

## Env (VPS `/etc/us-license-hook.env`)

```
US_SMTP_ENABLE=1
US_SMTP_HOST=smtp.hostinger.com
US_SMTP_PORT=465
US_SMTP_USER=support@kanus.org
US_SMTP_FROM=support@kanus.org
US_SMTP_PASS=***REDACTED***
```

Backup: `/opt/us-license-hook/backup-smtp-preenable-20261106-074711/` (licenses.json + env).

## Test kanıtı

1. Outbox worker: `SentAt.IsZero() && !Dead && !now.Before(NextTry)` → due.
2. SMTP açmadan önce 5 eski outbox (example.com + windows-smoke) neutralize: `dead=true` + `sentAt` set; `still_due=0`.
3. Synthetic paid: product `15473713348861` (10 kredi), email demirtumen08@gmail.com, name `#SYN-SMTP-20261106`.
4. Outbox Demir kaydı: attempts=1 (ilk dial timeout), sonra **sentAt dolu**, lastErr boş → **SEND PASS**.
5. Synthetic cancelled → **REVOKE PASS**.

## Notlar / residual risk

- Go dial timeout 20s; Hostinger’a ara sıra `context deadline exceeded` görüldü; outbox backoff retry ile düzeldi. İleride timeout artışı düşünülebilir.
- Box’tan SMTP auth yapılamıyor (egress); operasyonel testler VPS üzerinden.
- DKIM-b/c Hostinger tarafında `p=` boş (önceki rapor); DMARC `rua` hâlâ eksik (MCP DNS write timeout). SPF/MX/DKIM-a OK.
- Gerçek müşteri maili gönderilmedi.

## Sonuç

**SMTP teslimatı Demir test adresine PASS; revoke PASS; üretim SMTP açık (US_SMTP_ENABLE=1).**
