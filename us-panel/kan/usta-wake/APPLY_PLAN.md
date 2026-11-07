# usta-wake TLS + firewall APPLY planı

Tarih: 2026-11-06 Europe/Istanbul. Phase 1 staging tamam. **Option A uygulandı** (09:53–09:56 TSİ). kalfa-allow-* daraltılmadı.

**Proje:** `project-27082a40-6083-4952-9db`  
**VM:** `usta-wake` / `us-central1-c` / `34.41.186.170` / tag `kalfa-http`  
**Unit:** `usta-wake.service` → `/usr/bin/python3 /opt/usta-wake.py`

## Staging artefaktları (box)

| Dosya | Açıklama |
|---|---|
| `/workspace/us-panel/kan/usta-wake/usta-wake.py.new` | Yama: :8443 TLS yalnız `/a100/*`, :80'de `/a100/*` → 404, allowlist iki VPS |
| `/workspace/us-panel/kan/usta-wake/usta-wake.py.current` | Canlıdan çekilen kopya (2026-11-06) |
| `/workspace/us-panel/kan/usta-wake/usta-wake.py.bak-a100api` | VM'deki `/opt/usta-wake.py.bak-20261106-a100api` |
| `/workspace/us-panel/kan/usta-wake/usta-wake.py.diff` | current → new |
| `/workspace/secrets/usta-wake/cert.pem` | Self-signed ECDSA P-256, SAN IP:34.41.186.170, 2 yıl (chmod 600) |
| `/workspace/secrets/usta-wake/key.pem` | Private key (chmod 600) — **yazdırma** |
| `/workspace/secrets/usta-wake/a100.token` | Yeni Bearer token (chmod 600) — **yazdırma** |
| `/workspace/secrets/usta-wake/pin.txt` | `sha256//…` SPKI pin |
| `/workspace/us-panel/kan/usta-wake/audit/kalfa-allow-*.before.json` | Eski firewall sourceRanges yedeği |

**Pubkey pin (güvenli, paylaşılabilir):** dosyada `pin.txt`. Apply sonrası VPS'e verilir.

## UYARI — meşru istemciler

Son 7 günde TR konut IP'leri ve AWS IP'leri `:8788` `/api/durum` + `/v1/chat/completions` kullanıyor.
`kalfa-allow-http` / `kalfa-allow-worker` yalnız VPS `/32`'ye daraltılırsa **bu istemciler kesilir**.
Ayrıntı: `audit/CLIENTS_7D.md` § KRİTİK.

Önerilen güvenli sıra (A): önce yalnız **8443 + yeni tag**; :80/:8788 daraltmayı panel taşınınca yap.
Onaylı plan (B): aşağıdaki tam daraltma — KAN.US/Demir meşru istemci uyarısını okuduktan sonra.

`82.29.184.225` loglarda **hiç yok**; HOST outbound teyidi gelmeden `/32` daraltma **yapma**.

## Apply sonrası etki

- Box'tan `http://34.41.186.170:{80,8788}` ve `:8443` **timeout** (kaynak VPS değil).
- Cold-start / agent wake **gcloud compute instances start** ile yapılmalı; usta-wake :80 üzerinden değil.
- VPS'ten `https://34.41.186.170:8443/a100/status` pin'li curl → 200.
- Eski token geçersiz; VPS `/etc/us-model/wake.token` güncellenmeli.
- US uç: `https://34.41.186.170:8443/a100/{start,status}`.

---

## APPLY komutları (B — onaylı tam plan)

