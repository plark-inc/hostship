// Package hotreload implements the hot-reload HTTP server that listens for
// update requests and restarts the Docker service when triggered.
package hotreload

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/plark-inc/hostship/config"
	"github.com/plark-inc/hostship/docker"
)

// Load the configuration and starts the hot-reload HTTP server.
// Docker must already be installed and the container running.
func StartUpdateServer(verbose bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	upd := New(docker.NewComposeClient(false, verbose), verbose)
	return upd.Start(ctx, config.Path)
}

type Updater struct {
	compose *docker.ComposeClient
	file    string
	verbose bool
}

// New creates a new Updater instance using the provided Docker compose client.
// The returned updater is ready to be started.
func New(c *docker.ComposeClient, verbose bool) *Updater {
	return &Updater{
		compose: c,
		verbose: verbose,
	}
}

// Start launches the update HTTP server on port 8080
func (u *Updater) Start(ctx context.Context, cfgPath string) error {
	u.file = cfgPath

	// Start the HTTP server in a goroutine and report any error via a channel
	srv := &http.Server{Addr: ":8080", Handler: http.HandlerFunc(u.handle)}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()
	if u.verbose {
		fmt.Println("listening on :8080")
	}

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("listen error:", err)
		}
		return err
	case <-ctx.Done():
		err := srv.Shutdown(context.Background())
		if srvErr := <-errCh; srvErr != nil && srvErr != http.ErrServerClosed {
			fmt.Println("listen error:", srvErr)
			return srvErr
		}
		return err
	}
}

// handle processes incoming update requests.
func (u *Updater) handle(w http.ResponseWriter, r *http.Request) {
	if u.verbose {
		fmt.Printf("%s %s\n", r.Method, r.URL.Path)
	}
	if r.Method != http.MethodPost {
		u.unknownEndpoint(w)
		return
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 || parts[0] != "update" {
		if len(parts) == 1 && parts[0] == "update" {
			u.missingKey(w)
		} else {
			u.unknownEndpoint(w)
		}
		return
	}
	key := parts[1]
	loadEnv()
	raw := os.Getenv("DEPLOY_URL")
	if raw == "" {
		u.deployURLError(w, fmt.Errorf("DEPLOY_URL not set"))
		return
	}
	uParsed, err := url.Parse(raw)
	if err != nil {
		u.deployURLError(w, err)
		return
	}
	envParts := strings.Split(strings.Trim(uParsed.Path, "/"), "/")
	if len(envParts) != 2 || envParts[0] != "update" {
		u.deployURLError(w, fmt.Errorf("unexpected path %q", uParsed.Path))
		return
	}
	if key != envParts[1] {
		u.invalidKey(w)
		return
	}
	u.handleUpdate(w)
}

// handleUpdate downloads the compose file from the URL specified in x-metadata
// and restarts all services if updates are available.
func (u *Updater) handleUpdate(w http.ResponseWriter) {
	cfg, err := docker.Load(u.file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	url := docker.GetString(cfg, "x-metadata.url")
	if url == "" {
		http.Error(w, "missing x-metadata.url", http.StatusInternalServerError)
		return
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("%s: %s", url, resp.Status), http.StatusInternalServerError)
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := docker.ServiceNames(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := docker.Save(u.file, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err := u.compose.Pull(u.file, "hostship"); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	go func() {
		if err := u.compose.Up(u.file, "hostship"); err != nil {
			if u.verbose {
				fmt.Println("compose up:", err)
			}
		}
	}()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (u *Updater) unknownEndpoint(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "unknown endpoint"})
}

func (u *Updater) missingKey(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing key"})
}

func (u *Updater) invalidKey(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid key"})
}

func (u *Updater) deployURLError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	msg := "invalid DEPLOY_URL"
	if err != nil {
		msg = fmt.Sprintf("invalid DEPLOY_URL: %v", err)
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func loadEnv() {
	data, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if _, ok := os.LookupEnv(key); !ok {
			_ = os.Setenv(key, value)
		}
	}
}
