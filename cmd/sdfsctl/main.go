package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	switch cmd {
	case "enroll-request":
		runEnrollRequest(os.Args[2:])
	case "files-put":
		runFilesPut(os.Args[2:])
	case "files-get":
		runFilesGet(os.Args[2:])
	case "files-list":
		runFilesList(os.Args[2:])
	case "session-get":
		runSessionGet(os.Args[2:])
	case "session-heartbeat":
		runSessionHeartbeat(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("sdfsctl (client) commands:")
	fmt.Println("  enroll-request --server URL --invite-token TOKEN --device-name NAME --device-public-key KEY")
	fmt.Println("  files-put --server URL --session-id SID --local PATH --remote PATH")
	fmt.Println("  files-get --server URL --session-id SID --remote PATH --local PATH")
	fmt.Println("  files-list --server URL --session-id SID [--prefix PREFIX]")
	fmt.Println("  session-get --server URL --device-public-key KEY")
	fmt.Println("  session-heartbeat --server URL --session-id SID")
}

func runEnrollRequest(args []string) {
	fs := flag.NewFlagSet("enroll-request", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	invite := fs.String("invite-token", "", "invite token")
	deviceName := fs.String("device-name", "", "device name")
	devicePublicKey := fs.String("device-public-key", "", "device public key")
	_ = fs.Parse(args)

	require(*invite != "", "--invite-token is required")
	require(*deviceName != "", "--device-name is required")
	require(*devicePublicKey != "", "--device-public-key is required")
	mustJSON(doJSONRequest(http.MethodPost, *server, "/v1/enroll/request", "", map[string]string{
		"inviteToken":     *invite,
		"deviceName":      *deviceName,
		"devicePublicKey": *devicePublicKey,
	}))
}

func runFilesPut(args []string) {
	fs := flag.NewFlagSet("files-put", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	sessionID := fs.String("session-id", "", "session id")
	local := fs.String("local", "", "local file path")
	remote := fs.String("remote", "", "remote path")
	_ = fs.Parse(args)

	require(*sessionID != "", "--session-id is required")
	require(*local != "", "--local is required")
	require(*remote != "", "--remote is required")

	f, err := os.Open(*local)
	if err != nil {
		fatalf("open local file: %v", err)
	}
	defer f.Close()

	path := "/v1/client/files/object?path=" + url.QueryEscape(*remote)
	mustJSON(doRawRequest(http.MethodPut, *server, path, *sessionID, f, "application/octet-stream"))
}

func runFilesGet(args []string) {
	fs := flag.NewFlagSet("files-get", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	sessionID := fs.String("session-id", "", "session id")
	remote := fs.String("remote", "", "remote path")
	local := fs.String("local", "", "local file path")
	_ = fs.Parse(args)

	require(*sessionID != "", "--session-id is required")
	require(*remote != "", "--remote is required")
	require(*local != "", "--local is required")

	path := "/v1/client/files/object?path=" + url.QueryEscape(*remote)
	b := doBinaryGet(*server, path, *sessionID)
	if err := os.MkdirAll(filepath.Dir(*local), 0o755); err != nil {
		fatalf("create local dir: %v", err)
	}
	if err := os.WriteFile(*local, b, 0o644); err != nil {
		fatalf("write local file: %v", err)
	}
	fmt.Printf("downloaded %d bytes to %s\n", len(b), *local)
}

func runFilesList(args []string) {
	fs := flag.NewFlagSet("files-list", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	sessionID := fs.String("session-id", "", "session id")
	prefix := fs.String("prefix", "", "optional path prefix")
	_ = fs.Parse(args)

	require(*sessionID != "", "--session-id is required")
	path := "/v1/client/files/list"
	if strings.TrimSpace(*prefix) != "" {
		path += "?prefix=" + url.QueryEscape(*prefix)
	}
	mustJSON(doJSONRequest(http.MethodGet, *server, path, *sessionID, nil))
}

func runSessionGet(args []string) {
	fs := flag.NewFlagSet("session-get", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	devicePublicKey := fs.String("device-public-key", "", "device public key")
	_ = fs.Parse(args)

	require(*devicePublicKey != "", "--device-public-key is required")
	mustJSON(doJSONRequest(http.MethodPost, *server, "/v1/client/sessions/get", "", map[string]string{
		"devicePublicKey": *devicePublicKey,
	}))
}

func runSessionHeartbeat(args []string) {
	fs := flag.NewFlagSet("session-heartbeat", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	sessionID := fs.String("session-id", "", "session id")
	_ = fs.Parse(args)

	require(*sessionID != "", "--session-id is required")
	mustJSON(doJSONRequest(http.MethodPost, *server, "/v1/client/sessions/heartbeat", *sessionID, nil))
}

func doJSONRequest(method, baseURL, path, sessionID string, body any) []byte {
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			fatalf("marshal request: %v", err)
		}
		payload = bytes.NewReader(b)
	}
	return doRawRequest(method, baseURL, path, sessionID, payload, "application/json")
}

func doRawRequest(method, baseURL, path, sessionID string, body io.Reader, contentType string) []byte {
	url := strings.TrimRight(baseURL, "/") + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fatalf("create request: %v", err)
	}
	if contentType != "" && body != nil {
		req.Header.Set("Content-Type", contentType)
	}
	if sessionID != "" {
		req.Header.Set("X-Session-Id", sessionID)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fatalf("read response: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fatalf("http %d: %s", resp.StatusCode, string(b))
	}
	return b
}

func doBinaryGet(baseURL, path, sessionID string) []byte {
	url := strings.TrimRight(baseURL, "/") + path
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fatalf("create request: %v", err)
	}
	req.Header.Set("X-Session-Id", sessionID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		fatalf("read response: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fatalf("http %d: %s", resp.StatusCode, string(b))
	}
	return b
}

func mustJSON(b []byte) {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		fmt.Println(string(b))
		return
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(string(b))
		return
	}
	fmt.Println(string(out))
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

func require(cond bool, msg string) {
	if !cond {
		fatalf(msg)
	}
}
