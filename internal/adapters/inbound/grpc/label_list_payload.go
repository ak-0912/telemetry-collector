package grpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	pb "telemetry-collector/api/telemetry/v1"
)

// parseLabelListTelemetry parses DCGM / exporter lines like:
//
//	__name__="DCGM_FI_DEV_MEM_COPY_UTIL",gpu="1",UUID="GPU-...",Hostname="...",instance="host:9400",...
//
// (comma-separated key="value" pairs, metric in __name__). No trailing brace or sample value is required;
// numeric sample value may appear as the last space-separated token, or in a "value" label.
func parseLabelListTelemetry(payload []byte) (*pb.TelemetryMessage, error) {
	raw := strings.TrimSpace(string(bytes.TrimSpace(payload)))
	if raw == "" || !looksLikeCommaLabelListLine(raw) {
		return nil, nil
	}
	// Avoid stealing normal Prometheus lines (handled elsewhere).
	if strings.Contains(raw, "{") && strings.Contains(raw, "}") {
		return nil, nil
	}

	line := raw
	trailVal := 0.0
	if len(raw) > 0 {
		last := raw[len(raw)-1]
		if (last >= '0' && last <= '9') || last == '.' || last == 'e' || last == 'E' {
			if b, v, ok := splitTrailingSampleValue(raw); ok {
				line = b
				trailVal = v
			}
		}
	}

	labels, err := parseCommaQuotedLabelList(line)
	if err != nil {
		// Binary / wrong wire often trips this path; fall back to CSV instead of a noisy decode error.
		return nil, nil
	}
	metric := labelGetCI(labels, "__name__", "metric_name", "metricName", "name")
	if metric == "" {
		return nil, nil
	}

	if vStr := labelGetCI(labels, "value", "Value"); vStr != "" {
		if v, err := strconv.ParseFloat(strings.TrimSpace(vStr), 64); err == nil {
			trailVal = v
		}
	}

	tsNano := time.Now().UnixNano()
	if vStr := labelGetCI(labels, "processed_at_unix_nano", "processedAtUnixNano"); vStr != "" {
		if n, err := strconv.ParseInt(strings.TrimSpace(vStr), 10, 64); err == nil && n > 0 {
			tsNano = n
		}
	}

	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("label list json: %w", err)
	}

	gpuID := firstNonEmpty(
		labelGetCI(labels, "gpu_id", "gpuId"),
		labelGetCI(labels, "gpu", "GPU"),
		"unknown",
	)
	host := firstNonEmpty(
		labelGetCI(labels, "Hostname", "hostname", "host_name", "hostName"),
	)
	if host == "" {
		host = hostFromInstance(labelGetCI(labels, "instance", "Instance"))
	}
	if host == "" {
		host = "unknown"
	}
	uuid := firstNonEmpty(labelGetCI(labels, "UUID", "uuid"))
	if uuid == "" {
		uuid = stableUUID(metric + "\x00" + line)
	}

	return &pb.TelemetryMessage{
		MetricName:          metric,
		GpuId:               gpuID,
		Device:              labelGetCI(labels, "device", "Device"),
		Uuid:                uuid,
		ModelName:           labelGetCI(labels, "modelName", "model_name", "ModelName"),
		HostName:            host,
		Value:               trailVal,
		LabelsRaw:           string(labelsJSON),
		ProcessedAtUnixNano: tsNano,
	}, nil
}

// splitTrailingSampleValue splits "…labels… 12.5" when the last token is a float (OpenMetrics / scrape lines).
func splitTrailingSampleValue(line string) (before string, value float64, ok bool) {
	line = strings.TrimSpace(line)
	last := -1
	for j := len(line) - 1; j >= 0; j-- {
		if line[j] == ' ' || line[j] == '\t' {
			last = j
			break
		}
	}
	if last < 0 {
		return line, 0, false
	}
	tail := strings.TrimSpace(line[last+1:])
	if tail == "" {
		return line, 0, false
	}
	v, err := strconv.ParseFloat(tail, 64)
	if err != nil {
		return line, 0, false
	}
	return strings.TrimSpace(line[:last]), v, true
}

func parseCommaQuotedLabelList(s string) (map[string]string, error) {
	out := make(map[string]string)
	s = strings.TrimSpace(s)
	i := 0
	for i < len(s) {
		for i < len(s) && (s[i] == ' ' || s[i] == ',') {
			i++
		}
		if i >= len(s) {
			break
		}
		start := i
		for i < len(s) && s[i] != '=' {
			i++
		}
		key := strings.TrimSpace(s[start:i])
		if key == "" {
			break
		}
		if !isASCIIExporterLabelKey(key) {
			return nil, fmt.Errorf("invalid label key %q", key)
		}
		if i >= len(s) || s[i] != '=' {
			return nil, fmt.Errorf("expected = after key %q", key)
		}
		i++
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= len(s) {
			return nil, fmt.Errorf("missing value for key %q", key)
		}
		if s[i] == '"' {
			i++
			var b strings.Builder
			for i < len(s) {
				if s[i] == '\\' {
					i++
					if i >= len(s) {
						return nil, fmt.Errorf("dangling escape in %q", key)
					}
				} else if s[i] == '"' {
					break
				}
				b.WriteByte(s[i])
				i++
			}
			if i >= len(s) || s[i] != '"' {
				return nil, fmt.Errorf("unclosed quote for key %q", key)
			}
			i++
			out[key] = b.String()
			continue
		}
		startVal := i
		for i < len(s) && s[i] != ',' {
			i++
		}
		out[key] = strings.TrimSpace(s[startVal:i])
	}
	return out, nil
}

func labelGetCI(m map[string]string, keys ...string) string {
	for _, want := range keys {
		if v, ok := m[want]; ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	for mk, mv := range m {
		for _, want := range keys {
			if strings.EqualFold(mk, want) && strings.TrimSpace(mv) != "" {
				return strings.TrimSpace(mv)
			}
		}
	}
	return ""
}

// looksLikeCommaLabelListLine is a strict gate so random binary (which can contain "__name__" bytes)
// is not parsed as text. Requires UTF-8 and the Prometheus metric marker as an assigned quoted key.
func looksLikeCommaLabelListLine(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	t := strings.TrimSpace(s)
	if !strings.Contains(t, "=") || !strings.Contains(t, ",") {
		return false
	}
	return hasAssignedLabelKey(t, "__name__") ||
		hasAssignedLabelKey(t, "metric_name") ||
		hasAssignedLabelKey(t, "metricName") ||
		hasAssignedLabelKey(t, "name")
}

func hasAssignedLabelKey(s, key string) bool {
	for i := 0; i < len(s); i++ {
		if !strings.HasPrefix(s[i:], key) {
			continue
		}
		prev := byte(',')
		if i > 0 {
			prev = s[i-1]
		}
		if i > 0 && prev != ',' && prev != ' ' && prev != '\t' {
			continue
		}
		j := i + len(key)
		for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
			j++
		}
		if j >= len(s) || s[j] != '=' {
			continue
		}
		return true
	}
	return false
}

func isASCIIExporterLabelKey(key string) bool {
	if key == "" || len(key) > 256 {
		return false
	}
	for _, r := range key {
		if r < 0x20 || r > 0x7e {
			return false
		}
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_', r == ':', r == '.', r == '-':
		default:
			return false
		}
	}
	return true
}
