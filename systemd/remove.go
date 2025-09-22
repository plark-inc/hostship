package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Remove stops and disables the hostship service and removes the systemd unit.
// When dryRun is true the actions are only printed.
func Remove(dryRun, verbose bool) error {
	path := "/etc/systemd/system/hostship.service"
	cmds := [][]string{
		{"systemctl", "disable", "--now", "hostship"},
		{"systemctl", "daemon-reload"},
	}
	if err := runCommands(cmds, dryRun, verbose); err != nil {
		return err
	}
	return removeUnitFile(path, dryRun, verbose)
}

func runCommands(cmds [][]string, dryRun, verbose bool) error {
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
			if strings.Contains(err.Error(), "No such file or directory") {
				continue
			}
			return err
		}
	}
	return nil
}

func removeUnitFile(path string, dryRun, verbose bool) error {
	if verbose || dryRun {
		if os.Geteuid() != 0 {
			fmt.Printf("sudo rm -f %s\n", path)
		} else {
			fmt.Printf("rm -f %s\n", path)
		}
	}
	if dryRun {
		return nil
	}
	if os.Geteuid() != 0 {
		return exec.Command("sudo", "rm", "-f", path).Run()
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
