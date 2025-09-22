// Package docker provides helpers for interacting with Docker and Docker Compose.
// This file implements the Docker Compose specific functionality.
package docker

import (
	"fmt"
	"os"
	"os/exec"
)

type ComposeClient struct{ Runner }

func NewComposeClient(dryRun, verbose bool) *ComposeClient {
	return &ComposeClient{Runner: NewRunner(dryRun, verbose)}
}

func (c *ComposeClient) Pull(file, project string, services ...string) (string, error) {
	args := []string{"compose", "-f", file, "--project-name", project, "pull"}
	args = append(args, services...)
	cmd := exec.Command("docker", args...)
	cmd.Env = os.Environ()
	return c.Output(cmd)
}

func (c *ComposeClient) Up(file, project string, services ...string) error {
	args := []string{"compose", "-f", file, "--project-name", project, "up", "-d"}
	args = append(args, services...)
	cmd := exec.Command("docker", args...)
	cmd.Env = os.Environ()
	return c.Run(cmd)
}

func EnsureComposeInstalled(dryRun, verbose bool) error {
	if err := checkCompose(dryRun, verbose); err == nil {
		return nil
	}
	fmt.Println("docker compose not found, installing via convenience script")

	if err := EnsureInstalled(dryRun, verbose); err != nil {
		return err
	}
	return checkCompose(dryRun, verbose)
}

func checkCompose(dryRun, verbose bool) error {
	cmd := exec.Command("docker", "compose", "version")
	if dryRun {
		fmt.Println(cmd.String())
		return nil
	}
	if err := cmd.Run(); err == nil {
		return nil
	}
	if _, err := exec.LookPath("docker-compose"); err == nil {
		return nil
	}
	return fmt.Errorf("docker compose not found")
}
