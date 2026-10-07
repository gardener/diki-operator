// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package configmap_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfigMap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ConfigMap Output Test Suite")
}
