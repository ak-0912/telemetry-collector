{{- define "telemetry-collector.name" -}}
telemetry-collector
{{- end -}}

{{- define "telemetry-collector.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{ include "telemetry-collector.name" . }}
{{- end }}
{{- end -}}
