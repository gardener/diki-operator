// SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package reportexporter

import "time"

func SetReportFilePollInterval(d time.Duration) func() {
	old := reportFilePollInterval
	reportFilePollInterval = d
	return func() { reportFilePollInterval = old }
}
