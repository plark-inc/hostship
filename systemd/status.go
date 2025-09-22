package systemd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Status prints the systemctl status for the hostship service.
func Status(verbose bool) error {
	args := []string{"systemctl", "status", "hostship"}
	if os.Geteuid() != 0 {
		args = append([]string{"sudo"}, args...)
	}
	if verbose {
		fmt.Println(strings.Join(args, " "))
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
