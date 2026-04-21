# diki-operator

[![REUSE status](https://api.reuse.software/badge/github.com/gardener/diki-operator)](https://api.reuse.software/info/github.com/gardener/diki-operator)
[![Build](https://github.com/gardener/diki-operator/actions/workflows/non-release.yaml/badge.svg)](https://github.com/gardener/diki-operator/actions/workflows/non-release.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/gardener/diki-operator)](https://goreportcard.com/report/github.com/gardener/diki-operator)
[![GoDoc](https://godoc.org/github.com/gardener/diki-operator?status.svg)](https://godoc.org/github.com/gardener/diki-operator)
[![gardener compliance checker](https://badgen.net/badge/gardener/compliance-checker/009f76)](https://github.com/gardener)
[![status alpha](https://badgen.net/badge/status/alpha/d8624d)](https://badgen.net/badge/status/alpha/d8624d)
[![license apache 2.0](https://badgen.net/badge/license/apache-2.0/8ab803)](https://opensource.org/licenses/Apache-2.0)

A Kubernetes operator that orchestrates compliance scanning for Kubernetes clusters using the [Diki](https://github.com/gardener/diki) compliance checker. Part of the [Gardener](https://github.com/gardener/gardener) ecosystem.

## Overview

diki-operator automates the execution of compliance scans against Kubernetes clusters and manages the export of scan reports. It provides:

- **ComplianceScan CRD** -- define and trigger compliance scans
- **ScheduledComplianceScan CRD** -- schedule recurring compliance scans
- **Report export** -- configurable outputs for diki scan reports

### Custom Resources

#### ComplianceScan

Cluster-scoped resource that specifies which rulesets to run, references optional ConfigMaps for ruleset/rule options, and tracks scan status (Pending, Running, Completed, Failed).

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ComplianceScan
metadata:
  name: example-compliancescan
spec:
  rulesets:
    - id: disa-kubernetes-stig
      version: v2r4
      options:
        ruleset:
          configMapRef:
            name: diki-options
            namespace: kube-system
  outputs:
    - name: compliance-scan-report
```

#### ScheduledComplianceScan

Cluster-scoped resource that defines a cron schedule for recurring ComplianceScans, with configurable history limits.

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ScheduledComplianceScan
metadata:
  name: weekly-scan
spec:
  schedule: "0 0 * * 0"  # every Sunday at midnight
  successfulScansHistoryLimit: 3
  failedScansHistoryLimit: 1
  scanTemplate:
    spec:
      rulesets:
        - id: disa-kubernetes-stig
          version: v2r4
      outputs:
        - name: compliance-scan-report
```

#### ReportOutput

Cluster-scoped resource that defines where compliance reports should be stored.

```yaml
apiVersion: diki.gardener.cloud/v1alpha1
kind: ReportOutput
metadata:
  name: compliance-scan-report
spec:
  output:
    configMap:
      namespace: kube-system
      namePrefix: compliance-scan-report-
```

## Development

For local setup instructions, see the [Getting Started Locally](docs/getting-started-locally.md) guide.

## Feedback and Support

Feedback and contributions are always welcome!

Please report bugs or suggestions as [GitHub issues](https://github.com/gardener/diki-operator/issues) or reach out on [Slack](https://gardener-cloud.slack.com/) (join the workspace [here](https://gardener.cloud/community)).

<p align="center"><img alt="Bundesministerium für Wirtschaft und Energie (BMWE)-EU funding logo" src="https://apeirora.eu/assets/img/BMWK-EU.png" width="400"/></p>
