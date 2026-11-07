package main

import (
    "archive/zip"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "strings"
)

// Version and SHA-256 are pinned to the official Google Cloud archive.
const cloudArchiveURL = "https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-592.0.0-windows-x86_64-bundled-python.zip"
const cloudArchiveSHA256 = "65002e0d679fc2ade00163b97637f45656d6fd5250d6f5962e4a1a555f2628ea"
const cloudInstallVersion = "google-cloud-cli-592.0.0"
const maxArchiveBytes = int64(256 << 20)
const maxExtractedBytes = uint64(2 << 30)

func cliPaths(root string) (string, string) {
    return filepath.Join(root, "google-cloud-sdk", "bin", "gcloud.cmd"), filepath.Join(root, "google-cloud-sdk", "platform", "bundledpython", "python.exe")
}
func regularFile(name string) bool {
    info, err := os.Stat(name)
    return err == nil && info.Mode().IsRegular()
}
func cachedCloudCLI(root, sha string) (string, string, bool) {
    cli, py := cliPaths(root)
    marker, err := os.ReadFile(filepath.Join(root, ".us-sha256"))
    return cli, py, err == nil && strings.TrimSpace(string(marker)) == sha && regularFile(cli) && regularFile(py)
}

// installCloudArchive installs to a fresh, private version directory and never
// modifies PATH or overwrites another installation. The caller pins the hash.
func installCloudArchive(parent string, client *http.Client, archiveURL, expected string, progress func(string)) (string, string, error) {
    if len(expected) != 64 { return "", "", fmt.Errorf("kurulum doğrulama değeri geçersiz") }
    if _, err := hex.DecodeString(expected); err != nil { return "", "", fmt.Errorf("kurulum doğrulama değeri geçersiz") }
    root := filepath.Join(parent, cloudInstallVersion)
    if cli, py, ok := cachedCloudCLI(root, expected); ok { return cli, py, nil }
    if err := os.MkdirAll(parent, 0700); err != nil { return "", "", fmt.Errorf("kurulum klasörü açılamadı: %w", err) }
    if _, err := os.Lstat(root); err == nil { return "", "", fmt.Errorf("eksik bir US CLI klasörü bulundu; güvenli biçimde kaldırılamadı, kurulumu elle kontrol edin") } else if !os.IsNotExist(err) { return "", "", err }
    stage, err := os.MkdirTemp(parent, ".us-cloud-setup-")
    if err != nil { return "", "", err }
    defer os.RemoveAll(stage)
    archive := filepath.Join(stage, "download.zip")
    if progress != nil { progress("Google Cloud CLI indiriliyor (yaklaşık 105 MB).") }
    resp, err := client.Get(archiveURL)
    if err != nil { return "", "", fmt.Errorf("Google Cloud CLI indirilemedi: %w", err) }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK { return "", "", fmt.Errorf("Google Cloud CLI indirme hatası: HTTP %d", resp.StatusCode) }
    f, err := os.OpenFile(archive, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
    if err != nil { return "", "", err }
    hash := sha256.New()
    count, copyErr := io.Copy(io.MultiWriter(f, hash), io.LimitReader(resp.Body, maxArchiveBytes+1))
    closeErr := f.Close()
    if copyErr != nil { return "", "", fmt.Errorf("indirme tamamlanamadı: %w", copyErr) }
    if closeErr != nil { return "", "", closeErr }
    if count > maxArchiveBytes { return "", "", fmt.Errorf("indirilen paket beklenenden büyük") }
    if hex.EncodeToString(hash.Sum(nil)) != strings.ToLower(expected) { return "", "", fmt.Errorf("paketin SHA-256 doğrulaması başarısız") }
    if progress != nil { progress("Google Cloud CLI doğrulandı, dosyalar hazırlanıyor.") }
    content := filepath.Join(stage, "content")
    if err := os.Mkdir(content, 0700); err != nil { return "", "", err }
    if err := unpackCloudArchive(archive, content); err != nil { return "", "", fmt.Errorf("paket açılamadı: %w", err) }
    cli, py := cliPaths(content)
    if !regularFile(cli) || !regularFile(py) { return "", "", fmt.Errorf("pakette CLI veya Python eksik") }
    if err := os.WriteFile(filepath.Join(content, ".us-sha256"), []byte(expected+"\n"), 0600); err != nil { return "", "", err }
    if err := os.Rename(content, root); err != nil {
        if cli, py, ok := cachedCloudCLI(root, expected); ok { return cli, py, nil }
        return "", "", fmt.Errorf("kurulum tamamlanamadı: %w", err)
    }
    cli, py = cliPaths(root)
    return cli, py, nil
}

func unpackCloudArchive(archive, destination string) error {
    z, err := zip.OpenReader(archive)
    if err != nil { return err }
    defer z.Close()
    var total uint64
    for _, entry := range z.File {
        name := entry.Name
        if name == "" || strings.ContainsAny(name, "\\:") || strings.HasPrefix(name, "/") { return fmt.Errorf("güvenli olmayan dosya yolu") }
        clean := path.Clean(name)
        if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || !strings.HasPrefix(clean, "google-cloud-sdk/") && clean != "google-cloud-sdk" { return fmt.Errorf("güvenli olmayan dosya yolu") }
        // Google's pinned Windows archive has documentation/test symlinks only.
        if entry.Mode()&os.ModeSymlink != 0 {
            if strings.HasPrefix(clean, "google-cloud-sdk/platform/gsutil/third_party/") { continue }
            return fmt.Errorf("beklenmeyen sembolik bağlantı")
        }
        if entry.UncompressedSize64 > maxExtractedBytes-total { return fmt.Errorf("paket açılmış boyut sınırını aşıyor") }
        total += entry.UncompressedSize64
        target := filepath.Join(destination, filepath.FromSlash(clean))
        rel, err := filepath.Rel(destination, target)
        if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) { return fmt.Errorf("güvenli olmayan dosya yolu") }
        if entry.FileInfo().IsDir() { if err := os.MkdirAll(target, 0700); err != nil { return err }; continue }
        if !entry.Mode().IsRegular() { return fmt.Errorf("desteklenmeyen dosya türü") }
        if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil { return err }
        in, err := entry.Open(); if err != nil { return err }
        out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
        if err != nil { in.Close(); return err }
        written, copyErr := io.Copy(out, io.LimitReader(in, int64(entry.UncompressedSize64)+1))
        closeOut := out.Close(); closeIn := in.Close()
        if copyErr != nil { return copyErr }; if closeOut != nil { return closeOut }; if closeIn != nil { return closeIn }
        if written != int64(entry.UncompressedSize64) { return fmt.Errorf("paket dosyası eksik") }
    }
    return nil
}
