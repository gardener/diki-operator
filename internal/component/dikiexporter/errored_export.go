// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package dikiexporter

type erroredExport struct {
	ErrorMessage string `json:"errorMessage"`
}

func newErroredExport(err error) erroredExport {
	return erroredExport{
		ErrorMessage: err.Error(),
	}
}
