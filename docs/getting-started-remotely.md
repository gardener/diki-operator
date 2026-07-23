# Getting Started Remotely

## Remote KinD Setup

This document will walk you through running two KinD clusters on your local machine and installing the diki-operator in a remote scanning configuration. In this mode, the diki-operator and diki-run Job run on a **source** cluster while scanning a **target** cluster.

### 1. Create KinD clusters and deploy the diki-operator

```bash
make remote-kind-up
make remote-operator-up
```

You can now target the clusters individually.

```bash
export KUBECONFIG=$(pwd)/dev/local/remote-kind/source-kubeconfig
# or
export KUBECONFIG=$(pwd)/dev/local/remote-kind/target-kubeconfig
```

This setup will deploy the diki-operator in the `kube-system` namespace on the source cluster.

### 2. Verify setup

Verify that `ComplianceScan` resources are successfully processed.

```bash
kubectl --kubeconfig $(pwd)/dev/local/remote-kind/target-kubeconfig apply -f ./example/80-diki-options-configmap.yaml
kubectl --kubeconfig $(pwd)/dev/local/remote-kind/target-kubeconfig apply -f ./example/80-reportoutput-configmap.yaml
kubectl --kubeconfig $(pwd)/dev/local/remote-kind/target-kubeconfig apply -f ./example/90-compliancescan.yaml
```

Check that reconciliation has started by looking at the logs of the operator:

```bash
kubectl --kubeconfig $(pwd)/dev/local/remote-kind/source-kubeconfig -n kube-system logs -l app.kubernetes.io/instance=diki-operator
```

## Cleanup

To tear down the local environment:

```bash
make remote-kind-down
```
