package proto

// JSON message types used over WebSocket (informational for now)

type Hello struct {
    Type    string `json:"type"`
    Client  string `json:"client,omitempty"`
    Version string `json:"version,omitempty"`
    Token   string `json:"token,omitempty"`
}

type HelloOK struct {
    Type   string `json:"type"`
    Server string `json:"server"`
    Root   string `json:"root"`
}

type OpenRequest struct {
    Type string `json:"type"`
    Path string `json:"path"`
}

type FileMessage struct {
    Type    string `json:"type"`
    Path    string `json:"path"`
    Content string `json:"content"`
}
