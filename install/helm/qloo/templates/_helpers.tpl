{{/* vim: set filetype=mustache: */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "qloo-helm.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "qloo-helm.fullname" -}}
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
{{- define "qloo-helm.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* helpers for components */}}

{{/* function discovery */}}
{{- define "function_discovery.name" -}}
{{- default "function-discovery" .Values.function_discovery.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "function_discovery.fullname" -}}
{{- $name := default "function-discovery" .Values.function_discovery.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* ingress */}}
{{- define "ingress.name" -}}
{{- default "ingress" .Values.ingress.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "ingress.fullname" -}}
{{- $name := default "ingress" .Values.ingress.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* gloo */}}
{{- define "control_plane.name" -}}
{{- default "control-plane" .Values.control_plane.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "control_plane.fullname" -}}
{{- $name := default "control-plane" .Values.control_plane.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Jaeger related templates */}}
{{- define "jaeger.name" -}}
{{ printf "%s-%s" .Release.Name "jaeger" | trunc 63 | trimSuffix "-"}}
{{- end -}}

{{- define "jaeger.host" -}}
{{- if eq .Values.opentracing.status "configure" -}}
{{- .Values.opentracing.jaeger_host -}}
{{- else -}}
{{-  template "jaeger.name" . -}}.{{ .Release.Namespace }}.svc.cluster.local
{{- end -}}
{{- end -}}

{{- define "jaeger.port" -}}
{{- if eq .Values.opentracing.status "configure" -}}
{{- .Values.opentracing.jaeger_port -}}
{{- else -}}9411{{- end -}}
{{- end -}}

{{/* Statsd related templates */}}
{{- define "statsd.name" -}}
{{- printf "%s-%s" .Release.Name "statsd" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "statsd.host" -}}
{{- if eq .Values.metrics.status "statsd" -}}
{{- .Values.metrics.statsd_host -}}
{{- else -}}
{{- template "statsd.name" . -}}.{{ .Release.Namespace }}.svc.cluster.local
{{- end -}}
{{- end -}}

{{- define "statsd.port" -}}
{{- if eq .Values.metrics.status "statsd" -}}
{{- .Values.metrics.statsd_port -}}
{{- else -}}9125{{- end -}}
{{- end -}}