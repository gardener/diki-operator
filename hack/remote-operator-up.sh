#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -o errexit
set -o nounset
set -o pipefail

repo_root="$(readlink -f "$(dirname "${0}")/..")"
kubeconfig_dir="$repo_root/dev/local/remote-kind"
source_kubeconfig="$kubeconfig_dir/source-kubeconfig"
target_kubeconfig="$kubeconfig_dir/target-kubeconfig"

if [ ! -f "$source_kubeconfig" ] || [ ! -f "$target_kubeconfig" ]; then
  echo "ERROR: Kubeconfigs not found. Run 'make remote-kind-up' first."
  exit 1
fi

charts_dir="$repo_root/charts/diki/diki-operator"
temp_dir="$repo_root/dev/remote"
mkdir -p "$temp_dir"
values_file="$temp_dir/operator-values.yaml"
cp "$charts_dir/values.yaml" "$values_file"

# Generate certificates
cert_dir="$temp_dir/certs"

"$repo_root"/hack/generate-certs.sh \
  "$temp_dir" \
  "diki-operator.kube-system.svc.cluster.local" \
  "DNS:localhost,DNS:diki-operator,DNS:diki-operator.kube-system,DNS:diki-operator.kube-system.svc,DNS:diki-operator.kube-system.svc.cluster.local,IP:127.0.0.1"

# Inject TLS certs into values
yq -i ' .config.server.webhooks.tls.caBundle = load_str("'"$cert_dir/ca.crt"'") | (.config.server.webhooks.tls.caBundle style="literal") ' "$values_file"
yq -i ' .config.server.webhooks.tls.crt = load_str("'"$cert_dir/tls.crt"'") | (.config.server.webhooks.tls.crt style="literal") ' "$values_file"
yq -i ' .config.server.webhooks.tls.key = load_str("'"$cert_dir/tls.key"'") | (.config.server.webhooks.tls.key style="literal") ' "$values_file"

# Inject kubeconfig configuration for remote mode
yq -i '.config.controllers.complianceScan.dikiRunner.targetKubeconfig.secretRef.name = "target-kubeconfig"' "$values_file"

# Configure the operator to use the target cluster kubeconfig for its manager
yq -i '.targetKubeconfig.secretName = "operator-target-kubeconfig"' "$values_file"

# Deploy CRDs to target cluster
echo "Applying CRDs to target cluster..."
kubectl --kubeconfig "$target_kubeconfig" apply -f "$repo_root/charts/diki/crds/"

# Deploy operator and runner RBAC to target cluster using helm template
echo "Applying operator and runner RBAC to target cluster..."
helm template diki-operator "$charts_dir" --namespace kube-system \
  --show-only 'templates/rbac/operator/*.yaml' \
  --show-only 'templates/rbac/run/scanner/*.yaml' \
  --show-only 'templates/rbac/run/exporter/*.yaml' \
  | kubectl --kubeconfig "$target_kubeconfig" apply -f -

# Deploy operator to source cluster via skaffold
echo "Deploying operator to source cluster..."
export KUBECONFIG="$source_kubeconfig"
skaffold run -f "$repo_root/skaffold-remote.yaml"

echo ""
echo "Remote operator deployed successfully!"
echo "  Operator running on: diki-source (source cluster)"
echo "  Watching CRDs on: diki-target (target cluster)"
echo ""
echo "To create a ComplianceScan:"
echo "  kubectl --kubeconfig $target_kubeconfig apply -f example/90-compliancescan.yaml"
echo ""
echo "Done."
