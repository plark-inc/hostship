// This file contains the implementation of the `setup` subcommand. The command
// ensures Docker and Docker Compose are installed and retrieves the compose
// file from the provided URL, overwriting any existing configuration.
package setup

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/plark-inc/hostship/config"
	"github.com/plark-inc/hostship/docker"
	"github.com/spf13/cobra"
)

// defaultComposeURL is the location of the default compose configuration.
const defaultComposeURL = "https://cli.plark.com/compose.json"

// Command constructs the `setup` subcommand. It reads the environment variables
// defined above from the current process, saves them into the configuration
// file, and then launches the Docker service. The --dry-run and --verbose flags
// behave the same as in other commands.
func Command() *cobra.Command {
	var dryRun bool
	var verbose bool
	cmd := &cobra.Command{
		Use:   "setup [compose_url]",
		Short: "Install Docker and download the compose configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("expected at most one compose URL")
			}
			composeURL := defaultComposeURL
			if len(args) == 1 {
				composeURL = args[0]
			}
			return runSetup(dryRun, verbose, composeURL)
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print commands without executing")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	return cmd
}

// runSetup installs Docker if required and downloads the compose file,
// overwriting any existing configuration.
func runSetup(dryRun, verbose bool, composeURL string) error {
	cfgPath := config.Path
	if verbose {
		fmt.Printf("downloading compose file to %s\n", cfgPath)
	}
	resp, err := http.Get(composeURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch compose file: %s", resp.Status)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return err
	}
	if err := docker.EnsureInstalled(dryRun, verbose); err != nil {
		return err
	}
	if err := docker.EnsureComposeInstalled(dryRun, verbose); err != nil {
		return err
	}
	if _, err := os.Stat(".env"); errors.Is(err, os.ErrNotExist) {
		id := uuid.New().String()
		deployURL := fmt.Sprintf("http://172.17.0.1:8080/update/%s", id)
		content := fmt.Sprintf("DEPLOY_URL=%s\n", deployURL)
		if verbose {
			fmt.Printf("creating .env with DEPLOY_URL=%s\n", deployURL)
		}
		if err := os.WriteFile(".env", []byte(content), 0600); err != nil {
			return err
		}
	}
	return nil
}
