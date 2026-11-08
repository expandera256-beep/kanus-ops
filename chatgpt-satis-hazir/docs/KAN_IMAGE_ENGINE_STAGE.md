# KAN Image Engine — Pass 3: PRIVATE GCS STAGED / VERIFIED / NOT DEPLOYED

Timestamp: 2026-11-08T04:01:34.201242+03:00 TSİ (Europe/Istanbul).
Pass-2 gövdesi 2026-11-08T03:58:49.653758+03:00 TSİ’de yazıldı; bu sürüm tokenizer LFS uzlaştırmasını kanıta işler.

## GPU
GPU was not opened; ds-a100 remained TERMINATED at both read-only describe checks.
- Start: 8 Nov 2026 03:57:34 TSİ. End: 8 Nov 2026 03:58:49 TSİ.
- Exact evidence: image_stage_evidence/ds-a100-pass2-start.json and ds-a100-pass2-end.json. No continuous monitoring is claimed.
- No start/stop/SSH, BOOT6, /v1/chat, timer, policy or live deployment changes.

## Correction
Previous no-object listing means an empty virtual media-models/ prefix, not a missing bucket or staging prohibition. Bucket p27082-ds-istek exists. No new bucket was created.

## Local klein download
Pinned repo: black-forest-labs/FLUX.2-klein-4B@e7b7dc27f91deacad38e78976d1f2b499d76a294.
Local root: /workspace/tmp/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294
Disk before download: 120 GiB available (>20GB). Anonymous public HTTPS only; no credentials searched or used. No monolithic checkpoint downloaded.
4/4 weight files VERIFIED locally against both manifest sizes and SHA256. Total weights 15,964,212,614 bytes.

| Weight | Bytes | Local SHA256 |
|---|---:|---|
| text_encoder/model-00001-of-00002.safetensors | 4967215360 | 8c0506e7f4936fa7e26183a4fd8da4e2bdbc5990ba64ae441f965d51228f36ea |
| text_encoder/model-00002-of-00002.safetensors | 3077766632 | 82f2bd839378541b0557bfabaf37c7d3d637071fdcb73302dedd7cf61162ce07 |
| transformer/diffusion_pytorch_model.safetensors | 7751109744 | 9f29f9edcfdae452a653ffb51a534ca4decd389952c225724ff3b94042612a6e |
| vae/diffusion_pytorch_model.safetensors | 168120878 | ca70d2202afe6415bdbcb8793ba8cd99fd159cfe6192381504d6c4d3036e0f04 |

## Snapshot integrity — tokenizer LFS uzlaştırması (Pass 2b)
Pass 2, tokenizer/tokenizer.json için ekstra git-blob SHA1 karşılaştırmasında durdu ve dosyayı `tokenizer.json.partial` olarak karantinaya aldı:
- Quarantine/final boyut: 11422654
- Yerel SHA256: aeb13307a71acd8fe81861d94ad54ab689df773318809eed3cbe794b4492dae4
- Manifest/HF git oid: cd71f61a15a522601badb3dc960d800d9cb3766c
- Ölçülen git-blob SHA1 (dosya içeriği): 949e1ec83f61520a25c75426edc4a43acc36f29a
- Hata: AssertionError: git blob SHA1 mismatch

8 Kasım 2026 geçiş 2b’de Hugging Face tree API yeniden okundu:
`https://huggingface.co/api/models/black-forest-labs/FLUX.2-klein-4B/tree/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer`
`tokenizer/tokenizer.json` kaydı:
- `oid` = cd71f61a15a522601badb3dc960d800d9cb3766c (git LFS pointer; pointerSize=133). Bu, dosya içeriğinin SHA1’i değildir.
- `lfs.oid` = aeb13307a71acd8fe81861d94ad54ab689df773318809eed3cbe794b4492dae4
- `lfs.size` / `size` = 11422654

