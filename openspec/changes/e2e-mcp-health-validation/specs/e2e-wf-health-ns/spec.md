## ADDED Requirements

### Requirement: Namespace-wide health aggregation
The test suite SHALL verify that wf_health_ns correctly aggregates G/A/R status across all tentacular deployments in a namespace.

#### Scenario: Mixed health states produce correct summary counts
- **WHEN** a namespace contains 3 GREEN, 1 AMBER, and 1 RED workflow deployments
- **THEN** wf_health_ns SHALL return summary with green=3, amber=1, red=1 and 5 individual workflow entries

#### Scenario: All healthy namespace
- **WHEN** a namespace contains only GREEN workflow deployments
- **THEN** wf_health_ns SHALL return summary with green matching deployment count, amber=0, red=0

#### Scenario: Empty namespace returns zero counts
- **WHEN** a namespace contains no tentacular-managed deployments
- **THEN** wf_health_ns SHALL return summary with green=0, amber=0, red=0 and an empty workflows list

### Requirement: Limit parameter enforcement
The test suite SHALL verify that wf_health_ns respects the limit parameter.

#### Scenario: Limit caps the number of workflows checked
- **WHEN** wf_health_ns is called with limit=2 on a namespace with 5 deployments
- **THEN** the result SHALL contain at most 2 workflow entries

#### Scenario: Default limit is 20
- **WHEN** wf_health_ns is called without a limit parameter
- **THEN** the result SHALL check up to 20 deployments

### Requirement: Label selector filtering
The test suite SHALL verify that wf_health_ns only checks deployments managed by tentacular.

#### Scenario: Non-tentacular deployments are excluded
- **WHEN** a namespace contains both tentacular-managed and unrelated deployments
- **THEN** wf_health_ns SHALL only include deployments with the tentacular managed-by label in the results
