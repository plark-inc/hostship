package docker

import (
	"fmt"
	"os/exec"
	"strings"
)

// Runner executes shell commands while respecting dry-run and verbose modes.
type Runner struct {
	DryRun  bool
	Verbose bool
}

// New creates a new Runner instance.
func NewRunner(dryRun, verbose bool) Runner {
	return Runner{DryRun: dryRun, Verbose: verbose}
}

// Run executes the provided command and returns any error including stderr
// output. When verbose or dry-run mode is enabled the command is printed.
func (r Runner) Run(cmd *exec.Cmd) error {
	if r.Verbose || r.DryRun {
		fmt.Println(cmd.String())
	}
	if r.DryRun {
		return nil
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %v: %s", cmd.Path, err, string(out))
	}
	return nil
}

// Output executes the command and returns its trimmed stdout. Behavior for
// verbose and dry-run modes matches Run().
func (r Runner) Output(cmd *exec.Cmd) (string, error) {
	if r.Verbose || r.DryRun {
		fmt.Println(cmd.String())
	}
	if r.DryRun {
		return "", nil
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %v: %s", cmd.Path, err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}
