# Prompt requirements/differences

{
  "prompt": "/workspace/us-panel/prompt/KANUS_PROMPT_20261107.txt",
  "sha256": "3c71bea9f9bc9325d370bd667cc2706e0aec0ecfadae74c8d51db0f2e0963663",
  "old_to_new_sections": {
    "old11_real_project": "new7_KANUS_SaaS_Project_Manager",
    "old9_performance": "new11_load"
  },
  "new7_additions": [
    "explicit product name KANUS SaaS Project Manager",
    "file upload",
    "admin panel",
    "installation guide",
    "working application + logs + source files"
  ],
  "new11_metrics": [
    "p50/p95/p99",
    "success rate",
    "GPU memory",
    "queue time",
    "error rate",
    "credit consistency",
    "concurrent task isolation",
    "total cost"
  ],
  "removed_as_explicit_requirements_vs_old": [
    "old project section explicit CI/CD (still recommended)",
    "old load section GPU utilization and request cost now memory/total-cost emphasis"
  ],
  "numeric_latency_error_gpu_memory_SLO": null,
  "required_concurrency": [
    1,
    5,
    10,
    25,
    50
  ],
  "report_statuses": [
    "PASS",
    "FAIL",
    "BLOCKED"
  ],
  "report_fields": [
    "starting state",
    "changed files",
    "solution",
    "actual tests",
    "results",
    "remaining gaps",
    "needed approvals"
  ],
  "idle_note": "Both old and new prompt text say10min; intervening Demir operational decision was15min. Latest prompt acceptance target10min recorded, no live config changed; clarify before GPU run."
}
