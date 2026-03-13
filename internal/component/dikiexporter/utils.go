// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package dikiexporter

import (
	"slices"

	dikireport "github.com/gardener/diki/pkg/report"
	"github.com/gardener/diki/pkg/rule"

	"github.com/gardener/diki-operator/pkg/apis/diki/v1alpha1"
)

// createRulesetSummaries adds all rulesets of a report to the ComplianceScan's summary.
func createRulesetSummaries(report dikireport.Report) []v1alpha1.RulesetSummary {
	var rulesetSummaries []v1alpha1.RulesetSummary

	for _, provider := range report.Providers {
		for _, ruleset := range provider.Rulesets {
			rulesetSummary := v1alpha1.RulesetSummary{
				ID:      ruleset.ID,
				Version: ruleset.Version,
				Results: v1alpha1.RulesResults{},
			}

			rulesetSummary.Results.Summary = rulesetSummaryCount(&ruleset)
			rulesetSummary.Results.Rules = &v1alpha1.RulesFindings{
				Failed:  getRulesetFindingsByStatus(&ruleset, rule.Failed),
				Errored: getRulesetFindingsByStatus(&ruleset, rule.Errored),
				Warning: getRulesetFindingsByStatus(&ruleset, rule.Warning),
			}

			rulesetSummaries = append(rulesetSummaries, rulesetSummary)
		}
	}

	return rulesetSummaries
}

func getRulesetFindingsByStatus(ruleset *dikireport.Ruleset, status rule.Status) []v1alpha1.Rule {
	var ruleFindings []v1alpha1.Rule

	for _, rule := range ruleset.Rules {
		if slices.ContainsFunc(rule.Checks, func(check dikireport.Check) bool {
			return check.Status == status
		}) {
			ruleFindings = append(ruleFindings, v1alpha1.Rule{ID: rule.ID, Name: rule.Name})
		}
	}

	return ruleFindings
}

func rulesetSummaryCount(ruleset *dikireport.Ruleset) v1alpha1.RulesSummary {
	var summary v1alpha1.RulesSummary

	statuses := rule.Statuses()
	for _, status := range statuses {
		num := numOfRulesWithStatus(ruleset, status)
		switch status {
		case rule.Passed:
			summary.Passed = num
		case rule.Skipped:
			summary.Skipped = num
		case rule.Accepted:
			summary.Accepted = num
		case rule.Warning:
			summary.Warning = num
		case rule.Failed:
			summary.Failed = num
		case rule.Errored:
			summary.Errored = num
		}
	}

	return summary
}

func numOfRulesWithStatus(ruleset *dikireport.Ruleset, status rule.Status) int32 {
	var num int32
	for _, rule := range ruleset.Rules {
		for _, check := range rule.Checks {
			if check.Status == status {
				num++
				break
			}
		}
	}
	return num
}
