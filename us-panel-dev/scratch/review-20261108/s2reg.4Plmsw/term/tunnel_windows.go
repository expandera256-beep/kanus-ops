//go:build windows

package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "sync"
    "syscall"
    "time"
    "unsafe"
)

var cloudMu sync.Mutex
var cloudRunning bool

func lookupCloudCLI() (string, string) {
    if p, err := exec.LookPath("gcloud.cmd"); err == nil { return p, "" }
    for _, root := range []string{os.Getenv("LOCALAPPDATA"), os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles")} {
        if root == "" { continue }
        for _, rel := range []string{`Google\Cloud SDK\google-cloud-sdk\bin\gcloud.cmd`, `Google\Cloud SDK\bin\gcloud.cmd`} {
            path := filepath.Join(root, rel)
            if regularFile(path) { return path, "" }
        }
    }
    return "", ""
}
func cloudEnvironment(python string) []string {
    env := os.Environ()
    if python == "" { return env }
    clean := make([]string, 0, len(env)+1)
    for _, item := range env { if !strings.HasPrefix(strings.ToUpper(item), "CLOUDSDK_PYTHON=") { clean = append(clean, item) } }
    return append(clean, "CLOUDSDK_PYTHON="+python)
}
func cloudHasAccount(gcloud, python string) bool {
    // This command only checks whether the user's own gcloud profile has an
    // active account. It never prints or sends the account to US's servers.
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second); defer cancel()
    cmd := exec.CommandContext(ctx, "cmd.exe")
    cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000, CmdLine: `cmd.exe /d /c call "`+gcloud+`" auth list --filter=status:ACTIVE --format="value(account)"`}
    cmd.Env = cloudEnvironment(python)
    output, err := cmd.Output()
    return err == nil && strings.TrimSpace(string(output)) != ""
}
// Only fixed local disks are considered. Removable and network drives are skipped.
func windowsCloudLocations() []cloudInstallLocation {
    var candidates []cloudInstallLocation
    var preferredVolume string
    if cache, err := os.UserCacheDir(); err == nil && cache != "" {
        candidates = append(candidates, cloudInstallLocation{parent: filepath.Join(cache, "US"), probe: cache})
        preferredVolume = strings.ToUpper(filepath.VolumeName(cache))
    }
    kernel := syscall.NewLazyDLL("kernel32.dll")
    mask, _, _ := kernel.NewProc("GetLogicalDrives").Call()
    driveType := kernel.NewProc("GetDriveTypeW")
    for n:=0; n<26; n++ {
        if mask&(uintptr(1)<<uint(n))==0 { continue }
        volume := fmt.Sprintf("%c:", 'A'+n)
        if strings.EqualFold(volume, preferredVolume) { continue }
        root := volume + `\`
        ptr, err := syscall.UTF16PtrFromString(root); if err != nil { continue }
        kind, _, _ := driveType.Call(uintptr(unsafe.Pointer(ptr)))
        if kind != 3 { continue } // DRIVE_FIXED
        candidates = append(candidates, cloudInstallLocation{parent: filepath.Join(root, "US_Cache"), probe: root})
    }
    return candidates
}
func windowsFreeBytes(dir string) (uint64, error) {
    ptr, err := syscall.UTF16PtrFromString(dir); if err != nil { return 0, err }
    var available uint64
    ok, _, callErr := syscall.NewLazyDLL("kernel32.dll").NewProc("GetDiskFreeSpaceExW").Call(uintptr(unsafe.Pointer(ptr)), uintptr(unsafe.Pointer(&available)), 0, 0)
    if ok == 0 { return 0, fmt.Errorf("disk alanı okunamadı: %v", callErr) }
    return available, nil
}
func cloudInstallLocationForWindows() (string, string, string, error) {
    candidates:=windowsCloudLocations()
    // A previously installed verified copy can be reused without another download.
    for _, location := range candidates {
        if cli, py, ok := cachedCloudCLI(filepath.Join(location.parent, cloudInstallVersion), cloudArchiveSHA256); ok {
            return location.parent, cli, py, nil
        }
    }
    selected, err:=chooseCloudInstallLocation(candidates, windowsFreeBytes)
    if err != nil { return "", "", "", err }
    return selected.parent, "", "", nil
}

func startCloudSSH() error {
    cloudMu.Lock()
    defer cloudMu.Unlock()
    if cloudRunning { return fmt.Errorf("Google Cloud penceresi zaten açık; oradaki bağlantıyı kontrol edin") }
    gcloud, python := lookupCloudCLI()
    if gcloud == "" {
        parent, installedCLI, installedPython, err := cloudInstallLocationForWindows()
        if err != nil { return fmt.Errorf("Google Cloud CLI kurulumu başlatılamadı: %w", err) }
        if installedCLI != "" { gcloud, python = installedCLI, installedPython } else {
            autoTunnel.update("Google Cloud CLI " + filepath.VolumeName(parent) + " sürücüsüne hazırlanıyor.")
            client := &http.Client{Timeout: 15*time.Minute}
            gcloud, python, err = installCloudArchive(parent, client, cloudArchiveURL, cloudArchiveSHA256, autoTunnel.update)
            if err != nil { return fmt.Errorf("Google Cloud CLI kurulumu tamamlanamadı (%s): %w", filepath.VolumeName(parent), err) }
        }
    }
    command, err := cloudSSHCommand(gcloud)
    if err != nil { return err }
    autoTunnel.update("Google hesabı kontrol ediliyor. İlk kullanımda açılan pencerede oturum açın.")
    if !cloudHasAccount(gcloud, python) {
        command = `call "`+gcloud+`" auth login && ` + command
    }
    // CmdLine avoids Go's CommandLineToArgvW escaping, which cmd.exe does not use.
    cmd := exec.Command("cmd.exe")
    cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x00000010, CmdLine: `cmd.exe /d /k ` + command}
    cmd.Env = cloudEnvironment(python)
    if err := cmd.Start(); err != nil { return err }
    autoTunnel.update("Google Cloud penceresi açıldı. Gerekirse Google hesabınızla oturum açın ve pencereyi açık tutun.")
    cloudRunning = true
    go func() { _ = cmd.Wait(); cloudMu.Lock(); cloudRunning = false; cloudMu.Unlock() }()
    return nil
}

func autoTunnelOnThisPlatform() bool { return true }