Yerel dosya aynı SHA256 ve boyuta sahip; LFS oid ile birebir eşleşir. Git-blob sapması yanlış alarmdır: manifest `git_blob` alanı pointer oid’sidir. Dosya final yola alındı:
`/workspace/tmp/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer/tokenizer.json`
`.partial` artık yok. Kanıt: image_stage_evidence/tokenizer-lfs-reconciliation.json (SHA256 fe2b65acd96c9e855430a648b6e50f660e8af64c7ff35e3ba84909a6d3c27575).

Yerel klein diffusers ağacı bu doğrulamayla tamdır (4/4 ağırlık + destek dosyaları + tokenizer.json). Ready/publication marker YAZILMADI. Pass 2b sırasında GCS yüklemesi yoktu; güncel Pass 3 sonucu aşağıda.

## Helper / license files locally present
- model_index.json: 446 bytes; SHA256 51a76cb1cf3ed37423a1128c79c22faee8e6fbe7f5aaeb737f0a258930dbaac0
- LICENSE.md: 9584 bytes; SHA256 ca02bc51900ab07789d1b70283329e7137f5af98f5161c23a1c81fc38a4af1fe
- README.md: 9991 bytes; SHA256 bd447051cf2db4bf67b992e563d89c35ebf1123f009c063361a222ab14e1c4f7
- scheduler/scheduler_config.json: 486 bytes; SHA256 067afb012cef64553a763447d1efd93daeffcc0123ca7e25b09f8de20b90762e
- text_encoder/config.json: 1536 bytes; SHA256 214b4c29a0d975e9fddf9994a5673f22cb2c4c5750352f9227c2c3251ebeab40
- text_encoder/generation_config.json: 214 bytes; SHA256 4347b1aeed2b2b78bc059920a0b7f5fec71482e1344952b76d7665d638d71f13
- text_encoder/model.safetensors.index.json: 32855 bytes; SHA256 06b3d5319b6d76d1a4a2433419180016cfd54ed62d086a5e6567a809f8c82634
- tokenizer/added_tokens.json: 707 bytes; SHA256 c0284b582e14987fbd3d5a2cb2bd139084371ed9acbae488829a1c900833c680
- tokenizer/chat_template.jinja: 4168 bytes; SHA256 a55ee1b1660128b7098723e0abcd92caa0788061051c62d51cbe87d9cf1974d8
- tokenizer/merges.txt: 1671853 bytes; SHA256 8831e4f1a044471340f7c0a83d7bd71306a5b867e95fd870f74d0c5308a904d5
- tokenizer/special_tokens_map.json: 613 bytes; SHA256 76862e765266b85aa9459767e33cbaf13970f327a0e88d1c65846c2ddd3a1ecd
- tokenizer/tokenizer.json: 11422654 bytes; SHA256 aeb13307a71acd8fe81861d94ad54ab689df773318809eed3cbe794b4492dae4 (HF LFS oid ile eşleşti)
- tokenizer/tokenizer_config.json: 5404 bytes; SHA256 443bfa629eb16387a12edbf92a76f6a6f10b2af3b53d87ba1550adfcf45f7fa0
- tokenizer/vocab.json: 2776833 bytes; SHA256 ca10d7e9fb3ed18575dd1e277a2579c16d108e32f27439684afa0e10b1440910
- transformer/config.json: 541 bytes; SHA256 09733c74a3da6d17dd0a0472a091a8950c7c6935889c32c16cc800ede05029de
- vae/config.json: 821 bytes; SHA256 0d6dfb69ae95a5e2ac9836284bbb63d8b38ce67b25ba2dff380752b2a10ab948
- IMAGE_WEIGHT_MANIFEST.json: 10137 bytes; SHA256 6198c9a8d50475a1a236ae7c6baa2e9c1faa9eaeb26d673da1a232a0228aec7b
- licenses/Qwen3-4B-LICENSE.txt: 11343 bytes; SHA256 832dd9e00a68dd83b3c3fb9f5588dad7dcf337a0db50f7d9483f310cd292e92e
- licenses/FLUX2-VAE-PROVENANCE-README.md: 8466 bytes; SHA256 435753749320bbb395c15983e63e3dc73936c03f0f1870c628e7bf922c55eef7
- licenses/official-license-sources.json: 1757 bytes; SHA256 276fdc32a6f75a127bfdcf9eb3b2b6c3c6135d3fa313f1810f71ba575f9af4a3
No NOTICE file was invented. Source IMAGE_WEIGHT_MANIFEST.json is copied unchanged, including its historical PLAN ONLY status; current results are the separate local verification evidence.

