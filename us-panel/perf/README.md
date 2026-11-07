# PERF scratch harness — current policy v5

Product sources unchanged. Standard-library Python; cached/offline Go builds. All local fixtures set US_GPU_WAKE_DISABLE=1 and SMTPoff. Never log token/password/session/hash values.

## Current policy
Idle600sn/10dk, policyv5, Demir8Kasım2026 00:20TSİ kararı. Older15dk records are historical and superseded. Active SaaSacceptance/docs/logicalclock now600sn; actualGPUstop nottested.

## Local reproduction
```bash
python3 r20261108_local.py
python3 isolation_credit_bench.py
python3 saas/mock_project_contract.py
python3 idle_contract.py
python3 remote_selftest.py
python3 remote_phase_unit.py
```
Dated outputs:/workspace/us-panel/reports/r20261108/. Localmodel1/5/10/25/50mock;login1/5/10/20twoCPUcost12;creditstatusread-only. FixturePASS is not realmodel/PostgreSQL/UIacceptance. Source/binaryhashes included.

## Approved remote conditional phases
Normalhttps://api-test.kanus.org:443; SNI/Hostsame; nofixedIP. EveryDNSanswer mustbe82.29.184.225 elseSTOPbeforeHTTP. Shared600sn;phase1≤300sn,phase2≤min(remaining,300). Max20;healthbeforelevels;3sbetweenlevels. Read-onlyphase1andconditionalQAloginphase2; nomodel/register/webhook/reserve/settle/redeem.429expectedseparately; successful-onlyloginpercentiles. Health/transport/balancedecrease/consecutive502-504/>50%unexpectederrorsSTOP.

```bash
python3 remote_bench.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --dry-run
python3 remote_phases.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --execute-approved-phases
```
ActualapprovedattemptSTOPPED:boxDNS198.18.0.1notexpected82;0HTTP;QAandtokenfilesnotread. Do notautoretryorIPpin. Markeroptional;identityisHOSTdeclaration7d72f449…56f0fromcutoverreport§15. Demirapproved8Kasım00:31TSİ. Noaccountdeletion/restartorconfigurationchange.

Loginaccountdefault/workspace/secrets/perf_login_account.json,regularfileexactmode600,JSONemail/password/synthetic:true; missingfileclearerror. Token/workspace/secrets/perf_api_token.txt. SecretvaluesonlyRAM. CurrentreportexplainsQAIP5r/m,burst2,email10failed/900s,androtatingegresslimitation.

### Son onaylı uzak faz protokolü (8 Kasım, TSİ)
`remote_phases.py` iki kamu DoH sağlayıcısında yalnız82.29.184.225 A kaydını doğrular. Yerel198.18.0.1 boxproxy yanıtı tanıdır; TCPconnect originRTT değildir. Normal domain/SNI/Host ve doğrulanan TLS kullanılır;IPpin/k yok. Herfaz≤300sn, paylaşılan toplam≤600sn ve max20. Önceki harcanan süre configte korunur. FAZ2yalnıztemizFAZ1+health200sonrasıKAN.USsentetikhesapla;register/model/webhookyasak. Son iki faz tamamlandı; otomatiktekrar yok. Güncelgerçeksonuçlar reports/r20261108/REMOTE_PHASES_REPORT.md, hamresults/remote_phases_execution_raw.json. Finalizer yalnızağsız rapor üretir.
