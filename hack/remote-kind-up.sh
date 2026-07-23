#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

repo_root="$(readlink -f "$(dirname "${0}")/..")"
kubeconfig_dir="$repo_root/dev/local/remote-kind"
mkdir -p "$kubeconfig_dir"

SOURCE_CLUSTER="diki-source"
TARGET_CLUSTER="diki-target"
RUNNER_NAMESPACE="kube-system"
RUNNER_SA="diki-run"

# setup_kind_network is similar to kind's network creation logic.
# Ensures stable CIDRs for the local setup.
setup_kind_network() {
  local existing_network_id
  existing_network_id="$(docker network list --filter=name=^kind$ --format='{{.ID}}')"

  if [ -n "$existing_network_id" ] ; then
    local network network_options network_ipam expected_network_ipam
    network="$(docker network inspect $existing_network_id | yq '.[]')"
    network_options="$(echo "$network" | yq '.EnableIPv6 + "," + .Options["com.docker.network.bridge.enable_ip_masquerade"]')"
    network_ipam="$(echo "$network" | yq '.IPAM.Config' -o=json -I=0)"
    expected_network_ipam='[{"Subnet":"172.18.0.0/16","Gateway":"172.18.0.1"},{"Subnet":"fd00:10::/64","Gateway":"fd00:10::1"}]'

    if [ "$network_options" = 'true,true' ] && [ "$network_ipam" = "$expected_network_ipam" ] ; then
      return 0
    else
      echo "kind network is not configured correctly, recreating..."
      docker network rm $existing_network_id
    fi
  fi

  docker network create kind --driver=bridge \
    --subnet 172.18.0.0/16 --gateway 172.18.0.1 \
    --ipv6 --subnet fd00:10::/64 --gateway fd00:10::1 \
    --opt com.docker.network.bridge.enable_ip_masquerade=true
}

setup_kind_network

echo "Creating source cluster ($SOURCE_CLUSTER)..."
kind create cluster --name "$SOURCE_CLUSTER"

echo "Creating target cluster ($TARGET_CLUSTER)..."
kind create cluster --name "$TARGET_CLUSTER"

# Save kubeconfigs
kind get kubeconfig --name "$SOURCE_CLUSTER" > "$kubeconfig_dir/source-kubeconfig"
kind get kubeconfig --name "$TARGET_CLUSTER" > "$kubeconfig_dir/target-kubeconfig"

export KUBECONFIG="$kubeconfig_dir/target-kubeconfig"

echo "Setting up ServiceAccount and RBAC on target cluster..."

# Create ServiceAccount for diki-runner (used by the scan Job)
kubectl create serviceaccount "$RUNNER_SA" --namespace "$RUNNER_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create ServiceAccount for the operator (needs CRD access)
kubectl create serviceaccount "diki-operator" --namespace "$RUNNER_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create a long-lived token Secret for the operator ServiceAccount
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: diki-operator-token
  namespace: ${RUNNER_NAMESPACE}
  annotations:
    kubernetes.io/service-account.name: diki-operator
type: kubernetes.io/service-account-token
EOF

# Create a long-lived token Secret for the diki-runner ServiceAccount
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${RUNNER_SA}-token
  namespace: ${RUNNER_NAMESPACE}
  annotations:
    kubernetes.io/service-account.name: ${RUNNER_SA}
type: kubernetes.io/service-account-token
EOF

# Wait for the operator token to be populated
echo "Waiting for operator token to be populated..."
OPERATOR_TOKEN=""
for i in $(seq 1 30); do
  OPERATOR_TOKEN=$(kubectl get secret "diki-operator-token" -n "$RUNNER_NAMESPACE" -o jsonpath='{.data.token}' 2>/dev/null || true)
  if [ -n "$OPERATOR_TOKEN" ]; then
    break
  fi
  sleep 1
done

if [ -z "$OPERATOR_TOKEN" ]; then
  echo "ERROR: Operator token was not populated in time"
  exit 1
fi
OPERATOR_TOKEN=$(echo "$OPERATOR_TOKEN" | base64 -d)

# Wait for the runner token to be populated
echo "Waiting for runner token to be populated..."
RUNNER_TOKEN=""
for i in $(seq 1 30); do
  RUNNER_TOKEN=$(kubectl get secret "${RUNNER_SA}-token" -n "$RUNNER_NAMESPACE" -o jsonpath='{.data.token}' 2>/dev/null || true)
  if [ -n "$RUNNER_TOKEN" ]; then
    break
  fi
  sleep 1
done

if [ -z "$RUNNER_TOKEN" ]; then
  echo "ERROR: Runner token was not populated in time"
  exit 1
fi
RUNNER_TOKEN=$(echo "$RUNNER_TOKEN" | base64 -d)

# Get the target cluster's API server URL (docker-internal)
TARGET_CONTAINER="${TARGET_CLUSTER}-control-plane"
TARGET_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$TARGET_CONTAINER")
TARGET_API_SERVER="https://${TARGET_IP}:6443"

# Get the target cluster's CA certificate
CA_DATA=$(kubectl config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')

echo "Setting up Secrets on source cluster..."

export KUBECONFIG="$kubeconfig_dir/source-kubeconfig"

# Create the kubeconfig that the diki-run Job will use (with inline token)
JOB_KUBECONFIG_CONTENT=$(cat <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: ${CA_DATA}
    server: ${TARGET_API_SERVER}
  name: target
contexts:
- context:
    cluster: target
    user: diki-runner
  name: target
current-context: target
users:
- name: diki-runner
  user:
    token: ${RUNNER_TOKEN}
EOF
)

# Create the kubeconfig that the operator will use to connect to the target cluster (with inline token)
OPERATOR_KUBECONFIG_CONTENT=$(cat <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: ${CA_DATA}
    server: ${TARGET_API_SERVER}
  name: target
contexts:
- context:
    cluster: target
    user: diki-operator
  name: target
current-context: target
users:
- name: diki-operator
  user:
    token: ${OPERATOR_TOKEN}
EOF
)

# Create kubeconfig Secret for the diki-run Job on source cluster
kubectl create secret generic target-kubeconfig \
  --namespace "$RUNNER_NAMESPACE" \
  --from-literal=kubeconfig="$JOB_KUBECONFIG_CONTENT" \
  --dry-run=client -o yaml | kubectl apply -f -

# Create kubeconfig Secret for the operator to connect to target cluster
kubectl create secret generic operator-target-kubeconfig \
  --namespace "$RUNNER_NAMESPACE" \
  --from-literal=kubeconfig="$OPERATOR_KUBECONFIG_CONTENT" \
  --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "Remote kind setup complete!"
echo "  Source cluster kubeconfig: $kubeconfig_dir/source-kubeconfig"
echo "  Target cluster kubeconfig: $kubeconfig_dir/target-kubeconfig"
echo ""
echo "Next step: make remote-operator-up"
