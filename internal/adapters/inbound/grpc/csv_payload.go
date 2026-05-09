package grpc

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	pb "telemetry-collector/api/telemetry/v1"
)

// parseTelemetryPayload tries JSON (object or one-element array), Prometheus text, protobuf
// TelemetryMessage, DCGM-style comma-separated __name__="..." label lists, then CSV.
// CSV header row (metric_name,gpu_id,...) is optional; if the first field equals "metric_name" it is skipped.
func parseTelemetryPayload(payload []byte) (*pb.TelemetryMessage, error) {
	payload = bytes.TrimSpace(payload)
	payload = bytes.TrimPrefix(payload, []byte{0xEF, 0xBB, 0xBF}) // UTF-8 BOM
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty payload")
	}
	payload = unwrapJSONStringPayload(payload)
	payload = unwrapKnownJSONEnvelopes(payload)
	payload = bytes.TrimSpace(payload)
	if len(payload) == 0 {
		return nil, fmt.Errorf("empty payload")
	}
	if payload[0] == '[' {
		var arr []json.RawMessage
		if err := json.Unmarshal(payload, &arr); err == nil {
			for _, raw := range arr {
				raw = bytes.TrimSpace(raw)
				if len(raw) == 0 || raw[0] != '{' {
					continue
				}
				var msg pb.TelemetryMessage
				if err := json.Unmarshal(raw, &msg); err == nil && msg.GetMetricName() != "" {
					return &msg, nil
				}
			}
		}
	}
	if payload[0] == '{' {
		if json.Valid(payload) {
			var msg pb.TelemetryMessage
			if err := json.Unmarshal(payload, &msg); err == nil && msg.GetMetricName() != "" {
				return &msg, nil
			}
			return nil, fmt.Errorf("valid JSON object but not a TelemetryMessage (need metric_name); send Prometheus text without a leading JSON wrapper, or a JSON string containing the sample line")
		}
	}
	if msg, err := parsePrometheusSample(payload); err != nil {
		return nil, err
	} else if msg != nil {
		return msg, nil
	}
	if msg, ok := tryProtoTelemetryMessage(payload); ok {
		return msg, nil
	}
	if msg, err := parseLabelListTelemetry(payload); err != nil {
		return nil, err
	} else if msg != nil {
		return msg, nil
	}
	return parseTelemetryCSVRow(payload)
}

// unwrapJSONStringPayload decodes a top-level JSON string (MQ may store the sample as one JSON string).
func unwrapJSONStringPayload(payload []byte) []byte {
	p := bytes.TrimSpace(payload)
	if len(p) < 2 || p[0] != '"' {
		return payload
	}
	var s string
	if err := json.Unmarshal(p, &s); err != nil {
		return payload
	}
	return []byte(s)
}

// unwrapKnownJSONEnvelopes extracts a string from common MQ/HTTP JSON wrappers like {"payload":"..."}.
func unwrapKnownJSONEnvelopes(payload []byte) []byte {
	p := bytes.TrimSpace(payload)
	if len(p) == 0 || p[0] != '{' {
		return payload
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(p, &m); err != nil {
		return payload
	}
	if _, ok := m["metric_name"]; ok {
		return payload
	}
	if _, ok := m["metricName"]; ok {
		return payload
	}
	for _, k := range []string{"payload", "body", "message", "data"} {
		raw, ok := m[k]
		if !ok {
			continue
		}
		var s string
		if err := json.Unmarshal(raw, &s); err != nil || strings.TrimSpace(s) == "" {
			continue
		}
		return []byte(s)
	}
	return payload
}

func parseTelemetryCSVRow(payload []byte) (*pb.TelemetryMessage, error) {
	r := csv.NewReader(bytes.NewReader(payload))
	r.TrimLeadingSpace = true
	r.LazyQuotes = true
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("csv: %w", err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("csv: no rows")
	}
	idx := 0
	if len(records[0]) > 0 && strings.EqualFold(strings.TrimSpace(records[0][0]), "metric_name") {
		idx = 1
	}
	if idx >= len(records) {
		return nil, fmt.Errorf("csv: no data rows after header")
	}
	row := records[idx]
	// metric_name,gpu_id,device,uuid,model_name,host_name,value,labels_raw,processed_at_unix_nano
	if len(row) < 9 {
		return nil, fmt.Errorf("csv: need 9 columns, got %d", len(row))
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(row[6]), 64)
	if err != nil {
		return nil, fmt.Errorf("csv value: %w", err)
	}
	nano, err := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("csv processed_at_unix_nano: %w", err)
	}
	return &pb.TelemetryMessage{
		MetricName:          strings.TrimSpace(row[0]),
		GpuId:               strings.TrimSpace(row[1]),
		Device:              strings.TrimSpace(row[2]),
		Uuid:                strings.TrimSpace(row[3]),
		ModelName:           strings.TrimSpace(row[4]),
		HostName:            strings.TrimSpace(row[5]),
		Value:               val,
		LabelsRaw:           strings.TrimSpace(row[7]),
		ProcessedAtUnixNano: nano,
	}, nil
}
