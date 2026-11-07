//go:build !linux && !windows

package main

import (
	"fmt"
	"log"
	"os/exec"
)

func sandboxPreflight() error {
	err := fmt.Errorf("UYARI: bu işletim sisteminde dosya izolasyonu desteklenmiyor; komut çalıştırılmadı")
	log.Print(err)
	return err
}

func prepareChildSandbox(cmd *exec.Cmd, root string) (func(), func() error, error) {
	return nil, nil, sandboxPreflight()
}
