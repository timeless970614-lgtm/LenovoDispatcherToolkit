//go:build windows

package backend

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// hiddenCmd creates an exec.Cmd with the console window hidden on Windows.
// Use this instead of exec.Command to prevent conhost.exe window flashing.
// Sets working dir to LOCALAPPDATA\Temp for tools that need a writable dir.
func hiddenCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NEW_CONSOLE, // Give child a real console
	}
	if cmd.Dir == "" {
		cmd.Dir = os.Getenv("LOCALAPPDATA") + `\Temp`
	}
	return cmd
}

// visibleCmd creates an exec.Cmd with a visible console window on Windows.
func visibleCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    false,
		CreationFlags: windows.CREATE_NEW_CONSOLE,
	}
	return cmd
}

// wprCmd creates an exec.Cmd specifically for WPR commands.
// WPR.exe needs a real console to manage ETW sessions. Use CREATE_NEW_CONSOLE
// so WPR can access the ETW subsystem even though the window is hidden.
func wprCmd(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: windows.CREATE_NEW_CONSOLE,
	}
	if cmd.Dir == "" {
		cmd.Dir = os.Getenv("LOCALAPPDATA") + `\Temp`
	}
	return cmd
}
