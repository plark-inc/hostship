package logs

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/plark-inc/hostship/config"
	"github.com/plark-inc/hostship/docker"
	"github.com/spf13/cobra"
)

// Command creates the `logs` subcommand for displaying service logs.
func Command() *cobra.Command {
	var follow bool
	cmd := &cobra.Command{
		Use:   "logs [service]",
		Short: "Show live logs for a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogs(args[0], follow)
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", true, "follow log output")
	return cmd
}

func runLogs(service string, follow bool) error {
	cfgPath := config.Path
	cfg, err := docker.Load(cfgPath)
	if err != nil {
		return err
	}
	names, err := docker.ServiceNames(cfg)
	if err != nil {
		return err
	}
	found := false
	for _, n := range names {
		if n == service {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("service %s not found", service)
	}
	container := docker.GetString(cfg, fmt.Sprintf("services.%s.container_name", service))
	if container == "" {
		container = fmt.Sprintf("hostship-%s-1", service)
	}
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	args = append(args, container)
	c := exec.Command("docker", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}
