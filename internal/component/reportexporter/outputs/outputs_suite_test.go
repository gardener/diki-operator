// SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package outputs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOutputs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Outputs Test Suite")
}
