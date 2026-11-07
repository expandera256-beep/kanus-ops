# Windows .6 smoke — BLOCKED native Windows

```powershell
$exe = 'C:\US\US_Ticari_Aday_2026_11_06_6.exe'
$license = 'C:\US\test-license.txt' # UTF-8 single line; key stays in this private file
$env:US_LICENSE_URL = 'https://api-test.kanus.org'
function Invoke-USSmoke {
    param([string[]]$ModeArgs)
    $env:US_SMOKE_LOG = Join-Path $env:TEMP ('US-smoke-' + [guid]::NewGuid().ToString('N') + '.log')
    $p = Start-Process -FilePath $exe -ArgumentList $ModeArgs -Wait -PassThru
    Get-Content -LiteralPath $env:US_SMOKE_LOG -Encoding UTF8
    "exit=$($p.ExitCode) log=$env:US_SMOKE_LOG"
}
Get-FileHash -LiteralPath $exe -Algorithm SHA256
Invoke-USSmoke -ModeArgs @('--version')
Invoke-USSmoke -ModeArgs @('--selftest')
Invoke-USSmoke -ModeArgs @('--check-server')
Invoke-USSmoke -ModeArgs @('--check-license', ('"{0}"' -f $license))
```

Beklenen çıktı/exit kodları: ROUND6_RAPOR_TR.md. US_SMOKE_LOG her çağrıda benzersiz; key yalnız private dosyadadır.
