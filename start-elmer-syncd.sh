#!/usr/bin/env bash
ROOT="${1:-/var/lib/elmer-syncd}"
PORT="${2:-12345}"
TOKEN="${3:-mytoken}"
exec ./elmer-syncd --root "$ROOT" --port "$PORT" --token "$TOKEN"
