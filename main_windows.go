//go:build windows

package main

import (
	_ "embed"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex = kernel32.NewProc("CreateMutexW")
	procCloseHandle = kernel32.NewProc("CloseHandle")
	user32          = syscall.NewLazyDLL("user32.dll")
	procMessageBox  = user32.NewProc("MessageBoxW")
)

const (
	errorAlreadyExists syscall.Errno = 183
	mbOk               uint32        = 0x00000000
	mbIconWarning      uint32        = 0x00000030
)

//go:embed assets/icon.ico
var iconData []byte

func checkSingleInstance() (func(), bool) {
	// Check for single instance using Windows Named Mutex
	mutexName, _ := syscall.UTF16PtrFromString("Local\\AyayaBot-SingleInstance-Mutex")
	ret, _, err := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(mutexName)))

	if err != nil && err.(syscall.Errno) == errorAlreadyExists {
		if ret != 0 {
			_, _, _ = procCloseHandle.Call(ret)
		}
		return nil, false
	}

	if ret == 0 {
		return nil, true // Failed to create mutex, but we'll allow running
	}

	cleanup := func() {
		_, _, _ = procCloseHandle.Call(ret)
	}
	return cleanup, true
}

func showAlert(title, message string) {
	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(message)
	_, _, _ = procMessageBox.Call(0, uintptr(unsafe.Pointer(m)), uintptr(unsafe.Pointer(t)), uintptr(mbOk|mbIconWarning))
}

func showInputDialog(title, message, defaultAnswer string) (string, bool) {
	// Use PowerShell to show an input box via Microsoft.VisualBasic.Interaction
	psCommand := fmt.Sprintf(`[System.Reflection.Assembly]::LoadWithPartialName('Microsoft.VisualBasic') | Out-Null; $res = [Microsoft.VisualBasic.Interaction]::InputBox('%s', '%s', '%s'); if($res) { Write-Host $res }`, message, title, defaultAnswer)
	out, err := exec.Command("powershell", "-NoProfile", "-Command", psCommand).Output()
	if err != nil {
		return "", false
	}
	result := strings.TrimSpace(string(out))
	if result == "" {
		return "", false
	}
	return result, true
}
