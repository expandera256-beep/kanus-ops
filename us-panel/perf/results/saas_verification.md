# MOCK — SaaS verifier

```json
{
  "source": "MOCK",
  "not_real_saas_acceptance": true,
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
      "stderr": "test_health (test_backend.TestHealth.test_health) ... ok\n\n----------------------------------------------------------------------\nRan 1 test in 0.000s\n\nOK\n"
    }
  ],
  "pass": true
}
```
