# Report Outputs

Report Outputs define where compliance scan reports are exported after a scan completes. Each `ReportOutput` is a cluster-scoped resource that can be referenced by one or more `ComplianceScan` resources.

## Overview

A `ComplianceScan` references report outputs by name:

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ComplianceScan
metadata:
  name: my-scan
spec:
  rulesets:
    - id: disa-kubernetes-stig
      version: v2r6
  outputs:
    - name: my-configmap-output
    - name: my-webhook-output
```

Each referenced `ReportOutput` defines a single output type with its configuration. Multiple outputs can be used simultaneously, they are executed concurrently after the scan completes.

## Available Output Types

| Type | Description | Documentation |
|------|-------------|---------------|
| [ConfigMap](configmap.md) | Stores the report as a gzipped JSON in a Kubernetes ConfigMap | [configmap.md](configmap.md) |
| [Webhook](webhook.md) | POSTs the report as JSON to an HTTP(S) endpoint | [webhook.md](webhook.md) |

## Output Status

After a scan completes, the status of each output is reported in the `ComplianceScan` status:

```yaml
status:
  phase: Completed
  outputs:
    - outputName: my-configmap-output
      phase: Completed
      details:
        configMapRef:
          name: compliance-scan-report-abc123
          namespace: kube-system
    - outputName: my-webhook-output
      phase: Completed
      details:
        url: "https://compliance-api.corp.example.com/v1/reports"
        statusCode: 200
```

Each output reports either `Completed` or `Failed` independently. A scan is marked as `Failed` if any output fails.
