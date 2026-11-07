@echo off
setlocal EnableExtensions EnableDelayedExpansion
cd /d "%~dp0"
rem r13b native Windows test runner. Writes metadata + results + exit codes to r13b-results.log
set "LOG=r13b-results.log"
> "%LOG%" echo ==== R13B NATIVE RUN ====
>>"%LOG%" echo RUN_DATE=%DATE% %TIME%
>>"%LOG%" echo HOSTNAME=%COMPUTERNAME%
for /f "delims=" %%v in ('hostname') do >>"%LOG%" echo HOSTNAME_CMD=%%v
for /f "delims=" %%v in ('ver') do >>"%LOG%" echo OS_VERSION=%%v
>>"%LOG%" echo PROCESSOR_ARCHITECTURE=%PROCESSOR_ARCHITECTURE% PROCESSOR_ARCHITEW6432=%PROCESSOR_ARCHITEW6432%
net session >nul 2>&1
if "%ERRORLEVEL%"=="0" (set "ADMIN=1") else (set "ADMIN=0")
>>"%LOG%" echo ADMIN=%ADMIN%
whoami /groups | findstr /i /c:"Mandatory Label" >>"%LOG%"
>>"%LOG%" echo ---- SHA256 (certutil) ----
for %%f in (r13b_windows_amd64.test.exe r13b_windows_386.test.exe windows_handle_fileops_v2.diff BUILDINFO.txt) do (
  for /f "skip=1 delims=" %%h in ('certutil -hashfile "%%f" SHA256 ^| findstr /v /i "certutil"') do >>"%LOG%" echo SHA256 %%f %%h
)
>>"%LOG%" echo ---- expected (SHA256SUMS.txt) ----
type SHA256SUMS.txt >>"%LOG%"
type BUILDINFO.txt >>"%LOG%"

set "RC64=NOTRUN"
set "RC32=NOTRUN"
set "MAIN=r13b_windows_amd64.test.exe"
if /I "%PROCESSOR_ARCHITECTURE%"=="x86" if "%PROCESSOR_ARCHITEW6432%"=="" set "MAIN="
if defined MAIN (
  >>"%LOG%" echo ==== amd64 ====
  "r13b_windows_amd64.test.exe" -test.v -test.count=3 -test.timeout=600s -test.run="^TestR13" >>"%LOG%" 2>&1
  set "RC64=!ERRORLEVEL!"
  >>"%LOG%" echo AMD64_EXITCODE=!RC64!
)
>>"%LOG%" echo ==== 386 ====
"r13b_windows_386.test.exe" -test.v -test.count=3 -test.timeout=600s -test.run="^TestR13" >>"%LOG%" 2>&1
set "RC32=!ERRORLEVEL!"
>>"%LOG%" echo I386_EXITCODE=!RC32!

set "FINAL=0"
if defined MAIN if not "!RC64!"=="0" set "FINAL=1"
if not "!RC32!"=="0" set "FINAL=1"
>>"%LOG%" echo ---- SKIP lines ----
findstr /c:"SKIP-" "%LOG%" | findstr /v /c:"findstr" >>"%LOG%.skips.tmp"
type "%LOG%.skips.tmp" >>"%LOG%"
del "%LOG%.skips.tmp" >nul 2>&1
>>"%LOG%" echo FINAL_EXITCODE=!FINAL!
for /f "skip=1 delims=" %%h in ('certutil -hashfile "%LOG%" SHA256 ^| findstr /v /i "certutil"') do echo LOG_SHA256 %%h
type "%LOG%"
if "!FINAL!"=="0" (echo PASS - r13b Windows testleri. SKIP satirlarini README'ye gore inceleyin.) else (echo FAIL - %LOG% dosyasini PANEL DEV'e iletin.)
exit /b !FINAL!
