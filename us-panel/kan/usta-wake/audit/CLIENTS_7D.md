# usta-wake istemci denetimi (son 7 gün)

- Toplam istek satırı: **15844**
- Benzersiz kaynak IP: **492**
- Kaynak: `journalctl -u usta-wake.service` (BaseHTTPRequestHandler log_message)
- Port ayrımı yok; yol kalıbından çıkarım yapıldı. nginx yok.
- Box egress (bugün ölçülen): ~`140.248.50.0/24` (ör. 140.248.50.102)
- Cloud Run `us-commerce` / `us-watchdog`: env içinde `34.41.186.170` / usta-wake / :8788 /a100 **yok**.

## Önemli / işaretlenen istemciler

| IP | İstek | Etiket | İlk | Son | Üst yollar |
|---|---:|---|---|---|---|
| `213.14.161.140` | 5099 | LEGITIMATE NON-VPS? — FLAG | 2026-11-02T20:27:53+0000 | 2026-11-06T02:04:15+0000 | GET /api/durum×4988; POST /v1/chat/completions×97; GET /×5; GET /panel×3 |
| `24.133.236.116` | 1157 | LEGITIMATE NON-VPS? — FLAG | 2026-11-01T21:36:38+0000 | 2026-11-02T11:12:09+0000 | GET /api/durum×552; POST /v1/chat/completions×342; GET /api/nabiz×256; GET /panel×3 |
| `5.24.32.56` | 815 | LEGITIMATE NON-VPS? — FLAG | 2026-11-02T15:35:58+0000 | 2026-11-02T19:45:44+0000 | GET /api/durum×523; GET /api/nabiz×150; POST /v1/chat/completions×138; GET /panel×2 |
| `31.141.91.27` | 770 | LEGITIMATE NON-VPS? — FLAG | 2026-11-02T12:36:39+0000 | 2026-11-02T15:32:01+0000 | GET /api/durum×399; POST /v1/chat/completions×254; GET /api/nabiz×113; GET /×1 |
| `31.143.34.222` | 355 | LEGITIMATE NON-VPS? — FLAG | 2026-11-01T16:31:11+0000 | 2026-11-01T21:36:05+0000 | GET /api/nabiz×132; POST /v1/chat/completions×115; GET /api/durum×90; GET /panel×7 |
| `176.227.117.238` | 349 | LEGITIMATE NON-VPS? — FLAG | 2026-10-31T05:31:05+0000 | 2026-10-31T17:43:59+0000 | POST /v1/chat/completions×218; GET /api/nabiz×87; GET /×9; GET /api/durum×8 |
| `52.73.250.138` | 345 | LEGITIMATE NON-VPS? — FLAG | 2026-10-31T14:25:40+0000 | 2026-11-03T23:38:40+0000 | GET /api/durum×325; POST /v1/chat/completions×7; GET /health×4; GET /×2 |
| `3.214.111.192` | 339 | LEGITIMATE NON-VPS? — FLAG | 2026-10-31T16:16:08+0000 | 2026-11-03T23:47:40+0000 | GET /api/durum×321; POST /v1/chat/completions×8; GET /health×3; HEAD /indir/usta.exe×1 |
| `184.72.220.21` | 330 | LEGITIMATE NON-VPS? — FLAG | 2026-10-31T05:34:19+0000 | 2026-11-03T23:44:40+0000 | GET /api/durum×311; POST /v1/chat/completions×10; GET /health×2; GET /indir/usta.exe×1 |
| `34.232.125.200` | 327 | likely GCP/Google range — REVIEW | 2026-10-31T05:23:43+0000 | 2026-11-03T23:48:10+0000 | GET /api/durum×306; POST /v1/chat/completions×15; GET /health×2; GET /api/nabiz×1 |
| `34.195.13.132` | 313 | likely GCP/Google range — REVIEW | 2026-10-31T14:13:38+0000 | 2026-11-03T23:49:10+0000 | GET /api/durum×296; POST /v1/chat/completions×9; GET /health×4; GET /panel×1 |
| `54.236.202.137` | 302 | LEGITIMATE NON-VPS? — FLAG | 2026-10-31T14:25:32+0000 | 2026-11-03T23:43:10+0000 | GET /api/durum×288; POST /v1/chat/completions×11; GET /health×2; GET /×1 |
| `44.216.155.73` | 293 | LEGITIMATE NON-VPS? — FLAG | 2026-10-31T14:13:37+0000 | 2026-11-03T23:41:10+0000 | GET /api/durum×275; POST /v1/chat/completions×14; GET /×2; GET /health×1 |
| `31.220.63.191` | 228 | VPS (known) | 2026-11-06T04:29:17+0000 | 2026-11-06T05:41:14+0000 | GET /a100/status×226; POST /a100/start×2 |
| `127.0.0.1` | 125 | local-healthcheck | 2026-10-31T05:15:10+0000 | 2026-11-06T04:25:27+0000 | GET /×22; GET /health×9; [scanner-ish] /.env×5; GET /api/durum×5 |
| `45.148.10.247` | 61 | unknown/other | 2026-11-01T10:29:43+0000 | 2026-11-01T16:53:20+0000 | GET /login×6; GET /register×6; GET /admin/login×4; GET /×4 |
| `45.138.12.48` | 47 | unknown/other | 2026-11-01T10:56:43+0000 | 2026-11-02T08:46:24+0000 | GET /×6; GET /login×6; GET /register×6; GET /contact×4 |
| `45.198.224.125` | 44 | unknown/other | 2026-11-01T00:59:44+0000 | 2026-11-06T02:27:09+0000 | GET /×44 |
| `92.112.189.213` | 18 | LEGITIMATE NON-VPS? — FLAG | 2026-11-02T22:31:54+0000 | 2026-11-02T23:55:03+0000 | POST /v1/chat/completions×17; HEAD /×1 |
| `35.210.167.237` | 11 | likely GCP/Google range — REVIEW | 2026-11-01T06:21:03+0000 | 2026-11-01T06:21:14+0000 | OPTIONS /×2; GET /×2; IZHB /×1; HEAD /×1 |
| `35.210.198.165` | 9 | likely GCP/Google range — REVIEW | 2026-11-06T04:01:48+0000 | 2026-11-06T04:01:58+0000 | OPTIONS /×2; GET /×2; LYSS /×1; HEAD /×1 |
| `35.210.231.84` | 6 | likely GCP/Google range — REVIEW | 2026-11-04T06:24:38+0000 | 2026-11-04T06:24:40+0000 | OPTIONS /×2; MQEE /×1; GET /×1; HEAD /×1 |
| `164.52.24.185` | 5 | LEGITIMATE NON-VPS? — FLAG | 2026-11-01T01:47:48+0000 | 2026-11-01T01:47:57+0000 | GET /×1; GET /v1/models×1; GET /v1/embeddings×1; GET /v1/completions×1 |
| `186.244.201.29` | 5 | LEGITIMATE NON-VPS? — FLAG | 2026-11-05T02:37:45+0000 | 2026-11-05T12:23:14+0000 | GET /panel/status×3; GET /health×2 |
| `103.68.192.130` | 5 | LEGITIMATE NON-VPS? — FLAG | 2026-11-05T03:42:59+0000 | 2026-11-05T09:32:25+0000 | GET /panel/status×3; GET /health×2 |
| `54.189.162.162` | 5 | LEGITIMATE NON-VPS? — FLAG | 2026-11-06T01:41:28+0000 | 2026-11-06T06:42:06+0000 | GET /v1/models×2; GET /×1; GET /api/durum×1; POST /v1/chat/completions×1 |
| `44.227.25.49` | 5 | LEGITIMATE NON-VPS? — FLAG | 2026-11-06T02:30:58+0000 | 2026-11-06T06:32:40+0000 | GET /v1/models×3; GET /health×1; POST /v1/chat/completions×1 |
| `34.90.165.170` | 4 | likely GCP/Google range — REVIEW | 2026-10-31T14:29:08+0000 | 2026-10-31T14:29:17+0000 | GET /×3; HEAD /favicon.ico×1 |
| `100.21.173.32` | 4 | LEGITIMATE NON-VPS? — FLAG | 2026-11-06T02:53:17+0000 | 2026-11-06T04:41:55+0000 | GET /×1; GET /healthz×1; POST /a100/start×1; POST /account/login×1 |
| `64.118.143.136` | 3 | LEGITIMATE NON-VPS? — FLAG | 2026-11-03T20:31:49+0000 | 2026-11-03T20:37:42+0000 | GET /health×2; GET /gas-sponsor/info×1 |
| `52.35.237.101` | 3 | LEGITIMATE NON-VPS? — FLAG | 2026-11-06T02:53:15+0000 | 2026-11-06T04:14:35+0000 | GET /health×2; GET /v1/models×1 |
| `178.243.205.170` | 2 | LEGITIMATE NON-VPS? — FLAG | 2026-11-02T15:35:17+0000 | 2026-11-02T15:35:18+0000 | GET /api/nabiz×2 |
| `35.232.76.162` | 2 | likely GCP/Google range — REVIEW | 2026-11-03T05:13:04+0000 | 2026-11-03T17:16:12+0000 | [scanner-ish] /.git/config×2 |
| `52.10.120.60` | 2 | LEGITIMATE NON-VPS? — FLAG | 2026-11-06T01:41:28+0000 | 2026-11-06T04:25:28+0000 | GET /health×1; GET /a100/status×1 |
| `35.210.223.144` | 1 | likely GCP/Google range — REVIEW | 2026-10-31T06:41:12+0000 | 2026-10-31T06:41:12+0000 | GET /×1 |
| `34.38.65.98` | 1 | likely GCP/Google range — REVIEW | 2026-11-01T04:11:54+0000 | 2026-11-01T04:11:54+0000 | GET /×1 |
| `35.210.181.143` | 1 | likely GCP/Google range — REVIEW | 2026-11-01T06:20:59+0000 | 2026-11-01T06:20:59+0000 | GET /×1 |
| `34.56.6.168` | 1 | likely GCP/Google range — REVIEW | 2026-11-01T12:51:31+0000 | 2026-11-01T12:51:31+0000 | GET /×1 |
| `34.38.254.144` | 1 | likely GCP/Google range — REVIEW | 2026-11-01T13:26:12+0000 | 2026-11-01T13:26:12+0000 | GET /×1 |
| `34.24.52.92` | 1 | likely GCP/Google range — REVIEW | 2026-11-01T13:47:35+0000 | 2026-11-01T13:47:35+0000 | GET /×1 |
| `35.210.236.4` | 1 | likely GCP/Google range — REVIEW | 2026-11-02T06:14:30+0000 | 2026-11-02T06:14:30+0000 | GET /×1 |
| `34.178.254.199` | 1 | likely GCP/Google range — REVIEW | 2026-11-02T10:06:18+0000 | 2026-11-02T10:06:18+0000 | OPTIONS /×1 |
| `34.82.99.54` | 1 | likely GCP/Google range — REVIEW | 2026-11-02T15:32:33+0000 | 2026-11-02T15:32:33+0000 | GET /×1 |
| `35.210.135.215` | 1 | likely GCP/Google range — REVIEW | 2026-11-03T03:47:19+0000 | 2026-11-03T03:47:19+0000 | GET /×1 |
| `34.78.7.72` | 1 | likely GCP/Google range — REVIEW | 2026-11-03T06:18:56+0000 | 2026-11-03T06:18:56+0000 | GET /×1 |
| `34.48.174.155` | 1 | likely GCP/Google range — REVIEW | 2026-11-03T18:04:13+0000 | 2026-11-03T18:04:13+0000 | GET /×1 |
| `35.240.22.124` | 1 | likely GCP/Google range — REVIEW | 2026-11-03T21:24:50+0000 | 2026-11-03T21:24:50+0000 | GET /×1 |
| `35.210.152.25` | 1 | likely GCP/Google range — REVIEW | 2026-11-04T06:24:02+0000 | 2026-11-04T06:24:02+0000 | GET /×1 |
| `35.210.82.106` | 1 | likely GCP/Google range — REVIEW | 2026-11-05T06:30:38+0000 | 2026-11-05T06:30:38+0000 | GET /×1 |
| `184.33.84.196` | 1 | LEGITIMATE NON-VPS? — FLAG | 2026-11-06T01:06:38+0000 | 2026-11-06T01:06:38+0000 | GET /v1/models×1 |
| `34.215.96.239` | 1 | likely GCP probe/health (Google range) | 2026-11-06T01:54:27+0000 | 2026-11-06T01:54:27+0000 | GET /health×1 |
| `35.210.120.245` | 1 | likely GCP/Google range — REVIEW | 2026-11-06T04:01:43+0000 | 2026-11-06T04:01:43+0000 | GET /×1 |
| `54.201.40.3` | 1 | LEGITIMATE NON-VPS? — FLAG | 2026-11-06T04:43:34+0000 | 2026-11-06T04:43:34+0000 | GET /a100/status×1 |

