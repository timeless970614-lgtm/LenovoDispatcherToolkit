//go:build windows

package backend

import (
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// hiddenCmd creates an exec.Cmd with the console window hidden on Windows.
// Use this instead of exec.Command to prevent conhost.exe window flashing.
func hiddenCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd
}

// visibleCmd creates an exec.Cmd with a visible console window on Windows.
// Used for interactive commands where the user needs to see output.
func visibleCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    false,
		CreationFlags: windows.CREATE_NEW_CONSOLE,
	}
	return cmd
}
