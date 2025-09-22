package selfupdate

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/plark-inc/hostship/systemd"
)

const baseURL = "https://cli.hostship.com"

// releaseInfo represents the JSON describing the latest release.
type releaseInfo struct {
	Version string `json:"version"`
}

// Update checks the latest release info for the given channel and replaces the
// current executable if a newer version is available.
func Update(current, channel string, verbose bool) error {
	infoURL := fmt.Sprintf("%s/%s/metadata.json", baseURL, channel)
	if verbose {
		fmt.Printf("fetching release info from %s\n", infoURL)
	}
	info, err := fetchInfo(infoURL)
	if err != nil {
		return err
	}

	curV, err := semver.NewVersion(current)
	if err != nil {
		curV = &semver.Version{}
	}
	latestV, err := semver.NewVersion(info.Version)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("current version: %s, latest: %s\n", curV, latestV)
	}
	if !latestV.GreaterThan(curV) {
		fmt.Println("hostship is up to date")
		return nil
	}

	file := fmt.Sprintf("hostship_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	url := fmt.Sprintf("%s/%s/%s", baseURL, channel, file)

	tmpBin, err := downloadBinary(url, verbose)
	if err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := replaceBinary(exe, tmpBin, verbose); err != nil {
		return err
	}
	reinstallServiceIfActive(exe, verbose)
	return nil
}

func downloadBinary(url string, verbose bool) (string, error) {
	tmpArchive := filepath.Join(os.TempDir(), filepath.Base(url))
	if verbose {
		fmt.Printf("downloading %s to %s\n", url, tmpArchive)
	}
	if err := download(url, tmpArchive); err != nil {
		return "", err
	}
	tmpBin := filepath.Join(os.TempDir(), "hostship.new")
	if verbose {
		fmt.Printf("extracting binary to %s\n", tmpBin)
	}
	if err := extractBinary(tmpArchive, tmpBin); err != nil {
		return "", err
	}
	if err := os.Chmod(tmpBin, 0755); err != nil {
		return "", err
	}
	return tmpBin, nil
}

func replaceBinary(exe, newBin string, verbose bool) error {
	backup := exe + ".old"
	_ = os.Remove(backup)
	if verbose {
		fmt.Printf("replacing %s (backup %s)\n", exe, backup)
	}
	if err := os.Rename(exe, backup); err != nil {
		return err
	}
	if err := os.Rename(newBin, exe); err != nil {
		_ = os.Rename(backup, exe)
		return err
	}

	if verbose {
		fmt.Printf("running %s -v\n", exe)
	}
	out, err := exec.Command(exe, "-v").CombinedOutput()
	if err != nil {
		_ = os.Rename(backup, exe)
		return fmt.Errorf("verification failed: %v: %s", err, out)
	}
	fmt.Printf("updated %s to %s\n", exe, strings.TrimSpace(string(out)))
	return nil
}

func fetchInfo(infoURL string) (*releaseInfo, error) {
	req, err := http.NewRequest("GET", infoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", infoURL, resp.Status)
	}
	var info releaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}
	return &info, nil
}

func download(url, dest string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", url, resp.Status)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func extractBinary(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		name := filepath.Base(hdr.Name)
		if (hdr.Typeflag == tar.TypeReg || hdr.Typeflag == tar.TypeRegA) && (name == "hostship" || name == "hostship.exe") {
			out, err := os.Create(dest)
			if err != nil {
				return err
			}
			defer out.Close()
			if _, err := io.Copy(out, tr); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("binary not found in archive")
}

// reinstallServiceIfActive checks if the hostship systemd service is active and
// re-installs it so the updated binary takes effect. Any errors are ignored.
func reinstallServiceIfActive(bin string, verbose bool) {
	if _, err := exec.LookPath("systemctl"); err != nil {
		if verbose {
			fmt.Println("systemctl not found; skipping service reinstall")
		}
		return
	}
	check := exec.Command("systemctl", "is-active", "--quiet", "hostship")
	if err := check.Run(); err != nil {
		if verbose {
			fmt.Println("hostship service not active; skipping reinstall")
		}
		return
	}
	if err := systemd.Remove(false, verbose); err != nil && verbose {
		fmt.Printf("failed to remove service: %v\n", err)
	}
	if err := systemd.Install(bin, false, verbose); err != nil && verbose {
		fmt.Printf("failed to install service: %v\n", err)
	}
}