## Tüm IP özeti (ilk 40, istek sayısına göre)

| IP | İstek | Etiket | Son |
|---|---:|---|---|
| `213.14.161.140` | 5099 | LEGITIMATE NON-VPS? — FLAG | 2026-11-06T02:04:15+0000 |
| `24.133.236.116` | 1157 | LEGITIMATE NON-VPS? — FLAG | 2026-11-02T11:12:09+0000 |
| `144.91.74.145` | 936 | scanner/botnet | 2026-11-04T01:28:23+0000 |
| `5.24.32.56` | 815 | LEGITIMATE NON-VPS? — FLAG | 2026-11-02T19:45:44+0000 |
| `31.141.91.27` | 770 | LEGITIMATE NON-VPS? — FLAG | 2026-11-02T15:32:01+0000 |
| `74.207.252.24` | 506 | scanner/botnet | 2026-11-04T03:19:00+0000 |
| `31.143.34.222` | 355 | LEGITIMATE NON-VPS? — FLAG | 2026-11-01T21:36:05+0000 |
| `176.227.117.238` | 349 | LEGITIMATE NON-VPS? — FLAG | 2026-10-31T17:43:59+0000 |
| `52.73.250.138` | 345 | LEGITIMATE NON-VPS? — FLAG | 2026-11-03T23:38:40+0000 |
| `3.214.111.192` | 339 | LEGITIMATE NON-VPS? — FLAG | 2026-11-03T23:47:40+0000 |
| `18.235.67.204` | 338 | scanner/botnet | 2026-11-03T23:49:59+0000 |
| `184.72.220.21` | 330 | LEGITIMATE NON-VPS? — FLAG | 2026-11-03T23:44:40+0000 |
| `34.232.125.200` | 327 | likely GCP/Google range — REVIEW | 2026-11-03T23:48:10+0000 |
| `45.138.12.65` | 323 | scanner/botnet | 2026-11-05T07:16:17+0000 |
| `34.195.13.132` | 313 | likely GCP/Google range — REVIEW | 2026-11-03T23:49:10+0000 |
| `54.236.202.137` | 302 | LEGITIMATE NON-VPS? — FLAG | 2026-11-03T23:43:10+0000 |
| `44.216.155.73` | 293 | LEGITIMATE NON-VPS? — FLAG | 2026-11-03T23:41:10+0000 |
| `80.94.95.211` | 279 | scanner/botnet | 2026-11-02T20:57:52+0000 |
| `31.220.63.191` | 228 | VPS (known) | 2026-11-06T05:41:14+0000 |
| `45.138.12.16` | 179 | scanner/botnet | 2026-11-02T00:23:23+0000 |
| `45.138.12.51` | 178 | scanner/botnet | 2026-11-04T19:39:34+0000 |
| `79.124.58.82` | 150 | scanner/botnet | 2026-11-04T07:28:59+0000 |
| `188.165.236.183` | 128 | scanner/botnet | 2026-10-31T11:52:37+0000 |
| `111.10.163.7` | 128 | scanner/botnet | 2026-10-31T15:37:58+0000 |
| `127.0.0.1` | 125 | local-healthcheck | 2026-11-06T04:25:27+0000 |
| `45.134.142.204` | 87 | scanner/botnet | 2026-11-05T14:29:29+0000 |
| `45.148.10.247` | 61 | unknown/other | 2026-11-01T16:53:20+0000 |
| `45.148.10.5` | 57 | scanner/botnet | 2026-11-04T10:34:11+0000 |
| `185.177.72.64` | 56 | scanner/botnet | 2026-11-02T06:46:39+0000 |
| `45.140.193.140` | 49 | scanner/botnet | 2026-10-31T06:45:38+0000 |
| `103.160.205.37` | 49 | scanner/botnet | 2026-11-04T21:14:58+0000 |
| `45.138.12.48` | 47 | unknown/other | 2026-11-02T08:46:24+0000 |
| `45.198.224.125` | 44 | unknown/other | 2026-11-06T02:27:09+0000 |
| `194.180.49.41` | 40 | scanner/botnet | 2026-11-04T05:49:27+0000 |
| `129.213.151.234` | 38 | scanner/botnet | 2026-10-31T07:34:09+0000 |
| `85.11.167.192` | 38 | scanner/botnet | 2026-11-01T02:34:20+0000 |
| `138.68.144.227` | 31 | scanner/botnet | 2026-10-31T14:14:16+0000 |
| `45.156.87.34` | 30 | scanner/botnet | 2026-10-31T05:20:28+0000 |
| `174.138.37.24` | 26 | scanner/botnet | 2026-11-06T04:08:48+0000 |
| `91.92.40.12` | 18 | unknown/other | 2026-11-05T23:36:40+0000 |

