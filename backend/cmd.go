//go:build windows

package backend

import (
	"os/exec"
	"syscall"
)

// hiddenCmd creates an exec.Cmd with the console window hidden on Windows.
// Use this instead of exec.Command to prevent conhost.exe window flashing.
func hiddenCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}
