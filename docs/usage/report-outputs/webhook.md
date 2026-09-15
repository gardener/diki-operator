# Webhook Output

The Webhook output sends the compliance scan report as a JSON payload via a configurable HTTP method (POST, PUT) to a configured endpoint.

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
      method: "POST"
      credentialsRef:
        name: webhook-headers
        namespace: kube-system
      tls:
        caConfigMapRef:
          name: webhook-ca
          namespace: kube-system
```

## Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `url` | string | **Yes** | - | The destination endpoint URL. Must use the `https://` scheme. |
| `method` | string | No | `POST` | The HTTP method used to send the report. Allowed values: `POST`, `PUT`. |
| `credentialsRef` | [CredentialsSecretRef](#credentialssecretref) | No | - | Reference to a Secret containing HTTP headers to include in the request. |
| `tls` | [TLSConfig](#tlsconfig) | No | - | TLS settings for HTTPS connections. Only relevant when the URL uses the `https` scheme. |

### CredentialsSecretRef

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | **Yes** | - | Name of the Secret. |
| `namespace` | string | **Yes** | - | Namespace of the Secret. |
| `headersKey` | string | No | `headers` | The key within the Secret's data that contains the JSON-encoded headers. |

### TLSConfig

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `caConfigMapRef` | [CAConfigMapRef](#caconfigmapref) | No | - | Reference to a ConfigMap containing a PEM-encoded CA certificate bundle. If not set, the system root CA pool is used. |
| `mtlsSecretRef` | [MTLSSecretRef](#mtlssecretref) | No | - | Reference to a Kubernetes TLS Secret containing a client certificate and key for mutual TLS (mTLS) authentication. |

### CAConfigMapRef

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | **Yes** | - | Name of the ConfigMap. |
| `namespace` | string | **Yes** | - | Namespace of the ConfigMap. |
| `key` | string | No | `ca.crt` | The key within the ConfigMap's data that contains the PEM-encoded CA certificate(s). |

### MTLSSecretRef

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | **Yes** | - | Name of the Secret. |
| `namespace` | string | **Yes** | - | Namespace of the Secret. |
| `certKey` | string | No | `tls.crt` | The key within the Secret's data that contains the PEM-encoded client certificate. |
| `privateKey` | string | No | `tls.key` | The key within the Secret's data that contains the PEM-encoded client private key. |

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
- Credentials, TLS CA certificates and client certificates are resolved by the operator at reconciliation time - the report-exporter binary does not need access to Secrets or ConfigMaps.
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

### Minimal HTTPS

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: simple-webhook
spec:
  output:
    webhook:
      url: "https://report-collector.monitoring.svc.cluster.local:8443/reports"
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
        caConfigMapRef:
          name: internal-ca
          namespace: kube-system
          key: ca.crt
```

### HTTPS with mTLS (Mutual TLS)

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: mtls-webhook
spec:
  output:
    webhook:
      url: "https://compliance.internal.corp:443/api/v1/reports"
      tls:
        caConfigMapRef:
          name: internal-ca
          namespace: kube-system
        mtlsSecretRef:
          name: webhook-client-tls
          namespace: kube-system
```

The mTLS Secret must be a standard Kubernetes TLS Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: webhook-client-tls
  namespace: kube-system
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-client-certificate>
  tls.key: <base64-encoded-client-private-key>
```

Custom key names can be specified if the Secret uses non-standard keys:

```yaml
tls:
  mtlsSecretRef:
    name: webhook-client-tls
    namespace: kube-system
    certKey: client.crt
    privateKey: client.key
```
