# Local phase control tests

{
  "data_source": "LOCAL_CONTROL_FLOW_SIMULATION",
  "remote_requests": 0,
  "dns_lookups": 0,
  "real_token_or_QA_file_reads": 0,
  "checks": [
    {
      "name": "429_expected_no_abort",
      "pass": true
    },
    {
      "name": "exact50pct_not_abort",
      "pass": true
    },
    {
      "name": "over50pct_abort",
      "pass": true
    },
    {
      "name": "consecutive_5xx_abort",
      "pass": true
    },
    {
      "name": "health_failure",
      "pass": true,
      "status": "STOPPED",
      "stop_reason": "HEALTH_FAILED: HTTP503",
      "QA_file_loader_calls_simulated": 0
    },
    {
      "name": "balance_decrease",
      "pass": true,
      "status": "STOPPED",
      "stop_reason": "CREDIT_DECREASED",
      "QA_file_loader_calls_simulated": 0
    },
    {
      "name": "clean_both_phases",
      "pass": true,
      "status": "COMPLETED",
      "stop_reason": null,
      "QA_file_loader_calls_simulated": 1
    },
    {
      "name": "shared600_budget_exhausted_before_DNS",
      "pass": true
    },
    {
      "name": "shared_budget_caps_both_phase_timers",
      "pass": true,
      "timer_values_seconds": [
        2,
        1.9938559459997123,
        0
      ]
    }
  ],
  "pass": true
}
