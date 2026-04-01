//go:build windows

package backend

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const serviceName = "LenovoProcessManagement"

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
