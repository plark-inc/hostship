# Hostship

A Go-based CLI tool to run an application defined in a Docker Compose JSON file. 
It installs Docker, runs the containers, and exposes an endpoint to update the  compose services on request. 

Curious about the motivations behind Hostship? Check out the full story in the blog post: [Why I built Hostship](https://plark.com/hostship-dokku-alternative)
 
## Prerequisites

* Go 1.24+ installed

## Docker Compose definition 
First, define a single compose.json file and upload it to an S3 bucket. The file at the S3 URL points to the latest version. When an update is triggered, the CLI fetches this file and overrides the local one.

```json
{
  "x-metadata": {
    "url": "https://cli.plark.com/compose.json",
    "version": "0.9.8"
  },
  "services": {
    "app": {
      "image": "ghcr.io/myorg/hostship:latest",
      "restart": "unless-stopped",
      "environment": {
        "LOG_LEVEL": "debug",
        "MAX_CONN": "10"
      }
    }
  }
}
```


## The CLI

```Shell
hostship setup <compose-url>
```

- Installs Docker (if missing).
- Downloads compose.json
- Writes a `.env` that includes an DEPLOY_URL (ex: `DEPLOY_URL=http://172.17.0.1:8080/update/<KEY>`). Hitting this endpoint will update your compose.json file to the latest version.


```Shell
hostship start
```
- Starts the service

```Shell
hostship hotreload
```
- Runs the HTTP listener to trigger updates.
- The server listener validates the key before updating.

```Shell
hostship systemd install
```
- To ensure the service listener runs in the background and persists across reboots, this configures a systemd service.

## Usage
```bash
# Show help
hostship -h

# Print the version
hostship -v

# Run setup with a compose file
hostship setup https://example.com/compose.json

# Start the container
hostship start

# Dry-run
hostship start --dry-run

# Verbose logging
hostship start --verbose

# Hot-reload service listener (hidden command)
hostship hotreload --verbose

# Installs the hotreload as a systems service
hostship systemd install

# Removes the systemd service
hostship systemd remove

# View live logs of your service
hostship logs caddy
```


## Installing the CLI

A shell script is provided to download the latest CLI and installs `hostship` binary to `/usr/local/bin`. 

```bash
bash <(curl -fsSL https://cli.hostship.com/install.sh)
```

Once installed, you can run `hostship update` at any time to update the CLI.

## Releasing

This project uses [goreleaser](https://goreleaser.com/) to build binaries for Linux and upload them to an S3 bucket.

1. Install goreleaser [installation-guide](https://goreleaser.com/install/#npm).
2. Create a .env file and at minimum set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` for an IAM user with write access to the bucket. Optionally set `AWS_REGION`. The S3 endpoint and bucket name are configured in `.goreleaser.yaml`.

3. Run the release command specifying the version number, and the target channel (`prod` or `dev`):

   ```bash
   ./scripts/release.sh 1.0.0 prod
   ```

GoReleaser uploads each release to either the `prod/` or `dev/` directory along with a `metadata.json` file that the `hostship update` command reads to determine the newest version.
