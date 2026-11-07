# elmer-syncd

Minimal Go server for ElmerPad Desktop Sync (Sprint 1)

Run:

```bash
go run ./cmd/elmer-syncd --root "C:/path/to/notes" --port 12345 --token mytoken
```

Endpoints:
- GET /api/list?path=relative/path
- GET /api/file?path=relative/path
- POST /api/file  (JSON {path,content})
- WS  /ws         (JSON messages: hello, open, save)
