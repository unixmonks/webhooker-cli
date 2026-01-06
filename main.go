package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

var version = "dev"

type WebSocketMessage struct {
	Webhook     Webhook `json:"webhook"`
	BalanceSats int     `json:"balance_sats"`
}

type Webhook struct {
	ID        int    `json:"id"`
	AccountID int    `json:"account_id"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Headers   string `json:"headers"`
	Body      string `json:"body"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type LogEntry struct {
	Time        string `json:"time"`
	Event       string `json:"event"`
	Method      string `json:"method,omitempty"`
	Path        string `json:"path,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
	BalanceSats int    `json:"balance_sats,omitempty"`
	Error       string `json:"error,omitempty"`
	Response    string `json:"response,omitempty"`
}

func logJSON(entry LogEntry) {
	entry.Time = time.Now().Format(time.RFC3339)
	data, _ := json.Marshal(entry)
	fmt.Println(string(data))
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return nil
	}

	switch os.Args[1] {
	case "connect":
		return runConnect(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("webhooker-cli %s\n", version)
		return nil
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", os.Args[1])
	}
}

func printUsage() {
	fmt.Println(`webhooker-cli - Forward webhooks to your local server

Usage:
  webhooker connect <account-token> --forward <url> [options]

Commands:
  connect    Connect to webhooker server and forward webhooks
  version    Print version information

Options:
  --server   Server URL (default: wss://webhooker.site)
  --forward  Local URL to forward webhooks to (required)
  --verbose  Enable verbose output

Examples:
  webhooker connect abc123 --forward http://localhost:3000
  webhooker connect abc123 --forward http://localhost:8080/webhooks --verbose
  webhooker connect abc123 --server wss://webhooker.site --forward http://localhost:3000`)
}

func runConnect(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("account token required")
	}
	token := args[0]

	fs := flag.NewFlagSet("connect", flag.ExitOnError)
	serverURL := fs.String("server", "wss://webhooker.site", "Server URL")
	forwardURL := fs.String("forward", "", "Local URL to forward webhooks to")
	verbose := fs.Bool("verbose", false, "Enable verbose output")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	if *forwardURL == "" {
		return fmt.Errorf("--forward URL required")
	}

	if _, err := url.Parse(*forwardURL); err != nil {
		return fmt.Errorf("invalid forward URL: %w", err)
	}

	wsURL := *serverURL + "/api/v1/connect/" + token

	log.Printf("Connecting to %s...", *serverURL)
	log.Printf("Forwarding webhooks to %s", *forwardURL)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	for {
		err := connectAndListen(ctx, wsURL, *forwardURL, *verbose)
		if ctx.Err() != nil {
			log.Println("Shutting down...")
			return nil
		}
		if err != nil {
			log.Printf("Connection error: %v", err)
			log.Println("Reconnecting in 5 seconds...")
			select {
			case <-ctx.Done():
				log.Println("Shutting down...")
				return nil
			case <-time.After(5 * time.Second):
			}
		}
	}
}

func connectAndListen(ctx context.Context, wsURL, forwardURL string, verbose bool) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	log.Println("Connected! Waiting for webhooks...")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Read error: %v", err)
				}
				return
			}

			var msg WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Failed to parse message: %v", err)
				continue
			}

			go forwardWebhook(msg.Webhook, msg.BalanceSats, forwardURL, verbose)
		}
	}()

	select {
	case <-done:
		return fmt.Errorf("connection closed")
	case <-ctx.Done():
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		return ctx.Err()
	}
}

func forwardWebhook(webhook Webhook, balanceSats int, forwardURL string, verbose bool) {
	targetURL := forwardURL
	if webhook.Path != "/" && webhook.Path != "" {
		targetURL = strings.TrimSuffix(forwardURL, "/") + webhook.Path
	}

	logJSON(LogEntry{
		Event:       "webhook_received",
		Method:      webhook.Method,
		Path:        webhook.Path,
		BalanceSats: balanceSats,
	})

	req, err := http.NewRequest(webhook.Method, targetURL, bytes.NewBufferString(webhook.Body))
	if err != nil {
		logJSON(LogEntry{
			Event:       "forward_error",
			Method:      webhook.Method,
			Path:        webhook.Path,
			BalanceSats: balanceSats,
			Error:       err.Error(),
		})
		return
	}

	var headers map[string]string
	if err := json.Unmarshal([]byte(webhook.Headers), &headers); err == nil {
		for key, value := range headers {
			if strings.ToLower(key) != "host" && strings.ToLower(key) != "content-length" {
				req.Header.Set(key, value)
			}
		}
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logJSON(LogEntry{
			Event:       "forward_error",
			Method:      webhook.Method,
			Path:        webhook.Path,
			BalanceSats: balanceSats,
			Error:       err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	entry := LogEntry{
		Event:       "webhook_forwarded",
		Method:      webhook.Method,
		Path:        webhook.Path,
		StatusCode:  resp.StatusCode,
		BalanceSats: balanceSats,
	}

	if verbose {
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			entry.Response = truncate(string(body), 200)
		}
	}

	logJSON(entry)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
