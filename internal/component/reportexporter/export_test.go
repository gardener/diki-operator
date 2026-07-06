// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reportexporter

import "time"

func SetReportFilePollInterval(d time.Duration) func() {
	old := reportFilePollInterval
	reportFilePollInterval = d
	return func() { reportFilePollInterval = old }
}
