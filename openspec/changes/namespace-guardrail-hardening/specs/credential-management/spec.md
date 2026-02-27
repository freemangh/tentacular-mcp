## MODIFIED Requirements

### Requirement: Issue short-lived ServiceAccount token
The system SHALL issue a time-bound token for the `tentacular-workflow` ServiceAccount in a given namespace using the Kubernetes TokenRequest API. The TTL SHALL be specified in minutes and MUST be between 10 and 1440 (24 hours). The system SHALL reject the operation if the target namespace is in the protected set (`tentacular-system`, `kube-system`, `kube-public`, `kube-node-lease`, `default`) or if the namespace is not managed by tentacular.

#### Scenario: Issue token with valid TTL
- **WHEN** the `cred_issue_token` tool is called with `namespace: "dev-alice"` and `ttl_minutes: 60`
- **THEN** the system returns a JWT token string valid for 60 minutes scoped to the workflow ServiceAccount in `dev-alice`

#### Scenario: Reject TTL out of range
- **WHEN** the `cred_issue_token` tool is called with `ttl_minutes: 5`
- **THEN** the system returns an error indicating the TTL must be between 10 and 1440 minutes

#### Scenario: Reject system namespace
- **WHEN** the `cred_issue_token` tool is called with a namespace in the protected set
- **THEN** the system returns an error indicating operations on that namespace are not permitted

#### Scenario: Reject unmanaged namespace
- **WHEN** the `cred_issue_token` tool is called with a namespace that lacks the managed-by label
- **THEN** the system returns an error indicating the namespace is not managed by tentacular and includes the kubectl label command to adopt it

### Requirement: Generate scoped kubeconfig
The system SHALL generate a complete kubeconfig YAML string containing the cluster CA, API server URL, issued token, and target namespace. The kubeconfig SHALL use context name `tentacular` and user name `tentacular-workflow`. The system SHALL call the token issuance internally with the specified TTL. The system SHALL reject the operation if the target namespace is not managed by tentacular.

#### Scenario: Generate kubeconfig
- **WHEN** the `cred_kubeconfig` tool is called with `namespace: "dev-alice"` and `ttl_minutes: 120`
- **THEN** the system returns a valid kubeconfig YAML with the cluster endpoint, CA data, a token valid for 120 minutes, and the namespace set to `dev-alice`

#### Scenario: Reject unmanaged namespace
- **WHEN** the `cred_kubeconfig` tool is called with a namespace that lacks the managed-by label
- **THEN** the system returns an error indicating the namespace is not managed by tentacular

### Requirement: Rotate credentials
The system SHALL rotate credentials for a namespace by deleting and recreating the `tentacular-workflow` ServiceAccount, which invalidates all previously issued tokens. The system SHALL reject the operation if the target namespace is in the protected set or is not managed by tentacular.

#### Scenario: Successful credential rotation
- **WHEN** the `cred_rotate` tool is called with `namespace: "dev-alice"` and the namespace is managed
- **THEN** the system deletes the existing workflow ServiceAccount and creates a new one, and all previously issued tokens for that ServiceAccount become invalid

#### Scenario: ServiceAccount does not exist
- **WHEN** the `cred_rotate` tool is called for a managed namespace where the ServiceAccount was never created
- **THEN** the system creates the ServiceAccount (idempotent behavior)

#### Scenario: Reject unmanaged namespace
- **WHEN** the `cred_rotate` tool is called with a namespace that lacks the managed-by label
- **THEN** the system returns an error indicating the namespace is not managed by tentacular
