#!/usr/bin/env bash
# Make sure this file uses LF (Unix) line endings, not CRLF.
# In VS Code: check the bottom-right corner → change CRLF → LF.
set -euo pipefail

# This script downloads the latest binary and installs it to
# /usr/local/bin.

if [ "$(id -u)" != "0" ]; then
	echo "This script must be run as root" >&2
	exit 1
fi

# check if is running inside a container
if [ -f /.dockerenv ]; then
	echo "This script cannot run within docker" >&2
	exit 1
fi

# check if something is running on port 8080
if ss -tulnp | grep ':8080 ' >/dev/null; then
	echo "Error: something is already running on port 8080" >&2
	exit 1
fi

# Get machine hardware architecture (e.g., x86_64, arm64, aarch64)
arch=$(uname -m)

# Normalize architecture name to match release file naming conventions
case "$arch" in
  x86_64) arch=amd64 ;;          # Map x86_64 to amd64
  arm64 | aarch64) arch=arm64 ;; # Map arm64/aarch64 to arm64
  *)
    echo "Unsupported architecture: $arch"
    exit 1
    ;;
esac

# Get operating system name and convert it to lowercase (e.g., Linux -> linux, Darwin -> darwin)
os=$(uname | tr '[:upper:]' '[:lower:]')

channel=${1:-prod}
if [[ "$channel" != "prod" && "$channel" != "dev" ]]; then
	echo "Invalid channel: $channel (must be 'prod' or 'dev')" >&2
	exit 1
fi

file="hostship_${os}_${arch}.tar.gz"              # archive name on the server
url="https://cli.hostship.com/${channel}/${file}" # download location

echo "Downloading hostship from $url"

# fetch the archive
curl -sL "$url" -o "$file" 

# extract the binary
tar -xzf "$file" hostship             

# place it in PATH
sudo mv hostship /usr/local/bin/hostship 
sudo chmod +x /usr/local/bin/hostship

# clean up
rm "$file" 

echo "hostship installed to /usr/local/bin/hostship"

# You may extend this script, to add commands to install your app alongside the CLI 
# echo "Setting up docker and compose config"
# hostship setup

# echo "Installing systemd service"
# hostship systemd install

# echo "Starting the application"
# hostship start



