The server side for the browser is in place and the project is already validated end-to-end with the Android app. The next steps are:

1. Add Android browser model and request code
create the remote directory list request
parse JSON into a folder/file model
show folder entries and file entries in the UI

2. Implement navigation
current path state
folder click loads subdirectory
parent folder handling (..)
safe relative-path navigation only

3. Connect open/save to the server
GET /api/list?path=...
GET /api/read?path=...
POST /api/write with JSON payload

4. Bind the browser to the app’s existing sync flow
use the selected remote file path in the current editor
keep file opens and saves consistent with the existing remote-sync model

5. Validate end-to-end
browse a folder tree
open a remote file
edit and save it
confirm it appears at the same host path
test nested subfolders
verify root-scoped restrictions