# ConfigMap Output

The ConfigMap output stores the compliance scan report as a gzipped JSON file inside a Kubernetes ConfigMap.

## Configuration

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: example-configmap-output
spec:
  output:
    configMap:
      namespace: kube-system
      namePrefix: compliance-scan-report-
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `namespace` | string | No | `kube-system` | The namespace where the ConfigMap will be created. |
| `namePrefix` | string | No | `compliance-scan-report-` | Prefix for the generated ConfigMap name. A unique suffix is appended automatically. |

## Behavior

- The report is JSON-marshaled, gzip-compressed, and stored under the key `report.json.gz` in the ConfigMap's `binaryData`.
- A new ConfigMap is created for each scan execution (using `generateName`).
- The ConfigMap is labeled with the scan's name and UID for identification.

## Output Details

On success, the output status reports the created ConfigMap's name and namespace:

```yaml
outputs:
  - outputName: example-configmap-output
    phase: Completed
    details:
      configMapRef:
        name: compliance-scan-report-a1b2c3
        namespace: kube-system
```

## Reading the Report

To extract and read the report from the ConfigMap:

```bash
kubectl get configmap <name> -n kube-system -o jsonpath='{.binaryData.report\.json\.gz}' | base64 -d | gunzip | jq .
```

## Example

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: my-configmap-output
spec:
  output:
    configMap:
      namespace: monitoring
      namePrefix: diki-report-
```
