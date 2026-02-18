# Getting Started Locally

## Local KinD Setup

This document will walk you through running a KinD cluster on your local machine and installing the diki-operator in it.

### 1. Create KinD cluster and deploy the diki-operator

```bash
make kind-up
make operator-up
```

You can now target the KinD cluster.

```bash
export KUBECONFIG=$(pwd)/dev/local/kind/kubeconfig
```

This setup will deploy the diki-operator in the `kube-system` namespace.

### 2. Verify setup

Verify that `ComplianceRun` resources are successfully processed.

```bash
k apply -f ./example/80-dikioptionscm.yaml
k apply -f ./example/90-compliancerun.yaml
```

## Cleanup

To tear down the local environment:

```bash
make kind-down
```
