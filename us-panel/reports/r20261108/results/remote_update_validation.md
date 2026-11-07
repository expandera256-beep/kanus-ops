# DNS STOP / final safety validation

{
  "timestamp": "2026-11-08T00:41:27.877626+03:00",
  "target": "https://api-test.kanus.org:443",
  "actual_remote_attempt": {
    "DNS": {
      "addresses": [
        "198.18.0.1"
      ],
      "expected": "82.29.184.225",
      "elapsed_ms": 2.134217999810062,
      "pass": false
    },
    "elapsed_seconds": 0.002177548999497958,
    "HTTPS_requests": 0,
    "phase1": "STOPPED_DNS_MISMATCH",
    "phase2": "BLOCKED_PHASE1_NOT_CLEAN",
    "balance_before": null,
    "balance_after": null,
    "service_health": "NOT_TESTED",
    "429_rate": null,
    "5xx_rate": null,
    "QA_account_file_read": false,
    "real_license_token_file_read": false
  },
  "local_safety_checks_passed": 22,
  "phase_control_checks_passed": 7,
  "idle_checks_passed": 7,
  "idle_seconds": 600,
  "idle_policy": "v5",
  "remote_approval": "Demir2026-11-08 00:31TSİ",
  "idle_decision": "Demir2026-11-08 00:20TSİ",
  "HOST_binary_declaration": "7d72f449…56f0",
  "source": "HOOK_MOVE_PROD_VPS_20261106.md §15"
}
