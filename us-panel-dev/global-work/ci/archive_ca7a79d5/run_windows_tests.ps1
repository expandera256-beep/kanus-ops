# Local runner for the GLOBAL i18n Windows tests. No uploads, no installs.
# Windows PowerShell 5.1 and PowerShell 7 compatible.
# Binary mode (default) needs no Go/GCC; the cross-built binary is NOT race-enabled.
# Source mode (-ModuleDir) needs Go + MinGW-w64 GCC and runs ./i18n/ with -race.
# Exit: 0 all PASS, 1 test FAIL, 2 PARTIAL (skip/missing), 3 hash mismatch/setup error.
[CmdletBinding()]
param(
  [string]$Binary = '',
  [string]$ModuleDir = '',
  [string]$LogDir = '',
  [string]$SumsFile = ''
)
$ErrorActionPreference = 'Stop'

# PS 5.1 leaves $PSScriptRoot empty inside param defaults; resolve here.
$here = $PSScriptRoot
if (-not $here -and $MyInvocation.MyCommand.Path) { $here = Split-Path -Parent $MyInvocation.MyCommand.Path }
if (-not $here) { $here = (Get-Location).Path }
if (-not $Binary)   { $Binary   = Join-Path $here 'i18n_windows_amd64.test.exe' }
if (-not $LogDir)   { $LogDir   = Join-Path $here 'windows-test-results' }
if (-not $SumsFile) { $SumsFile = Join-Path $here 'SHA256SUMS.txt' }
$expectedWin = 6   # TestWinPref* cases in this build
New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
$LogDir = (Resolve-Path -LiteralPath $LogDir).Path

# --- environment report: symlink privilege / Developer Mode ---
$priv = 'unknown'
try {
  $w = (whoami /priv) 2>$null | Select-String 'SeCreateSymbolicLinkPrivilege'
  if ($w) { $priv = ($w.ToString() -split '\s+' | Where-Object { $_ })[-1] } else { $priv = 'absent' }
} catch { $priv = 'unknown' }
$dev = 'off'
try {
  $v = Get-ItemPropertyValue -Path 'HKLM:\SOFTWARE\Microsoft\Windows\CurrentVersion\AppModelUnlock' -Name AllowDevelopmentWithoutDevLicense -ErrorAction Stop
  if ($v -eq 1) { $dev = 'on' }
} catch { $dev = 'off/unknown' }
$os = [System.Environment]::OSVersion.VersionString
$envLine = "host=$env:COMPUTERNAME os=$os ps=$($PSVersionTable.PSVersion) SeCreateSymbolicLinkPrivilege=$priv DeveloperMode=$dev"
Write-Host $envLine
$envLine | Set-Content -Encoding utf8 (Join-Path $LogDir 'environment.txt')
if ($priv -ne 'Enabled' -and $dev -ne 'on') {
  Write-Host 'NOTE: no symlink privilege and Developer Mode off -> TestWinPrefSymlinkFileRejected will SKIP (documented; hardlink/junction cases still run).'
}

function Show-Summary([string[]]$lines, [int]$code, [string]$mode) {
  $pass = @($lines | Where-Object { $_ -match '^\s*--- PASS: ' }).Count
  $fail = @($lines | Where-Object { $_ -match '^\s*--- FAIL: ' }).Count
  $skip = @($lines | Where-Object { $_ -match '^\s*--- SKIP: ' }).Count
  $winPass = @($lines | Where-Object { $_ -match '^\s*--- PASS: TestWinPref' }).Count
  $winSkip = @($lines | Where-Object { $_ -match '^\s*--- SKIP: TestWinPref' }).Count
  Write-Host "$mode exit=$code total pass=$pass fail=$fail skip=$skip | TestWinPref pass=$winPass skip=$winSkip (expected $expectedWin)"
  if ($fail -gt 0) {
    Write-Host 'FAILED TESTS:'
    for ($i = 0; $i -lt $lines.Count; $i++) {
      if ($lines[$i] -match '^\s*--- FAIL: (\S+)') {
        $name = $Matches[1]; $msg = @()
        for ($j = $i - 1; $j -ge 0 -and $lines[$j] -match '^\s{4}'; $j--) { $msg = ,$lines[$j].Trim() + $msg }
        Write-Host ("  {0}: {1}" -f $name, ($msg -join ' | '))
      }
    }
  }
  if ($winSkip -gt 0) {
    $lines | Where-Object { $_ -match '^\s*--- SKIP: ' } | ForEach-Object { Write-Host "  $_" }
  }
  if ($code -ne 0 -or $fail -gt 0) { exit 1 }
  if ($winPass + $winSkip -ne $expectedWin) { Write-Warning 'PARTIAL: Windows cases missing'; exit 2 }
  if ($winSkip -gt 0) { Write-Warning 'PARTIAL: only the symlink case may skip without SeCreateSymbolicLinkPrivilege/Developer Mode'; exit 2 }
}

if ($ModuleDir) {
  $env:CGO_ENABLED = '1'
  gcc --version
  if ($LASTEXITCODE -ne 0) { Write-Host ('ERROR: ' + 'Race requires MinGW-w64 GCC; use Binary mode otherwise.') -ForegroundColor Red; exit 3 }
  Push-Location -LiteralPath $ModuleDir
  try {
    $lines = @(go test -race -count=1 -v ./i18n/ 2>&1 | ForEach-Object { "$_" })
    $code = $LASTEXITCODE
    $lines | Set-Content -Encoding utf8 (Join-Path $LogDir 'go-test.log')
    $lines | Write-Output
    Show-Summary $lines $code 'source(race)'
  } finally { Pop-Location }
} else {
  if (-not (Test-Path -LiteralPath $Binary)) { Write-Host ('ERROR: ' + "binary not found: $Binary") -ForegroundColor Red; exit 3 }
  $Binary = (Resolve-Path -LiteralPath $Binary).Path
  $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $Binary).Hash.ToLowerInvariant()
  if (-not (Test-Path -LiteralPath $SumsFile)) { Write-Host ('ERROR: ' + "SHA256SUMS not found: $SumsFile") -ForegroundColor Red; exit 3 }
  $leaf = Split-Path -Leaf $Binary
  $want = $null
  foreach ($l in Get-Content -LiteralPath $SumsFile) {
    $f = $l.Trim() -split '\s+', 2
    if ($f.Count -eq 2 -and $f[1].TrimStart('*') -eq $leaf) { $want = $f[0].ToLowerInvariant() }
  }
  Write-Host "sha256 $leaf actual=$actual expected=$want"
  if (-not $want -or $want -ne $actual) { Write-Host ('ERROR: ' + 'HASH MISMATCH: binary does not match SHA256SUMS.txt; stopping.') -ForegroundColor Red; exit 3 }
  $lines = @(& $Binary '-test.v' '-test.count=1' '-test.timeout=5m' 2>&1 | ForEach-Object { "$_" })
  $code = $LASTEXITCODE
  $lines | Set-Content -Encoding utf8 (Join-Path $LogDir 'binary-tests.log')
  $lines | Write-Output
  Show-Summary $lines $code 'binary(non-race)'
}
Write-Host "PASS: all $expectedWin Windows-specific cases ran, no skip, no fail. Logs: $LogDir"
exit 0
