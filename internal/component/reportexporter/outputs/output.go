// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package outputs

import (
	"context"

	dikireport "github.com/gardener/diki/pkg/report"

	"github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

// Output is the interface for exporting Diki reports to different output types.
type Output interface {
	Type() v1alpha1.OutputType
	Export(ctx context.Context, report dikireport.Report) (exportDetails any, err error)
}
