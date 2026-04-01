// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reportexporter

type exportError struct {
	Error string `json:"error"`
}

func newExportError(err error) exportError {
	return exportError{
		Error: err.Error(),
	}
}
