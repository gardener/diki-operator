// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reportexporter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	dikireport "github.com/gardener/diki/pkg/report"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dikioutputs "github.com/gardener/diki-operator/internal/component/reportexporter/outputs"
	dikiv1alpha1 "github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
	"github.com/gardener/diki-operator/pkg/apis/reportexporter/v1alpha1"
)

// ReportExporter is responsible for exporting compliance scan data.
type ReportExporter struct {
	Client client.Client
	Config v1alpha1.ReportExporterConfiguration
}

// NewReportExporter creates a new instance of ReportExporter.
func NewReportExporter(
	client client.Client,
	config v1alpha1.ReportExporterConfiguration,
) *ReportExporter {
	return &ReportExporter{
		Client: client,
		Config: config,
	}
}

// Export exports the compliance scan data.
func (d *ReportExporter) Export(ctx context.Context) error {
	complianceScan := &dikiv1alpha1.ComplianceScan{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.Config.ComplianceScanName,
		},
	}

	if err := d.Client.Get(ctx, client.ObjectKey{Name: d.Config.ComplianceScanName}, complianceScan); err != nil {
		return fmt.Errorf("error retrieving complianceScan: %w", err)
	}

	// If the ComplianceScan is already completed, we should not export the report again
	// This is a safety check to prevent overwriting the report in case the exporter is restarted after the ComplianceScan has completed
	if complianceScan.Status.Phase == dikiv1alpha1.ComplianceScanCompleted {
		return fmt.Errorf("complianceScan is already completed, cannot export report")
	}

	report, err := d.readDikiReport()
	if err != nil {
		return fmt.Errorf("error reading diki report: %w", err)
	}

	outputs, err := d.createOutputs()
	if err != nil {
		return fmt.Errorf("error creating outputs: %w", err)
	}

	var wg sync.WaitGroup
	outputStatusChan := make(chan dikiv1alpha1.OutputStatus, len(outputs))

	for name, output := range outputs {
		wg.Add(1)
		go func(exp dikioutputs.Output) {
			defer wg.Done()
			var (
				phase   dikiv1alpha1.OutputStatusPhase
				details runtime.RawExtension
			)

			if d, err := exp.Export(ctx, *report); err != nil {
				phase = dikiv1alpha1.OutputStatusFailed
				details = toRawExtension(newExportError(err))
			} else {
				phase = dikiv1alpha1.OutputStatusCompleted
				details = toRawExtension(d)
			}

			outputStatusChan <- dikiv1alpha1.OutputStatus{
				OutputName: name,
				Phase:      phase,
				Details:    details,
			}
		}(output)
	}

	wg.Wait()
	close(outputStatusChan)

	var outputStatuses []dikiv1alpha1.OutputStatus
	for status := range outputStatusChan {
		outputStatuses = append(outputStatuses, status)
	}

	patch := client.MergeFrom(complianceScan.DeepCopy())
	complianceScan.Status.Rulesets = createRulesetSummaries(report)
	complianceScan.Status.Outputs = outputStatuses

	if err := d.Client.Status().Patch(ctx, complianceScan, patch); err != nil {
		return fmt.Errorf("failed to patch ComplianceScan status: %w", err)
	}

	return nil
}

func (d *ReportExporter) createOutputs() (map[string]dikioutputs.Output, error) {
	outputs := make(map[string]dikioutputs.Output)

	for _, output := range d.Config.Outputs {
		switch output.Type {
		case v1alpha1.ExporterTypeConfigMap:
			var configMapOutput dikiv1alpha1.OutputConfigMap
			if err := json.Unmarshal(output.Config.Raw, &configMapOutput); err != nil {
				return nil, fmt.Errorf("failed to unmarshal ConfigMapOutput: %w", err)
			}

			outputs[output.Name] = dikioutputs.NewConfigMapExporter(d.Client, configMapOutput)
		default:
			return nil, fmt.Errorf("unsupported output type: %s", output.Type)
		}
	}

	return outputs, nil
}

func (d *ReportExporter) readDikiReport() (*dikireport.Report, error) {
	reportData, err := os.ReadFile(d.Config.ReportPath)
	if err != nil {
		return nil, fmt.Errorf("error reading report file: %w", err)
	}

	var report dikireport.Report
	if err := json.Unmarshal(reportData, &report); err != nil {
		return nil, fmt.Errorf("error unmarshaling report: %w", err)
	}

	return &report, nil
}

func toRawExtension(v any) runtime.RawExtension {
	if v == nil {
		return runtime.RawExtension{}
	}

	data, err := json.Marshal(v)
	if err != nil {
		errData, _ := json.Marshal(map[string]string{"error": fmt.Sprintf("failed to marshal details: %v", err)})
		return runtime.RawExtension{Raw: errData}
	}

	return runtime.RawExtension{Raw: data}
}