## GCS — historical Pass 2b status (superseded below)
NOT UPLOADED in Pass 2b. 0 uploaded files/bytes. No upload was attempted this pass, per explicit instruction; direct-user storage confirmation remains separate. No IAM/ACL/lifecycle/traffic changes.
Proposed destination only: gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/

## Code and CPU tests
No executable package written. /workspace/us-panel/kan/image-engine/README.md remains the prior review-handoff checklist, NOT DEPLOYED.
The developer-provided code-changes skill mandates CloudAgent for non-trivial code work. The delegated correction cannot override that requirement. CloudAgent is absent; no CloudAgent/Origin/git/deploy action was taken.
CPU tests: NOT RUN (0 passed, 0 failed; implementation unavailable). Weight/support hash checks above are artifact validation, not supervisor CPU tests.

## Schnell
Still BLOCKED (prior anonymous HTTP 401), not retried this pass. Prior exact response: Access to model black-forest-labs/FLUX.1-schnell is restricted. You must have access to it and be authenticated to access it. Please log in.

## NOT TESTED
No GPU smoke, from_pretrained, image generation, NVML/context handoff, VRAM/latency, sequential supervisor tests or deployment. No customer data used.

## Evidence
- image_stage_evidence/klein-local-verification.json
- image_stage_evidence/klein-local-provenance.json
- image_stage_evidence/tokenizer-mismatch.json (geçiş 2 karantina kaydı; geçersiz kılınmadı, üstüne uzlaştırma eklendi)
- image_stage_evidence/tokenizer-lfs-reconciliation.json
- image_stage_evidence/ds-a100-pass2-{start,end}.json
- Image plan SHA256 unchanged: 6d44566c9c0f20679d3c7a7f0d95b1a625a602a95eea9e13da1d4053b1d70a38.


## Pass 3 — executable LOCAL supervisor package / NOT DEPLOYED

Timestamp: 2026-11-08T04:05:39.281238+03:00 TSİ (Europe/Istanbul, UTC+3).
Append-only update; the Pass 2b tokenizer LFS reconciliation above is preserved.
Previous report SHA256 before this append: `018c2cbea77c36723bec81b6877be2afac2a4d5638e4f58066f52a18c3df0f03`.
This section supersedes the historical Pass-2 "No executable package / tests not run"
statements above. No CloudAgent, Origin or git operation was used.

### Actual code and behavior

Package root: `/workspace/us-panel/kan/image-engine/`.
**10 actual Python files written**, not just a README.
- Explicit state machine, serialized image admission, drain active chat without
  interruption, chat precedence before image load and at each serial image boundary.
- Linux OS advisory flock, durable fsynced monotonic fencing generation, boot-id,
  PID and kernel process-start identity; stale/replaced identities are rejected.
- Fixed model/revision allowlist, explicit boolean fallback/access permission,
  no automatic fallback/retry and OOM/memory fallback refusal; sizes restricted,
  n=1..4 via serial one-image calls; local snapshot SHA256 verification and no
  runtime HF downloads (`local_files_only=True`, offline child environment).
- Separate-venv local worker transport and owned process-group termination code;
  fail-closed physical-UUID NVML compute/graphics-context probe; existing-local
  chat reload/readiness/synthetic-smoke contract. None bound to live services.
- Durable idempotency/one local terminal settlement, bounded boot/PID/fence/job
  work lease, hard absolute deadline and pre-reserved restore/encode/stop budget.
  Restore failure or crash barrier is persistent FAILED_CLOSED; no new boot.
- Health/ready/list/poll are inert; polling reads the local journal only, without
  wake, lease, idle refresh or adapter calls. Policy values remain unchanged:
  daily_runtime=3600, max_uptime=1800, idle=600.

### CPU evidence

