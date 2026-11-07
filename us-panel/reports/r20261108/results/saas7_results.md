# MOCK — §7 contract fixture

{
  "data_source": "MOCK_FIXED_CONTRACT_NOT_MODEL_OUTPUT",
  "pass": true,
  "steps": [
    {
      "command": [
        "-m",
        "compileall",
        "-q",
        "."
      ],
      "exit_code": 0,
      "stdout": "",
      "stderr": ""
    },
    {
      "command": [
        "-m",
        "unittest",
        "discover",
        "-v"
      ],
      "exit_code": 0,
      "stdout": "",
      "stderr": "test_admin (test_backend.Contracts.test_admin) ... ok\ntest_auth (test_backend.Contracts.test_auth) ... ok\ntest_cross_tenant (test_backend.Contracts.test_cross_tenant) ... ok\ntest_project_task (test_backend.Contracts.test_project_task) ... ok\ntest_upload_cross_tenant (test_backend.Contracts.test_upload_cross_tenant) ... ok\ntest_upload_download (test_backend.Contracts.test_upload_download) ... ok\ntest_upload_limit (test_backend.Contracts.test_upload_limit) ... ok\ntest_upload_traversal (test_backend.Contracts.test_upload_traversal) ... ok\n\n----------------------------------------------------------------------\nRan 8 tests in 0.000s\n\nOK\n"
    }
  ],
  "fixture_file_manifest": [
    "backend.py",
    "test_backend.py",
    "frontend/index.html",
    "README.md"
  ],
  "mock_scope": "auth lookup/RBAC/project/task/tenant/upload contracts only; in-memory service, not real backend HTTP/UI/DB",
  "real_model_acceptance": "BLOCKED: .11 deployment + GPU window pending",
  "not_tested": [
    "PostgreSQL",
    "real frontend/browser",
    "actual backend HTTP integration",
    "Docker",
    "model-generated code",
    "agent repair cycle"
  ]
}