```bash
set -euo pipefail
export PATH="$HOME/google-cloud-sdk/bin:$PATH"
P=project-27082a40-6083-4952-9db
Z=us-central1-c
VM=usta-wake
VPS1=31.220.63.191
VPS2=82.29.184.225   # HOST teyidi şart
STAMP=$(date +%Y%m%d-%H%M%S)
STAGE=/workspace/us-panel/kan/usta-wake
SEC=/workspace/secrets/usta-wake

# --- 0) Önkoşul: HOST IP teyidi + meşru istemci kararı ---
# VPS2 boşsa veya panel istemcileri hâlâ :80/:8788 kullanıyorsa DUR.

# --- 1) Firewall yedek (tekrar) ---
gcloud compute firewall-rules describe kalfa-allow-http --project=$P --format=json \
  > "$STAGE/audit/kalfa-allow-http.before.$STAMP.json"
gcloud compute firewall-rules describe kalfa-allow-worker --project=$P --format=json \
  > "$STAGE/audit/kalfa-allow-worker.before.$STAMP.json"
# Eski sourceRanges (2026-11-06 Phase1): her ikisi de ["0.0.0.0/0"]

# --- 2) Yeni tag usta-a100-api + /32 kuralları (önce izin, sonra daralt) ---
gcloud compute instances add-tags $VM --project=$P --zone=$Z --tags=usta-a100-api

gcloud compute firewall-rules create usta-a100-api-allow-8443 \
  --project=$P --network=default --direction=INGRESS --priority=900 \
  --allow=tcp:8443 \
  --source-ranges=${VPS1}/32,${VPS2}/32 \
  --target-tags=usta-a100-api \
  --description="usta-wake /a100 TLS :8443 from VPS only"

gcloud compute firewall-rules create usta-a100-api-allow-80 \
  --project=$P --network=default --direction=INGRESS --priority=900 \
  --allow=tcp:80 \
  --source-ranges=${VPS1}/32,${VPS2}/32 \
  --target-tags=usta-a100-api \
  --description="usta-wake :80 from VPS only (parallel to narrowed kalfa-allow-http)"

gcloud compute firewall-rules create usta-a100-api-allow-8788 \
  --project=$P --network=default --direction=INGRESS --priority=900 \
  --allow=tcp:8788 \
  --source-ranges=${VPS1}/32,${VPS2}/32 \
  --target-tags=usta-a100-api \
  --description="usta-wake :8788 from VPS only (parallel to narrowed kalfa-allow-worker)"

# --- 3) Dosyaları VM'e kopyala (servis henüz restart yok) ---
gcloud compute ssh $VM --project=$P --zone=$Z --command="sudo mkdir -p /etc/usta-wake && sudo chmod 700 /etc/usta-wake"
gcloud compute scp "$STAGE/usta-wake.py.new" $VM:/tmp/usta-wake.py.new --project=$P --zone=$Z
gcloud compute scp "$SEC/cert.pem" "$SEC/key.pem" "$SEC/a100.token" $VM:/tmp/ --project=$P --zone=$Z
gcloud compute ssh $VM --project=$P --zone=$Z --command="
  set -e
  sudo cp -a /opt/usta-wake.py /opt/usta-wake.py.bak-tls-$STAMP
  sudo cp -a /etc/usta-wake/a100.token /etc/usta-wake/a100.token.bak-$STAMP 2>/dev/null || true
  sudo install -m 600 /tmp/cert.pem /etc/usta-wake/cert.pem
  sudo install -m 600 /tmp/key.pem /etc/usta-wake/key.pem
  sudo install -m 600 /tmp/a100.token /etc/usta-wake/a100.token
  sudo install -m 755 /tmp/usta-wake.py.new /opt/usta-wake.py
  sudo rm -f /tmp/cert.pem /tmp/key.pem /tmp/a100.token /tmp/usta-wake.py.new
  # syntax check before restart
  python3 -m py_compile /opt/usta-wake.py
"

# --- 4) Restart (kesinti ~1-2 sn) ---
gcloud compute ssh $VM --project=$P --zone=$Z --command="
  sudo systemctl restart usta-wake.service
  sleep 1
  systemctl is-active usta-wake.service
  ss -lntp | grep -E ':80|:8788|:8443'
"

# --- 5) Eski geniş kuralları daralt (SİLME) ---
gcloud compute firewall-rules update kalfa-allow-http --project=$P \
  --source-ranges=${VPS1}/32,${VPS2}/32
gcloud compute firewall-rules update kalfa-allow-worker --project=$P \
  --source-ranges=${VPS1}/32,${VPS2}/32

# --- 6) Doğrulama ---
# Box'tan (bu makine) — timeout beklenir:
timeout 5 bash -c 'echo >/dev/tcp/34.41.186.170/80'  && echo FAIL_80_OPEN || echo OK_80_BLOCKED
timeout 5 bash -c 'echo >/dev/tcp/34.41.186.170/8788' && echo FAIL_8788_OPEN || echo OK_8788_BLOCKED
timeout 5 bash -c 'echo >/dev/tcp/34.41.186.170/8443' && echo FAIL_8443_OPEN || echo OK_8443_BLOCKED

# VPS'ten (HOST çalıştırır; GPU BAŞLATMA):
#   PIN=$(cat /etc/us-model/usta-wake.pin)   # sha256//...
#   curl -sS --max-time 15 --pinnedpubkey \"$PIN\" \\
#     -H \"Authorization: Bearer \$(cat /etc/us-model/wake.token)\" \\
#     https://34.41.186.170:8443/a100/status
# Beklenen: HTTP 200 JSON ok:true (instance TERMINATED olabilir — start etme)
```

