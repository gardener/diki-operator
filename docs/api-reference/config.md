<p>Packages:</p>
<ul>
<li>
<a href="#diki.gardener.cloud%2fv1alpha1">diki.gardener.cloud/v1alpha1</a>
</li>
</ul>

<h2 id="diki.gardener.cloud/v1alpha1">diki.gardener.cloud/v1alpha1</h2>
<p>

</p>
Resource Types:
<ul>
<li>
<a href="#compliancescan">ComplianceScan</a>
</li>
<li>
<a href="#reportoutput">ReportOutput</a>
</li>
<li>
<a href="#scheduledcompliancescan">ScheduledComplianceScan</a>
</li>
</ul>

<h3 id="compliancescan">ComplianceScan
</h3>


<p>
ComplianceScan describes a compliance scan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta">ObjectMeta</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#compliancescanspec">ComplianceScanSpec</a>
</em>
</td>
<td>
<p>Spec contains the specification of this compliance scan.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#compliancescanstatus">ComplianceScanStatus</a>
</em>
</td>
<td>
<p>Status contains the status of this compliance scan.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="compliancescanphase">ComplianceScanPhase
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#compliancescanstatus">ComplianceScanStatus</a>)
</p>

<p>
ComplianceScanPhase is an alias for string representing the phase of a ComplianceScan.
</p>


<h3 id="compliancescanspec">ComplianceScanSpec
</h3>


<p>
(<em>Appears on:</em><a href="#compliancescan">ComplianceScan</a>, <a href="#scheduledcompliancescantemplate">ScheduledComplianceScanTemplate</a>)
</p>

<p>
ComplianceScanSpec is the specification of a ComplianceScan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>rulesets</code></br>
<em>
<a href="#rulesetconfig">RulesetConfig</a> array
</em>
</td>
<td>
<p>Rulesets describe the rulesets to be applied during the compliance scan.</p>
</td>
</tr>
<tr>
<td>
<code>outputs</code></br>
<em>
<a href="#reportoutputref">ReportOutputRef</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Outputs describe the outputs of the compliance scan.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="compliancescanstatus">ComplianceScanStatus
</h3>


<p>
(<em>Appears on:</em><a href="#compliancescan">ComplianceScan</a>)
</p>

<p>
ComplianceScanStatus contains the status of a ComplianceScan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>conditions</code></br>
<em>
<a href="#condition">Condition</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Conditions contains the conditions of the ComplianceScan.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code></br>
<em>
<a href="#compliancescanphase">ComplianceScanPhase</a>
</em>
</td>
<td>
<p>Phase represents the current phase of the ComplianceScan.</p>
</td>
</tr>
<tr>
<td>
<code>rulesets</code></br>
<em>
<a href="#rulesetsummary">RulesetSummary</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Rulesets contains the ruleset summaries of the ComplianceScan.</p>
</td>
</tr>
<tr>
<td>
<code>outputs</code></br>
<em>
<a href="#outputstatus">OutputStatus</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Outputs contain the output statuses of the ComplianceScan.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="condition">Condition
</h3>


<p>
(<em>Appears on:</em><a href="#compliancescanstatus">ComplianceScanStatus</a>)
</p>

<p>
Condition describes a condition of a ComplianceScan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>type</code></br>
<em>
<a href="#conditiontype">ConditionType</a>
</em>
</td>
<td>
<p>Type is the type of the condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#conditionstatus">ConditionStatus</a>
</em>
</td>
<td>
<p>Status is the status of the condition.</p>
</td>
</tr>
<tr>
<td>
<code>lastUpdateTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#time-v1-meta">Time</a>
</em>
</td>
<td>
<p>LastUpdateTime is the last time the condition was updated.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#time-v1-meta">Time</a>
</em>
</td>
<td>
<p>LastTransitionTime is the last time the condition transitioned from one status to another.</p>
</td>
</tr>
<tr>
<td>
<code>reason</code></br>
<em>
string
</em>
</td>
<td>
<p>Reason is a brief reason for the condition's last transition.</p>
</td>
</tr>
<tr>
<td>
<code>message</code></br>
<em>
string
</em>
</td>
<td>
<p>Message is a human-readable message indicating details about the last transition.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="conditionstatus">ConditionStatus
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#condition">Condition</a>)
</p>

