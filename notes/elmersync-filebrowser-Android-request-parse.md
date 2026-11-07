# ElmerSync File Browser: Android Request/Parse and Go handleList

## Go server: `internal/server/handlers.go`

```go
package server

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

func (s *Server) checkAuth(r *http.Request) bool {
    if s.token == "" {
        return true
    }
    return r.Header.Get("X-API-Token") == s.token
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
    if !s.checkAuth(r) {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    path := r.URL.Query().Get("path")
    target, err := s.resolvePath(path)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    info, err := os.Stat(target)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    if !info.IsDir() {
        http.Error(w, "not a directory", http.StatusBadRequest)
        return
    }

    entries, err := os.ReadDir(target)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    result := make([]DirEntry, 0, len(entries))
    for _, entry := range entries {
        fi, err := entry.Info()
        if err != nil {
            log.Printf("read dir info: %v", err)
            continue
        }

        result = append(result, DirEntry{
            Name:  entry.Name(),
            IsDir: entry.IsDir(),
            Size:  fi.Size(),
            Mtime: fi.ModTime(),
        })
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{
        "path":    path,
        "entries": result,
    })
}

#This is the exact response shape returned by the server:
{
  "path": ".",
  "entries": [
    {
      "name": "subdir",
      "isDir": true,
      "size": 0,
      "mtime": "2026-11-05T12:00:00Z"
    },
    {
      "name": "file.txt",
      "isDir": false,
      "size": 1234,
      "mtime": "2026-11-05T12:00:00Z"
    }
  ]
}

#Android: request directory list
Use HttpURLConnection or any HTTP client already in the app. This example uses HttpURLConnection :

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.ArrayList;
import java.util.List;

public class RemoteBrowserClient {
    private final String baseUrl;
    private final String token;

    public RemoteBrowserClient(String baseUrl, String token) {
        this.baseUrl = baseUrl;
        this.token = token;
    }

    public static class RemoteEntry {
        public String name;
        public boolean isDir;
        public long size;
        public long mtime;

        public RemoteEntry(String name, boolean isDir, long size, long mtime) {
            this.name = name;
            this.isDir = isDir;
            this.size = size;
            this.mtime = mtime;
        }
    }

    public List<RemoteEntry> listDir(String path) throws Exception {
        String urlString = baseUrl + "/api/list?path=" + java.net.URLEncoder.encode(path, "UTF-8");
        URL url = new URL(urlString);

        HttpURLConnection conn = (HttpURLConnection) url.openConnection();
        conn.setRequestMethod("GET");
        conn.setConnectTimeout(15000);
        conn.setReadTimeout(15000);

        if (token != null && !token.isEmpty()) {
            conn.setRequestProperty("X-API-Token", token);
        }

        int status = conn.getResponseCode();
        if (status != 200) {
            throw new RuntimeException("HTTP " + status);
        }

        BufferedReader reader = new BufferedReader(new InputStreamReader(conn.getInputStream()));
        StringBuilder sb = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            sb.append(line);
        }
        reader.close();

        JSONObject root = new JSONObject(sb.toString());
        JSONArray entries = root.getJSONArray("entries");

        List<RemoteEntry> results = new ArrayList<>();
        for (int i = 0; i < entries.length(); i++) {
            JSONObject obj = entries.getJSONObject(i);
            results.add(new RemoteEntry(
                    obj.optString("name"),
                    obj.optBoolean("isDir"),
                    obj.optLong("size", 0L),
                    obj.optLong("mtime", 0L)
            ));
        }

        return results;
    }
}


#Android usage example :

new Thread(() -> {
    try {
        RemoteBrowserClient client = new RemoteBrowserClient("http://192.168.1.50:12345", "demo");
        List<RemoteBrowserClient.RemoteEntry> entries = client.listDir(".");

        for (RemoteBrowserClient.RemoteEntry e : entries) {
            System.out.println(e.name + " | dir=" + e.isDir + " | size=" + e.size);
        }
    } catch (Exception e) {
        e.printStackTrace();
    }
}).start();

#Android: parse folder list into the UI
#A simple model for the UI :

public class RemoteFolderModel {
    public String currentPath;
    public List<RemoteBrowserClient.RemoteEntry> entries = new ArrayList<>();
}

#Load and parse:

public RemoteFolderModel loadFolder(String path) throws Exception {
    RemoteBrowserClient client = new RemoteBrowserClient("http://192.168.1.50:12345", "demo");
    List<RemoteBrowserClient.RemoteEntry> items = client.listDir(path);

    RemoteFolderModel model = new RemoteFolderModel();
    model.currentPath = path;
    model.entries.addAll(items);
    return model;
}

#Then render:
folders as expandable rows
files as selectable items
.. row for parent folder if path is not root
tap folder -> call listDir(newPath)

#Android: open a file

public String readFile(String relativePath) throws Exception {
    String urlString = "http://192.168.1.50:12345/api/read?path="
            + java.net.URLEncoder.encode(relativePath, "UTF-8");
    URL url = new URL(urlString);

    HttpURLConnection conn = (HttpURLConnection) url.openConnection();
    conn.setRequestMethod("GET");
    conn.setConnectTimeout(15000);
    conn.setReadTimeout(15000);
    conn.setRequestProperty("X-API-Token", "demo");

    int status = conn.getResponseCode();
    if (status != 200) {
        throw new RuntimeException("read failed: " + status);
    }

    BufferedReader reader = new BufferedReader(new InputStreamReader(conn.getInputStream()));
    StringBuilder sb = new StringBuilder();
    String line;
    while ((line = reader.readLine()) != null) {
        sb.append(line).append('\n');
    }
    reader.close();
    return sb.toString();
}

#Android: write a file

public void writeFile(String relativePath, String content) throws Exception {
    String urlString = "http://192.168.1.50:12345/api/write";
    URL url = new URL(urlString);

    JSONObject payload = new JSONObject();
    payload.put("path", relativePath);
    payload.put("content", content);

    HttpURLConnection conn = (HttpURLConnection) url.openConnection();
    conn.setRequestMethod("POST");
    conn.setDoOutput(true);
    conn.setConnectTimeout(15000);
    conn.setReadTimeout(15000);
    conn.setRequestProperty("Content-Type", "application/json");
    conn.setRequestProperty("X-API-Token", "demo");

    OutputStream os = conn.getOutputStream();
    os.write(payload.toString().getBytes("UTF-8"));
    os.close();

    int status = conn.getResponseCode();
    if (status != 200) {
        throw new RuntimeException("write failed: " + status);
    }
}