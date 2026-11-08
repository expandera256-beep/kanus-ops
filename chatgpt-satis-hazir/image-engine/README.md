# KAN image-engine supervisor — IMPLEMENTED LOCALLY / NOT DEPLOYED

GPU was not opened. No GPU/VM start/stop, SSH, GCS upload, Cloud Run deployment,
systemd/timer, policy, secret, firewall or live-service change was made.
Schnell remains gated/BLOCKED; no download or token search was attempted.

## Run CPU tests (stdlib suite; optional local tokenizer diagnostic)

```sh
cd /workspace/us-panel/kan/image-engine
python3 -m unittest discover -s tests -v
```

These tests use temporary local files and synthetic adapters/output references.
No torch/diffusers/live NVML is imported; worker tests use mocked providers.
The optional real-tokenizer CPU test skips honestly if `tokenizers` is absent.
For this pass only, official tokenizers 0.22.0 was installed with `--no-deps` into
`/workspace/tmp/kan-image-tokenizer-cpu` (not the system/chat environment):
`CUDA_VISIBLE_DEVICES= PYTHONPATH=/workspace/tmp/kan-image-tokenizer-cpu python3 -m unittest discover -s tests -v`.
No CUDA or diffusion stack was installed.
There is no auto-start CLI, HTTP binding, background daemon or installation hook.

## Package

- `image_engine/supervisor.py`: explicit state machine, chat drain/priority,
  serial admission and generation, bounded work TTL, absolute deadline,
  pre-reserved encode/reload/stop budget, safe restore and durable fail-closed
  crash barrier. Restore failure rejects all subsequent image jobs; no new boot.
- `ownership.py`: Linux `flock` held for the supervisor lifetime, durable/fsynced
  monotonic generation, kernel boot-id + PID + `/proc` process-start identity.
  Each operation checks the fence and the original lock inode. File existence or
  PID existence alone is never authority. Releasing revokes worker authorization.
- `ledger.py`: SQLite FULL/WAL idempotency journal, at most one local terminal
  settlement, payload mismatch rejection, no automatic replay after a crash.
  This is NOT a payment/refund service or replacement for the existing shared
  lifecycle store. Shared reservation/lease release must be atomic and fenced.
- `models.py`: exact model/revision allowlist, sizes 1024x1024 / 1344x768 /
  768x1344, integer n=1..4, pinned-tokenizer prompt admission callback, immutable
  externally anchored SHA256 manifest verification of every local file/shard.
- `process.py`: explicit separate-venv worker, credential-free offline environment,
  bounded IPC, owned process-group TERM/KILL with boot/PID/start validation.
- `worker.py`: second snapshot verification; local path only,
  `local_files_only=True`, `trust_remote_code=False`, BF16, four steps,
  klein guidance 1 / schnell guidance 0, exactly one image per pipeline call.
  PNG encoding/fsync precedes the response. No runtime Hugging Face download.
- `integration.py`: concrete local subprocess backend with injected, reviewed
  chat/lifecycle adapters; fail-closed physical-GPU UUID NVML context probe.
  All adapters remain unbound to live services.
- `adapters.py`: injectable `SandboxChatAdapter`, `SandboxLifecycleAdapter`, and
  read-only `LocalStagingAdapter`. No process/service discovery or cloud writes.
  Sandbox lifecycle is in-memory, atomically fenced, deadline/TTL bounded and
  non-payment; restart durability and the live shared store are NOT implemented.
- `prompt.py`: lazy pinned `tokenizers` loading with all seven tokenizer sibling
  hashes checked against an external manifest. Counts have truncation/padding
  disabled. Raw diagnostics are NOT formatted pipeline admission counts.
  Admission fails closed until both sourced limit and reviewed formatter exist.
- `klein.snapshot.manifest.json`: external runtime-schema manifest derived from
  prior `image_stage_evidence/upload-local-manifest.json`, preserving all 24
  staged files (15,980,183,023 bytes), including the historical inner manifest.
  Its hash belongs in reviewed host configuration, not client input.
  `LocalStagingAdapter.verified_snapshot()` performs complete local SHA256 reads;
  `staged_location()` returns provenance only, not ready/served status.
  Schnell has no local staged snapshot and remains rejected even if sandbox
  permission flags are true.

## Host contract and safety

1. Host creates one `Ownership` on a private, persistent LOCAL directory. Never
   unlink/replace its lock file or roll back its fencing state. All chat and
   media CUDA starts must pass the same supervisor authority; independent
   systemd auto-restart is not safe. The supplied chat adapter must suppress
   restart/tunnel races, drain without interrupting existing chat, then resume.
2. Host injects `Ledger`, reviewed immutable `Snapshot` manifest/hash anchors,
   pinned-tokenizer validator, existing lifecycle adapters and `Budget` from
   authoritative runtime balances. Never trust a client to supply permissions,
   budgets, file paths, UUID, output directory, manifest hash or fallback reason.
3. Start only acquires local authority and checks existing chat readiness. It
   never wakes a VM, loads a model or restores chat. `health`, `ready`,
   `list_models`, `poll` have no adapter side effects, leases or idle refresh.
