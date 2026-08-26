// SPDX-FileCopyrightText: Copyright Contributors to the Gardener project
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
