package queue

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"telemetry-collector/internal/infrastructure/config"
)

func TestHTTPClientPullAndAck(t *testing.T) {
	var acked string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/pull":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]string{
					{"id": "m1", "body": `{"metric_name":"gpu.temperature","gpu_id":"g1","device":"d1","uuid":"u1","model_name":"M1","host_name":"h1","value":1,"labels_raw":"{}","processed_at_unix_nano":1735689600000000000}`},
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/ack":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			acked = body["id"]
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c, err := NewHTTPClient(config.Config{
		MQHTTPBase:     srv.URL,
		MQHTTPPullPath: "/pull",
		MQHTTPAckPath:  "/ack",
	})
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := c.Pull(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 1 {
		t.Fatalf("messages: %d", len(msgs))
	}
	if string(msgs[0].Body()) == "" {
		t.Fatal("empty body")
	}
	if err := msgs[0].Ack(context.Background()); err != nil {
		t.Fatal(err)
	}
	if acked != "m1" {
		t.Fatalf("ack id: %q", acked)
	}
}

func TestHTTPClientPullEmpty(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	c, err := NewHTTPClient(config.Config{MQHTTPBase: srv.URL, MQHTTPPullPath: "/pull", MQHTTPAckPath: "/ack"})
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := c.Pull(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(msgs))
	}
}
