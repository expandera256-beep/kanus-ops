# Accounts-mode security QA (api-test only)

```bash
python3 /workspace/us-panel/qa/accounts/run.py
```

Runtime-only files: acct-test-credentials.json, acct-test-password2.txt, acct-test-license.txt, prior revoked qa-license-a/b.txt under /workspace/secrets. Never reads webhook/session signing secrets; never creates orders/accounts; never calls reset/verify/mail/model endpoints. Issued sessions remain in RAM only. No password/token/license values in stdout, reports, URL or recorded bodies.

Hard caps:60 total requests,15 login requests; sequential except2-parallel same-owner redeem;120s per-request timeout. First network exception or5xx stops all further requests, no retries. Limits are not raised. Rotating XFF uses two documentation-range fake addresses on the SAME service, never another network/origin IP. Token tests and redeems run before the shared-IP limiter test so later429 does not invalidate earlier results.

Phases:
1. Two real logins and balances.
2. Tampered signature; changed exp with invalid signature (not proof of valid signed expiry); changed account subject; A token plus B email must retain A scoped balance/history.
3. Existing valid owner1 license idempotent double redeem and parallel2 repeat; owner2 reject; revoked A/B reject; balances unchanged. Fresh first-claim race requires coordinator-provided unbound synthetic test license, not webhook access.
4. Login enumeration:2 interleaved samples/class (known QA vs nonexistent example.invalid), generic status/body fingerprint comparison. Timing samples only; heavy-load noise means no side-channel absence/existence claim.
5. A few wrong-password attempts on QA1 only, one correct-password observation; bounded short-invalid-password probes to reach429 cheaply; then two rotated-XFF probes and an unchanged-header control. Per-IP429 is NOT proof of per-account lockout.

Logs/results: reports/qa-20261106/accounts-security.log, accounts-security-events.json, accounts-security-results.json. Report: accounts-security.md; deploy report receives an appended QA section, not an overwrite.

WARNING: This routine can temporarily limit the shared test IP for15minutes. It does not lock/reset/delete accounts or touch server files/services. Do not run concurrently with other account-login QA routines. No automatic repeat on failure/timeouts.

## Koordinatörün açık5-paralel istisnası — tek takip koşumu
```bash
python3 /workspace/us-panel/qa/accounts/burst_probe.py --mode baseline
# Sadece ilk sonuç inconclusive ise ve en az60s sonra; en fazla bir kez:
python3 /workspace/us-panel/qa/accounts/burst_probe.py --mode baseline --repeat-baseline
# Son burst bittikten en az60s sonra:
python3 /workspace/us-panel/qa/accounts/burst_probe.py --mode spoof
```
Her burst tam5 istek,5 worker, barrier ile birlikte başlatılır. Parola bilerek kısa-invaliddır: allowed requests400le bcrypti çalıştırmadan çıkar; nginx rate-limit body validationdan bağımsız olduğu için frontend limiter yüzeyi yine denenir. Bu auth/brute-force KDF performans testi değildir. Spoof her istekte farklı TEST-NET IPv4ü hem X-Forwarded-For hem X-Real-IPye koyar.120s timeout, herhangi5xx/network error sonrası sonraki burstler durur. Script önceki burst bitiminden60s dolmadan ağ çağrısı yapmaz ve aynı adlandırılmış burstü yeniden koşmaz. Baseline429 yoksa bypass var/yok iddiası kurulmaz; kaynakIP bucketı bağımsız görülmemiştir.

### Fresh ilk-claim yarışı
```bash
# Tüm burstlerden sonra ve gerçek loginlerden önce en az60s cooldown:
python3 /workspace/us-panel/qa/accounts/fresh_claim.py
```
Yalnız `/workspace/secrets/acct-qa-fresh-license.txt` varsa çalışır; dosya yoksa HTTP yapmadan BLOCKED. Gerçek QA1/QA2 login, iki farklı subject, pre-balances, raw license valid/nonrevoked/unreserved kontrolü; tam2-paralel first redeem → bir200/diğer409; winner repeat200, loser409; source credits harcanmadan tek hesapta tek quota artışı ve reserved0 doğrulaması. Yeni key/order/webhook üretmez, model/e-posta/admin/VPS işlemi yapmaz. Race bir kez gerçekleştiyse otomatik tekrar etmez. Koordinatör taze test keyini sonradan revoke edebilir.
