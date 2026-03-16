{{- define "food-delivery-system.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "food-delivery-system.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "food-delivery-system.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "food-delivery-system.labels" -}}
helm.sh/chart: {{ include "food-delivery-system.chart" . }}
{{ include "food-delivery-system.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.global.labels }}
{{- toYaml . | nindent 0 }}
{{- end }}
{{- end }}

{{- define "food-delivery-system.selectorLabels" -}}
app.kubernetes.io/name: {{ include "food-delivery-system.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "food-delivery-system.image" -}}
{{- if .global.imageRegistry }}
{{- printf "%s/%s:%s" .global.imageRegistry .image .tag }}
{{- else }}
{{- printf "%s:%s" .image .tag }}
{{- end }}
{{- end }}

{{- define "food-delivery-system.namespace" -}}
{{- .Values.global.namespace | default .Release.Namespace }}
{{- end }}
