# ElmerSyncd Linux Go Build and Run Notes

This guide is a general, configuration-aware procedure for installing Go, building ElmerSyncd, launching the server, and checking access from an Android device on the same local network. It is also intended as a hand-off document that can be supplied to ChatGPT or another technical assistant when help is needed on a particular Linux distribution.

The examples use a Debian/Ubuntu-style Linux host and the repository at `~/Code/elmer-syncd`. Adjust paths and package-manager commands for other distributions.

> **Important:** Building the executable and running the server are separate steps. A successful build does not mean that the server is running, and a server bound only to `127.0.0.1` cannot accept connections from Android or other computers.

## 1. Requirements and repository layout

Requirements:

- A Linux computer.
- Git.
- Go 1.23.0 or newer (the module currently declares `go 1.23`).
- Network access between the Linux host and Android device for LAN use.

The main executable package is in:

```text
cmd/elmer-syncd/main.go
```

Other Go source files are under `internal/`. Consequently, running plain `go build` from the repository root may report `no Go files`; that does not mean the source is missing. The build command must identify the main package under `cmd/elmer-syncd`.

## 2. Install or verify Go

### 2.1 Check whether Go is already installed

Run:

```sh
go version
go env GOTOOLCHAIN
which go
```

If `go version` itself fails while trying to download a toolchain, inspect the locally installed compiler without automatic toolchain switching:

```sh
GOTOOLCHAIN=local go version
GOTOOLCHAIN=local go env GOVERSION GOROOT
```

The repository's `go.mod` currently starts with:

```go
go 1.23
```

This declares the minimum Go language/toolchain version required by the module. Use Go 1.23.0 or later. A newer compatible Go release is generally preferable to installing an old release solely to match the minimum version.

### 2.2 Debian/Ubuntu package installation

The distribution package may be older than the required version. It is fine to check the package version, but do not assume `apt install golang-go` supplies a sufficiently recent Go release:

```sh
sudo apt update
sudo apt install git golang-go
 go version
```

If the installed version is older than 1.23, install a current Go release from the official Go downloads page instead:

<https://go.dev/dl/>

Choose the Linux archive for the host architecture (commonly `linux-amd64` for 64-bit Intel/AMD PCs, or `linux-arm64` for 64-bit ARM systems). Follow the current official installation instructions. Avoid mixing multiple Go installations in `PATH`.

A common official tarball installation uses `/usr/local/go`, but first check how Go was installed. Do not blindly overwrite a package-managed installation. If replacing an existing manual `/usr/local/go` installation, remove that directory before extracting the new archive, as directed by the official instructions. Ensure `/usr/local/go/bin` is on `PATH`, then open a new shell or reload the shell configuration.

Verify the selected compiler and its location:

```sh
which go
go version
go env GOTOOLCHAIN
```

If multiple Go installations exist, `which go` and `type -a go` help identify which one the shell is invoking:

```sh
type -a go
```

### 2.3 Understanding the `toolchain not available` error

An error such as:

```text
go: downloading go1.23 (linux/amd64)
go: download go1.23 for linux/amd64: toolchain not available
```

usually means that the `go` command being invoked is attempting automatic toolchain selection/download, but cannot obtain the toolchain it needs. Common causes include an old distribution-provided Go installation, a `PATH` that selects a different Go installation than expected, or a failed/unavailable toolchain download.

Check the actual local compiler with `GOTOOLCHAIN=local go version`, and check `which go` / `type -a go`. The durable fix is normally to install and select Go 1.23.0 or newer from a trusted source. `GOTOOLCHAIN=auto` permits automatic toolchain selection; it does not make an obsolete local Go installation satisfy the module requirement if the requested toolchain cannot be downloaded.

Do not change the module's Go version merely to suppress the error. The `go 1.23` line records the project's requirement. Change it only if the project maintainer has deliberately changed the supported minimum and verified the code with that older version.

## 3. Obtain the repository

Clone the public repository, or use the existing checkout:

```sh
mkdir -p ~/Code
cd ~/Code
git clone https://github.com/ElmerWorks/elmer-syncd.git
cd ~/Code/elmer-syncd
```

For an existing checkout:

```sh
cd ~/Code/elmer-syncd
git status
git pull
```

If the repository is already downloaded elsewhere, use its actual path. Do not clone a second copy unless that is intended.

Confirm the main package and module file exist:

```sh
ls -l go.mod cmd/elmer-syncd/main.go
find . -name '*.go' -type f
```

## 4. Test and build

