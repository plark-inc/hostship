package selfupdate

import (
	"github.com/spf13/cobra"
)

// Command returns a cobra command that updates the hostship executable.
// It keeps the update channel ("prod" or "dev") consistent with the current
// binary so updates do not switch environments.
func Command(current, channel *string) *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for updates and replace the hostship binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Update(*current, *channel, verbose)
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return cmd
}