Command: `cd /workspace/us-panel/kan/image-engine && python3 -m unittest discover -s tests -v`.
Result: **70 passed, 0 failed, 0 errors, 0 skipped**, Python 3.13.5,
0.567 seconds in the final run. Compile checks also passed.
Coverage includes transition ordering; OS-lock contention in independent processes;
crashed owner, stale boot/PID/start/fence and lock inode; active chat drain/timeout;
preemption; failed context release; lost fence during inference; durable restore
failure; fallback/OOM and allowlist refusal; serial n=4; duplicate settlement;
lease loss/deadline; missing/untrusted/corrupt snapshot files and shards; inert
health/list/poll including restart lookup. NVML/Diffusers/Torch are mocked.
A real LOCAL CPU-only Python worker IPC/owned-stop smoke was run with
`CUDA_VISIBLE_DEVICES` empty and NO load/generate operation. Synthetic encode test
writes placeholder bytes, not a real generated image. No customer data used.

### Tokenizer preserved

Reconfirmed final tokenizer path, size 11,422,654 and SHA256
`aeb13307a71acd8fe81861d94ad54ab689df773318809eed3cbe794b4492dae4`.
This matches the reconciled HF LFS oid from Pass 2b. The git pointer oid is NOT a
file-content SHA1. No snapshot file was renamed, modified or re-quarantined.

### Safety / remaining gates

**GCS still NOT UPLOADED (0 files/bytes). GPU was not opened.**
No GPU/VM start OR stop, ds-a100 SSH, BOOT6, Cloud Run deployment, timer/policy,
secret, Hostinger, DNS, Shopify, firewall, bucket or ACL change was attempted.
No new live GPU status check is claimed by this local-code pass.
Schnell remains BLOCKED/gated; no download, credential search or gate bypass.

**NOT TESTED / NOT DEPLOYED:** real from_pretrained/CUDA inference and image output,
NVML on hardware, VRAM/latency/SLO, local-chat reload and production restart/tunnel
coordination. Production chat/lifecycle adapters remain deliberately unbound.
An exact dependency/wheel-hash lock, immutable snapshot mount/trust-anchor review,
shared atomic reservation integration, durable output policy and operator recovery
runbook are still deployment gates. No guessed dependency versions were installed.
Cross-model replay under an existing reservation is deliberately refused; no
second attempt or charge is silently created. Local terminal journal is not a
payment/refund service. CPU PASS does not claim commercial readiness.

### Written source and evidence SHA256

All paths below are relative to `/workspace/us-panel/kan/image-engine/`.

| File | SHA256 |
|---|---|
| `README.md` | `e2a0b452632193e45031a7196f709c5d05380f4ac9214fe6316d5d16cc2a5e43` |
| `image_engine/__init__.py` | `7016a3f7e36b1dd16c13a4b91474cc5fdac959a58fb74887b00e5288b22ff7f4` |
| `image_engine/integration.py` | `e5ee5c439d5ca53484a169b397314be0c6a10fc3c6d825db4b0537e8db845690` |
| `image_engine/ledger.py` | `11dfd2883fbdc1aa3a4b15f029d94527a8626b29d578585fcdf5e4c9ad55c72e` |
| `image_engine/models.py` | `107dbca8180c8d2bafff845fb6533257311678f8fcda7716c8ba9ebc9a2de767` |
| `image_engine/ownership.py` | `b0c28f0c9893c6f730f88d03077d0b83ee9ce0a850058ba4fb2ec7a0f054d1b5` |
| `image_engine/process.py` | `9bd6970bc210ab873e6db14cde609b59daaf4168da15d4d6ab846f4e062e84d3` |
| `image_engine/supervisor.py` | `80b41b07f61fa2228be187cafbfe0d6241954f6d2ebc3be20932c0e7b7487c40` |
| `image_engine/worker.py` | `8a2e36b6f49e357c723125b2a04405316045611655a60e6dec561df90c5393cb` |
| `tests/test_engine.py` | `b4924c9198c1c402d658939e2ac8a7456a28dd56459c96f83e12f8080e1dc591` |
| `tests/test_integration.py` | `1d44332d155d5cabceee4dd65a03bc32b180db1a14e36d42bda27a0f872421b8` |
| `cpu-tests.log` | `22d0f42cd0c9cde96a151f669344f3729f24ed8c1d467df546e87cf68e5714bb` |
| `cpu-test-results.json` | `46aff514ed98944bea31bef30bf7024830d00c947445192504dc76a3f9f955ca` |

