# Webhook Output

The Webhook output sends the compliance scan report as a JSON payload via an HTTP POST request to a configured endpoint.

## Configuration

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: example-webhook-output
spec:
  output:
    webhook:
      url: "https://compliance-api.corp.example.com/v1/reports"
      credentialsRef:
        name: webhook-headers
        namespace: kube-system
      tls:
        caSecretRef:
          name: webhook-ca
          namespace: kube-system
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | string | **Yes** | - | The destination endpoint URL. The scheme (`http://` or `https://`) determines whether TLS is used. |
| `credentialsRef` | [SecretReference](#secretreference) | No | - | Reference to a Secret containing HTTP headers to include in the request. |
| `tls` | [TLSConfig](#tlsconfig) | No | - | TLS settings for HTTPS connections. Only relevant when the URL uses the `https` scheme. |

### SecretReference

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | **Yes** | - | Name of the Secret. |
| `namespace` | string | **Yes** | - | Namespace of the Secret. |
| `key` | string | No | `headers` (for credentials) / `ca.crt` (for CA) | The key within the Secret's data to read. |

### TLSConfig

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `insecureSkipVerify` | bool | No | `false` | Disables TLS certificate verification. Use with caution. |
| `caSecretRef` | [SecretReference](#secretreference) | No | - | Reference to a Secret containing a PEM-encoded CA certificate bundle. If not set, the system root CA pool is used. |

## Credentials Secret Format

The Secret referenced by `credentialsRef` must contain a JSON object at the specified key, where keys are HTTP header names and values are the corresponding header values:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: webhook-headers
  namespace: kube-system
stringData:
  headers: |
    {
      "Authorization": "Bearer <token>",
      "X-Custom-Header": "custom-value"
    }
```

This supports any authentication scheme (Bearer tokens, Basic Auth, API keys, etc.) by setting the appropriate headers.

## Behavior

- The report is JSON-marshaled and sent as the request body with `Content-Type: application/json`.
- Credentials and TLS CA certificates are resolved by the operator at reconciliation time - the report-exporter binary does not need access to Secrets.
- A response with HTTP status 2xx is considered successful. Any other status code results in a failed output.

## Output Details

On success, the output status reports the URL and HTTP status code:

```yaml
outputs:
  - outputName: example-webhook-output
    phase: Completed
    details:
      url: "https://compliance-api.corp.example.com/v1/reports"
      statusCode: 200
```

On failure:

```yaml
outputs:
  - outputName: example-webhook-output
    phase: Failed
    details:
      error: "webhook request failed with status 401: {\"error\":\"unauthorized\"}"
```

## Examples

### Plain HTTP (no auth, no TLS)

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: simple-webhook
spec:
  output:
    webhook:
      url: "http://report-collector.monitoring.svc.cluster.local:8080/reports"
```

### HTTPS with Custom CA

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: tls-webhook
spec:
  output:
    webhook:
      url: "https://compliance.internal.corp:443/api/v1/reports"
      tls:
        caSecretRef:
          name: internal-ca
          namespace: kube-system
          key: ca.crt
```
