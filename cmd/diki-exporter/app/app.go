package app

import (
	"context"
	goflag "flag"
	"fmt"

	"github.com/gardener/gardener/pkg/logger"
	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/pkg/version"
	"k8s.io/component-base/version/verflag"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki-operator/internal/component/dikiexporter"
	dikiinstall "github.com/gardener/diki-operator/pkg/apis/diki/install"
	"github.com/gardener/diki-operator/pkg/apis/dikiexporter/v1alpha1"
)

// AppName is the name of the application.
const AppName = "diki-exporter"

// NewCommand is the root command for the Diki operator.
func NewCommand() *cobra.Command {
	opt := newOptions()

	cmd := &cobra.Command{
		Use:   AppName,
		Short: "Launch the " + AppName,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := opt.Complete(); err != nil {
				return err
			}

			log, err := logger.NewZapLogger(logger.InfoLevel, logger.FormatJSON)
			if err != nil {
				return fmt.Errorf("error instantiating zap logger: %w", err)
			}

			log.Info("Starting application", "app", AppName, "version", version.Get())
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				log.Info("Flag", "name", flag.Name, "value", flag.Value, "default", flag.DefValue)
			})

			return run(cmd.Context(), log, opt.config)
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			verflag.PrintAndExitIfRequested()
			return nil
		},
	}

	flags := cmd.Flags()
	opt.addFlags(flags)
	flags.AddGoFlagSet(goflag.CommandLine)

	return cmd
}

func run(ctx context.Context, log logr.Logger, cfg *v1alpha1.DikiExporterConfiguration) error {
	conf, err := ctrl.GetConfig()
	if err != nil {
		return err
	}

	scheme := runtime.NewScheme()
	if err := dikiinstall.AddToScheme(scheme); err != nil {
		return fmt.Errorf("could not update scheme: %w", err)
	}
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return fmt.Errorf("could not update scheme: %w", err)
	}

	c, err := client.New(conf, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return err
	}

	log.Info("Setting up diki-exporter")
	dikiExporter := dikiexporter.NewDikiExporter(c, *cfg)

	log.Info("Starting diki-exporter")
	return dikiExporter.Export(ctx)
}
