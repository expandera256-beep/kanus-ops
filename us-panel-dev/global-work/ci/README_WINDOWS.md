# GLOBAL i18n Windows tests — how to run (local, no upload)
Files in this folder: `i18n_windows_amd64.test.exe`, `run_windows_tests.ps1`, `SHA256SUMS.txt` (the script checks the hash and stops on mismatch).

cmd.exe (echo must be a SEPARATE command; `... & echo %ERRORLEVEL%` on the same line shows 0 because cmd expands %VAR% when the line is read):
```
powershell -NoProfile -ExecutionPolicy Bypass -File .\run_windows_tests.ps1
echo %ERRORLEVEL%
```
cmd.exe, one line (delayed expansion):
```
cmd /v:on /c "powershell -NoProfile -ExecutionPolicy Bypass -File .\run_windows_tests.ps1 & echo !ERRORLEVEL!"
```
PowerShell:
```
powershell -NoProfile -ExecutionPolicy Bypass -File .\run_windows_tests.ps1; $LASTEXITCODE
```
Exit codes: 0 all PASS · 1 a test FAILED · 2 PARTIAL (only TestWinPrefSymlinkFileRejected may SKIP without SeCreateSymbolicLinkPrivilege / Developer Mode) · 3 hash mismatch or setup error.
Expected without symlink privilege: 23 PASS / 0 FAIL / 1 SKIP, `RESULT exit=2`.
The runner also writes `RESULT exit=N` to `<LogDir>\result.txt` (script sha256 32c3034e…; older runs: console only).
