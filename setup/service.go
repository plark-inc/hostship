// Helpers for starting and updating the Docker service described in the
// configuration file.
package setup

import (
	"github.com/plark-inc/hostship/config"
	"github.com/plark-inc/hostship/docker"
)

// StartService launches the compose stack defined in the configuration file.
// Docker and Docker Compose are verified to be installed before the containers
// are started. When dryRun is true Docker commands are printed but not
// executed.
func StartService(dryRun, verbose bool) error {
	cfgPath := config.Path

	if err := ensureDockerAvailable(dryRun, verbose); err != nil {
		return err
	}

	c := docker.NewComposeClient(dryRun, verbose)
	if _, err := c.Pull(cfgPath, "hostship"); err != nil {
		return err
	}
	return c.Up(cfgPath, "hostship")
}

func ensureDockerAvailable(dryRun, verbose bool) error {
	if err := docker.EnsureInstalled(dryRun, verbose); err != nil {
		return err
	}
	return docker.EnsureComposeInstalled(dryRun, verbose)
}
