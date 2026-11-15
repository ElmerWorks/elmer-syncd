# ElmerSyncd Linux Build and Run Guide

This guide explains how to clone the public repository, test and build the Go server on Linux, and run it with a directory safely selected as its file root.

## Requirements

- A Linux host (these commands use Ubuntu/Debian package commands where applicable).
- Git.
- Go 1.23 or newer. Check with `go version`.
- Network access from the Android device to the Linux host if using the server over a LAN.

Install Git and Go on Ubuntu/Debian if needed:

```sh
sudo apt update
sudo apt install git golang-go
```

If the distribution package provides a Go version older than 1.23, install a newer Go release from the official Go downloads page, then check `go version` again.

## 1. Clone the repository

```sh
cd ~
git clone https://github.com/ElmerWorks/elmer-syncd.git
cd elmer-syncd
```

To update a previously cloned copy later:

```sh
cd ~/elmer-syncd
git pull
```

## 2. Run the tests

From the repository root:

```sh
go test ./...
```

## 3. Build the Linux executable

Build a native Linux executable with CGO disabled:

```sh
CGO_ENABLED=0 go build -o elmer-syncd ./cmd/elmer-syncd
```

The executable is `./elmer-syncd` (Linux binaries normally do not use the `.exe` extension). Confirm that it was created:

```sh
file ./elmer-syncd
```

Optionally install it system-wide after building:

```sh
sudo install -m 0755 ./elmer-syncd /usr/local/bin/elmer-syncd
```

If installed system-wide, run `/usr/local/bin/elmer-syncd` in place of `./elmer-syncd` below.

## 4. Choose and prepare the server root

The server only intends to serve files beneath its configured root. Create a dedicated directory for files that the Android client may access; do not use `/` or a directory containing private files.

```sh
mkdir -p "$HOME/ElmerSync"
```

The server process user must have permission to read and write the root and its contents.

## 5. Start the server

For local-only testing on the Linux host, bind to the default loopback address:

```sh
./elmer-syncd --root "$HOME/ElmerSync" --port 12345 --token 'replace-with-a-long-random-token'
```

For Android/LAN access, explicitly allow LAN connections. This binds the service to all network interfaces, so use a non-empty token and allow access only from trusted networks:

```sh
./elmer-syncd --root "$HOME/ElmerSync" --port 12345 --token 'replace-with-a-long-random-token' --allow-lan
```

Keep the terminal open while the server is running. Press `Ctrl+C` to stop it. Do not expose this service directly to the public internet. If a firewall is enabled, allow TCP port 12345 only from the trusted LAN as appropriate for that firewall.

To print LAN connection examples and then exit:

```sh
./elmer-syncd --port 12345 --show-ip
```

Use the Linux host's reachable LAN IP address in the Android app, together with port `12345` and the same token passed to `--token`.

## 6. Check the HTTP endpoint

While the server is running, open a second terminal on the Linux host and request a listing of the configured root:

```sh
curl -H 'X-API-Token: replace-with-a-long-random-token' \
  'http://127.0.0.1:12345/api/list?path=.'
```

For a LAN-bound server, this request still works locally. The current server listing response is a JSON object with `path` and `entries`. The implemented file endpoints are `GET /api/read?path=...` and `POST /api/write`; use those paths rather than the older `/api/file` paths mentioned in some repository notes.

## 7. Build a Linux binary on another operating system

If compiling from Windows or macOS with Go installed, cross-compile for Linux amd64 from the repository root:

```sh
# PowerShell
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o elmer-syncd-linux ./cmd/elmer-syncd
```

Copy `elmer-syncd-linux` to the Linux host, then run it there as described above. This example targets 64-bit x86 Linux (`amd64`). For an ARM Linux device, build with `GOARCH=arm64` instead.

If the transfer method did not preserve executable permissions, enable them on the Linux host:

```sh
chmod +x ./elmer-syncd-linux
```

## Notes

- The default bind address is `127.0.0.1`, which is not reachable from other devices. Use `--allow-lan` for LAN access.
- `--token` is optional in the program, but should be set whenever the host is reachable over a network.
- Replace the example token with a unique, hard-to-guess value. Do not commit the real token. Command-line arguments may be visible to other users on the same host, so use a dedicated account and restrict host access.
- Treat the server root as an access boundary and keep it limited to files you intend to share. Avoid symlinks to data outside the root.
- The repository's checked-in Linux systemd unit is an example, not a ready-to-use secure service configuration; it contains placeholder paths and a sample token. Do not enable it as-is. For a persistent service, configure a dedicated restricted service account and keep the real token out of public source control.
