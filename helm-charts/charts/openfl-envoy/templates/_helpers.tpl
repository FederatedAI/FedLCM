# Copyright 2022 VMware, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

{{/* Helm required labels */}}
{{- define "openfl-envoy.labels" -}}
name: {{ .Values.name | quote  }}
owner: kubefate
cluster: openfl-envoy
heritage: {{ .Release.Service }}
release: {{ .Release.Name }}
chart: {{ .Chart.Name }}
{{- end -}}

{{/* matchLabels */}}
{{- define "openfl-envoy.matchLabels" -}}
name: {{ .Values.name | quote  }}
{{- end -}}

{{/*
Create the name of the controller service account to use
*/}}
{{- define "serviceAccountName" -}}
{{- if .Values.podSecurityPolicy.enabled -}}
    {{ default .Values.name }}
{{- else -}}
    {{ default "default" }}
{{- end -}}
{{- end -}}