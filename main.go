// Command hostship is a small CLI tool that manages running a single Docker
// service defined in a JSON configuration file. It can install itself as a
// systemd service, start the container, and run an HTTP listener that allows
// the service to be updated at runtime.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/plark-inc/hostship/hotreload"
	"github.com/plark-inc/hostship/logs"
	"github.com/plark-inc/hostship/selfupdate"
	"github.com/plark-inc/hostship/setup"
	"github.com/plark-inc/hostship/start"
	"github.com/plark-inc/hostship/systemd"
)

// version is injected at build time using -ldflags. It is displayed when
// running the application with the -v flag.
var version = "development"

// channel indicates whether this build is a "prod" or "dev" release. It is
// also injected at build time using -ldflags and defaults to "dev".
var channel = "dev"

// main wires together the CLI commands using cobra and executes the root
// command. It parses the global flags and delegates to the subcommands for any
// actual work.
func main() {
	var showVersion bool

	root := &cobra.Command{
		Use:          "hostship",
		Short:        "Docker Service Manager",
		SilenceUsage: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if showVersion {
				fmt.Printf("%s %s\n", channel, version)
				os.Exit(0)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.Flags().BoolVarP(&showVersion, "version", "v", false, "print version and exit")

	root.AddCommand(systemd.Command())
	root.AddCommand(setup.Command())
	root.AddCommand(start.Command())
	root.AddCommand(hotreload.Command())
	root.AddCommand(logs.Command())
	root.AddCommand(selfupdate.Command(&version, &channel))

	// Hide the default 'help' subcommand to keep the usage output concise.
	root.SetHelpCommand(&cobra.Command{Use: "no-help", Hidden: true})

	root.CompletionOptions.DisableDefaultCmd = true

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
