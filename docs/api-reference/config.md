<p>Packages:</p>
<ul>
<li>
<a href="#diki.gardener.cloud%2fv1alpha1">diki.gardener.cloud/v1alpha1</a>
</li>
</ul>
<h2 id="diki.gardener.cloud/v1alpha1">diki.gardener.cloud/v1alpha1</h2>
<p>
<p>Package v1alpha1 is a version of the API.</p>
</p>
Resource Types:
<ul></ul>
<h3 id="diki.gardener.cloud/v1alpha1.ComplianceRun">ComplianceRun
</h3>
<p>
<p>ComplianceRun describes a compliance run.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
<p>Standard object metadata.</p>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRunSpec">
ComplianceRunSpec
</a>
</em>
</td>
<td>
<p>Spec contains the specification of this compliance run.</p>
<br/>
<br/>
<table>
<tr>
<td>
<code>rulesets</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.RulesetConfig">
[]RulesetConfig
</a>
</em>
</td>
<td>
<p>Rulesets describe the rulesets to be applied during the compliance run.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRunStatus">
ComplianceRunStatus
</a>
</em>
</td>
<td>
<p>Status contains the status of this compliance run.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.ComplianceRunPhase">ComplianceRunPhase
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRunStatus">ComplianceRunStatus</a>)
</p>
<p>
<p>ComplianceRunPhase is an alias for string representing the phase of a ComplianceRun.</p>
</p>
<h3 id="diki.gardener.cloud/v1alpha1.ComplianceRunSpec">ComplianceRunSpec
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRun">ComplianceRun</a>)
</p>
<p>
<p>ComplianceRunSpec is the specification of a ComplianceRun.</p>
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
<a href="#diki.gardener.cloud/v1alpha1.RulesetConfig">
[]RulesetConfig
</a>
</em>
</td>
<td>
<p>Rulesets describe the rulesets to be applied during the compliance run.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.ComplianceRunStatus">ComplianceRunStatus
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRun">ComplianceRun</a>)
</p>
<p>
<p>ComplianceRunStatus contains the status of a ComplianceRun.</p>
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
<a href="#diki.gardener.cloud/v1alpha1.Condition">
[]Condition
</a>
</em>
</td>
<td>
<p>Conditions contains the conditions of the ComplianceRun.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRunPhase">
ComplianceRunPhase
</a>
</em>
</td>
<td>
<p>Phase represents the current phase of the ComplianceRun.</p>
</td>
</tr>
<tr>
<td>
<code>rulesets</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.RulesetSummary">
[]RulesetSummary
</a>
</em>
</td>
<td>
<p>Rulesets contains the ruleset summaries of the ComplianceRun.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.Condition">Condition
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRunStatus">ComplianceRunStatus</a>)
</p>
<p>
<p>Condition described a condition of a ComplianceRun.</p>
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
<a href="#diki.gardener.cloud/v1alpha1.ConditionType">
ConditionType
</a>
</em>
</td>
<td>
<p>Type of condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.ConditionStatus">
ConditionStatus
</a>
</em>
</td>
<td>
<p>Status of the condition.</p>
</td>
</tr>
<tr>
<td>
<code>lastUpdateTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>Last time the condition was updated.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code></br>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
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
<p>Reason is a brief reason for the condition&rsquo;s last transition.</p>
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
<h3 id="diki.gardener.cloud/v1alpha1.ConditionStatus">ConditionStatus
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.Condition">Condition</a>)
</p>
<p>
<p>ConditionStatus is an alias for string representing the status of a condition.</p>
</p>
<h3 id="diki.gardener.cloud/v1alpha1.ConditionType">ConditionType
(<code>string</code> alias)</p></h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.Condition">Condition</a>)
</p>
<p>
<p>ConditionType is an alias for string representing the type of a condition.</p>
</p>
<h3 id="diki.gardener.cloud/v1alpha1.Findings">Findings
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.RulesetSummary">RulesetSummary</a>)
</p>
<p>
<p>Findings contains information about the specific rules that have errored/warned/failed.</p>
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
<a href="#diki.gardener.cloud/v1alpha1.Rule">
[]Rule
</a>
</em>
</td>
<td>
<p>Failed contains information about the rules that contain a Failed checkResult.</p>
</td>
</tr>
<tr>
<td>
<code>errored</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.Rule">
[]Rule
</a>
</em>
</td>
<td>
<p>Errored contains information about the rules that contain a Errored checkResult.</p>
</td>
</tr>
<tr>
<td>
<code>warning</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.Rule">
[]Rule
</a>
</em>
</td>
<td>
<p>Warning contains information about the rules that contain a Warning checkResult.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.Rule">Rule
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.Findings">Findings</a>)
</p>
<p>
<p>Rule contains information about the ID and the name of the rule that contains the findings.</p>
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
<code>ruleID</code></br>
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
<code>ruleName</code></br>
<em>
string
</em>
</td>
<td>
<p>Name is name of the rule which contains the finding.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.RuleOptions">RuleOptions
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.RulesetConfig">RulesetConfig</a>)
</p>
<p>
<p>RuleOptions describes the rule options for a ruleset.</p>
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
<a href="#diki.gardener.cloud/v1alpha1.RuleOptionsConfigMapRef">
RuleOptionsConfigMapRef
</a>
</em>
</td>
<td>
<p>ConfigMapRef references a ConfigMap containing rule options for the ruleset.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.RuleOptionsConfigMapRef">RuleOptionsConfigMapRef
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.RuleOptions">RuleOptions</a>)
</p>
<p>
<p>RuleOptionsConfigMapRef references a ConfigMap containing rule options for the ruleset.</p>
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
<p>Key is the key in the ConfigMap.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.RulesetConfig">RulesetConfig
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRunSpec">ComplianceRunSpec</a>)
</p>
<p>
<p>RulesetConfig describes the configuration of a ruleset.</p>
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
<code>ruleOptions</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.RuleOptions">
RuleOptions
</a>
</em>
</td>
<td>
<p>RuleOptions describes the rule options for a ruleset.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.RulesetSummary">RulesetSummary
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.ComplianceRunStatus">ComplianceRunStatus</a>)
</p>
<p>
<p>RulesetSummary contains the identifiers and the summary for a specific ruleset.</p>
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
<code>summary</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.Summary">
Summary
</a>
</em>
</td>
<td>
<p>Summary contains information about the amount of rules per each status.</p>
</td>
</tr>
<tr>
<td>
<code>findings</code></br>
<em>
<a href="#diki.gardener.cloud/v1alpha1.Findings">
Findings
</a>
</em>
</td>
<td>
<p>Findings contains information about the specific rules that have errored/warned/failed</p>
</td>
</tr>
</tbody>
</table>
<h3 id="diki.gardener.cloud/v1alpha1.Summary">Summary
</h3>
<p>
(<em>Appears on:</em>
<a href="#diki.gardener.cloud/v1alpha1.RulesetSummary">RulesetSummary</a>)
</p>
<p>
<p>Summary contains information about the amount of rules per each status.</p>
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
int
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
int
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
int
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
int
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
int
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
int
</em>
</td>
<td>
<p>Errored counts the amount of rules in a specific ruleset that have errored.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <a href="https://github.com/ahmetb/gen-crd-api-reference-docs">gen-crd-api-reference-docs</a>
</em></p>
