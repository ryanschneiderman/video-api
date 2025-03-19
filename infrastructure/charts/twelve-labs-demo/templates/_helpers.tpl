{{/*
Expand the name of the chart.
*/}}
{{- define "twelve-labs-demo.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
This value is used to generate resource names and is limited to 63 characters.
*/}}
{{- define "twelve-labs-demo.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "twelve-labs-demo.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels that should be attached to all resources.
*/}}
{{- define "twelve-labs-demo.labels" -}}
helm.sh/chart: {{ include "twelve-labs-demo.chart" . | quote }}
{{ include "twelve-labs-demo.selectorLabels" . }}
{{- with .Chart.AppVersion }}
app.kubernetes.io/version: {{ . | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
{{- end -}}

{{/*
Labels used by the selector.
*/}}
{{- define "twelve-labs-demo.selectorLabels" -}}
app.kubernetes.io/name: {{ include "twelve-labs-demo.name" . | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
{{- end -}}

{{/*
Generate a standard name for a Kubernetes resource based on the chart name.
*/}}
{{- define "twelve-labs-demo.fullnameOverride" -}}
{{- default (include "twelve-labs-demo.fullname" .) .Values.fullnameOverride -}}
{{- end -}}
