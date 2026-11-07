# Local draft. Runs on an approved Windows amd64 computer; no uploads/install.
# Binary mode needs no Go/GCC; this cross-built binary is NOT race-enabled.
# Source mode needs Go + modern MinGW-w64 GCC and runs only ./i18n/ with -race.
[CmdletBinding()]
param(
  [string]$Binary = (Join-Path $PSScriptRoot 'i18n_windows_amd64.test.exe'),
  [string]$ModuleDir = '',
  [string]$LogDir = (Join-Path $PSScriptRoot 'windows-test-results')
)
$ErrorActionPreference = 'Stop'
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
$LogDir = (Resolve-Path -LiteralPath $LogDir).Path
if ($ModuleDir) {
  $env:CGO_ENABLED = '1'
  gcc --version
  if ($LASTEXITCODE -ne 0) { throw 'Race requires MinGW-w64 GCC; use Binary mode otherwise.' }
  $lib = (gcc --print-file-name libsynchronization.a).Trim()
  if ($lib -eq 'libsynchronization.a' -or !(Test-Path $lib)) { throw 'Unsupported MinGW-w64 runtime' }
  Push-Location -LiteralPath $ModuleDir
  try {
    $lines = @(go test -race -count=1 -json ./i18n/)
    $code = $LASTEXITCODE
    $lines | Set-Content -Encoding utf8 (Join-Path $LogDir 'go-test.jsonl')
    $events = @($lines | ForEach-Object { $_ | ConvertFrom-Json })
    $winPass = @($events | Where-Object { $_.Test -like 'TestWinPref*' -and $_.Action -eq 'pass' }).Count
    $winSkip = @($events | Where-Object { $_.Test -like 'TestWinPref*' -and $_.Action -eq 'skip' }).Count
    Write-Host "source exit=$code Windows pass=$winPass skip=$winSkip"
    if ($code -ne 0) { exit $code }
    if ($winPass -ne 4 -or $winSkip -gt 0) { Write-Warning 'PARTIAL: Windows cases skipped or missing'; exit 2 }
  } finally { Pop-Location }
} else {
  $Binary = (Resolve-Path -LiteralPath $Binary).Path
  Get-FileHash -Algorithm SHA256 -LiteralPath $Binary | Format-List
  $lines = @(& $Binary '-test.v' '-test.count=1' '-test.timeout=5m')
  $code = $LASTEXITCODE
  $lines | Tee-Object -FilePath (Join-Path $LogDir 'binary-tests.log') | Write-Output
  $winPass = @($lines | Where-Object { $_ -match '^--- PASS: TestWinPref' }).Count
  $winSkip = @($lines | Where-Object { $_ -match '^--- SKIP: TestWinPref' }).Count
  Write-Host "binary (non-race) exit=$code Windows pass=$winPass skip=$winSkip"
  if ($code -ne 0) { exit $code }
  if ($winPass -ne 4 -or $winSkip -gt 0) { Write-Warning 'PARTIAL: enable symlink privilege/Developer Mode and rerun'; exit 2 }
}
Write-Host 'PASS: all four Windows-specific cases ran, no skip. Keep logs locally.'
exit 0
