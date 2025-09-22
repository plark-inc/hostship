// Package docker contains helpers that use the docker CLI directly. This file
// ensures the command line tool is available. Docker Compose is installed using
// the same convenience script when missing.
package docker

import (
	"fmt"
	"os"
	"os/exec"
)

// EnsureInstalled verifies the docker CLI is available on the system. If it is
// missing, a best-effort attempt is made to install it using whatever common
// package manager is detected. When dryRun is true the install commands are
// printed but not executed.
func EnsureInstalled(dryRun, verbose bool) error {
	if _, err := exec.LookPath("docker"); err == nil {
		return nil
	}
	fmt.Println("docker not found, installing via convenience script")
	script := "curl -sSL https://get.docker.com | sh"
	if verbose || dryRun {
		fmt.Println(script)
	}
	if dryRun {
		return nil
	}
	cmd := exec.Command("sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	if _, err := exec.LookPath("docker"); err != nil {
		return err
	}
	if verbose {
		fmt.Println("docker installed")
	}
	return nil
}
