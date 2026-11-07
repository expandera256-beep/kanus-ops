# REAL_NETWORK_NOT_GPU — düşük hızlı health ölçümü

Kaynak zamanı: 2026-11-06T09:31:17.942789+03:00 (TSİ)

| HTTP | DNS ms | TCP connect ms | TLS ms | HTTP TTFB ms | Toplam ms |
|---|---:|---:|---:|---:|---:|
| 200 | 1.263 | 0.245 | 409.921 | 1107.959 | 1108.067 |
| 200 | 1.402 | 0.258 | 465.809 | 863.709 | 863.761 |
| 200 | 1.125 | 0.285 | 496.227 | 1095.206 | 1095.261 |

Egress: `{"error": "curl_exit_35", "detail": "curl: (35) TLS connect error: error:0A000126:SSL routines::unexpected eof while reading"}`

- TCP connect is connect completion minus DNS, NOT pure RTT.
- TLS increment can include network/proxy overhead; no packet-level RTT inferred.
- Health HTTP TTFB is NOT LLM TTFT.
- Single box vantage does not measure US/Europe end-user latency.
- Egress region is ipinfo geolocation estimate, not verified hosting location.
- Egress lookup failed; no alternate network route used.
