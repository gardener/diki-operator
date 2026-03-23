{{- define "diki-operator.name" -}}
diki-operator
{{- end -}}

{{- define "leaderelectionid" -}}
diki-operator-leader-election
{{- end -}}

{{- define "labels.app.key" -}}
app.kubernetes.io/name
{{- end -}}
{{- define "labels.app.value" -}}
{{- include "diki-operator.name" . }}
{{- end -}}

{{- define "labels" -}}
{{- include "labels.app.key" . }}: {{ include "labels.app.value" . }}
helm.sh/chart: {{ include "labels.app.value" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{-  define "image" -}}
  {{- if .Values.image.ref -}}
  {{ .Values.image.ref }}
  {{- else -}}
  {{- if hasPrefix "sha256:" .Values.image.tag }}
  {{- printf "%s@%s" .Values.image.repository .Values.image.tag }}
  {{- else }}
  {{- printf "%s:%s" .Values.image.repository .Values.image.tag }}
  {{- end }}
  {{- end -}}
{{- end }}


{{- define "diki-operator.config.data" -}}
config.yaml: |
{{ include "diki-operator.config" . | indent 2 }}
{{- end -}}

{{- define "diki-operator.config" -}}
apiVersion: config.diki.gardener.cloud/v1alpha1
kind: DikiOperatorConfiguration
log:
   level: {{ .Values.config.log.level }}
   format: {{ .Values.config.log.format }}
controllers:
  complianceScan:
    syncPeriod: {{ .Values.config.controllers.complianceScan.syncPeriod }}
    dikiRunner:
      waitInterval: {{ .Values.config.controllers.complianceScan.dikiRunner.waitInterval }}
      podCompletionTimeout: {{ .Values.config.controllers.complianceScan.dikiRunner.podCompletionTimeout }}
      execTimeout: {{ .Values.config.controllers.complianceScan.dikiRunner.execTimeout }}
      {{- if .Values.config.controllers.complianceScan.dikiRunner.namespace }}
      namespace: {{ .Values.config.controllers.complianceScan.dikiRunner.namespace }}
      {{- else }}
      namespace: {{ .Release.Namespace }}
      {{- end }}
server:
  healthProbes:
    port: {{ .Values.config.server.healthProbes.port }}
  metrics:
    {{- if .Values.config.server.metrics.bindAddress }}
    bindAddress: {{ .Values.config.server.metrics.bindAddress }}
    {{- end }}
    port: {{ .Values.config.server.metrics.port }}
  webhooks:
    port: {{ .Values.config.server.webhooks.port }}
    tls:
      serverCertDir: /etc/diki-operator/webhooks/tls
leaderElection:
  resourceName: {{ include "leaderelectionid" . }}
  resourceNamespace: {{ .Release.Namespace }}
  {{- if .Values.config.leaderElection.leaderElect }}
  leaderElect: {{ .Values.config.leaderElection.leaderElect }}
  {{- end }}
  {{- if .Values.config.leaderElection.leaseDuration }}
  leaseDuration: {{ .Values.config.leaderElection.leaseDuration }}
  {{- end }}
  {{- if .Values.config.leaderElection.renewDeadline }}
  renewDeadline: {{ .Values.config.leaderElection.renewDeadline }}
  {{- end }}
  {{- if .Values.config.leaderElection.retryPeriod }}
  retryPeriod: {{ .Values.config.leaderElection.retryPeriod }}
  {{- end }}
  {{- if .Values.config.leaderElection.resourceLock }}
  resourceLock: {{ .Values.config.leaderElection.resourceLock }}
  {{- end }}
{{- end -}}
