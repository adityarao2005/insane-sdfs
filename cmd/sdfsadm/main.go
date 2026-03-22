package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
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
	case "invite":
		runInvite(os.Args[2:])
	case "pending":
		runPending(os.Args[2:])
	case "approve":
		runApprove(os.Args[2:])
	case "devices":
		runDevices(os.Args[2:])
	case "revoke":
		runRevoke(os.Args[2:])
	case "sessions-active":
		runSessionsActive(os.Args[2:])
	case "sessions-start":
		runSessionsStart(os.Args[2:])
	case "sessions-heartbeat":
		runSessionsHeartbeat(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("sdfsadm commands:")
	fmt.Println("  invite --server URL --token TOKEN [--ttl-seconds 600]")
	fmt.Println("  pending --server URL --token TOKEN")
	fmt.Println("  approve --server URL --token TOKEN --request-id ID")
	fmt.Println("  devices --server URL --token TOKEN")
	fmt.Println("  revoke --server URL --token TOKEN --device-id ID")
	fmt.Println("  sessions-active --server URL --token TOKEN")
	fmt.Println("  sessions-start --server URL --token TOKEN --device-id ID")
	fmt.Println("  sessions-heartbeat --server URL --token TOKEN --session-id ID")
}

func runInvite(args []string) {
	fs := flag.NewFlagSet("invite", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	token := fs.String("token", "", "admin token")
	ttl := fs.Int("ttl-seconds", 600, "invite ttl in seconds")
	_ = fs.Parse(args)
	require(*token != "", "--token is required")
	mustJSON(doRequest(http.MethodPost, *server, "/v1/admin/invites", *token, map[string]int{"ttlSeconds": *ttl}))
}

func runPending(args []string) {
	fs := flag.NewFlagSet("pending", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	token := fs.String("token", "", "admin token")
	_ = fs.Parse(args)
	require(*token != "", "--token is required")
	mustJSON(doRequest(http.MethodGet, *server, "/v1/admin/enrollments/pending", *token, nil))
}

func runApprove(args []string) {
	fs := flag.NewFlagSet("approve", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	token := fs.String("token", "", "admin token")
	requestID := fs.String("request-id", "", "enrollment request id")
	_ = fs.Parse(args)
	require(*token != "", "--token is required")
	require(*requestID != "", "--request-id is required")
	mustJSON(doRequest(http.MethodPost, *server, "/v1/admin/enrollments/approve", *token, map[string]string{"requestId": *requestID}))
}

func runDevices(args []string) {
	fs := flag.NewFlagSet("devices", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	token := fs.String("token", "", "admin token")
	_ = fs.Parse(args)
	require(*token != "", "--token is required")
	mustJSON(doRequest(http.MethodGet, *server, "/v1/admin/devices", *token, nil))
}

func runRevoke(args []string) {
	fs := flag.NewFlagSet("revoke", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	token := fs.String("token", "", "admin token")
	deviceID := fs.String("device-id", "", "device id")
	_ = fs.Parse(args)
	require(*token != "", "--token is required")
	require(*deviceID != "", "--device-id is required")
	mustJSON(doRequest(http.MethodPost, *server, "/v1/admin/devices/revoke", *token, map[string]string{"deviceId": *deviceID}))
}

func runSessionsActive(args []string) {
	fs := flag.NewFlagSet("sessions-active", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	token := fs.String("token", "", "admin token")
	_ = fs.Parse(args)
	require(*token != "", "--token is required")
	mustJSON(doRequest(http.MethodGet, *server, "/v1/admin/sessions/active", *token, nil))
}

func runSessionsStart(args []string) {
	fs := flag.NewFlagSet("sessions-start", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	token := fs.String("token", "", "admin token")
	deviceID := fs.String("device-id", "", "device id")
	_ = fs.Parse(args)
	require(*token != "", "--token is required")
	require(*deviceID != "", "--device-id is required")
	mustJSON(doRequest(http.MethodPost, *server, "/v1/admin/sessions/start", *token, map[string]string{"deviceId": *deviceID}))
}

func runSessionsHeartbeat(args []string) {
	fs := flag.NewFlagSet("sessions-heartbeat", flag.ExitOnError)
	server := fs.String("server", "http://127.0.0.1:8080", "server base URL")
	token := fs.String("token", "", "admin token")
	sessionID := fs.String("session-id", "", "session id")
	_ = fs.Parse(args)
	require(*token != "", "--token is required")
	require(*sessionID != "", "--session-id is required")
	mustJSON(doRequest(http.MethodPost, *server, "/v1/admin/sessions/heartbeat", *token, map[string]string{"sessionId": *sessionID}))
}

func doRequest(method, baseURL, path, adminToken string, body any) []byte {
	url := strings.TrimRight(baseURL, "/") + path
	var payload io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			fatalf("marshal request: %v", err)
		}
		payload = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		fatalf("create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Admin-Token", adminToken)

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