<p>
ConditionStatus is an alias for string representing the status of a condition.
</p>


<h3 id="conditiontype">ConditionType
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#condition">Condition</a>)
</p>

<p>
ConditionType is an alias for string representing the type of a condition.
</p>


<h3 id="options">Options
</h3>


<p>
(<em>Appears on:</em><a href="#rulesetoptions">RulesetOptions</a>)
</p>

<p>
Options contains references to options.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>configMapRef</code></br>
<em>
<a href="#optionsconfigmapref">OptionsConfigMapRef</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ConfigMapRef is a reference to a ConfigMap containing options.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="optionsconfigmapref">OptionsConfigMapRef
</h3>


<p>
(<em>Appears on:</em><a href="#options">Options</a>)
</p>

<p>
OptionsConfigMapRef references a ConfigMap containing rule options for the ruleset.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the ConfigMap.</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<p>Namespace is the namespace of the ConfigMap.</p>
</td>
</tr>
<tr>
<td>
<code>key</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Key is the key within the ConfigMap, where the options are stored.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="output">Output
</h3>


<p>
(<em>Appears on:</em><a href="#reportoutputspec">ReportOutputSpec</a>)
</p>

<p>
Output describes a specific output of a compliance scan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>configMap</code></br>
<em>
<a href="#outputconfigmap">OutputConfigMap</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>ConfigMap contains the configuration for exporting the report to a ConfigMap.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="outputconfigmap">OutputConfigMap
</h3>


<p>
(<em>Appears on:</em><a href="#output">Output</a>)
</p>

<p>
OutputConfigMap contains the configuration for exporting the report to a ConfigMap.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>namespace</code></br>
<em>
string
</em>
</td>
<td>
<p>Namespace is the namespace where the ConfigMap will be created.<br />Defaults to `kube-system`.</p>
</td>
</tr>
<tr>
<td>
<code>namePrefix</code></br>
<em>
string
</em>
</td>
<td>
<p>NamePrefix is the prefix for the generated ConfigMap name.<br />Defaults to "compliance-scan-report-".</p>
</td>
</tr>

</tbody>
</table>


<h3 id="outputstatus">OutputStatus
</h3>


<p>
(<em>Appears on:</em><a href="#compliancescanstatus">ComplianceScanStatus</a>)
</p>

<p>
OutputStatus contains the status of a specific output of a compliance scan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>outputName</code></br>
<em>
string
</em>
</td>
<td>
<p>OutputName is the name of the report output.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code></br>
<em>
<a href="#outputstatusphase">OutputStatusPhase</a>
</em>
</td>
<td>
<p>Phase represents the final phase of the output after the exporter has processed it.</p>
</td>
</tr>
<tr>
<td>
<code>details</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#rawextension-runtime-pkg">RawExtension</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Details contains details about the output.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="outputstatusphase">OutputStatusPhase
</h3>
<p><em>Underlying type: string</em></p>


<p>
(<em>Appears on:</em><a href="#outputstatus">OutputStatus</a>)
</p>

<p>
OutputStatusPhase is an alias for string representing the phase of an output after processing by the exporter.
</p>


<h3 id="reportoutput">ReportOutput
</h3>


<p>
ReportOutput describes a report output.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta">ObjectMeta</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#reportoutputspec">ReportOutputSpec</a>
</em>
</td>
<td>
<p>Spec contains the specification of this report output.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="reportoutputref">ReportOutputRef
</h3>


<p>
(<em>Appears on:</em><a href="#compliancescanspec">ComplianceScanSpec</a>)
</p>

<p>
ReportOutputRef describes a reference to a report output.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the report output.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="reportoutputspec">ReportOutputSpec
</h3>


<p>
(<em>Appears on:</em><a href="#reportoutput">ReportOutput</a>)
</p>

<p>
ReportOutputSpec is the specification of a ReportOutput.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>output</code></br>
<em>
<a href="#output">Output</a>
</em>
</td>
<td>
<p>Output describes a specific output of a compliance scan.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="rule">Rule
</h3>


