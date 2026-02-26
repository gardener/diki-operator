// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package reconciler

import (
	"bytes"
	"context"
	"fmt"

	dikiconfig "github.com/gardener/diki/pkg/config"
	"github.com/gardener/diki/pkg/provider/managedk8s"
	"github.com/gardener/diki/pkg/provider/managedk8s/ruleset/disak8sstig"
	"github.com/gardener/diki/pkg/provider/managedk8s/ruleset/securityhardenedk8s"
	"go.yaml.in/yaml/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

func (r *Reconciler) deployDikiConfigMap(ctx context.Context, complianceRun *v1alpha1.ComplianceRun) (*corev1.ConfigMap, error) {
	managedk8sProvider := dikiconfig.ProviderConfig{
		ID:   managedk8s.ProviderID,
		Name: managedk8s.ProviderName,
	}

	for _, ruleset := range complianceRun.Spec.Rulesets {
		switch ruleset.ID {
		case disak8sstig.RulesetID:
			ruleOptions, err := r.getRuleOptions(ctx, ruleset.Options, disak8sstig.RulesetID)
			if err != nil {
				return nil, fmt.Errorf("failed to get rule options: %w", err)
			}
			rulesetOptions, err := r.getRulesetOptions(ctx, ruleset.Options, disak8sstig.RulesetID)
			if err != nil {
				return nil, fmt.Errorf("failed to get ruleset options: %w", err)
			}

			dikiRuleset := dikiconfig.RulesetConfig{
				ID:          disak8sstig.RulesetID,
				Name:        disak8sstig.RulesetName,
				Version:     ruleset.Version,
				Args:        rulesetOptions,
				RuleOptions: ruleOptions,
			}

			managedk8sProvider.Rulesets = append(managedk8sProvider.Rulesets, dikiRuleset)
		case securityhardenedk8s.RulesetID:
			ruleOptions, err := r.getRuleOptions(ctx, ruleset.Options, securityhardenedk8s.RulesetID)
			if err != nil {
				return nil, fmt.Errorf("failed to get rule options: %w", err)
			}
			rulesetOptions, err := r.getRulesetOptions(ctx, ruleset.Options, securityhardenedk8s.RulesetID)
			if err != nil {
				return nil, fmt.Errorf("failed to get ruleset options: %w", err)
			}

			dikiRuleset := dikiconfig.RulesetConfig{
				ID:          securityhardenedk8s.RulesetID,
				Name:        securityhardenedk8s.RulesetName,
				Version:     ruleset.Version,
				Args:        rulesetOptions,
				RuleOptions: ruleOptions,
			}

			managedk8sProvider.Rulesets = append(managedk8sProvider.Rulesets, dikiRuleset)
		default:
		}
	}

	dikiConfig := dikiconfig.DikiConfig{
		Providers: []dikiconfig.ProviderConfig{managedk8sProvider},
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(dikiConfig); err != nil {
		return nil, fmt.Errorf("failed to marshal diki config: %w", err)
	}
	dikiConfigYAML := buf.Bytes()

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: ConfigMapGenerateNamePrefix,
			Namespace:    r.Config.DikiRunner.Namespace,
			//OwnerReferences: r.getOwnerReference(job),
			Labels: r.getLabels(complianceRun),
		},
		Data: map[string]string{
			DikiConfigKey: string(dikiConfigYAML),
		},
	}

	if err := r.Client.Create(ctx, configMap); err != nil {
		return nil, fmt.Errorf("failed to create diki config configMap: %w", err)
	}

	return configMap, nil
}

func (r *Reconciler) getRuleOptions(ctx context.Context, options *v1alpha1.RulesetOptions, rulesetID string) ([]dikiconfig.RuleOptionsConfig, error) {
	if options == nil || options.Rules == nil || options.Rules.ConfigMapRef == nil {
		return nil, nil
	}

	ruleOptionsYAML, err := r.getConfigMapKeyValue(ctx, *options.Rules.ConfigMapRef, rulesetID+RuleOptionsSuffix)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule options from configMap: %w", err)
	}

	var ruleOptions []dikiconfig.RuleOptionsConfig
	if err := yaml.Unmarshal([]byte(ruleOptionsYAML), &ruleOptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rule options from configMap %s/%s: %w", options.Rules.ConfigMapRef.Namespace, options.Rules.ConfigMapRef.Name, err)
	}

	return ruleOptions, nil
}

func (r *Reconciler) getRulesetOptions(ctx context.Context, options *v1alpha1.RulesetOptions, rulesetID string) (any, error) {
	if options == nil || options.Ruleset == nil || options.Ruleset.ConfigMapRef == nil {
		return nil, nil
	}

	rulesetOptionsYAML, err := r.getConfigMapKeyValue(ctx, *options.Ruleset.ConfigMapRef, rulesetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ruleset options from configMap: %w", err)
	}

	var rulesetOptions any
	if err := yaml.Unmarshal([]byte(rulesetOptionsYAML), &rulesetOptions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ruleset options from configMap %s/%s: %w", options.Ruleset.ConfigMapRef.Namespace, options.Ruleset.ConfigMapRef.Name, err)
	}

	return rulesetOptions, nil
}

func (r *Reconciler) getConfigMapKeyValue(ctx context.Context, configMapRef v1alpha1.OptionsConfigMapRef, defaultKey string) (string, error) {
	key := defaultKey
	if configMapRef.Key != nil {
		key = *configMapRef.Key
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapRef.Name,
			Namespace: configMapRef.Namespace,
		},
	}

	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(configMap), configMap); err != nil {
		return "", fmt.Errorf("failed to get configMap %s: %w", client.ObjectKeyFromObject(configMap), err)
	}

	ruleOptions, exists := configMap.Data[key]
	if !exists {
		return "", fmt.Errorf("key '%s' does not exist in configMap %s", key, client.ObjectKeyFromObject(configMap))
	}

	return ruleOptions, nil
}
