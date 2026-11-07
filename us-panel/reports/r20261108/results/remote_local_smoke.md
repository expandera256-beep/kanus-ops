# LOCAL_FIXTURE — login + READ-ONLY credit load

| Scenario | Users | n | p50/p95/p99 ms | Success% | Expected429% | Unexpected error% | Credits expected/actual |
|---|---:|---:|---|---:|---:|---:|---|
| login | 1 | 1 | 31.292/31.292/31.292 | 100.0 | 0.0 | 0.0 | 0/0 |
| login | 5 | 5 | 31.899/31.931/31.934 | 100.0 | 0.0 | 0.0 | 0/0 |
| login | 10 | 10 | 32.097/32.781/32.797 | 100.0 | 0.0 | 0.0 | 0/0 |
| login | 20 | 20 | 32.379/34.075/34.284 | 100.0 | 0.0 | 0.0 | 0/0 |
| credit_status | 1 | 1 | 31.246/31.246/31.246 | 100.0 | 0.0 | 0.0 | 0/0 |
| credit_status | 5 | 5 | 31.157/31.309/31.326 | 100.0 | 0.0 | 0.0 | 0/0 |
| credit_status | 10 | 10 | 32.182/32.464/32.474 | 100.0 | 0.0 | 0.0 | 0/0 |
| credit_status | 20 | 20 | 32.789/33.799/34.451 | 100.0 | 0.0 | 0.0 | 0/0 |

## Comparable workload table (429 separately expected)
| Scenario | Users | Current box p50/p95 ms | Old test VPS | api-test remote p50/p95 ms | Status |
|---|---:|---|---|---|---|
| login | 1 | 226.671/280.938 | NOT MEASURED | NOT MEASURED | NOT RUN |
| login | 5 | 612.044/1195.376 | NOT MEASURED | NOT MEASURED | NOT RUN |
| login | 10 | 658.54/1103.099 | NOT MEASURED | NOT MEASURED | NOT RUN |
| login | 20 | 1266.715/2188.818 | NOT MEASURED | NOT MEASURED | NOT RUN |
| login | 25 | NOT MEASURED | NOT MEASURED | NOT MEASURED | BLOCKED approval cap20 |
| login | 50 | NOT MEASURED | NOT MEASURED | NOT MEASURED | BLOCKED approval cap20 |
| credit_status | 1 | 0.241/0.427 | NOT MEASURED | NOT MEASURED | NOT RUN |
| credit_status | 5 | 0.58/2.189 | NOT MEASURED | NOT MEASURED | NOT RUN |
| credit_status | 10 | 1.111/1.605 | NOT MEASURED | NOT MEASURED | NOT RUN |
| credit_status | 20 | 1.738/2.471 | NOT MEASURED | NOT MEASURED | NOT RUN |
| credit_status | 25 | NOT MEASURED | NOT MEASURED | NOT MEASURED | BLOCKED approval cap20 |
| credit_status | 50 | NOT MEASURED | NOT MEASURED | NOT MEASURED | BLOCKED approval cap20 |

Old VPS login/credit measurements were not supplied; previous health TTFB is not comparable. Box login isolates IP rate limits; remote never spoofs trusted-proxy headers. Read-only credit status is not model debit throughput.
