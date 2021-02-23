package systemd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type Service struct {
	Name    string
	Active  bool
	Enabled bool
}

func unitExists(name string) (bool, error) {
	unitPaths := []string{
		"/usr/lib/systemd/system/",
		"/etc/systemd/system/",
		"/usr/local/lib/systemd/system/",
		"/etc/systemd/user/",
		"/etc/systemd/system.control/",
		"/run/systemd/system.control/",
		"/run/systemd/transient/",
		"/run/systemd/generator.early/",
		"/etc/systemd/systemd.attached/",
		"/run/systemd/system/",
		"/run/systemd/systemd.attached/",
		"/run/systemd/generator/",
		"/lib/systemd/system/",
		"/run/systemd/generator.late/",
		"/usr/lib/systemd/user/"}

	for _, unitPath := range unitPaths {
		if _, err := os.Stat(unitPath); os.IsNotExist(err) {
			continue
		}

		files, err := ioutil.ReadDir(unitPath)

		if err != nil {
			return false, err
		}

		for _, file := range files {
			if file.Name() == name {
				return true, nil
			}
		}
	}

	return false, nil
}

func Unit(name string) (Service, error) {
	if exist, err := unitExists(name); err != nil {
		return Service{}, err
	} else if exist != true {
		return Service{}, fmt.Errorf("unit not exist: %s", name)
	}

	active, err := IsActive(name)

	if err != nil {
		return Service{}, err
	}

	enabled, err := IsEnabled(name)

	if err != nil {
		return Service{}, err
	}

	return Service{Name: name, Active: active, Enabled: enabled}, nil
}

func IsActive(name string) (bool, error) {
	output, err := exec.Command("/usr/bin/systemctl", "is-active", name).CombinedOutput()

	if err != nil {
		return false, fmt.Errorf("failed to run systemctl: %s %s", output, err)
	}

	switch string(output) {
	case "active\n":
		return true, nil
	case "inactive\n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid response: %s", string(output))
	}
}

func IsEnabled(name string) (bool, error) {
	output, err := exec.Command("/usr/bin/systemctl", "is-enabled", name).CombinedOutput()

	if err != nil {
		return false, fmt.Errorf("failed to run systemctl: %s %s", output, err)
	}

	switch string(output) {
	case "enabled\n":
		return true, nil
	case "disabled\n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid response: %s", string(output))
	}
}

func (s *Service) Enable() error {
	output, err := exec.Command("/usr/bin/systemctl", "enable", s.Name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func (s *Service) Disable() error {
	output, err := exec.Command("/usr/bin/systemctl", "disable", s.Name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func (s *Service) Start() error {
	output, err := exec.Command("/usr/bin/systemctl", "start", s.Name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func (s *Service) Stop() error {
	output, err := exec.Command("/usr/bin/systemctl", "stop", s.Name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func (s *Service) Restart() error {
	output, err := exec.Command("/usr/bin/systemctl", "restart", s.Name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func (s *Service) Reload() error {
	output, err := exec.Command("/usr/bin/systemctl", "reload", s.Name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func EnableService(name string) error {
	if exist, err := unitExists(name); err != nil {
		return err
	} else if exist != true {
		return fmt.Errorf("unit not exist: %s", name)
	}

	output, err := exec.Command("/usr/bin/systemctl", "enable", name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func DisableService(name string) error {
	if exist, err := unitExists(name); err != nil {
		return err
	} else if exist != true {
		return fmt.Errorf("unit not exist: %s", name)
	}

	output, err := exec.Command("/usr/bin/systemctl", "disable", name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func StartService(name string) error {
	if exist, err := unitExists(name); err != nil {
		return err
	} else if exist != true {
		return fmt.Errorf("unit not exist: %s", name)
	}

	output, err := exec.Command("/usr/bin/systemctl", "start", name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func StopService(name string) error {
	if exist, err := unitExists(name); err != nil {
		return err
	} else if exist != true {
		return fmt.Errorf("unit not exist: %s", name)
	}

	output, err := exec.Command("/usr/bin/systemctl", "stop", name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func RestartService(name string) error {
	if exist, err := unitExists(name); err != nil {
		return err
	} else if exist != true {
		return fmt.Errorf("unit not exist: %s", name)
	}

	output, err := exec.Command("/usr/bin/systemctl", "restart", name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func ReloadService(name string) error {
	if exist, err := unitExists(name); err != nil {
		return err
	} else if exist != true {
		return fmt.Errorf("unit not exist: %s", name)
	}

	output, err := exec.Command("/usr/bin/systemctl", "reload", name).CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}

func Link(service string, source string) error {
	os.Remove(filepath.Join("/etc/systemd/system", service))

	return os.Symlink(source, filepath.Join("/etc/systemd/system", service))
}

func UnLink(service string) error {
	return os.Remove(filepath.Join("/etc/systemd/system", service))
}

func Reload() error {
	output, err := exec.Command("/usr/bin/systemctl", "daemon-reload").CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s %s", output, err)
	}

	return nil
}