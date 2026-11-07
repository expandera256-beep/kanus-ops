@echo off
cd /d "%~dp0"
where go >nul 2>nul
if errorlevel 1 (
  echo Go derleyicisi bulunamadi. Go 1.22 veya daha yeni surumu kurun.
  pause
  exit /b 1
)
echo Testler calistiriliyor...
go test ./...
if errorlevel 1 (
  echo Testler basarisiz. EXE uretilmedi.
  pause
  exit /b 1
)
set "CGO_ENABLED=0"
set "GOOS=windows"
set "GOARCH=amd64"
go build -ldflags "-H windowsgui -s -w" -o US.exe .
if errorlevel 1 (
  echo Derleme basarisiz.
  pause
  exit /b 1
)
echo US.exe olusturuldu.
pause