<p>
(<em>Appears on:</em><a href="#rulesfindings">RulesFindings</a>)
</p>

<p>
Rule contains information about the ID and the name of the rule that contains the findings.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>id</code></br>
<em>
string
</em>
</td>
<td>
<p>ID is the unique identifier of the rule which contains the finding.</p>
</td>
</tr>
<tr>
<td>
<code>name</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the rule which contains the finding.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="rulesfindings">RulesFindings
</h3>


<p>
(<em>Appears on:</em><a href="#rulesresults">RulesResults</a>)
</p>

<p>
RulesFindings contains information about the specific rules that have errored/warned/failed.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>failed</code></br>
<em>
<a href="#rule">Rule</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Failed contains information about the rules that have a Failed status.</p>
</td>
</tr>
<tr>
<td>
<code>errored</code></br>
<em>
<a href="#rule">Rule</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Errored contains information about the rules that have an Errored status.</p>
</td>
</tr>
<tr>
<td>
<code>warning</code></br>
<em>
<a href="#rule">Rule</a> array
</em>
</td>
<td>
<em>(Optional)</em>
<p>Warning contains information about the rules that have a Warning status.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="rulesresults">RulesResults
</h3>


<p>
(<em>Appears on:</em><a href="#rulesetsummary">RulesetSummary</a>)
</p>

<p>
RulesResults contains the results of the rules in a ruleset.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>summary</code></br>
<em>
<a href="#rulessummary">RulesSummary</a>
</em>
</td>
<td>
<p>Summary contains information about the amount of rules per each status.</p>
</td>
</tr>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#rulesfindings">RulesFindings</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Rules contains information about the specific rules that have errored/warned/failed.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="rulessummary">RulesSummary
</h3>


<p>
(<em>Appears on:</em><a href="#rulesresults">RulesResults</a>)
</p>

<p>
RulesSummary contains information about the amount of rules per each status.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>passed</code></br>
<em>
integer
</em>
</td>
<td>
<p>Passed counts the amount of rules in a specific ruleset that have passed.</p>
</td>
</tr>
<tr>
<td>
<code>skipped</code></br>
<em>
integer
</em>
</td>
<td>
<p>Skipped counts the amount of rules in a specific ruleset that have been skipped.</p>
</td>
</tr>
<tr>
<td>
<code>accepted</code></br>
<em>
integer
</em>
</td>
<td>
<p>Accepted counts the amount of rules in a specific ruleset that have been accepted.</p>
</td>
</tr>
<tr>
<td>
<code>warning</code></br>
<em>
integer
</em>
</td>
<td>
<p>Warning counts the amount of rules in a specific ruleset that have returned a warning.</p>
</td>
</tr>
<tr>
<td>
<code>failed</code></br>
<em>
integer
</em>
</td>
<td>
<p>Failed counts the amount of rules in a specific ruleset that have failed.</p>
</td>
</tr>
<tr>
<td>
<code>errored</code></br>
<em>
integer
</em>
</td>
<td>
<p>Errored counts the amount of rules in a specific ruleset that have errored.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="rulesetconfig">RulesetConfig
</h3>


<p>
(<em>Appears on:</em><a href="#compliancescanspec">ComplianceScanSpec</a>)
</p>

<p>
RulesetConfig describes the configuration of a ruleset.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>id</code></br>
<em>
string
</em>
</td>
<td>
<p>ID is the identifier of the ruleset.</p>
</td>
</tr>
<tr>
<td>
<code>version</code></br>
<em>
string
</em>
</td>
<td>
<p>Version is the version of the ruleset.</p>
</td>
</tr>
<tr>
<td>
<code>options</code></br>
<em>
<a href="#rulesetoptions">RulesetOptions</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Options are options for a ruleset.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="rulesetoptions">RulesetOptions
</h3>


<p>
(<em>Appears on:</em><a href="#rulesetconfig">RulesetConfig</a>)
</p>

<p>
RulesetOptions are options for a ruleset.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>ruleset</code></br>
<em>
<a href="#options">Options</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Ruleset contains global options for the ruleset.</p>
</td>
</tr>
<tr>
<td>
<code>rules</code></br>
<em>
<a href="#options">Options</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Rules contains references to rule options.<br />Users can use these to configure the behaviour of specific rules.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="rulesetsummary">RulesetSummary
</h3>


