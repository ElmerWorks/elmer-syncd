package main

import (
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"

    "net"
    "bufio"
    "strings"
    "path/filepath"

    "github.com/elmerpad/elmer-syncd/internal/server"
)

func main() {
    var (
        root     string
        bind     string
        port     int
        token    string
        allowLan bool
        showIP   bool
    )

    flag.StringVar(&root, "root", ".", "root directory to expose")
    flag.StringVar(&bind, "bind", "127.0.0.1", "bind address (use 0.0.0.0 with --allow-lan)")
    flag.IntVar(&port, "port", 12345, "port to listen on")
    flag.StringVar(&token, "token", "", "API token (optional)")
    flag.BoolVar(&allowLan, "allow-lan", false, "allow listening on 0.0.0.0")
    flag.BoolVar(&showIP, "show-ip", false, "print LAN IP addresses and exit")
    flag.Parse()

    if showIP {
        addrs, err := localIPs()
        if err != nil {
            fmt.Fprintf(os.Stderr, "error enumerating IPs: %v\n", err)
            os.Exit(2)
        }
        fmt.Println("ElmerSyncd connection examples:")
        for _, a := range addrs {
            fmt.Printf("  WebSocket: ws://%s:%d/ws\n", a, port)
            fmt.Printf("  REST list: http://%s:%d/api/list?path=\n", a, port)
        }
        fmt.Println("If you run with --allow-lan, bind will be 0.0.0.0 to accept LAN connections.")
        os.Exit(0)
    }

    if allowLan {
        bind = "0.0.0.0"
    }

    // If user didn't provide a root (left as "."), try to read a bundled ini config
    // placed next to the executable by the installer: elmer-syncd.ini with a line `root=...`
    if root == "." || root == "" {
        if exepath, err := os.Executable(); err == nil {
            exedir := filepath.Dir(exepath)
            cfg := filepath.Join(exedir, "elmer-syncd.ini")
            if info, err := os.Stat(cfg); err == nil && !info.IsDir() {
                // simple parse: look for `root=` prefix
                if f, err := os.Open(cfg); err == nil {
                    scanner := bufio.NewScanner(f)
                    for scanner.Scan() {
                        line := strings.TrimSpace(scanner.Text())
                        if strings.HasPrefix(line, "root=") {
                            root = strings.TrimSpace(strings.TrimPrefix(line, "root="))
                            break
                        }
                    }
                    f.Close()
                }
            }
        }
    }

    s, err := server.New(root, token)
    if err != nil {
        log.Fatalf("server init: %v", err)
    }

    addr := fmt.Sprintf("%s:%d", bind, port)
    log.Printf("starting elmer-syncd on %s (root=%s)", addr, root)

    srv := &http.Server{
        Addr:    addr,
        Handler: s.Router(),
    }

    // start a goroutine to listen for single-key input on stdin
    go func() {
        reader := bufio.NewReader(os.Stdin)
        for {
            line, err := reader.ReadString('\n')
            if err != nil {
                return
            }
            line = strings.TrimSpace(line)
            if len(line) == 0 {
                continue
            }
            if strings.EqualFold(line, "h") {
                fmt.Println("ElmerSyncd console help:")
                fmt.Println("  H    : show this help")
                fmt.Printf("  Port : %d\n", port)
                fmt.Printf("  Root : %s\n", root)
                fmt.Println("  Use --show-ip to list LAN addresses and example URLs")
            }
        }
    }()

    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Printf("server stopped: %v", err)
        os.Exit(1)
    }
}

// localIPs returns IPv4 addresses for non-loopback interfaces
func localIPs() ([]string, error) {
    var out []string
    ifaces, err := net.Interfaces()
    if err != nil {
        return nil, err
    }
    for _, ifi := range ifaces {
        if ifi.Flags&net.FlagUp == 0 {
            continue
        }
        if ifi.Flags&net.FlagLoopback != 0 {
            continue
        }
        addrs, err := ifi.Addrs()
        if err != nil {
            continue
        }
        for _, a := range addrs {
            var ip net.IP
            switch v := a.(type) {
            case *net.IPNet:
                ip = v.IP
            case *net.IPAddr:
                ip = v.IP
            }
            if ip == nil || ip.IsLoopback() {
                continue
            }
            ip = ip.To4()
            if ip == nil {
                continue
            }
            s := ip.String()
            // skip link-local
            if s == "" || s[:4] == "169." {
                continue
            }
            out = append(out, s)
        }
    }
    if len(out) == 0 {
        out = append(out, "127.0.0.1")
    }
    return out, nil
}
