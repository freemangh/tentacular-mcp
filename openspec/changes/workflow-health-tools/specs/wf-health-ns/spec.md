## ADDED Requirements

### Requirement: wf_health_ns tool registration
The MCP server SHALL register a wf_health_ns tool that accepts namespace and optional limit parameters. The tool SHALL validate the namespace via guard.CheckNamespace before proceeding.

#### Scenario: Tool is registered
- **WHEN** the MCP server starts
- **THEN** the wf_health_ns tool SHALL be available with name "wf_health_ns" and description mentioning aggregate G/A/R health

#### Scenario: Invalid namespace rejected
- **WHEN** wf_health_ns is called with a namespace that fails guard.CheckNamespace
- **THEN** the tool SHALL return an error without querying the cluster

### Requirement: wf_health_ns namespace scan
The wf_health_ns tool SHALL list all tentacular-managed deployments in the namespace using the label selector app.kubernetes.io/managed-by=tentacular and probe each for health status.

#### Scenario: Lists managed deployments only
- **WHEN** wf_health_ns is called for a namespace containing both tentacular and non-tentacular deployments
- **THEN** only deployments with label app.kubernetes.io/managed-by=tentacular SHALL be included

#### Scenario: Default limit of 20
- **WHEN** wf_health_ns is called without a limit parameter
- **THEN** at most 20 workflows SHALL be checked

#### Scenario: Custom limit
- **WHEN** wf_health_ns is called with limit=5
- **THEN** at most 5 workflows SHALL be checked

#### Scenario: Truncation reported
- **WHEN** the namespace contains more deployments than the limit
- **THEN** the result SHALL have truncated=true and total set to the actual count

### Requirement: wf_health_ns G/A/R summary
The wf_health_ns tool SHALL return a summary with counts of green, amber, and red workflows alongside the individual workflow entries.

#### Scenario: Summary counts match entries
- **WHEN** wf_health_ns checks 3 workflows: 2 green, 1 red
- **THEN** the summary SHALL have green=2, amber=0, red=1
- **AND** the workflows array SHALL contain 3 entries with matching statuses

#### Scenario: All healthy namespace
- **WHEN** all workflows in the namespace are healthy
- **THEN** summary.green SHALL equal the workflow count and summary.amber and summary.red SHALL be 0