<p>
(<em>Appears on:</em><a href="#compliancescanstatus">ComplianceScanStatus</a>)
</p>

<p>
RulesetSummary contains the identifiers and the summary for a specific ruleset.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>id</code></br>
<em>
string
</em>
</td>
<td>
<p>ID is the identifier of the ruleset that is summarized.</p>
</td>
</tr>
<tr>
<td>
<code>version</code></br>
<em>
string
</em>
</td>
<td>
<p>Version is the version of the ruleset that is summarized.</p>
</td>
</tr>
<tr>
<td>
<code>results</code></br>
<em>
<a href="#rulesresults">RulesResults</a>
</em>
</td>
<td>
<p>Results contains the results of the ruleset.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="scheduledcompliancescan">ScheduledComplianceScan
</h3>


<p>
ScheduledComplianceScan describes a scheduled compliance scan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>metadata</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectmeta-v1-meta">ObjectMeta</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#scheduledcompliancescanspec">ScheduledComplianceScanSpec</a>
</em>
</td>
<td>
<p>Spec contains the specification of this scheduled compliance scan.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#scheduledcompliancescanstatus">ScheduledComplianceScanStatus</a>
</em>
</td>
<td>
<p>Status contains the status of this scheduled compliance scan.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="scheduledcompliancescanspec">ScheduledComplianceScanSpec
</h3>


<p>
(<em>Appears on:</em><a href="#scheduledcompliancescan">ScheduledComplianceScan</a>)
</p>

<p>
ScheduledComplianceScanSpec is the specification of a ScheduledComplianceScan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>schedule</code></br>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Schedule is a cron expression defining when the compliance scan should run.</p>
</td>
</tr>
<tr>
<td>
<code>successfulScansHistoryLimit</code></br>
<em>
integer
</em>
</td>
<td>
<em>(Optional)</em>
<p>SuccessfulScansHistoryLimit is the number of completed compliance scans to keep.</p>
</td>
</tr>
<tr>
<td>
<code>failedScansHistoryLimit</code></br>
<em>
integer
</em>
</td>
<td>
<em>(Optional)</em>
<p>FailedScansHistoryLimit is the number of failed compliance scans to keep.</p>
</td>
</tr>
<tr>
<td>
<code>scanTemplate</code></br>
<em>
<a href="#scheduledcompliancescantemplate">ScheduledComplianceScanTemplate</a>
</em>
</td>
<td>
<p>ScanTemplate is the template for the ComplianceScan that will be created on each scheduled scan.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="scheduledcompliancescanstatus">ScheduledComplianceScanStatus
</h3>


<p>
(<em>Appears on:</em><a href="#scheduledcompliancescan">ScheduledComplianceScan</a>)
</p>

<p>
ScheduledComplianceScanStatus contains the status of a ScheduledComplianceScan.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>active</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#objectreference-v1-core">ObjectReference</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Active is a reference to the currently active ComplianceScan, if any.</p>
</td>
</tr>
<tr>
<td>
<code>lastScheduleTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#time-v1-meta">Time</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LastScheduleTime is the last time a ComplianceScan was scheduled.</p>
</td>
</tr>
<tr>
<td>
<code>lastCompletionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#time-v1-meta">Time</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>LastCompletionTime is the last time a scheduled ComplianceScan completed.</p>
</td>
</tr>

</tbody>
</table>


<h3 id="scheduledcompliancescantemplate">ScheduledComplianceScanTemplate
</h3>


<p>
(<em>Appears on:</em><a href="#scheduledcompliancescanspec">ScheduledComplianceScanSpec</a>)
</p>

<p>
ScheduledComplianceScanTemplate is the template for the ComplianceScan that will be created.
</p>

<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>

<tr>
<td>
<code>spec</code></br>
<em>
<a href="#compliancescanspec">ComplianceScanSpec</a>
</em>
</td>
<td>
<p>Spec is the spec of the ComplianceScan that will be created.</p>
</td>
</tr>

</tbody>
</table>