Run the module tests from the repository root:

```sh
go test ./...
```

Build the Linux executable:

```sh
CGO_ENABLED=0 go build -o elmer-syncd ./cmd/elmer-syncd
```

Why this command:

- `CGO_ENABLED=0` requests a build without CGO, avoiding a dependency on a C toolchain for this build configuration.
- `-o elmer-syncd` names the output executable. Linux executables normally do not need an `.exe` suffix.
- `./cmd/elmer-syncd` identifies the actual `main` package. Plain `go build` at the repository root is not appropriate because the root contains no Go source files.

Verify the result:

```sh
ls -l ./elmer-syncd
file ./elmer-syncd
```

The file should be an executable Linux ELF binary for the host architecture. If it is not executable, run:

```sh
chmod +x ./elmer-syncd
```

An optional system-wide installation is:

```sh
sudo install -m 0755 ./elmer-syncd /usr/local/bin/elmer-syncd
```

If installed there, use `/usr/local/bin/elmer-syncd` in the launch commands below. Otherwise, run `./elmer-syncd` from the repository directory.

## 5. Prepare a dedicated server root

Choose a directory containing only files that the Android client is intended to access. Do not use `/`, the whole home directory, or a directory containing unrelated private files.

For example:

```sh
mkdir -p "$HOME/ElmerSync"
```

The Linux account running ElmerSyncd must have the necessary read/write permissions for this directory and its contents. Keep the server root narrow and avoid symlinks that lead outside it.

## 6. Understand the API token

The server accepts a command-line option named `--token`. The Android client sends the matching secret with each authenticated API request in an HTTP header named:

```text
X-API-Token
```

The name **API token** distinguishes this value from a Linux login password, Wi-Fi password, or a general account credential. It is a shared secret used by the ElmerSyncd HTTP API to recognize requests from an authorized client. The `X-API-Token` header makes the purpose explicit: this is the token for access to the application's API. It is not an OAuth access token and should not be sent as a URL query parameter.

The value configured with `--token` on the server must match the token configured in DOSwriter. The header name is fixed by the API; the token value is chosen by the operator.

Generate a strong random token for network use:

```sh
openssl rand -hex 32
```

Copy the generated value into the server launch command and the Android application's connection settings. Do not use the example token below for a real network. Do not commit a real token to Git, put it in screenshots, or share it in support messages. Command-line arguments may be visible to other local users through process-inspection tools, and shell commands may be retained in shell history; on a multi-user or sensitive host, consider a more protected service configuration rather than treating command-line arguments as a secret store.

## 7. Launch the server

### 7.1 Local-only test

For a test from the Linux host itself, omit `--allow-lan`:

```sh
./elmer-syncd \
  --root "$HOME/ElmerSync" \
  --port 12345 \
  --token 'replace-with-a-strong-random-token'
```

The default bind address is `127.0.0.1`. This is loopback-only and is not reachable from Android or another computer.

### 7.2 Android access over the local network

Stop a currently running instance with `Ctrl+C`, then start it with `--allow-lan`:

```sh
./elmer-syncd \
  --root "$HOME/ElmerSync" \
  --port 12345 \
  --token 'replace-with-a-strong-random-token' \
  --allow-lan
```

`--allow-lan` changes the bind behavior so the server listens on the host's network interfaces rather than only on loopback. Keep the terminal open while using the server. Press `Ctrl+C` to stop it.

Do not expose this service directly to the public internet or configure router port forwarding for it. Use it on a trusted local network with a strong, private token.

To print LAN connection examples and exit, the program also supports:

```sh
./elmer-syncd --port 12345 --show-ip
```

Use the actual LAN address reachable by the Android device. The address may change when the laptop reconnects to Wi-Fi or receives a new DHCP lease.

### 7.3 Persistent/background launch varies by Linux configuration

The simplest and most portable initial approach is to run the executable in a terminal and leave that terminal open. This makes startup errors visible and is recommended while testing.

For ongoing use, Linux distributions may use different service managers and conventions. On a systemd-based distribution, a systemd service can be created after the executable path, service account, server root, token handling, working directory, and network policy have been chosen. The repository's checked-in `packaging/linux/elmer-syncd.service` is an example, not a ready-to-enable production service: it contains placeholder paths and a sample token. Do not enable it unchanged. Use a dedicated restricted service account where practical, grant it access only to the intended server root, and keep the real token out of source control. On non-systemd distributions, use the service manager appropriate to that system.