### HOST'un VPS'lere yerleştireceği dosyalar

Box'tan HOST'a iletilecek (içerik yazdırılmadan scp):

- `/workspace/secrets/usta-wake/cert.pem` → VPS `/etc/us-model/usta-wake.pem` (veya pin-only)
- `/workspace/secrets/usta-wake/pin.txt` → VPS `/etc/us-model/usta-wake.pin`
- `/workspace/secrets/usta-wake/a100.token` → VPS `/etc/us-model/wake.token` (chmod 600)
- Hook env: `US_GPU_WAKE_URL=https://34.41.186.170:8443/a100/start`  
  `US_GPU_STATUS_URL=https://34.41.186.170:8443/a100/status`

---

## ROLLBACK

```bash
set -euo pipefail
export PATH="$HOME/google-cloud-sdk/bin:$PATH"
P=project-27082a40-6083-4952-9db
Z=us-central1-c
VM=usta-wake
# STAMP= apply sırasında kullanılan değer

# 1) Firewall'u eski haline (0.0.0.0/0)
gcloud compute firewall-rules update kalfa-allow-http --project=$P --source-ranges=0.0.0.0/0
gcloud compute firewall-rules update kalfa-allow-worker --project=$P --source-ranges=0.0.0.0/0

# 2) Yeni /32 kurallarını kaldır (ayrı onay gerekebilir; disable alternatifi)
gcloud compute firewall-rules delete usta-a100-api-allow-8443 --project=$P --quiet
gcloud compute firewall-rules delete usta-a100-api-allow-80 --project=$P --quiet
gcloud compute firewall-rules delete usta-a100-api-allow-8788 --project=$P --quiet

# 3) Tag kaldır
gcloud compute instances remove-tags $VM --project=$P --zone=$Z --tags=usta-a100-api

# 4) Kod + token geri
gcloud compute ssh $VM --project=$P --zone=$Z --command="
  set -e
  # bak-tls-\$STAMP apply'de oluşturulur; yoksa bak-20261106-a100api
  sudo cp -a /opt/usta-wake.py.bak-tls-STAMP /opt/usta-wake.py
  sudo cp -a /etc/usta-wake/a100.token.bak-STAMP /etc/usta-wake/a100.token
  sudo systemctl restart usta-wake.service
  systemctl is-active usta-wake.service
  ss -lntp | grep -E ':80|:8788|:8443' || true
"
# VPS wake.token'ı eski değere döndür (HOST).
```

## Apply öncesi kontrol listesi

- [ ] HOST: `31.220.63.191` ve `82.29.184.225` outbound teyidi
- [ ] KAN.US/Demir: VPS dışı panel istemcileri için :80/:8788 daraltma onayı (veya yalnız 8443)
- [ ] Cold-start job usta-wake kullanmıyor (parent: evet)
- [ ] GPU start yok (doğrulama yalnız /a100/status)
- [ ] Token/key/cert değerleri loglara yazılmayacak


---

## APPLIED (A) — 2026-11-06 09:53–09:56 Europe/Istanbul

Option A only. `kalfa-allow-http` and `kalfa-allow-worker` **unchanged** (`0.0.0.0/0`) because audit found legitimate non-VPS clients on `:80`/`:8788`. Grace period keeps old Bearer token + `:80 /a100/*` working for VPS allowlist until HOST installs new token/pin.

### What changed

| Item | Before | After |
|---|---|---|
| VM tags | `kalfa-http` | `kalfa-http`, `usta-a100-api` |
| FW `usta-a100-api-allow` | (none) | INGRESS tcp:8443 from `31.220.63.191/32`,`82.29.184.225/32` → tag `usta-a100-api`, priority 900 |
| FW `kalfa-allow-http` | `0.0.0.0/0` tcp:80 | **unchanged** `0.0.0.0/0` |
| FW `kalfa-allow-worker` | `0.0.0.0/0` tcp:8788 | **unchanged** `0.0.0.0/0` |
| `/opt/usta-wake.py` | pre-TLS (bak-tls-20261106-095309) | staged `.new` + grace patch |
| Listeners | :80, :8788 | :80, :8788, **:8443 TLS** |
| `/etc/usta-wake/a100.token` | old token | **new** token |
| `/etc/usta-wake/a100.token.old` | — | old token (grace) |
| `/etc/usta-wake/grace_old_token` | — | **present** (empty flag file) |
| `/etc/usta-wake/{cert,key}.pem` | — | self-signed ECDSA, SAN IP:34.41.186.170 |

