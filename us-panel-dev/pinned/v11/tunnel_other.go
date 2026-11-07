//go:build !windows

package main
import "fmt"
func startCloudSSH() error { return fmt.Errorf("otomatik bağlantı yalnızca Windows sürümünde kullanılabilir") }

func autoTunnelOnThisPlatform() bool { return false }
