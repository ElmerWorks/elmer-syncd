# ElmerSync File Browser Communication Model

## 1. Server contract

The file browser is implemented as a thin host-side directory API layered on the existing sync server. The server root is the allowed browse boundary.

### Authentication

All browser requests use the same token guard as the sync server.

Header:
```http
X-API-Token: <token>
```

If the server token is empty, auth is disabled.

---

### 1.1 Directory listing

Method:
```http
GET /api/list?path=<relative-or-dot-path>
```

Example:
```http
GET /api/list?path=.
GET /api/list?path=subdir
GET /api/list?path=documents/project
```

Response:
```json
{
  "path": ".",
  "entries": [
    {
      "name": "documents",
      "isDir": true,
      "size": 0,
      "mtime": "2026-11-05T12:00:00Z"
    },
    {
      "name": "notes.txt",
      "isDir": false,
      "size": 1840,
      "mtime": "2026-11-05T12:00:00Z"
    }
  ]
}
```

Go model:
```go
type DirEntry struct {
    Name  string    `json:"name"`
    IsDir bool      `json:"isDir"`
    Size  int64     `json:"size"`
    Mtime time.Time `json:"mtime"`
}

type ListResponse struct {
    Path    string     `json:"path"`
    Entries []DirEntry `json:"entries"`
}
```

Server-side behavior:
- path is resolved under the server root
- `.` means current root
- `..` is rejected if it would leave the root
- directory entries are sorted by name
- non-text files may be included unless filtered later

---

### 1.2 File read

Method:
```http
GET /api/read?path=<relative-path>
```

Example:
```http
GET /api/read?path=notes.txt
GET /api/read?path=subdir/file.md
```

Response:
- raw file content as `application/octet-stream`
- HTTP 200 for success
- HTTP 400 for invalid path
- HTTP 404 for missing file
- HTTP 500 for read failure

---

### 1.3 File write

Method:
```http
POST /api/write
```

Body:
```json
{
  "path": "subdir/new-file.txt",
  "content": "hello from android"
}
```

Response:
```json
{
  "status": "ok"
}
```

Go model:
```go
type WriteRequest struct {
    Path    string `json:"path"`
    Content string `json:"content"`
}

type WriteResponse struct {
    Status string `json:"status"`
}
```

Server-side behavior:
- path is resolved under the server root
- directories are created automatically if needed
- file is written atomically if possible
- `fs_event` is broadcast to active clients after success

---

### 1.4 WebSocket event stream

URL:
```http
/ws
```

Hello message:
```json
{
  "type": "hello",
  "client": "DOSwriter-Android",
  "version": "1.0.0",
  "token": "demo-token"
}
```

Server hello ack:
```json
{
  "type": "hello_ok",
  "server": "elmer-syncd",
  "root": "C:\\sync-root"
}
```

File-system event:
```json
{
  "type": "fs_event",
  "path": "subdir/file.txt"
}
```

Go message model:
```go
type Hello struct {
    Type    string `json:"type"`
    Client  string `json:"client,omitempty"`
    Version string `json:"version,omitempty"`
    Token   string `json:"token,omitempty"`
}

type FsEvent struct {
    Type string `json:"type"`
    Path string `json:"path"`
}
```

---

## 2. Android client model

The Android app should map the server response into a browser tree model.

### 2.1 Remote entry model

```java
public class RemoteEntry {
    public String name;
    public boolean isDir;
    public long size;
    public long mtime;
}
```

JSON:
```json
{
  "name": "notes.txt",
  "isDir": false,
  "size": 1840,
  "mtime": "2026-11-05T12:00:00Z"
}
```

### 2.2 Directory response model

```java
public class RemoteListResponse {
    public String path;
    public List<RemoteEntry> entries;
}
```

### 2.3 Write request model

```java
public class RemoteWriteRequest {
    public String path;
    public String content;
}
```

### 2.4 Write response model

```java
public class RemoteWriteResponse {
    public String status;
}
```

---

## 3. Android browser flow

1. Connect to host using IP, port, and token.
2. Request:
   ```http
   GET /api/list?path=.
   ```
3. Render the returned entries in a remote file tree.
4. If user selects a folder:
   ```http
   GET /api/list?path=<selected-folder>
   ```
5. If user selects a file:
   ```http
   GET /api/read?path=<selected-file>
   ```
6. If user saves:
   ```http
   POST /api/write
   {
     "path": "<selected-file>",
     "content": "<edited-text>"
   }
   ```

---

## 4. Browser safety rules

- The server root is the only allowed browse area.
- `..` and absolute paths are blocked.
- Absolute paths on the desktop host are not exposed.
- All remote operations are relative to the root.
- A valid `X-API-Token` is required if the server is configured with one.

---

## 5. Minimal Go implementation summary

```go
func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
    // resolve path under s.root
    // read dir
    // build []DirEntry
    // JSON encode ListResponse
}
```

```go
func (s *Server) handleRead(w http.ResponseWriter, r *http.Request) {
    // resolve path under s.root
    // os.ReadFile
    // write raw bytes
}
```

```go
func (s *Server) handleWrite(w http.ResponseWriter, r *http.Request) {
    // decode WriteRequest
    // resolve path under s.root
    // create dirs if required
    // os.WriteFile
    // broadcast fs_event
    // respond ok
}
```

This is the exact browser communication model consistent with the current server architecture and the existing sync server design.