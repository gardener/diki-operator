{{- define "name" -}}
diki-operator
{{- end -}}

{{- define "leaderelectionid" -}}
diki-operator-leader-election
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
  complianceRun:
    syncPeriod: {{ .Values.config.controllers.complianceRun.syncPeriod }}
    dikiRunner:
      waitInterval: {{ .Values.config.controllers.complianceRun.dikiRunner.waitInterval }}
      podCompletionTimeout: {{ .Values.config.controllers.complianceRun.dikiRunner.podCompletionTimeout }}
      execTimeout: {{ .Values.config.controllers.complianceRun.dikiRunner.execTimeout }}
      {{- if .Values.config.controllers.complianceRun.dikiRunner.namespace }}
      namespace: {{ .Values.config.controllers.complianceRun.dikiRunner.namespace }}
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
leaderElection:
  resourceName: {{ include "leaderelectionid" . }}
  {{- if .Values.config.leaderElection.resourceNamespace}}
  resourceNamespace: {{ .Values.config.leaderElection.resourceNamespace }}
  {{- else }}
  resourceNamespace: {{ .Release.Namespace }}
  {{- end }}
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
