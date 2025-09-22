package systemd

import (
	"os"

	"github.com/spf13/cobra"
)

// Command constructs the `systemd` command group which manages installation and
// removal of the systemd service.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "systemd",
		Short: "Manage the hostship systemd service",
	}
	cmd.AddCommand(installCmd())
	cmd.AddCommand(removeCmd())
	cmd.AddCommand(statusCmd())
	return cmd
}

func installCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	c := &cobra.Command{
		Use:   "install",
		Short: "Install hostship as a systemd service",
		RunE: func(cmd *cobra.Command, args []string) error {
			bin, err := os.Executable()
			if err != nil {
				return err
			}
			return Install(bin, dryRun, verbose)
		},
	}
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print commands without executing")
	c.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return c
}

func removeCmd() *cobra.Command {
	var dryRun bool
	var verbose bool
	c := &cobra.Command{
		Use:   "remove",
		Short: "Remove the hostship systemd service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Remove(dryRun, verbose)
		},
	}
	c.Flags().BoolVar(&dryRun, "dry-run", false, "print commands without executing")
	c.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return c
}

func statusCmd() *cobra.Command {
	var verbose bool
	c := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the systemd service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Status(verbose)
		},
	}
	c.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return c
}