A ChatGPT assistant helping with a particular host should first identify the distribution/version, init or service manager, Go installation method, executable location, desired run user, and whether access is local-only or LAN. It should then tailor the launch/service steps rather than assume every Linux host is configured alike.

## 8. Verify the server is listening

In a second terminal, run:

```sh
ss -ltnp | grep 12345
```

Typical local-only output includes:

```text
127.0.0.1:12345
```

Typical LAN-enabled output may show:

```text
*:12345
```

or an equivalent wildcard address such as `0.0.0.0:12345` (and possibly an IPv6 wildcard, depending on the host and Go runtime).

Interpretation:

- `127.0.0.1:12345`: local connections only; Android cannot connect using the laptop's LAN IP.
- `*:12345` or `0.0.0.0:12345`: listening on all IPv4 interfaces / wildcard interfaces; proceed to firewall and network checks.
- No output: the process is not listening on that port. Check the server terminal for startup errors, verify the port, and confirm the process is running.

## 9. Test the HTTP API from Linux

First test loopback if the server is running locally:

```sh
curl -v \
  -H 'X-API-Token: replace-with-the-same-token-used-at-launch' \
  'http://127.0.0.1:12345/api/list?path=.'
```

For LAN-enabled testing, find the host's LAN address:

```sh
hostname -I
```

This may print more than one address. Select the address on the same network as the Android device (often an address beginning with `192.168.`, `10.`, or `172.16` through `172.31`). Confirm the correct interface and route with:

```sh
ip -br address
ip route
```

Then test using that address, for example:

```sh
curl -v \
  -H 'X-API-Token: replace-with-the-same-token-used-at-launch' \
  'http://192.168.1.110:12345/api/list?path=.'
```

A successful response resembles:

```json
{"entries":[],"path":"."}
```

An empty `entries` array is normal if the selected server root is empty. An HTTP `200 OK` confirms that the host can reach the server and the supplied token was accepted.

## 10. Firewall and network checks

Firewall configuration depends on the Linux distribution. Do not assume UFW is installed or active. Check the applicable firewall manager. On Ubuntu/Debian systems using UFW:

```sh
sudo ufw status
```

If UFW is active, allow TCP port 12345 only from the trusted LAN subnet where possible. For example, if the actual LAN subnet is `192.168.1.0/24`:

```sh
sudo ufw allow from 192.168.1.0/24 to any port 12345 proto tcp
```

Do not copy that subnet without checking the host's actual network. If UFW is inactive, it is not the cause of a UFW block, but another firewall, router policy, VPN, or Wi-Fi isolation setting may still matter.

For Android-to-Linux access, verify:

- Android and Linux host are connected to the same trusted LAN/Wi-Fi network.
- Android is not on a guest Wi-Fi network that blocks access to other clients.
- The access point/router does not have client isolation or AP isolation enabled.
- A VPN or security app is not preventing local-network access.
- The server is running with `--allow-lan` and `ss` shows a wildcard/LAN listening address.
- The Android device is using the current laptop LAN IP, not `127.0.0.1` and not an old DHCP address.
- The firewall permits inbound TCP connections to port 12345 from the Android device's subnet.

## 11. Test from Android

With the server running in LAN mode, open a browser on Android and visit:

```text
http://<LINUX-LAN-IP>:12345/api/list?path=.
```

For example:

```text
http://192.168.1.110:12345/api/list?path=.
```

A browser address bar does not normally send the custom `X-API-Token` header. Therefore, an authentication response such as HTTP 401 can still demonstrate that Android reached the server; it does not prove the token is wrong in DOSwriter. A browser connection timeout or connection-refused message indicates a network/listener problem instead.

In DOSwriter's remote-sync configuration, use:

| Setting | Value |
|---|---|
| Host / server address | Linux host's current LAN IP, e.g. `192.168.1.110` |
| Port | `12345` |
| Protocol | `HTTP` |
| API token | The exact value supplied to the server's `--token` option |

Do not enter `127.0.0.1` as the host on Android: on Android, that address refers to the Android device itself.

## 12. API endpoints and request details

The current server implementation uses these endpoints:

- `GET /api/list?path=...` — list entries under a path in the configured root.
- `GET /api/read?path=...` — read a file.
- `POST /api/write` — write a file.

Requests are authenticated with the `X-API-Token` HTTP header. Use these current endpoints rather than older `/api/file` paths that may appear in earlier design notes.

For a simple list test from Linux:

```sh
curl -H 'X-API-Token: YOUR_TOKEN' \
  'http://127.0.0.1:12345/api/list?path=.'
```

