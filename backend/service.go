//go:build windows

package backend

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "LenovoProcessManagement"

// ServiceControlCode contains common Lenovo service control codes
var ServiceControlCodes = map[uint32]string{
	0x94: "OEM_EXTRA_PERFORMANCE_MODE (Group 2/3/4 Turbo)",
	0xA3: "SET_INTELLIGENT",
	0xA4: "SET_BSM",
	0xA5: "SET_EPM",
	0xA7: "SET_GPU_OC_ON",
	0xA9: "SET_GPU_OC_OFF",
	0xAA: "SET_iEPM_ENABLE",
	0xAB: "SET_iEPM_DISABLE",
	0xAC: "SET_iGEEK_ENABLE",
	0xAD: "SET_AUTO_MAXPERFORMANCE_ON",
	0xAE: "SET_AUTO_MAXPERFORMANCE_OFF",
	0xAF: "SET_AUTO_MAXBATTERYLIFE_ON",
	0xA6: "SET_AUTO_MAXBATTERYLIFE_OFF",
	0xB2: "OEM_FGAPP_CHANGE (Foreground mode)",
	0xB3: "OEM_FOCUSUI_CHANGE (Focus UI change)",
	0xB4: "OEM_AQM_BACKGROUND (Foreground mode)",
	0xB7: "OEM_CAMERAINUSE (Not whitelist mode)",
	0xB8: "OEM_CAMERANOTINUSE (Not whitelist mode)",
	0xB9: "OEM_AUDIO_MIC_INUSE (Microphone)",
	0xBA: "OEM_AUDIO_MIC_NOTINUSE (Microphone)",
	0xBB: "SET_QUIETPERFORMANCE_ON",
	0xBC: "SET_QUIETPERFORMANCE_OFF",
	0xBD: "SET_IGPUPRIORITYMODE_ON",
	0xBE: "SET_AUTOIGPUONBATTERY_ON",
	0xBF: "SET_HYBRIDGRAPHICMODE_ON",
	0xC0: "SET_EPMGEEK",
	0xC1: "SET_GEEKCUSTOMIZE",
	0xC3: "SET_ADAPTIVEGRAPHICMODE_RESTORE_ON",
}

// GetServiceStatus returns the current status of the service
func GetServiceStatus() (string, error) {
	m, err := mgr.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return "Not Installed", nil
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return "", fmt.Errorf("failed to query service status: %w", err)
	}
	return serviceStateToString(status.State), nil
}

// StartService starts the LenovoProcessManagement service
func StartService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return fmt.Errorf("failed to query service: %w", err)
	}
	if status.State == svc.Running {
		return nil
	}
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}
	return waitForState(s, svc.Running, 10*time.Second)
}

// StopService stops the LenovoProcessManagement service
func StopService() error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return fmt.Errorf("failed to query service: %w", err)
	}
	if status.State == svc.Stopped {
		return nil
	}
	_, err = s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}
	return waitForState(s, svc.Stopped, 30*time.Second)
}

// RestartService restarts the LenovoProcessManagement service
func RestartService() error {
	if err := StopService(); err != nil {
		return fmt.Errorf("failed to stop service during restart: %w", err)
	}
	time.Sleep(1 * time.Second)
	if err := StartService(); err != nil {
		return fmt.Errorf("failed to start service during restart: %w", err)
	}
	return nil
}

// SendServiceControl sends a custom service control code to the LenovoProcessManagement service.
// Equivalent to: sc control LenovoProcessManagement <code>
func SendServiceControl(code uint32) (string, error) {
	m, err := mgr.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return "", fmt.Errorf("service %s not found: %w", serviceName, err)
	}
	defer s.Close()

	status, err := s.Control(svc.Cmd(code))
	if err != nil {
		return "", fmt.Errorf("sc control %s %d failed: %w", serviceName, code, err)
	}

	codeName := fmt.Sprintf("0x%X", code)
	if name, ok := ServiceControlCodes[code]; ok {
		codeName = fmt.Sprintf("0x%X (%s)", code, name)
	}

	return fmt.Sprintf("Service control sent: code=%s, service state=%s", codeName, serviceStateToString(status.State)), nil
}

// GetServiceControlCodes returns the map of known service control codes for UI display
func GetServiceControlCodes() map[uint32]string {
	return ServiceControlCodes
}

// waitForState waits for the service to reach the desired state
func waitForState(s *mgr.Service, desiredState svc.State, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, err := s.Query()
		if err != nil {
			return fmt.Errorf("failed to query service state: %w", err)
		}
		if status.State == desiredState {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for service to reach state %s", serviceStateToString(desiredState))
}

// serviceStateToString converts a service state to a human-readable string
func serviceStateToString(state svc.State) string {
	switch state {
	case svc.Stopped:
		return "Stopped"
	case svc.StartPending:
		return "Starting"
	case svc.StopPending:
		return "Stopping"
	case svc.Running:
		return "Running"
	case svc.ContinuePending:
		return "Resuming"
	case svc.Paused:
		return "Paused"
	default:
		return fmt.Sprintf("Unknown (%d)", state)
	}
}