… ve 452 IP daha (`clients_7d.tsv`).


## KRİTİK — VPS dışı meşru istemciler (daraltmadan önce)

`kalfa-allow-http` (:80) ve `kalfa-allow-worker` (:8788) şu an `0.0.0.0/0`.
Son 7 günde **VPS olmayan** gerçek kullanım trafiği var. Bunları yalnız iki VPS `/32`'sine daraltmak **usta panel / sohbet istemcilerini keser**.

| IP | İstek | Son (UTC) | Ne yapıyor |
|---|---:|---|---|
| `213.14.161.140` | 5099 | 2026-11-06 02:04 | TR konut — `/api/durum` + `/v1/chat/completions` + `/panel` |
| `24.133.236.116` | 1157 | 2026-11-02 11:12 | TR — durum + nabız + sohbet |
| `5.24.32.56` | 815 | 2026-11-02 19:45 | TR — durum + nabız + sohbet |
| `31.141.91.27` | 770 | 2026-11-02 15:32 | TR — durum + sohbet |
| `31.143.34.222` | 355 | 2026-11-01 21:36 | TR — nabız + sohbet |
| `176.227.117.238` | 349 | 2026-10-31 17:43 | TR — sohbet + `/indir/usta.exe` |
| `52.73.250.138` vb. AWS `3/18/44/54.*` | ~300'er | 2026-11-03 | `/api/durum` poll + az sohbet (panel veya izleme) |
| `34.232.125.200`, `34.195.13.132` | ~320 | 2026-11-03 | GCP aralığı — `/api/durum` + sohbet (Cloud Run değil; us-commerce/us-watchdog usta-wake'e bağlanmıyor) |

**VPS:**
- `31.220.63.191` — 228 istek, yalnız `/a100/status` (226) + `/a100/start` (2). Son: 2026-11-06 05:41 UTC.
- `82.29.184.225` — **7 günde sıfır istek** (HOST outbound teyidi bekleniyor).

**Cloud Run:** `us-commerce` ve `us-watchdog` (europe-west4) env değerlerinde `34.41.186.170`, usta-wake, `:8788`, `/a100` **yok**.

**Box egress:** bugün ölçülen `140.248.50.102` (aynı /24 dönüyor). Loglarda `140.248.50.*` **yok** (cold-start artık usta-wake kullanmıyor; apply sonrası box zaten kesilecek).

**Öneri (apply öncesi karar):**
1. **Daraltmayı yalnız `/a100` TLS yüzeyi için yap** (yeni tag + tcp:8443 /32) — panel :80/:8788 açık kalsın; VEYA
2. Onaylı plana göre :80/:8788'i de daralt — o zaman yukarıdaki TR/AWS istemcileri başka bir yola (us-gateway) taşınmış olmalı.

KAN.US talimatı: VPS dışı meşru istemci varsa daraltmadan önce bildir → **bu bölüm o bildirimdir.**
