Release process and packaging

Overview
- We use goreleaser to build cross-platform binaries and produce archives (zip/tar.gz).
- Releases are signed with GPG and uploaded to GitHub Releases.

Prerequisites
- goreleaser in PATH
- gpg configured for signing
- Inno Setup (ISCC) installed to build Windows installer

Quick local release (snapshot):
```bash
./scripts/release.sh v0.1.0
```

Create a real release (requires repo push/tag and GitHub token):
```bash
git tag v0.1.0
git push --tags
goreleaser release
```

Windows installer
- Use Inno Setup to compile `installer/ElmerSyncdInstaller.iss`. The provided PowerShell script `installer/build-windows-installer.ps1` can call ISCC.

Linux packaging
- Use `make cross` to build linux binary, then run `packaging/linux/install-elmer-syncd.sh /path/to/elmer-syncd` to install and start systemd service.

Code signing
- GPG is used to sign archives via goreleaser. For Windows PE code signing (Authenticode), use signtool after building; include cert management separately.
