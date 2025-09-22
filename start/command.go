package start

import (
	"github.com/plark-inc/hostship/setup"
	"github.com/spf13/cobra"
)

// Command constructs the `start` subcommand which starts the Docker services
// defined in the compose configuration.
func Command() *cobra.Command {
	var dryRun bool
	var verbose bool
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Docker compose services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return setup.StartService(dryRun, verbose)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print commands without executing")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return cmd
}
