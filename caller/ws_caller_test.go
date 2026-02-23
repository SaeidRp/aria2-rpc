package caller_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/saeidrp/aria2-rpc/caller"
)

func TestWSCallerLargeResponseOver32KiB(t *testing.T) {
	const payloadSize = 64 * 1024
	payload := strings.Repeat("a", payloadSize)

	result, err := callWSRPC(t, payload, nil)
	if err != nil {
		t.Fatalf("expected no error for large payload, got: %v", err)
	}
	if len(result) != payloadSize {
		t.Fatalf("unexpected result length: got %d want %d", len(result), payloadSize)
	}
}

func TestWSCallerDefaultReadLimitIs8MiB(t *testing.T) {
	if caller.DefaultWSReadLimit != 8*1024*1024 {
		t.Fatalf("unexpected default read limit: got %d want %d", caller.DefaultWSReadLimit, 8*1024*1024)
	}

	const payloadSize = 256 * 1024
	payload := strings.Repeat("b", payloadSize)

	_, err := callWSRPC(t, payload, nil)
	if err != nil {
		t.Fatalf("expected default limit to allow payload, got: %v", err)
	}
}

func TestWSCallerCustomReadLimitOverride(t *testing.T) {
	const payloadSize = 8 * 1024
	payload := strings.Repeat("c", payloadSize)

	_, err := callWSRPC(t, payload, &caller.Options{WSReadLimit: 1024})
	if err == nil {
		t.Fatal("expected error when payload exceeds custom read limit")
	}
	if !strings.Contains(err.Error(), "message too big") {
		t.Fatalf("expected message too big error, got: %v", err)
	}

	_, err = callWSRPC(t, payload, &caller.Options{WSReadLimit: 16 * 1024})
	if err != nil {
		t.Fatalf("expected overridden limit to allow payload, got: %v", err)
	}
}

func TestWSCallerSmallResponseNoRegression(t *testing.T) {
	const payload = "ok"

	result, err := callWSRPC(t, payload, nil)
	if err != nil {
		t.Fatalf("expected no error for small payload, got: %v", err)
	}
	if result != payload {
		t.Fatalf("unexpected response: got %q want %q", result, payload)
	}
}

func callWSRPC(t *testing.T, resultPayload string, opts *caller.Options) (string, error) {
	t.Helper()

	srv := newWSRPCServer(t, resultPayload)
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	u.Scheme = "ws"
	u.Path = "/jsonrpc"

	c, err := caller.NewCallerWithOptions(u, opts)
	if err != nil {
		t.Fatalf("create caller: %v", err)
	}
	defer c.Close()

	var out string
	err = c.Call("aria2.getVersion", nil, &out)
	return out, err
}

func newWSRPCServer(t *testing.T, resultPayload string) *httptest.Server {
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