Artifact index: `ARTIFACT_SHA256.json`, SHA256 `14f2d1face322c9d724fb70f903951af7687381d67527e453f377d674ee74848`.
The report's new SHA256 is returned separately to avoid self-referential hashing.

## Pass 3 — approved private GCS staging
Completed: 2026-11-08T04:07:11.639374+03:00 TSİ (Europe/Istanbul).
Direct user approval independently verified in KAN.US conversation: widget t72s266, 8 Nov 2026 04:01 TSİ. Scoped to this existing private prefix only.
24 snapshot objects uploaded, 15,980,183,023 bytes. Initial destination listing matched no objects; final inventory contains exactly 24 objects plus virtual directory prefixes. Each upload used `gcloud storage cp --if-generation-match=0` (atomic create-only). No existing snapshot object was overwritten or deleted. No new bucket, ACL/IAM or public sharing change. Existing bucket public_access_prevention=enforced and uniform_bucket_level_access=true.
GPU was not opened; ds-a100 remained TERMINATED at read-only checks 2026-11-08T04:03:49.315223+03:00 and 2026-11-08T04:05:59.989310+03:00 TSİ. Point-in-time checks, not continuous monitoring. No GPU/VM start/stop, BOOT6, timer, daily_runtime/max_uptime, Cloud Run or deployment action.
All 24 GCS sizes match local sizes. Local SHA256 recomputed for all 24 files. Four weights and tokenizer.json additionally have matching local/GCS CRC32C; GCS MD5 absent for composite weight objects and present/matching for tokenizer.json. These are staging integrity checks, not a serving test.

### Key objects
| Full GCS path | Bytes | Local SHA256 | GCS CRC32C (base64) |
|---|---:|---|---|
| gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/text_encoder/model-00001-of-00002.safetensors | 4967215360 | 8c0506e7f4936fa7e26183a4fd8da4e2bdbc5990ba64ae441f965d51228f36ea | sB2QGg== |
| gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/text_encoder/model-00002-of-00002.safetensors | 3077766632 | 82f2bd839378541b0557bfabaf37c7d3d637071fdcb73302dedd7cf61162ce07 | nDNsfw== |
| gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer/tokenizer.json | 11422654 | aeb13307a71acd8fe81861d94ad54ab689df773318809eed3cbe794b4492dae4 | t3YzMQ== |
| gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/transformer/diffusion_pytorch_model.safetensors | 7751109744 | 9f29f9edcfdae452a653ffb51a534ca4decd389952c225724ff3b94042612a6e | FPY9HA== |
| gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/vae/diffusion_pytorch_model.safetensors | 168120878 | ca70d2202afe6415bdbcb8793ba8cd99fd159cfe6192381504d6c4d3036e0f04 | Uw7SfQ== |

