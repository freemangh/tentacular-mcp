## MODIFIED Requirements

### Requirement: Create managed namespace
The system SHALL create a Kubernetes namespace with the `app.kubernetes.io/managed-by: tentacular` label and Pod Security Admission labels set to `restricted` profile at `latest` version. The system SHALL also create a default-deny NetworkPolicy, a DNS-allow NetworkPolicy, a ResourceQuota (from a named preset), a LimitRange with default container resource requests/limits, a workflow ServiceAccount, a workflow Role, and a workflow RoleBinding in the new namespace. The workflow Role SHALL grant `create,update,delete,patch,get,list,watch` on apps/deployments, core/services+configmaps+secrets, batch/cronjobs+jobs, networking.k8s.io/networkpolicies+ingresses; `get,list,watch` on core/pods+pods/log+events; and `get,list,patch,update` on core/serviceaccounts (for imagePullSecrets management). The system SHALL reject the operation if the target namespace is in the protected set: `tentacular-system`, `kube-system`, `kube-public`, `kube-node-lease`, or `default`.

#### Scenario: Successful namespace creation with small quota
- **WHEN** the `ns_create` tool is called with `name: "dev-alice"` and `quota_preset: "small"`
- **THEN** the system creates namespace `dev-alice` with managed-by label, PSA restricted labels, default-deny and DNS-allow NetworkPolicies, a ResourceQuota with CPU=2, Mem=2Gi, Pods=10, a LimitRange, and the workflow ServiceAccount/Role/RoleBinding

#### Scenario: Reject creation of system namespace
- **WHEN** the `ns_create` tool is called with any name in the protected set (`tentacular-system`, `kube-system`, `kube-public`, `kube-node-lease`, `default`)
- **THEN** the system returns an error indicating operations on that namespace are not permitted

#### Scenario: Namespace already exists
- **WHEN** the `ns_create` tool is called with a name that already exists
- **THEN** the system returns an error indicating the namespace already exists

### Requirement: Delete managed namespace
The system SHALL delete a namespace only if it carries the `app.kubernetes.io/managed-by: tentacular` label. The system SHALL reject deletion of any namespace in the protected set: `tentacular-system`, `kube-system`, `kube-public`, `kube-node-lease`, or `default`. Deleting a namespace removes all resources within it (Kubernetes garbage collection).

#### Scenario: Successful deletion of managed namespace
- **WHEN** the `ns_delete` tool is called with `name: "dev-alice"` and the namespace has the managed-by label
- **THEN** the system deletes the namespace

#### Scenario: Reject deletion of unmanaged namespace
- **WHEN** the `ns_delete` tool is called with a namespace that lacks the managed-by label
- **THEN** the system returns an error indicating the namespace is not managed by tentacular and includes the kubectl label command to adopt it

#### Scenario: Reject deletion of system namespace
- **WHEN** the `ns_delete` tool is called with any name in the protected set
- **THEN** the system returns an error indicating operations on that namespace are not permitted

### Requirement: Get namespace details
The system SHALL retrieve a single namespace by name and return its metadata, labels, annotations, status, resource quota usage, and limit range configuration. The system SHALL reject the operation if the target is in the protected set.

#### Scenario: Get existing namespace
- **WHEN** the `ns_get` tool is called with `name: "dev-alice"`
- **THEN** the system returns the namespace metadata, labels, status phase, quota summary, and limit range summary

#### Scenario: Namespace not found
- **WHEN** the `ns_get` tool is called with a name that does not exist
- **THEN** the system returns an error indicating the namespace was not found

### Requirement: List managed namespaces
The system SHALL list all namespaces with the `app.kubernetes.io/managed-by: tentacular` label, returning name, status, creation timestamp, and quota preset for each.

#### Scenario: List with managed namespaces present
- **WHEN** the `ns_list` tool is called and managed namespaces exist
- **THEN** the system returns a list of managed namespaces with their metadata

#### Scenario: List with no managed namespaces
- **WHEN** the `ns_list` tool is called and no managed namespaces exist
- **THEN** the system returns an empty list

## ADDED Requirements

### Requirement: Adopt pre-existing namespace
A pre-existing namespace (not created by `ns_create`) MAY be brought under tentacular management by applying the `app.kubernetes.io/managed-by: tentacular` label manually. Once labeled, the namespace SHALL be visible to `ns_list` and accessible to all tentacular write tools. The system SHALL NOT provide an automated adoption tool in this release; adoption is performed via `kubectl label namespace <name> app.kubernetes.io/managed-by=tentacular`.

#### Scenario: Adopted namespace visible in list
- **WHEN** a pre-existing namespace is manually labeled with `app.kubernetes.io/managed-by=tentacular`
- **THEN** `ns_list` includes it and all write tools (module_apply, cred_issue_token, etc.) accept it as a valid target

#### Scenario: Unadopted namespace rejected by write tools
- **WHEN** any write tool is called with a namespace that lacks the managed-by label
- **THEN** the system returns an error indicating the namespace is not managed by tentacular and includes the kubectl label command to adopt it
