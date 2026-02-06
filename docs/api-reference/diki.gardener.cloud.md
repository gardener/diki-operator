# API Reference

## Packages
- [diki.gardener.cloud/v1alpha1](#dikigardenercloudv1alpha1)


## diki.gardener.cloud/v1alpha1

Package v1alpha1 is a version of the API.



#### ComplianceRun



ComplianceRun describes a compliance run.



_Appears in:_
- [ComplianceRunList](#compliancerunlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `kind` _string_ | Kind is a string value representing the REST resource this object represents.<br />Servers may infer this from the endpoint the client submits requests to.<br />Cannot be updated.<br />In CamelCase.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds |  | Optional: \{\} <br /> |
| `apiVersion` _string_ | APIVersion defines the versioned schema of this representation of an object.<br />Servers should convert recognized schemas to the latest internal value, and<br />may reject unrecognized values.<br />More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources |  | Optional: \{\} <br /> |
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |  |  |
| `spec` _[ComplianceRunSpec](#compliancerunspec)_ | Spec contains the specification of this compliance run. |  |  |
| `status` _[ComplianceRunStatus](#compliancerunstatus)_ | Status contains the status of this compliance run. |  |  |




#### ComplianceRunPhase

_Underlying type:_ _string_

ComplianceRunPhase is an alias for string representing the phase of a ComplianceRun.



_Appears in:_
- [ComplianceRunStatus](#compliancerunstatus)

| Field | Description |
| --- | --- |
| `Pending` | ComplianceRunPending means that the ComplianceRun is pending execution.<br /> |
| `Running` | ComplianceRunRunning means that the ComplianceRun is running.<br /> |
| `Completed` | ComplianceRunCompleted means that the ComplianceRun has completed successfully.<br /> |
| `Failed` | ComplianceRunFailed means that the ComplianceRun has failed.<br /> |


#### ComplianceRunSpec



ComplianceRunSpec is the specification of a ComplianceRun.



_Appears in:_
- [ComplianceRun](#compliancerun)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `rulesets` _[RulesetConfig](#rulesetconfig) array_ | Rulesets describe the rulesets to be applied during the compliance run. |  |  |


#### ComplianceRunStatus



ComplianceRunStatus contains the status of a ComplianceRun.



_Appears in:_
- [ComplianceRun](#compliancerun)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `conditions` _[Condition](#condition) array_ | Conditions contains the conditions of the ComplianceRun. |  |  |
| `phase` _[ComplianceRunPhase](#compliancerunphase)_ | Phase represents the current phase of the ComplianceRun. |  |  |
| `rulesets` _[RulesetSummary](#rulesetsummary) array_ | Rulesets contains the ruleset summaries of the ComplianceRun. |  |  |


#### Condition



Condition described a condition of a ComplianceRun.



_Appears in:_
- [ComplianceRunStatus](#compliancerunstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `type` _[ConditionType](#conditiontype)_ | Type of condition. |  |  |
| `status` _[ConditionStatus](#conditionstatus)_ | Status of the condition. |  |  |
| `lastUpdateTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#time-v1-meta)_ | Last time the condition was updated. |  |  |
| `lastTransitionTime` _[Time](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#time-v1-meta)_ | LastTransitionTime is the last time the condition transitioned from one status to another. |  |  |
| `reason` _string_ | Reason is a brief reason for the condition's last transition. |  |  |
| `message` _string_ | Message is a human-readable message indicating details about the last transition. |  |  |


#### ConditionStatus

_Underlying type:_ _string_

ConditionStatus is an alias for string representing the status of a condition.



_Appears in:_
- [Condition](#condition)

| Field | Description |
| --- | --- |
| `True` | ConditionTrue means a resource is in the condition.<br /> |
| `False` | ConditionFalse means a resource is not in the condition.<br /> |
| `Unknown` | ConditionUnknown means diki-operator cannot decide if a resource is in the condition or not.<br /> |


#### ConditionType

_Underlying type:_ _string_

ConditionType is an alias for string representing the type of a condition.



_Appears in:_
- [Condition](#condition)

| Field | Description |
| --- | --- |
| `Completed` | ConditionTypeCompleted indicates whether the ComplianceRun has completed.<br /> |
| `Failed` | ConditionTypeFailed indicates whether the ComplianceRun has failed.<br /> |


#### Findings



Findings contains information about the specific rules that have errored/warned/failed.



_Appears in:_
- [RulesetSummary](#rulesetsummary)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `failed` _[Rule](#rule) array_ | Failed contains information about the rules that contain a Failed checkResult. |  |  |
| `errored` _[Rule](#rule) array_ | Errored contains information about the rules that contain a Errored checkResult. |  |  |
| `warning` _[Rule](#rule) array_ | Warning contains information about the rules that contain a Warning checkResult. |  |  |


#### Rule



Rule contains information about the ID and the name of the rule that contains the findings.



_Appears in:_
- [Findings](#findings)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `ruleID` _string_ | ID is the unique identifier of the rule which contains the finding. |  |  |
| `ruleName` _string_ | Name is name of the rule which contains the finding. |  |  |


#### RuleOptionsConfigMapRef



RuleOptionsConfigMapRef references a ConfigMap containing rule options for the ruleset.



_Appears in:_
- [RulesetConfig](#rulesetconfig)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `name` _string_ | Name is the name of the ConfigMap. |  |  |
| `namespace` _string_ | Namespace is the namespace of the ConfigMap. |  |  |
| `key` _string_ | Key is the key in the ConfigMap. |  |  |


#### RulesetConfig



RulesetConfig describes the configuration of a ruleset.



_Appears in:_
- [ComplianceRunSpec](#compliancerunspec)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `id` _string_ | ID is the identifier of the ruleset. |  |  |
| `version` _string_ | Version is the version of the ruleset. |  |  |
| `ruleOptionsConfigMapRef` _[RuleOptionsConfigMapRef](#ruleoptionsconfigmapref)_ | RuleOptionsConfigMapRef references a ConfigMap containing rule options for the ruleset. |  |  |


#### RulesetSummary



RulesetSummary contains the identifiers and the summary for a specific ruleset.



_Appears in:_
- [ComplianceRunStatus](#compliancerunstatus)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `id` _string_ | ID is the identifier of the ruleset that is summarized. |  |  |
| `version` _string_ | Version is the version of the ruleset that is summarized. |  |  |
| `summary` _[Summary](#summary)_ | Summary contains information about the amount of rules per each status. |  |  |
| `findings` _[Findings](#findings)_ | Findings contains information about the specific rules that have errored/warned/failed |  |  |


#### Summary



Summary contains information about the amount of rules per each status.



_Appears in:_
- [RulesetSummary](#rulesetsummary)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `passed` _integer_ | Passed counts the amount of rules in a specific ruleset that have passed. |  |  |
| `skipped` _integer_ | Skipped counts the amount of rules in a specific ruleset that have been skipped. |  |  |
| `accepted` _integer_ | Accepted counts the amount of rules in a specific ruleset that have been accepted. |  |  |
| `warning` _integer_ | Warning counts the amount of rules in a specific ruleset that have returned a warning. |  |  |
| `failed` _integer_ | Failed counts the amount of rules in a specific ruleset that have failed. |  |  |
| `errored` _integer_ | Errored counts the amount of rules in a specific ruleset that have errored. |  |  |