Replace the address with the Linux host's LAN IP when testing the LAN interface.

## 13. Troubleshooting guide

| Symptom | Likely cause / next check |
|---|---|
| `go: downloading go1.23 ... toolchain not available` | Check `GOTOOLCHAIN=local go version`, `which go`, and `type -a go`; install/select Go 1.23.0 or newer. Check network/proxy access if automatic download is expected. |
| `no Go files in .../elmer-syncd` | Plain `go build` was run at the repository root. Build the main package with `CGO_ENABLED=0 go build -o elmer-syncd ./cmd/elmer-syncd`. |
| Build reports missing packages or module errors | From the repository root, check `go version`, network access, and `go.mod`; try `go mod download`, then `go test ./...` and the documented build command. Do not delete or casually rewrite `go.mod` / `go.sum`. |
| `ss` shows no listener | Server is not running, failed at startup, or is using a different port. Read the server terminal output. |
| `ss` shows `127.0.0.1:12345` | Server was started without `--allow-lan`. Stop it and restart with that option for Android access. |
| `curl` to `127.0.0.1` fails | Check whether the server is running, the port, startup errors, and the token. |
| `curl` to the laptop's LAN IP succeeds but Android cannot connect | Check same Wi-Fi/LAN, guest-network restrictions, AP/client isolation, VPN, correct current IP, and inbound firewall policy. |
| Android browser reaches server but reports unauthorized | Expected if the browser did not send `X-API-Token`; use DOSwriter or a test client that sends the header. |
| Android reports unauthorized | Confirm the Android token exactly matches the server `--token` value; check for leading/trailing spaces and ensure the server was restarted with the intended token. |
| Android connection times out | Usually network path, firewall, guest Wi-Fi/client isolation, VPN, wrong IP, or server not listening on LAN. |
| Connection refused | Usually no process listening at that address/port, or the server is bound only to loopback. |
| HTTP 404 | The host was reached, but the requested path may not be implemented. Check that the current endpoint is `/api/list`, `/api/read`, or `/api/write` as appropriate. |

## 14. Useful diagnostic bundle for support

When asking ChatGPT or another assistant for help, provide the Linux distribution/version, CPU architecture, exact command and error, and the output of the following commands. Remove or redact real tokens, private IP details if desired, and any sensitive paths before sharing.

```sh
cat /etc/os-release
uname -m
which go
type -a go
GOTOOLCHAIN=local go version
go env GOTOOLCHAIN
cd /path/to/elmer-syncd && sed -n '1,40p' go.mod
cd /path/to/elmer-syncd && find . -name '*.go' -type f
cd /path/to/elmer-syncd && go test ./...
cd /path/to/elmer-syncd && ls -l ./elmer-syncd
ss -ltnp | grep 12345 || true
hostname -I
ip -br address
ip route
```

Also report:

- Whether the server is started manually, by systemd, or by another service manager.
- The exact server launch command with the token replaced by `<REDACTED>`.
- Whether the goal is local-only access or Android/LAN access.
- The result of a `curl -v` request from the Linux host to loopback and, if LAN mode is intended, to the host's LAN IP. Redact the token from the command/output.
- What Android reports: timeout, connection refused, authentication failure, HTTP status, or successful response.

## 15. Quick-start checklist

For a typical 64-bit Ubuntu/Debian laptop that already has Go 1.23.0 or newer:

```sh
# Enter the repository
cd ~/Code/elmer-syncd

# Confirm Go version
go version

# Test and build
go test ./...
CGO_ENABLED=0 go build -o elmer-syncd ./cmd/elmer-syncd

# Prepare a dedicated file root
mkdir -p "$HOME/ElmerSync"

# Generate a strong token separately and use the same value in DOSwriter
openssl rand -hex 32

# Start for Android/LAN access (replace the example token)
./elmer-syncd --root "$HOME/ElmerSync" --port 12345 \
  --token 'replace-with-a-strong-random-token' --allow-lan
```

In another terminal:

```sh
ss -ltnp | grep 12345
hostname -I
```

Confirm a wildcard listener, then test the API using the laptop's actual LAN IP and the same token:

```sh
curl -v -H 'X-API-Token: YOUR_TOKEN' \
  'http://LINUX-LAN-IP:12345/api/list?path=.'
```

Configure DOSwriter with that LAN IP, port `12345`, HTTP, and the matching API token. Keep the service on a trusted LAN; do not expose the port to the public internet.
