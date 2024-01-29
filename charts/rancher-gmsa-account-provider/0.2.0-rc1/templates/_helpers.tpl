# Rancher
{{- define "system_default_registry" -}}
{{- if .Values.global.cattle.systemDefaultRegistry -}}
{{- printf "%s/" .Values.global.cattle.systemDefaultRegistry -}}
{{- else -}}
{{ template "windows_default_registry" . }}
{{- end -}}
{{- end -}}

# Windows Support

{{- define "windows_default_registry" -}}
{{- $image := default .Values.image (get . "imageOverride") -}}
{{- if has $image.repository (list "windows/servercore" "windows/nanoserver" "windows/server" "windows" "windows/servercore/iis" "oss/kubernetes/windows-host-process-containers-base-image") -}}
{{- printf "%s/" "mcr.microsoft.com" -}}
{{- end -}}
{{- end -}}

{{/*
Windows cluster will add default taint for linux nodes,
add below linux tolerations to workloads could be scheduled to those linux nodes
*/}}

{{- define "linux-node-tolerations" -}}
- key: "cattle.io/os"
  value: "linux"
  effect: "NoSchedule"
  operator: "Equal"
{{- end -}}

{{- define "linux-node-selector" -}}
{{- if semverCompare "<1.14-0" .Capabilities.KubeVersion.GitVersion -}}
beta.kubernetes.io/os: linux
{{- else -}}
kubernetes.io/os: linux
{{- end -}}
{{- end -}}

# Rancher GMSA Account Provider

{{/* vim: set filetype=mustache: */}}
{{/* Expand the name of the chart. This is suffixed with -alertmanager, which means subtract 13 from longest 63 available */}}
{{- define "rancher-gmsa-account-provider.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 50 | trimSuffix "-" -}}
{{- end }}

{{/*
Allow the release namespace to be overridden for multi-namespace deployments in combined charts
*/}}
{{- define "rancher-gmsa-account-provider.namespace" -}}
  {{- if .Values.namespaceOverride -}}
    {{- .Values.namespaceOverride -}}
  {{- else -}}
    {{- .Release.Namespace -}}
  {{- end -}}
{{- end -}}

{{/* Create chart name and version as used by the chart label. */}}
{{- define "rancher-gmsa-account-provider.chartref" -}}
{{- replace "+" "_" .Chart.Version | printf "%s-%s" .Chart.Name -}}
{{- end }}

{{/* Generate basic labels */}}
{{- define "rancher-gmsa-account-provider.labels" -}}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: "{{ replace "+" "_" .Chart.Version }}"
app.kubernetes.io/part-of: {{ template "rancher-gmsa-account-provider.name" . }}
chart: {{ template "rancher-gmsa-account-provider.chartref" . }}
release: {{ $.Release.Name | quote }}
heritage: {{ $.Release.Service | quote }}
{{- end -}}

# Cert Manager

{{/* Determine apiVersion for cert-manager */}}
{{- define "cert-manager.apiversion" -}}
  {{- $certmanagerVer :=  split "." .Values.certificates.certManager.version -}}
  {{- if or (.Capabilities.APIVersions.Has "cert-manager.io/v1") (and (gt (len $certmanagerVer._0) 0) (eq (int $certmanagerVer._0) 1) (ge (int $certmanagerVer._1) 0)) }}
apiVersion: cert-manager.io/v1
  {{- else if or (.Capabilities.APIVersions.Has "cert-manager.io/v1beta1") (and (gt (len $certmanagerVer._0) 0) (eq (int $certmanagerVer._0) 0) (ge (int $certmanagerVer._1) 16)) }}
apiVersion: cert-manager.io/v1beta1
  {{- else if or (.Capabilities.APIVersions.Has "cert-manager.io/v1alpha2") (and (gt (len $certmanagerVer._0) 0) (eq (int $certmanagerVer._0) 0) (ge (int $certmanagerVer._1) 11)) }}
apiVersion: cert-manager.io/v1alpha2
  {{- else if or (.Capabilities.APIVersions.Has "certmanager.k8s.io/v1alpha1") (and (gt (len $certmanagerVer._0) 0) (eq (int $certmanagerVer._0) 0) (lt (int $certmanagerVer._1) 11)) }}
apiVersion: cert-manager.io/v1alpha1
  {{- else }}
apiVersion: cert-manager.io/v1
  {{- end }}
{{- end }}

