// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"github.com/gardener/diki-operator/pkg/apis/dikiexporter/v1alpha1"
)

var configDecoder runtime.Decoder

func init() {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	configDecoder = serializer.NewCodecFactory(scheme).UniversalDecoder(v1alpha1.SchemeGroupVersion)
}

type options struct {
	configFile string
	config     *v1alpha1.DikiExporterConfiguration
}

// newOptions return options with default values.
func newOptions() *options {
	return &options{}
}

// addFlags binds the command options to a given flagset.
func (o *options) addFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.configFile, "config", o.configFile, "Path to configuration file.")
}

// Complete adapts from the command line args to the data required.
func (o *options) Complete() error {
	if len(o.configFile) == 0 {
		return fmt.Errorf("missing config file")
	}

	data, err := os.ReadFile(o.configFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	o.config = &v1alpha1.DikiExporterConfiguration{}
	if err = runtime.DecodeInto(configDecoder, data, o.config); err != nil {
		return fmt.Errorf("error decoding config: %w", err)
	}

	return nil
}
