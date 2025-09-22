// Package systemd provides helpers to install the hostship binary as a systemd
// service on Linux systems.
package systemd

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed hostship.service
var unitTemplate string

// Install writes the systemd unit file and enables it so the hostship update
// listener starts automatically on boot. The provided binary and configuration
// paths are embedded into the unit file. When dryRun is true the steps are only
// printed.
func Install(binPath string, dryRun, verbose bool) error {
	path := "/etc/systemd/system/hostship.service"
	unit := strings.ReplaceAll(unitTemplate, "/usr/local/bin/hostship", binPath)

	if verbose || dryRun {
		fmt.Printf("installing unit file to %s\n", path)
	}

	if !dryRun {
		if err := writeUnitFile(path, unit); err != nil {
			return err
		}
	}

	return enableService(dryRun, verbose)
}

// writeUnitFile writes the hostship systemd unit file to the given path. When
// running as non-root the file is copied using sudo.
func writeUnitFile(path, unit string) error {
	if err := os.WriteFile(path, []byte(unit), 0644); err != nil {
		if os.Geteuid() != 0 {
			tmp := filepath.Join(os.TempDir(), "hostship.service")
			if err2 := os.WriteFile(tmp, []byte(unit), 0644); err2 != nil {
				return fmt.Errorf("write temp unit: %w", err2)
			}
			if err2 := exec.Command("sudo", "cp", tmp, path).Run(); err2 != nil {
				return fmt.Errorf("install unit: %w", err2)
			}
			return nil
		}
		return fmt.Errorf("install unit: %w", err)
	}
	return nil
}

// enableService reloads systemd and enables the hostship service. When dryRun is
// true, the commands are printed without executing.
func enableService(dryRun, verbose bool) error {
	cmds := [][]string{
		{"systemctl", "daemon-reload"},
		{"systemctl", "enable", "--now", "hostship"},
		{"systemctl", "restart", "hostship"},
	}
	for _, args := range cmds {
		if os.Geteuid() != 0 {
			args = append([]string{"sudo"}, args...)
		}
		if verbose || dryRun {
			fmt.Println(strings.Join(args, " "))
		}
		if dryRun {
			continue
		}
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
