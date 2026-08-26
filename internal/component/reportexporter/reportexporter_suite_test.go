// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package reportexporter_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestReportExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ReportExporter Test Suite")
}
