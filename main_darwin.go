//go:build darwin

package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func showInputDialog(title, message, defaultAnswer string) (string, bool) {
	script := fmt.Sprintf("text returned of (display dialog %q default answer %q with title %q buttons {\"取消\", \"確定\"} default button \"確定\")", message, defaultAnswer, title)
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", false
	}
	return string(out[:len(out)-1]), true // Remove trailing newline
}

func showConfirmDialog(title, message string) bool {
	script := fmt.Sprintf("button returned of (display dialog %q with title %q buttons {\"取消\", \"開始更新\"} default button \"開始更新\")", message, title)
	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "開始更新"
}