4. `daily_runtime=3600`, `max_uptime=1800`, `idle=600` are constants, not policy
   updates. Admission reserves drain/load/inference/encode/reload/stop before
   work. Estimates are conservative configuration, NOT measured guarantees.
   Shared lifecycle must enforce balances atomically. Lease deadline is absolute;
   TTL/heartbeat loss cannot resurrect work. Callbacks must honor deadlines and
   progress at <=20 seconds; supervisor cannot make an arbitrary blocking chat
   adapter bounded by itself. Cleanup has separately reserved time.
5. A switch is
   `CHAT_READY -> CHAT_DRAINING -> IMAGE_LOADING -> IMAGE_READY -> RUNNING ->
   IMAGE_READY -> IMAGE_DRAINING -> CHAT_RELOADING -> CHAT_READY`.
   Chat is rechecked before loading and at every serial image boundary (even
   inside n=4). Active chat is never forcibly cut off. Failure stops only the
   owned media group, proves compute AND graphics CUDA contexts absent via NVML,
   reloads the EXISTING local NVMe chat pack, checks readiness and synthetic smoke.
   Unknown context state or failed restore leaves a persistent FAILED_CLOSED
   barrier and operator alert; restart does not clear it or initiate a boot.
6. Schnell can be explicitly selected ONLY with fallback permission, authorized
   gated access AND a verified local snapshot. No automatic fallback or retry
   exists, so klein OOM never selects schnell. One model/attempt is associated
   with one idempotency reservation. Same-key retries replay the terminal result;
   changing a key's model/payload is rejected, preventing a second settlement.
   A future cross-model retry under the same reservation needs an explicit
   reviewed product contract; the package deliberately refuses such replay.
7. Snapshot manifest format:
   `{"model":"repo", "revision":"sha", "files":{"path":{"bytes":N,
   "sha256":"hex"}}}`. Keep the trust anchor outside the snapshot. Every file
   must be listed, including licenses; symlinks, extra/partial files, missing
   required configs/shards and mismatched hashes fail closed. Snapshot storage
   must be immutable/read-only to workers to prevent verify/load TOCTOU. An
   arbitrary writable manifest is NOT trust. No files are renamed/quarantined.
8. Local outputs are private; this package neither uploads nor applies public
   ACLs or retention/deletion rules. External durable output/partial refund
   semantics and cancellation policy still require US integration review.

## Deployment gates — still BLOCKED / NOT TESTED

- `dependencies.lock.json` is a PARTIAL evidence lock, NOT production-installable.
  transformers 4.56.1 is sourced from the pinned encoder config; released diffusers
  0.37.0 is a candidate with Flux2KleinPipeline verified in the official wheel
  archive, not imported. Both official PyPI wheel SHA256s are recorded. Snapshot
  `_diffusers_version=0.37.0.dev0` is DEV provenance, not a floating git lock.
  tokenizers 0.22.0 is sourced from the transformers metadata lower bound and
  official wheel; isolated CPU-only diagnostics succeeded. Exact torch, torchvision,
  CUDA, pynvml, Python/platform ABI and remaining transitive pins are UNKNOWN.
  Sourced ranges are NOT exact pins. No GPU compatibility is established.
- Prompt admission remains UNKNOWN / BLOCKED. No `max_sequence_length` occurs in
  the pinned snapshot JSON. 40960 encoder positions and 131072 tokenizer max length
  are NOT admission caps. Released diffusers 0.37.0 source defaults to 512 and
  uses chat-template formatting + truncation; this is recorded only as candidate
  source evidence, NOT adopted as a validated cap for the pinned snapshot.
  The host must review/anchor formatter + limit and prove no silent truncation.
- Live chat restart/tunnel coordination and shared lifecycle bridge are supplied
  sandbox implementations plus contracts, NOT installed integrations. Crash/operator reconciliation needs an
  approved runbook; there is no automatic unblocking method.
- No real `from_pretrained`, CUDA inference, output image, NVML handoff,
  local-chat reload, VRAM/latency or GPU smoke was run. Mocked CPU tests do not
  establish GPU compatibility, 30-second latency, or commercial readiness.
- BOOT6 remains separate. No timer masking, new boot or limit increase is present.

Plan: `/workspace/us-panel/reports/r20261108/KAN_IMAGE_ENGINE_PLAN.md` §5.
Tokenizer LFS reconciliation belongs to the stage report and is preserved;
`tokenizer.json` was not re-quarantined or renamed.

## CPU adapter/lock pass scope

Private klein staging already exists at the approved GCS prefix (24 objects);
this pass performed NO GCS read/write, upload, publication or ready-marker action.
The tokenizer git oid is the 133-byte LFS pointer oid, not a content hash; the
11,422,654-byte local tokenizer SHA256 is
`aeb13307a71acd8fe81861d94ad54ab689df773318809eed3cbe794b4492dae4`.
Raw diagnostic counts and source hashes are in
`/workspace/us-panel/reports/r20261108/image_stage_evidence/adapter-lock-cpu-evidence.json`.
No Go `/v1/image` handler was edited. GPU untouched by this pass; VM status was
not queried. Any separately authorized GPU task/report is outside this result.
