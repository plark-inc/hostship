package hotreload

import (
	"github.com/spf13/cobra"
)

// Command constructs the `hotreload` subcommand which only runs the hot-reload
// listener without starting the container.
func Command() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:    "hotreload",
		Short:  "Run only the hot-reload listener",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartUpdateServer(verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return cmd
}
