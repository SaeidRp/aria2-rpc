package ario_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/coder/websocket"
	ario "github.com/saeidrp/aria2-rpc"
)

func TestNewClientBackwardCompatibleForWS(t *testing.T) {
	client, closeServer := newWSClient(t, "ok", nil)
	defer closeServer()
	defer client.Close()

	var out string
	if err := client.Call("aria2.getVersion", nil, &out); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if out != "ok" {
		t.Fatalf("unexpected response: got %q want %q", out, "ok")
	}
}

func TestNewClientWithOptionsWSReadLimitOverride(t *testing.T) {
	payload := strings.Repeat("x", 8*1024)

	client, closeServer := newWSClient(t, payload, &ario.ClientOptions{WSReadLimit: 1024})
	defer closeServer()
	defer client.Close()

	var out string
	err := client.Call("aria2.getVersion", nil, &out)
	if err == nil {
		t.Fatal("expected error when payload exceeds configured WS read limit")
	}
	if !strings.Contains(err.Error(), "message too big") {
		t.Fatalf("expected message too big error, got: %v", err)
	}
}

func newWSClient(t *testing.T, responsePayload string, opts *ario.ClientOptions) (*ario.Client, func()) {
	t.Helper()

	srv := newWSClientServer(t, responsePayload)

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	u.Scheme = "ws"
	u.Path = "/jsonrpc"

	client, err := ario.NewClientWithOptions(u.String(), "", false, opts)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	return client, srv.Close
}

func newWSClientServer(t *testing.T, resultPayload string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/jsonrpc" {
			http.NotFound(w, r)
			return
		}

		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("websocket accept: %v", err)
			return
		}
		defer func() { _ = conn.Close(websocket.StatusNormalClosure, "bye") }()

		_, reqBits, err := conn.Read(r.Context())
		if err != nil {
			t.Errorf("websocket read request: %v", err)
			return
		}

		var req struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.Unmarshal(reqBits, &req); err != nil {
			t.Errorf("unmarshal request: %v", err)
			return
		}

		respBits, err := json.Marshal(struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id"`
			Result  string          `json:"result"`
		}{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  resultPayload,
		})
		if err != nil {
			t.Errorf("marshal response: %v", err)
			return
		}

		if err := conn.Write(context.Background(), websocket.MessageText, respBits); err != nil {
			t.Errorf("websocket write response: %v", err)
		}
	}))
}
