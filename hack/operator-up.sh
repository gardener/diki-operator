#!/usr/bin/env bash

# SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

DIKI_OPERATOR_NAME="diki-operator"

repo_root="$(readlink -f "$(dirname "${0}")/..")"

charts_dir="$repo_root/charts/diki/$DIKI_OPERATOR_NAME"
temp_dir="$repo_root/dev/diki-operator"
mkdir -p "$temp_dir"
values_file="$temp_dir/values.yaml"
cp "$charts_dir/values.yaml" "$values_file"

# Generate certificates
dev_diki_operator_dir="$repo_root/dev/diki-operator"
cert_dir="$dev_diki_operator_dir/certs"

"$repo_root"/hack/generate-certs.sh \
  "$repo_root/dev/diki-operator" \
  "diki-operator.kube-system.svc.cluster.local" \
  "DNS:localhost,DNS:diki-operator,DNS:diki-operator.kube-system,DNS:diki-operator.kube-system.svc,DNS:diki-operator.kube-system.svc.cluster.local,IP:127.0.0.1"

# Finish generating certificates
yq -i ' .config.server.webhooks.tls.caBundle = load_str("'"$cert_dir/ca.crt"'") | (.config.server.webhooks.tls.caBundle style="literal") ' "$values_file" 
yq -i ' .config.server.webhooks.tls.crt = load_str("'"$cert_dir/tls.crt"'") | (.config.server.webhooks.tls.crt style="literal") ' "$values_file" 
yq -i ' .config.server.webhooks.tls.key = load_str("'"$cert_dir/tls.key"'") | (.config.server.webhooks.tls.key style="literal") ' "$values_file"

# Deploy the operator charts
kubectl apply -f ./charts/diki/crds/
skaffold run

echo "diki-operator installed successfully in the runtime cluster."

echo "Done."
