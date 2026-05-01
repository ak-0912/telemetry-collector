{{- define "telemetry-collector.name" -}}
telemetry-collector
{{- end -}}

{{- define "telemetry-collector.fullname" -}}
{{ include "telemetry-collector.name" . }}
{{- end -}}
