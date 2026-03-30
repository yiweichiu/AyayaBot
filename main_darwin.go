//go:build darwin

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

//go:embed assets/icon.ico
var iconData []byte

func checkSingleInstance() (func(), bool) {
	lockFile := filepath.Join(os.TempDir(), "AyayaBot-SingleInstance.lock")
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		// Log error but allow running if we can't create lock file
		return nil, true
	}

	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		f.Close()
		return nil, false
	}

	cleanup := func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
		_ = os.Remove(lockFile)
	}
	return cleanup, true
}

func showAlert(title, message string) {
	script := fmt.Sprintf("display alert %q message %q", title, message)
	_ = exec.Command("osascript", "-e", script).Run()
}
