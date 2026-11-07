# LOCAL ONLY — remote runner safety

{
  "data_source": "LOCAL_FIXTURE_ONLY",
  "remote_requests_sent": 0,
  "remote_dns_lookups": 0,
  "checks": [
    {
      "name": "remote_dry_run_sockets_dns_and_credentials_forbidden",
      "pass": true
    },
    {
      "name": "reject_cap21",
      "pass": true
    },
    {
      "name": "reject_level25",
      "pass": true
    },
    {
      "name": "reject_duration601",
      "pass": true
    },
    {
      "name": "normal_domain_HTTPS_SNI_no_IP_pin",
      "pass": true
    },
    {
      "name": "blocked_consuming_route_/v1/chat/completions",
      "pass": true
    },
    {
      "name": "blocked_consuming_route_/v1/credits/reserve",
      "pass": true
    },
    {
      "name": "blocked_consuming_route_/v1/credits/settle",
      "pass": true
    },
    {
      "name": "blocked_consuming_route_/account/register",
      "pass": true
    },
    {
      "name": "blocked_consuming_route_/account/redeem",
      "pass": true
    },
    {
      "name": "blocked_consuming_route_/webhooks/shopify/orders-paid",
      "pass": true
    },
    {
      "name": "local_proxy198_diagnostic_not_gate",
      "pass": true
    },
    {
      "name": "mock_dual_DoH_both_correct",
      "pass": true
    },
    {
      "name": "mock_dual_DoH_cloudflare_wrong",
      "pass": true
    },
    {
      "name": "mock_dual_DoH_google_wrong",
      "pass": true
    },
    {
      "name": "429_expected_not_unexpected_error",
      "pass": true
    },
    {
      "name": "account_mode600_load",
      "pass": true
    },
    {
      "name": "account_mode644_rejected",
      "pass": true
    },
    {
      "name": "account_missing_clear_error",
      "pass": true
    },
    {
      "name": "loopback_login_credit_scenarios",
      "pass": true,
      "exit_code": 0
    },
    {
      "name": "all_local_status_200_and_zero_debit",
      "pass": true
    },
    {
      "name": "transport_peak_lte20",
      "pass": true,
      "observed_peak": 20
    },
    {
      "name": "hard_process_cutoff_SIGALRM",
      "pass": true,
      "exit_code": 124,
      "wall_seconds_including_startup": 0.3226621130006606,
      "configured_seconds": 0.25
    }
  ],
  "pass": true
}
