Yes — this is highly feasible, and the current architecture already points in that direction.

Why it is feasible

- The server already has a root directory concept:
  - `Server.root` is set in `internal/server/server.go`
  - path checks are enforced so the server stays inside the configured root
- The code already defines a directory listing model:
  - `DirEntry` in `internal/server/handlers.go`
  - fields: `Name`, `IsDir`, `Size`, `Mtime`
- The server is already designed to watch a directory tree with `fsnotify`
  - this is exactly the basis for live file/folder changes
- The app is already meant to browse and operate on remote files
  - the Android app is described as a “remote file browser” that links to a user folder, not the full Android filesystem

In other words, the host side already has the right building blocks for:
- a server-root folder
- recursive directory scanning
- folder/file metadata responses
- client-side navigation through subfolders

The key design point is this:

- If the server is started in a terminal, all transfers happen relative to that terminal’s working directory
- That is not a bug; it is how the current server behaves
- To make the app browser usable, the server should either:
  - start in the desired root folder, or
  - be launched with an explicit root flag/config, e.g. `--root "<chosen folder>"`

How it maps to the current code

From the attached code:
- `internal/server/handlers.go` already includes a `handleList`
- `internal/server/server.go` already includes root restriction logic and filesystem watchers
- `main.go` likely starts the server with root and bind settings

So the project is already close to a proper remote host browser:
- browse current folder
- show child folders and files
- navigate deeper
- open/read/write files
- filter to text files if desired

What the server should return

A simple response model is enough:

```json
{
  "path": ".",
  "entries": [
    { "name": "folder1", "isDir": true, "size": 0, "mtime": "2026-11-05T12:00:00Z" },
    { "name": "file1.txt", "isDir": false, "size": 1234, "mtime": "2026-11-05T12:00:00Z" }
  ]
}
```

Then the Android app can:
- show the folder list
- navigate into subfolders
- allow open/save/upload/delete actions relative to the server root

Practical recommendation

Use the server launch folder as the remote root, and expose a listing endpoint like:
- `GET /api/list?path=.` for current folder
- `GET /api/list?path=subfolder` for nested folders

Then the client can browse exactly the host’s launch folder and its descendants.

Conclusion

- Feasibility: very high
- Current codebase: already structurally ready
- Missing piece: connect the existing `handleList` and root enforcement to a real client navigation flow
- Best practice: set a fixed root folder and list only that subtree

So yes: the server can absolutely return host folder/subfolder directory listings for the Android app to navigate and operate on, without exposing the entire host filesystem.