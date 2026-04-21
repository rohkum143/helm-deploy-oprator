{{/*
Expand the name of the chart.
*/}}
{{- define "nginx-helm-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "nginx-helm-operator.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "nginx-helm-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "nginx-helm-operator.labels" -}}
helm.sh/chart: {{ include "nginx-helm-operator.chart" . }}
{{ include "nginx-helm-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "nginx-helm-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "nginx-helm-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "nginx-helm-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "nginx-helm-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the name of the namespace to use
*/}}
{{- define "nginx-helm-operator.namespaceName" -}}
{{- if .Values.namespace.name }}
{{- .Values.namespace.name }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "nginx-helm-operator.annotations" -}}
{{- with .Values.commonAnnotations }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Image pull secrets
*/}}
{{- define "nginx-helm-operator.imagePullSecrets" -}}
{{- with .Values.imagePullSecrets }}
imagePullSecrets:
{{- toYaml . | nindent 2 }}
{{- end }}
{{- end }}

{{/*
Operator image
*/}}
{{- define "nginx-helm-operator.image" -}}
{{- printf "%s:%s" .Values.image.repository (.Values.image.tag | default .Chart.AppVersion) }}
{{- end }}

{{/*
Validate required values
*/}}
{{- define "nginx-helm-operator.validateValues" -}}
{{- if not .Values.image.repository }}
{{- fail "image.repository is required" }}
{{- end }}
{{- if not .Values.image.tag }}
{{- fail "image.tag is required" }}
{{- end }}
{{- end }}