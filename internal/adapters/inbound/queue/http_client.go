package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"telemetry-collector/internal/infrastructure/config"
)

// HTTPClient pulls messages from the telemetry message queue HTTP API.
//
// Expected pull response (GET {MQ_HTTP_BASE}{MQ_HTTP_PULL_PATH}?limit=N), 200 JSON:
//
//	{"items":[{"id":"optional","body":"one CSV row or JSON object string"}]}
//
// Ack (optional): POST {MQ_HTTP_BASE}{MQ_HTTP_ACK_PATH} with body {"id":"<message id>"} when id was returned.
type HTTPClient struct {
	baseURL  string // no trailing slash, e.g. http://host.docker.internal:9002
	pullPath string // begins with /
	ackPath  string // begins with /
	http     *http.Client
}

func NewHTTPClient(cfg config.Config) (*HTTPClient, error) {
	raw := strings.TrimSpace(cfg.MQHTTPBase)
	if raw == "" {
		return nil, fmt.Errorf("MQ_HTTP_BASE is required for http queue backend")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid MQ_HTTP_BASE: %q", raw)
	}
	baseURL := strings.TrimRight(u.String(), "/")
	pull := cfg.MQHTTPPullPath
	if pull == "" {
		pull = "/pull"
	}
	if !strings.HasPrefix(pull, "/") {
		pull = "/" + pull
	}
	ack := cfg.MQHTTPAckPath
	if ack == "" {
		ack = "/ack"
	}
	if !strings.HasPrefix(ack, "/") {
		ack = "/" + ack
	}
	return &HTTPClient{
		baseURL:  baseURL,
		pullPath: pull,
		ackPath:  ack,
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

type pullEnvelope struct {
	Items []struct {
		ID   string `json:"id"`
		Body string `json:"body"`
	} `json:"items"`
}

func (c *HTTPClient) Pull(ctx context.Context, batchSize int) ([]Message, error) {
	if batchSize <= 0 {
		batchSize = 1
	}
	u := c.baseURL + c.pullPath + "?limit=" + strconv.Itoa(batchSize)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNoContent || len(bytes.TrimSpace(body)) == 0 {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("queue pull: status %d: %s", resp.StatusCode, string(bytes.TrimSpace(body)))
	}
	var env pullEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("queue pull decode: %w", err)
	}
	out := make([]Message, 0, len(env.Items))
	for _, it := range env.Items {
		b := []byte(it.Body)
		if len(bytes.TrimSpace(b)) == 0 {
			continue
		}
		out = append(out, &httpMessage{client: c, id: it.ID, payload: append([]byte(nil), b...)})
	}
	return out, nil
}

type httpMessage struct {
	client  *HTTPClient
	id      string
	payload []byte
}

func (m *httpMessage) Body() []byte { return m.payload }

func (m *httpMessage) Ack(ctx context.Context) error {
	if m.id == "" || m.client == nil {
		return nil
	}
	u := m.client.baseURL + m.client.ackPath
	payload, _ := json.Marshal(map[string]string{"id": m.id})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("queue ack: status %d: %s", resp.StatusCode, string(bytes.TrimSpace(b)))
	}
	return nil
}

func (m *httpMessage) Retry(context.Context, time.Duration) error { return nil }
func (m *httpMessage) Reject(context.Context) error               { return nil }
