# Gap Tracker (§1–§15)

Güncelleme: 06.11.2026 07:58 TSİ (Europe/Istanbul). Etiketler: PASS=kanıtlı, FAIL, BLOCKED, NOT TESTED. Üretime hazır DEĞİL.

| § | Başlık | Durum | Kanıt |
|---|---|---|---|
| 1 | Keşif | DONE | STATUS_TR.md, MERGE_20261106_TR.md |
| 2 | US panel işlevleri | IN_PROGRESS | TEST_RESULTS.md (birim); canlı ajan/terminal NOT TESTED |
| 3 | DeepSeek model | BLOCKED | gerçek model testleri opt-in kapalı; SSH tüneli doğrulanmadı |
| 4 | İstem-başı GPU | IN_PROGRESS | GPU_SOGUK_ACILIS_TR.md; idle 15 dk; soğuk açılış >10 dk |
| 5 | Cloud Run/watchdog | IN_PROGRESS | HOOK_VPS_STATUS; Cloud Run kapsam dışı; watchdog NOT TESTED |
| 6 | Shopify/kredi | IN_PROGRESS | KANUS_SHOPIFY_REPORT.md; webhook + SMTP test PASS (Demir) |
| 7 | Marka kimliği | IN_PROGRESS | KANUS_BRAND_REPORT.md |
| 8 | Çok kullanıcı güvenlik/perf | IN_PROGRESS | race PASS; LOAD_TEST; model proxy tunnel |
| 9 | Gerçek proje üretim testi | NOT_STARTED | skip |
| 10 | Windows EXE | IN_PROGRESS | native smoke parent PASS bildirildi |
| 11 | Maliyet/ticari | IN_PROGRESS | Seçenek A seçildi, canlıya uygulanmadı |
| 12 | Zorunlu testler | IN_PROGRESS | KANUS_TEST_RESULTS.md |
| 13 | Canlı işlem sınırları | DONE | gerçek müşteri maili yok; purchase yok |
| 14 | Bitiş raporu | DONE | KANUS_*.md |
| 15 | Nihai emir | IN_PROGRESS | SMTP Hostinger: **PASS** (Demir test) — SMTP_HOSTINGER_20261106.md |

## SMTP / Hostinger Mail (06.11.2026 07:58 TSİ)

- Mailbox `support@kanus.org`; VPS auth 465 PASS; env ENABLE=1.
- Outbox neutralize PASS; test mail demirtumen08@gmail.com **PASS**; cancelled revoke **PASS**.
- Box auth FAIL (egress). DKIM-b/c + DMARC rua hâlâ açık (DNS MCP).
- Kanıt: `SMTP_HOSTINGER_20261106.md`
