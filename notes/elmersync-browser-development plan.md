GitHub Copilot

Here is the current development plan and milestone sequence for the file browser update, based on the spec and the fact that the server already runs successfully with the Android app.

1. Baseline validation complete
- Confirmed the Go server builds and runs.
- Confirmed the Windows host binary works with the Android client.
- Current status: sync path works for direct file transfer; file browser is the next feature layer.

2. Milestone 1 — Root-scoped host browser
Goal:
- Restrict the remote browser to a single host folder and its descendants.
- Avoid exposing the full host filesystem.

Tasks:
- Define a fixed root directory on the server, such as the terminal launch folder or `--root`.
- Reject `..` traversal outside the root.
- Ensure all file operations are resolved relative to that root.

Acceptance:
- `/api/list?path=.` returns only the current root contents.
- Subfolders are visible.
- Files outside the root are not accessible.

3. Milestone 2 — Directory listing API
Goal:
- Let the Android app browse the remote folder tree.

API contract:
- `GET /api/list?path=<relative-or-dot-path>`
- Return JSON:
  - name
  - isDir
  - size
  - mtime
- Example response:
  - current folder
  - child folders
  - files
  - parent folder via `..`

Acceptance:
- App can navigate folder-to-folder.
- Data is stable across repeated requests.
- Empty folders return valid empty arrays.

4. Milestone 3 — Android browser UI
Goal:
- Use the returned host directory data in the app browser.

Tasks:
- Replace raw path-only transfer behavior with a browser model.
- Add current path display.
- Add parent navigation.
- Add folder expand/collapse.
- Add file open/select actions.

Acceptance:
- User can browse a remote folder tree on the host.
- File selection matches the server root.
- No full-device filesystem access is shown.

5. Milestone 4 — File operations over the browser
Goal:
- Make the browser operational, not just visual.

Add support for:
- open file
- save file
- upload file to selected remote path
- create new file/folder
- delete file/folder (optional, phase 2)
- rename file/folder (optional, phase 2)

Acceptance:
- App can open and edit a remote file.
- Save preserves the path under the server root.
- File writes are consistent with the existing sync model.

6. Milestone 5 — Safety and UX hardening
Goal:
- Turn the browser into a production-safe feature.

Tasks:
- token auth checks on every request
- root validation on all read/write operations
- clear errors for invalid paths
- optional file-type filtering
- optional hidden-file filtering
- optional async refresh when host files change

Acceptance:
- Browser cannot escape the allowed root.
- Invalid file paths fail cleanly.
- App remains stable under folder refresh.

7. Milestone 6 — Test and release
Goal:
- Verify the feature works end-to-end.

Recommended checks:
- Build server:
  - `go build -o elmer-syncd.exe ./cmd/elmer-syncd`
- Run server in a test folder
- Browse folder tree via HTTP
- Open/save a file through the Android client
- Verify nested folder navigation
- Verify lock and overwrite protections remain intact

Acceptance:
- `go test `dev26`.` passes
- server binary runs on Windows and Linux
- Android app can browse and edit files in a chosen host root folder

Suggested timeline

- Week 1: root-scoped directory API and validation
- Week 2: Android browser navigation and UI
- Week 3: file open/save and nested folder operations
- Week 4: security checks, error handling, packaging
- Week 5: release candidate and field testing

Recommended implementation order

1. Add list API
2. Add path validation
3. Add browser UI in Android
4. Add open/save against selected remote path
5. Add refresh notifications
6. Add more advanced file operations

The practical conclusion is:
- yes, the server can absolutely return folder and subfolder listings for the Android app
- the design already supports it
- the browser update is a natural next step after the current sync path works
- the safest implementation is to bind all browser operations to a single host root directory and its descendants only

If you want, I can next turn this into:
- a concrete backlog with issue titles and acceptance criteria, or
- the exact Go API and Android request/response model for the file browser.