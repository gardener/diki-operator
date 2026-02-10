// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

//go:generate gen-crd-api-reference-docs -api-dir . -config ../../../../hack/api-reference/config.json -template-dir "$GARDENER_HACK_DIR/api-reference/template" -out-file ../../../../docs/api-reference/config.md

// +k8s:deepcopy-gen=package
// +k8s:defaulter-gen=TypeMeta
// +k8s:conversion-gen=github.com/gardener/diki-operator/pkg/apis/diki
// +groupName=diki.gardener.cloud

// Package v1alpha1 is a version of the API.
package v1alpha1
