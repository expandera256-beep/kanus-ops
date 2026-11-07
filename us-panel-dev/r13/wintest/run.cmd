@echo off
setlocal
cd /d "%~dp0"
set "TESTEXE=r13_windows_amd64.test.exe"
if /I "%PROCESSOR_ARCHITECTURE%"=="x86" if "%PROCESSOR_ARCHITEW6432%"=="" set "TESTEXE=r13_windows_386.test.exe"
"%TESTEXE%" -test.v -test.count=3 -test.timeout=180s -test.run="^TestR13" > windows-results.log 2>&1
set "RC=%ERRORLEVEL%"
type windows-results.log
if not "%RC%"=="0" (
 echo FAIL - windows-results.log dosyasini inceleyin.
 exit /b %RC%
)
echo PASS - Windows dosya katmani testleri tamamlandi. SKIP satirlarini da inceleyin.
exit /b 0
