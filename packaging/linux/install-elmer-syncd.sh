#!/usr/bin/env bash
set -euo pipefail

# Usage: install-elmer-syncd.sh <path-to-binary> [--create-user]
BINARY=${1:-}
CREATE_USER=${2:-}
if [ -z "$BINARY" ]; then
  echo "Usage: $0 <path-to-elmer-syncd-binary> [--create-user]" >&2
  exit 2
fi

if [ ! -f "$BINARY" ]; then
  echo "Binary not found: $BINARY" >&2
  exit 3
fi

sudo cp "$BINARY" /usr/local/bin/elmer-syncd
sudo chmod 0755 /usr/local/bin/elmer-syncd

if [ "$CREATE_USER" = "--create-user" ]; then
  sudo useradd --system --no-create-home --shell /usr/sbin/nologin elmer || true
fi

sudo mkdir -p /var/lib/elmer-syncd
sudo chown root:root /var/lib/elmer-syncd

sudo cp packaging/linux/elmer-syncd.service /etc/systemd/system/elmer-syncd.service
sudo systemctl daemon-reload
sudo systemctl enable --now elmer-syncd.service
echo "elmer-syncd installed and started"
