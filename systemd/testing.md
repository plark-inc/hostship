# Testing the Hostship systemd service

After running `hostship systemd install`, you can verify the service is active with:

```bash
hostship systemd status
```

If the service is running, it listens on port 8080. You can trigger an update by including the deployment key in the URL:

```bash
curl -X POST http://172.17.0.1:8080/update/<KEY>
```

Replace `<KEY>` with the value stored in the .env file.

The installed unit executes `hostship hotreload` so the update listener starts automatically on boot.


