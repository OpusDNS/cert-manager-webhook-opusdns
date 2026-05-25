{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "cert-manager-webhook-opusdns.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "cert-manager-webhook-opusdns.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cert-manager-webhook-opusdns.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "cert-manager-webhook-opusdns.labels" -}}
app: {{ include "cert-manager-webhook-opusdns.name" . }}
chart: {{ include "cert-manager-webhook-opusdns.chart" . }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
{{- end -}}

{{- define "cert-manager-webhook-opusdns.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "cert-manager-webhook-opusdns.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-opusdns.rootCAIssuer" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-opusdns.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-opusdns.rootCACertificate" -}}
{{ printf "%s-ca" (include "cert-manager-webhook-opusdns.fullname" .) }}
{{- end -}}

{{- define "cert-manager-webhook-opusdns.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "cert-manager-webhook-opusdns.fullname" .) }}
{{- end -}}
