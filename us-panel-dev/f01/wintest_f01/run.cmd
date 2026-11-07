@echo off
setlocal
cd /d "%~dp0"
set "ARCH=amd64"
if /I "%PROCESSOR_ARCHITECTURE%"=="x86" set "ARCH=386"
if "%~1"=="386" set "ARCH=386"
"f01_%ARCH%.test.exe" -test.v -test.run="^TestF01Windows" -test.timeout=3m > "sonuc_%ARCH%.log" 2>&1
set "RC=%ERRORLEVEL%"
type "sonuc_%ARCH%.log"
echo exit_code=%RC%
echo exit_code=%RC%>>"sonuc_%ARCH%.log"
exit /b %RC%