Backup on VM: `/opt/usta-wake.py.bak-tls-20261106-095309`, `/etc/usta-wake/a100.token.bak-20261106-095309`.  
Firewall JSON backups: `audit/kalfa-allow-*.before.20261106-095309.json`.

Grace behaviour (while `/etc/usta-wake/grace_old_token` exists):
- Tokens accepted: new (`a100.token`) **or** old (`a100.token.old`)
- `:80 /a100/{start,status}` served for allowlisted VPS IPs (same handler as TLS)
- Without the flag file: `:80 /a100/*` → 404; only new token on `:8443`

### Verification (no GPU start)

| Check | Result |
|---|---|
| Box → tcp:8443 | **timeout / blocked** (OK) |
| Box → `:80/nope` | **404** (connectivity OK; kalfa not narrowed) |
| Box → `:8788/health` | **200** `{"ok":true,"port":8788,"model":"usta"}` |
| Box → `:80 /a100/status` + valid token | **403** (non-allowlisted source) |
| VPS srv1662550 → `https://:8443/a100/status` new token + `--cacert` | **200** `instance=TERMINATED` |
| VPS → same with `--pinnedpubkey` + `-k` | **200** |
| VPS → `:8443` old token (grace) | **200** |
| VPS → `:80 /a100/status` old token (grace) | **200** |
| VPS → `:80 /a100/status` new token (grace) | **200** |
| VPS → `:8443` bad/missing token | **403** |

Note: `:80 /` and `:80 /api/durum` still block while workers are down (`worth_waking` → `ensure`); use `:8788/api/durum` or `:8788/health` for liveness. Pre-existing.

### Box secrets for HOST (chmod 600, do not log)

- `/workspace/secrets/usta-wake/wake.pin` — copy of `cert.pem` (for `--cacert` / pin install)
- `/workspace/secrets/usta-wake.token` — new Bearer token (= `usta-wake/a100.token`)
- `/workspace/secrets/usta-wake.pin` — SPKI pin string (= `usta-wake/pin.txt`)
- Also: `usta-wake/{cert.pem,key.pem,a100.token,pin.txt}`

### End the grace period (after HOST installs new token+pin on both VPS)

No code edit. Flag is checked per request:

```bash
export PATH="$HOME/google-cloud-sdk/bin:$PATH"
gcloud compute ssh usta-wake --project=project-27082a40-6083-4952-9db --zone=us-central1-c   --command='sudo rm -f /etc/usta-wake/grace_old_token && sudo rm -f /etc/usta-wake/a100.token.old && echo grace_off'
```

After that: `:80 /a100/*` → 404; only new token works on `:8443`. Restart not required.

### Rollback (option A)

```bash
export PATH="$HOME/google-cloud-sdk/bin:$PATH"
P=project-27082a40-6083-4952-9db; Z=us-central1-c; VM=usta-wake
# restore code + old token
gcloud compute ssh $VM --project=$P --zone=$Z --command='
  sudo cp -a /opt/usta-wake.py.bak-tls-20261106-095309 /opt/usta-wake.py
  sudo cp -a /etc/usta-wake/a100.token.bak-20261106-095309 /etc/usta-wake/a100.token
  sudo rm -f /etc/usta-wake/grace_old_token /etc/usta-wake/a100.token.old
  sudo systemctl restart usta-wake.service
  systemctl is-active usta-wake.service
  ss -lntp | grep -E ":80|:8788|:8443" || true
'
# remove 8443 rule + tag (kalfa rules were never changed)
gcloud compute firewall-rules delete usta-a100-api-allow --project=$P --quiet
gcloud compute instances remove-tags $VM --project=$P --zone=$Z --tags=usta-a100-api
```

### Not done (deferred)

- Narrowing `kalfa-allow-http` / `kalfa-allow-worker` to VPS `/32` (would break TR residential + AWS panel clients — see `audit/CLIENTS_7D.md`)
- Creating `usta-a100-api-allow-80` / `-8788` parallel rules
- HOST install of wake.token / pin on VPS (HOST's job; paths above)
