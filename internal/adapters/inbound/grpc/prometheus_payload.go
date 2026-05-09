package grpc

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	pb "telemetry-collector/api/telemetry/v1"
)

// parsePrometheusSample parses one Prometheus/OpenMetrics text exposition sample line
// (e.g. `metric{label="v"} 1.5 [ts_ms]`) before falling back to CSV.
// Returns (nil, nil) if the payload does not look like a Prometheus sample.
func parsePrometheusSample(payload []byte) (*pb.TelemetryMessage, error) {
	line := prometheusSampleText(payload)
	if line == "" {
		return nil, nil
	}
	metric, labelsStr, remainder, ok := splitPrometheusMetricLine(line)
	if !ok || metric == "" || !isPrometheusMetricName(metric) {
		return nil, nil
	}
	labels, err := parsePrometheusLabels(labelsStr)
	if err != nil {
		return nil, fmt.Errorf("prometheus labels: %w", err)
	}
	fields := strings.Fields(remainder)
	if len(fields) < 1 {
		return nil, nil
	}
	val, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return nil, nil
	}
	var tsNano int64
	if len(fields) >= 2 {
		tsMs, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("prometheus timestamp: %w", err)
		}
		tsNano = tsMs * 1_000_000
	} else {
		tsNano = time.Now().UnixNano()
	}
	if tsNano <= 0 {
		tsNano = time.Now().UnixNano()
	}

	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("prometheus labels json: %w", err)
	}

	gpuID := firstNonEmpty(labels["gpu_id"], labels["gpu"], labels["device_id"], "unknown")
	host := firstNonEmpty(labels["host_name"], labels["hostname"])
	if host == "" {
		host = hostFromInstance(labels["instance"])
	}
	if host == "" {
		host = "unknown"
	}
	uuid := firstNonEmpty(labels["uuid"], labels["UUID"])
	if uuid == "" {
		uuid = stableUUID(metric + "\x00" + labelsStr + "\x00" + fields[0])
	}

	return &pb.TelemetryMessage{
		MetricName:          metric,
		GpuId:               gpuID,
		Device:              labels["device"],
		Uuid:                uuid,
		ModelName:           labels["model_name"],
		HostName:            host,
		Value:               val,
		LabelsRaw:           string(labelsJSON),
		ProcessedAtUnixNano: tsNano,
	}, nil
}

// prometheusSampleText returns one logical exposition line: merges `metric{labels}\n value`
// when the value is on the following line (common scrape / file formats).
func prometheusSampleText(payload []byte) string {
	s := string(bytes.TrimSpace(payload))
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return ""
	}
	first := lines[0]
	_, _, rem, ok := splitPrometheusMetricLine(first)
	if !ok {
		return first
	}
	if strings.TrimSpace(rem) != "" {
		return first
	}
	if len(lines) < 2 {
		return first
	}
	second := strings.TrimSpace(lines[1])
	if second == "" || strings.HasPrefix(second, "#") {
		return first
	}
	return first + " " + second
}

// splitPrometheusMetricLine splits `name{labels} rest` or `name rest` (no labels).
func splitPrometheusMetricLine(line string) (metricName, labelsInner, remainder string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", "", false
	}
	i := strings.IndexByte(line, '{')
	if i < 0 {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			return "", "", "", false
		}
		return fields[0], "", strings.Join(fields[1:], " "), true
	}
	metricName = line[:i]
	if metricName == "" {
		return "", "", "", false
	}
	depth := 0
	for j := i; j < len(line); j++ {
		switch line[j] {
		case '"':
			j++
			for j < len(line) {
				if line[j] == '\\' {
					j++
					if j >= len(line) {
						break
					}
				} else if line[j] == '"' {
					break
				}
				j++
			}
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				labelsInner = line[i+1 : j]
				remainder = strings.TrimSpace(line[j+1:])
				return metricName, labelsInner, remainder, true
			}
		}
	}
	return "", "", "", false
}

func parsePrometheusLabels(inner string) (map[string]string, error) {
	out := make(map[string]string)
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return out, nil
	}
	i := 0
	for i < len(inner) {
		for i < len(inner) && (inner[i] == ' ' || inner[i] == ',') {
			i++
		}
		if i >= len(inner) {
			break
		}
		start := i
		for i < len(inner) && inner[i] != '=' {
			i++
		}
		key := strings.TrimSpace(inner[start:i])
		if i >= len(inner) || inner[i] != '=' {
			return nil, fmt.Errorf("label parse: expected = after %q", key)
		}
		i++
		for i < len(inner) && inner[i] == ' ' {
			i++
		}
		if i >= len(inner) || inner[i] != '"' {
			return nil, fmt.Errorf("label parse: expected opening quote for %q", key)
		}
		i++
		var b strings.Builder
		for i < len(inner) {
			if inner[i] == '\\' {
				i++
				if i >= len(inner) {
					return nil, fmt.Errorf("label parse: dangling escape in %q", key)
				}
			} else if inner[i] == '"' {
				break
			}
			b.WriteByte(inner[i])
			i++
		}
		if i >= len(inner) || inner[i] != '"' {
			return nil, fmt.Errorf("label parse: unclosed value for %q", key)
		}
		i++
		out[key] = b.String()
	}
	return out, nil
}

func isPrometheusMetricName(s string) bool {
	if s == "" {
		return false
	}
	r0, w0 := utf8.DecodeRuneInString(s)
	if r0 == utf8.RuneError && w0 == 1 {
		return false
	}
	if !(r0 == '_' || r0 == ':' || unicode.IsLetter(r0)) {
		return false
	}
	for _, r := range s[w0:] {
		if r == '_' || r == ':' || r == '.' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		return false
	}
	return true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func hostFromInstance(instance string) string {
	instance = strings.TrimSpace(instance)
	if instance == "" {
		return ""
	}
	host, _, err := net.SplitHostPort(instance)
	if err == nil {
		return host
	}
	return instance
}

func stableUUID(fingerprint string) string {
	sum := sha256.Sum256([]byte(fingerprint))
	h := hex.EncodeToString(sum[:16])
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}