### Other uploaded objects
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/IMAGE_WEIGHT_MANIFEST.json — 10137 bytes; SHA256 6198c9a8d50475a1a236ae7c6baa2e9c1faa9eaeb26d673da1a232a0228aec7b
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/LICENSE.md — 9584 bytes; SHA256 ca02bc51900ab07789d1b70283329e7137f5af98f5161c23a1c81fc38a4af1fe
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/README.md — 9991 bytes; SHA256 bd447051cf2db4bf67b992e563d89c35ebf1123f009c063361a222ab14e1c4f7
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/licenses/FLUX2-VAE-PROVENANCE-README.md — 8466 bytes; SHA256 435753749320bbb395c15983e63e3dc73936c03f0f1870c628e7bf922c55eef7
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/licenses/Qwen3-4B-LICENSE.txt — 11343 bytes; SHA256 832dd9e00a68dd83b3c3fb9f5588dad7dcf337a0db50f7d9483f310cd292e92e
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/licenses/official-license-sources.json — 1757 bytes; SHA256 276fdc32a6f75a127bfdcf9eb3b2b6c3c6135d3fa313f1810f71ba575f9af4a3
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/model_index.json — 446 bytes; SHA256 51a76cb1cf3ed37423a1128c79c22faee8e6fbe7f5aaeb737f0a258930dbaac0
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/scheduler/scheduler_config.json — 486 bytes; SHA256 067afb012cef64553a763447d1efd93daeffcc0123ca7e25b09f8de20b90762e
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/text_encoder/config.json — 1536 bytes; SHA256 214b4c29a0d975e9fddf9994a5673f22cb2c4c5750352f9227c2c3251ebeab40
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/text_encoder/generation_config.json — 214 bytes; SHA256 4347b1aeed2b2b78bc059920a0b7f5fec71482e1344952b76d7665d638d71f13
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/text_encoder/model.safetensors.index.json — 32855 bytes; SHA256 06b3d5319b6d76d1a4a2433419180016cfd54ed62d086a5e6567a809f8c82634
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer/added_tokens.json — 707 bytes; SHA256 c0284b582e14987fbd3d5a2cb2bd139084371ed9acbae488829a1c900833c680
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer/chat_template.jinja — 4168 bytes; SHA256 a55ee1b1660128b7098723e0abcd92caa0788061051c62d51cbe87d9cf1974d8
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer/merges.txt — 1671853 bytes; SHA256 8831e4f1a044471340f7c0a83d7bd71306a5b867e95fd870f74d0c5308a904d5
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer/special_tokens_map.json — 613 bytes; SHA256 76862e765266b85aa9459767e33cbaf13970f327a0e88d1c65846c2ddd3a1ecd
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer/tokenizer_config.json — 5404 bytes; SHA256 443bfa629eb16387a12edbf92a76f6a6f10b2af3b53d87ba1550adfcf45f7fa0
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/tokenizer/vocab.json — 2776833 bytes; SHA256 ca10d7e9fb3ed18575dd1e277a2579c16d108e32f27439684afa0e10b1440910
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/transformer/config.json — 541 bytes; SHA256 09733c74a3da6d17dd0a0472a091a8950c7c6935889c32c16cc800ede05029de
- gs://p27082-ds-istek/media-models/black-forest-labs/FLUX.2-klein-4B/e7b7dc27f91deacad38e78976d1f2b499d76a294/vae/config.json — 821 bytes; SHA256 0d6dfb69ae95a5e2ac9836284bbb63d8b38ce67b25ba2dff380752b2a10ab948

### Restrictions / remaining gaps
Schnell still BLOCKED (prior HF 401), not uploaded, not retried or bypassed. Tokenizer Pass 2b LFS reconciliation remains valid; final tokenizer.json uploaded, never re-quarantined. No .partial, unrelated workspace files, ready/public/served marker or verification JSON uploaded. Verification evidence is local only. No customer data used. GPU generation, model load, VRAM, latency, supervisor execution and deployment remain NOT TESTED.

### Execution note
Initial CLI invocation combined --no-clobber and --if-generation-match=0; gcloud rejected the combination before any write. Corrected to atomic --if-generation-match=0 alone. Final upload completed with zero object failures. gcloud used parallel composite uploads for four weight objects (component_count/CRC32C metadata recorded); no manual object deletion command was issued.

### Pass 3 evidence
- image_stage_evidence/upload-local-manifest.json
- image_stage_evidence/upload-results.json
- image_stage_evidence/upload-key-verification.json
- image_stage_evidence/upload-local-gcloud-hashes.json
- image_stage_evidence/upload-prefix-before.txt and upload-prefix-after.json
- image_stage_evidence/upload-bucket.json
- image_stage_evidence/ds-a100-upload-{start,end}.json
- image_stage_evidence/gcs-*.json (per-object metadata)
- image_stage_evidence/upload-initial-cli-error.txt


## Pass 4 — CPU adapter/lock ONLY (live-unbound; no GPU test claim)

