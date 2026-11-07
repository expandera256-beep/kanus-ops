# Güncel runbook — policy v5 ve kademeli fazlar

Demir8Kasım2026 00:31TSİ doğrudan onayı. Idle8Kasım00:20TSİpolicyv5=600sn. Hedefhttps://api-test.kanus.org:443; normalDNS/SNI/Hostaynıalanadı, IPpin yok. HOSTbeyanı binary7d72f449…56f0, cutoverraporu§15.

```bash
python3 /workspace/us-panel/perf/remote_bench.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --dry-run
python3 /workspace/us-panel/perf/remote_selftest.py
python3 /workspace/us-panel/perf/remote_phase_unit.py
```

Onaylıkoşugirişimi:
```bash
python3 /workspace/us-panel/perf/remote_phases.py --config /workspace/us-panel/reports/r20261108/remote_82_config.json --execute-approved-phases
```
Bu girişimDNS198.18.0.1nedeniyleSTOPPED; tekrar yapılmadı. Beklenen82.29.184.225olmadanHTTPyok.31hedefi yok. Ortak600sn,herfazmax300sn,20concurrency;seviyeöncesihealth,seviyearası≥3sn. FAZ1yalnızGEThealthz/creditsstatus;FAZ2ancaktemizsonrasılogin,successful-onlylatency;429ayrıexpected. Hiçbirmodel/GPU/register/webhook/creditmutationçağrısı yok. Health/transport/bakiyeazalma/süre/5xxve>%50unexpectedhataSTOP.

TokenveQAhesapyalnızRAM;QAhesap yolu/workspace/secrets/perf_login_account.json exact600. İçerikloglanmaz. Markerbirengeldeğil;HOSTbeyanı olarakraporlanır. GerçekDNSgirişiminde token/QAhesap dosyalarıokunmadı.
