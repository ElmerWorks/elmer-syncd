The client code in my app allows sending and receiving files with the server once connected. After I developed the server I added a simplified file browser function to my app which only shows a user-linked folder and its subfolders and not the entire Android file system.

If the server is started in a terminal any transfers happen in the folder where it was started.

From the architecture and code, state how feasible it would be for the server to return folder and subfolder file directory listings? In other words provide server host folder data so the Android app can use its file browser function to navigate and operate on the remote host server terminal launch folder and its subfolders.



# ElmerSync File Browser Specification

## Background

The Android client app allows sending and receiving files with the host server once connected. After developing the Go host server, a simplified remote file browser was added to the app. That browser shows only a user-linked folder and its subfolders instead of exposing the entire Android filesystem.

The server is launched from a terminal, and any transfers happen relative to the folder where the server was started. This is currently a useful default, but for a real remote browser it needs a controlled root folder so the app can navigate within that host folder and its subfolders safely.

This project includes a Go server in `elmer-syncd`, plus architecture docs describing a remote sync system between DOSwriter on Android and a desktop host.

## Summary

Yes — it is highly feasible for the server to return folder and subfolder directory listings so the Android app can browse, navigate, and operate on files in the server launch folder and its descendants.

The architecture and code already point in this direction:
- the server has a root path concept
- the code defines a directory entry model
- server-side listing logic is already implied by the architecture
- file-system watchers and root restrictions are already present
- the app behavior is designed around linked remote folders rather than full host filesystem exposure

## Goal

Provide a remote host browser that lets the Android app:
- connect to the server
- request a directory listing for a folder on the host
- move through subfolders
- open, save, upload, and manage files within the host root
- avoid exposing files outside the chosen root

## Required behavior

1. Server chooses a root folder
   - Use the terminal working directory by default
   - Or support an explicit root such as `--root "C:\some\sync\folder"` or `root=/path/to/folder`
   - This becomes the only browseable host area for the Android app

2. Server responds with directory listings
   - `GET /api/list?path=.`
   - `GET /api/list?path=subfolder`
   - `GET /api/list?path=../somefolder` should be normalized and blocked outside root

3. Response includes folder and file metadata
   - `name`
   - `isDir`
   - `size`
   - `mtime`

4. Client navigates the remote tree
   - current directory
   - parent folder (`..`)
   - child folders
   - files

5. Optional text-file filtering
   - For DOSwriter, restrict to text-friendly files like `.txt`, `.md`, `.json`, `.html`, etc., if desired

## Recommended server response shape

```json
{
  "path": ".",
  "entries": [
    {
      "name": "docs",
      "isDir": true,
      "size": 0,
      "mtime": "2026-11-05T12:00:00Z"
    },
    {
      "name": "notes.txt",
      "isDir": false,
      "size": 1680,
      "mtime": "2026-11-05T12:00:00Z"
    }
  ]
}
```

## Security and safety requirements

- Never expose the entire host filesystem
- Enforce a root directory boundary
- Reject path traversal above the root
- Allow only the chosen subtree to be browsed
- Optionally restrict file types
- Require token auth for the connection

## Mapping to the current code

From the project architecture and Go source:
- `internal/server/server.go` already contains the root-based path handling and fsnotify filesystem watchers
- `internal/server/handlers.go` already defines `DirEntry` and appears intended for directory listing
- `cmd/elmer-syncd/main.go` already supports a root directory and host/server runtime configuration
- `internal/proto/messages.go` supports remote file message payloads for the sync protocol

This means the project is already structurally close to a proper host-side file browser.

## Feasibility assessment

Very feasible.

The required pieces are already present or implied:
- root directory management
- path validation
- file listing support
- file open/save behavior
- client connection and token auth
- remote file transfer logic

The main missing piece is wiring the listing result into the Android UI and making the server root fixed to a chosen directory instead of the process working directory.

## Recommended design

Use the server launch folder as the browser root:
- if started in `C:\Users\me\Documents\elmer-sync`, that becomes the remote root
- the app can browse that folder and subfolders only
- transfers occur relative to that root
- no access is granted outside it

Example launch:
```powershell
.\elmer-syncd.exe --root "C:\Users\me\Documents\elmer-sync" --port 12345 --token mytoken
```

Client flow:
1. Connect to host
2. Request `GET /api/list?path=.`
3. Display entries
4. User taps a subfolder -> request listing for that folder
5. Open or save file relative to current server root

## Conclusion

The server can absolutely return folder and subfolder directory listings for the Android app to navigate and operate on the host terminal launch folder and its subfolders. This is well within the current architecture and implementation pattern.

The practical recommendation is:
- enforce a fixed root directory
- keep the browser limited to that root and descendants
- expose a clean directory listing endpoint
- use relative paths for all open/save/upload operations

This gives the Android app a safe and useful remote file browser without exposing the entire host filesystem.