**Time:** 2026-11-08 04:49:43 TSİ (Europe/Istanbul, UTC+3).
**Scope:** local box adapter + partial dependency lock only. This pass did NOT open,
start/stop/query any GPU/VM, use SSH/BOOT6, bind live services, deploy, write GCS,
or write any ready/publication/served marker. GPU status was NOT queried in this
pass; any separate GPU worker/report is outside this result. Policy constants
`daily_runtime=3600`, `max_uptime=1800`, `idle=600` unchanged. No Go `/v1/image`
handler was edited.

Inputs verified before use:
- Contract r2 SHA256 `e4ee10ce930bd00b2da7e62de5c57d8844f732314ca7ecf93b7538b21b65e8ce`.
- Plan SHA256 `6d44566c9c0f20679d3c7a7f0d95b1a625a602a95eea9e13da1d4053b1d70a38`.
- Original stage input SHA256 `fec7ddf6fdb19a69a39e0479fbbccf109b8f774d251ce5e328e1186f034cb088`.
- Stage immediately before this APPEND: SHA256 `fec7ddf6fdb19a69a39e0479fbbccf109b8f774d251ce5e328e1186f034cb088`.
  All preexisting bytes preserved; Pass 2b and Pass 3 are not rewritten.

### Implemented locally / CPU evidence

- `SandboxChatAdapter`: admission pause, active drain through injected bounded
  sandbox callback, pending priority, stop/reload/readiness/resume and synthetic
  smoke. Active chat is never forcibly cut off; no systemd or worker discovery.
- `SandboxLifecycleAdapter`: atomic in-memory fenced reserve/heartbeat/release,
  cancellation and operator-code alerts. TTL cannot resurrect; work deadline
  cannot extend; duplicate cleanup cannot charge twice. NOT a payment service,
  NOT live license SQL; crash durability and live bridge NOT TESTED / unbound.
- `LocalStagingAdapter`: approved private prefix provenance + full read-only local
  tree verification using an external manifest derived from prior verified upload
  evidence. Actual complete SHA256 verification: **24 files / 15,980,183,023 bytes**.
  Private GCS staging from Pass 3 remains current; no additional GCS operation.
- `PinnedTokenizerPromptAdapter`: lazy CPU-only local tokenizer loading; seven
  sibling hashes checked, padding/truncation disabled. Raw diagnostic counts:
  English `A red cat on a blue chair.` = **8 tokens**; Turkish
  `İstanbul’da yağmurlu bir akşam, sulu boya.` = **16 tokens**.
  Method: `tokenizers==0.22.0`, `Tokenizer.from_file`, `encode(add_special_tokens=True)`;
  these are NOT pipeline-formatted counts or quality measurements. Official wheel
  installed `--no-deps` only in `/workspace/tmp/kan-image-tokenizer-cpu`, not system/chat env.
- Prompt admission cap remains **UNKNOWN / NOT TESTED / fail-closed**. No
  `max_sequence_length` key found in pinned snapshot JSON. Encoder positions
  `40960` and tokenizer `model_max_length=131072` are NOT admission caps.
  Released diffusers 0.37.0 pipeline source has default 512 with chat-template
  formatting and truncation; that candidate default is documented, **NOT adopted**
  as a pinned-snapshot admission cap. Reviewed formatter + limit remain gates.
- Existing fail-closed NVML unknown=not-empty behavior unchanged. No live probe.

### Tests (FAIL/SKIP/NOT TESTED are not promoted to PASS)

Command: `python3 -m unittest discover -s tests -v`, `CUDA_VISIBLE_DEVICES` empty.
- Isolated tokenizer CPU environment: **106 passed / 0 failed / 0 errors / 0 skipped**.
- Default interpreter without optional tokenizers: **105 passed / 0 failed / 0 errors / 1 skipped**
  (real-tokenizer test only; skip reason in its log).
- Prior 70 tests passed unchanged; 36 focused adapter/lock tests added, including
  sandbox LocalBackend + Supervisor wiring with synthetic media and fake NVML.
- Real diffusion load/generation, CUDA, GPU compatibility, NVML handoff, live chat
  reload/smoke, live license lifecycle, and production HTTP binding: **NOT TESTED**.

