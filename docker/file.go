// Package docker wraps reading and updating the Docker Compose JSON configuration
// used by the CLI. Service orchestration itself is implemented in compose.go.
package docker

import (
	"fmt"
	"os"

	"github.com/tidwall/gjson"
)

// Load reads and validates the compose JSON file, returning the raw bytes.
func Load(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	if !gjson.GetBytes(data, "services").Exists() {
		return nil, fmt.Errorf("compose file must define services")
	}
	return data, nil
}

// Save writes the given JSON bytes back to disk.
func Save(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// GetString retrieves a string value from the JSON using dot notation.
func GetString(data []byte, path string) string { return gjson.GetBytes(data, path).String() }

// ServiceNames returns the names of all services defined in the compose file
// in the order they appear.
func ServiceNames(data []byte) ([]string, error) {
	svcs := gjson.GetBytes(data, "services")
	if !svcs.Exists() {
		return nil, fmt.Errorf("compose file must define services")
	}
	names := make([]string, 0)
	svcs.ForEach(func(key, value gjson.Result) bool {
		names = append(names, key.String())
		return true
	})
	if len(names) == 0 {
		return nil, fmt.Errorf("compose file must define services")
	}
	return names, nil
}
