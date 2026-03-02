# security-audit Specification

## Purpose
TBD - created by archiving change in-cluster-mcp-server. Update Purpose after archive.
## Requirements
### Requirement: Audit RBAC for over-permissions
The system SHALL scan all Roles and RoleBindings in a given namespace and flag any rules that grant wildcard verbs (`*`), wildcard resources (`*`), or access to sensitive resources (secrets with `list`/`watch`, pods/exec, nodes). The system SHALL return a list of findings, each with the role name, the problematic rule, and a severity level (high/medium/low). The system SHALL reject the operation if the target namespace is `tentacular-system`.

#### Scenario: Namespace with over-permissioned role
- **WHEN** the `audit_rbac` tool is called with `namespace: "dev-alice"` and a Role grants `*` verbs on secrets
- **THEN** the system returns a finding with severity `high`, the role name, and the flagged rule

#### Scenario: Namespace with clean RBAC
- **WHEN** the `audit_rbac` tool is called and all Roles follow least-privilege principles
- **THEN** the system returns an empty findings list

#### Scenario: Also inspect ClusterRoleBindings targeting namespace
- **WHEN** the `audit_rbac` tool is called and a ClusterRoleBinding grants a ClusterRole to a ServiceAccount in the target namespace
- **THEN** the system includes any over-permissioned rules from the bound ClusterRole in the findings

### Requirement: Audit network policy coverage
The system SHALL verify that a namespace has at least a default-deny NetworkPolicy (denying all ingress and egress) and report on all NetworkPolicies present. The system SHALL flag namespaces that allow unrestricted egress or have no NetworkPolicies at all. The system SHALL reject the operation if the target namespace is `tentacular-system`.

#### Scenario: Namespace with default-deny policy
- **WHEN** the `audit_netpol` tool is called with `namespace: "dev-alice"` and a default-deny policy exists
- **THEN** the system returns `default_deny: true` and lists all NetworkPolicies with their policy types and pod selectors

#### Scenario: Namespace without network policies
- **WHEN** the `audit_netpol` tool is called and no NetworkPolicies exist in the namespace
- **THEN** the system returns `default_deny: false` with a finding flagging the namespace as having unrestricted network access

#### Scenario: Namespace with partial coverage
- **WHEN** the `audit_netpol` tool is called and the namespace has ingress policies but no egress restriction
- **THEN** the system returns `default_deny: false` with a finding noting unrestricted egress

### Requirement: Audit Pod Security Admission labels
The system SHALL check the Pod Security Admission labels on a namespace and report the enforce, audit, and warn levels. The system SHALL flag namespaces that do not have PSA enforce set to `restricted` or that have no PSA labels at all. The system SHALL treat `privileged` enforce level as high severity (all checks disabled) and `baseline` as medium severity. The system SHALL detect when audit or warn levels are weaker than the enforce level. The system SHALL return findings with remediation suggestions. The system SHALL reject the operation if the target namespace is `tentacular-system`.

#### Scenario: Namespace with restricted PSA
- **WHEN** the `audit_psa` tool is called with `namespace: "dev-alice"` and PSA enforce is `restricted` with matching audit and warn levels
- **THEN** the system returns `compliant: true` with the enforce, audit, and warn levels and no findings

#### Scenario: Namespace with privileged PSA
- **WHEN** the `audit_psa` tool is called and PSA enforce is `privileged`
- **THEN** the system returns `compliant: false` with a high-severity finding indicating all pod security checks are disabled

#### Scenario: Namespace with baseline PSA
- **WHEN** the `audit_psa` tool is called and PSA enforce is `baseline`
- **THEN** the system returns `compliant: false` with a medium-severity finding recommending upgrade to `restricted`

#### Scenario: Namespace with no PSA labels
- **WHEN** the `audit_psa` tool is called and no PSA labels exist on the namespace
- **THEN** the system returns `compliant: false` with a high-severity finding noting the absence of PSA configuration

#### Scenario: Audit level weaker than enforce
- **WHEN** the `audit_psa` tool is called and the audit level is weaker than the enforce level (e.g. enforce=restricted, audit=baseline)
- **THEN** the system returns a medium-severity finding noting the audit log will not capture all violations

#### Scenario: Warn level weaker than enforce
- **WHEN** the `audit_psa` tool is called and the warn level is weaker than the enforce level
- **THEN** the system returns a medium-severity finding noting users will not see warnings for all enforced restrictions