### Dependency evidence lock — PARTIAL / production install BLOCKED

- **PINNED sourced:** transformers **4.56.1** (pinned encoder config + official PyPI
  wheel SHA256), diffusers **0.37.0** (official released wheel inspected as zip;
  contains `Flux2KleinPipeline`; wheel SHA256), tokenizers **0.22.0** (sourced
  transformers metadata lower bound + official manylinux wheel SHA256).
- Snapshot `_diffusers_version=0.37.0.dev0` is **DEV-ONLY provenance**. Released
  0.37.0 is a sourced successor candidate, NOT a verified runtime-compatibility
  claim. No floating git main is locked or installed.
- **UNKNOWN:** exact torch, torchvision, CUDA variant, pynvml, Python/platform ABI,
  and remaining transitive package pins. Official ranges retained as constraints,
  NOT invented exact versions. Upstream FLUX.2-dev CUDA test environment is not
  borrowed as a klein runtime pin.
- **Schnell remains BLOCKED:** HF 401 not bypassed, no token/credential search,
  download, or fallback. Adapter rejects absent snapshot even with permission flags.

### Changed package files and SHA256

| Path (under `/workspace/us-panel/kan/image-engine/`) | SHA256 |
|---|---|
| `image_engine/adapters.py` | `3f65f130fa4341c937140ca24f52a36e78df2262ec7f26efdcc526c45a0b2658` |
| `image_engine/prompt.py` | `ea09a72038e5f04a7ebda1abc7ddbca1e817c5ebe81c7fe5c9d8e281b630a703` |
| `tests/test_adapters.py` | `fb39c88c795d79f36941f0b650dab982ef64a729c657434422c335d9e660b636` |
| `dependencies.lock.json` | `9c8c0387dad55914c7ba7e81bcbebceb6f9afe0ba6c2f9d678655d449e9b1252` |
| `klein.snapshot.manifest.json` | `3874f2f513e66fe4b308f84aab4a48d4511733d4969f5d6b7c21883a70444a21` |
| `README.md` | `3654bd92a9f433fe3a503a5cb539a9b0f50007c7294882e523711f477d694db1` |
| `cpu-tests.log` | `ac8a3449b4aea3f59ccdba59a34704f6263dfc3819002d5debd02cd6c849698a` |
| `cpu-test-results.json` | `1fc37e8d36067854b4cd07fda7d69e02166ab1281127d2e8133bd55955f5f0d1` |
| `ARTIFACT_SHA256.json` | `60983e788a05f39ac2fd14d376ad628d3ed286a9068d016e1573fc071df35de1` |

### Evidence files and SHA256

- `/workspace/us-panel/reports/r20261108/image_stage_evidence/adapter-lock-cpu-evidence.json` — `078bc09f306c331cc61444e72207d6121fbf6d17a87a27a74e7eaa63febc7781`
- `/workspace/us-panel/reports/r20261108/image_stage_evidence/adapter-lock-source-sha256.json` — `a30298107b9b3d3793b067471606eead0cdaa35e2beffd6f11244ad830cbcbf2`
- `/workspace/us-panel/reports/r20261108/image_stage_evidence/adapter-default-interpreter-tests.log` — `2545cb8120050a7c7cacc83fcb4e95b4f26b34b908173883500a212baf803b34`

### Remaining gates / no readiness claim

Complete exact transitive + platform/CUDA wheel lock and review; pinned pipeline
formatter/token admission and no silent truncation; immutable local snapshot /
external trust-anchor configuration; live fenced chat restart/tunnel + durable
shared lifecycle integration and crash/operator reconciliation; US HTTP/ledger/
output wiring and its product gates; separately authorized GPU handoff/inference
measurements. No deploy or commercial/GPU readiness claim is made.

Tokenizer LFS reconciliation is preserved: git oid
`cd71f61a15a522601badb3dc960d800d9cb3766c` identifies the **133-byte LFS pointer**;
the **11,422,654-byte content** SHA256 is
`aeb13307a71acd8fe81861d94ad54ab689df773318809eed3cbe794b4492dae4`.
