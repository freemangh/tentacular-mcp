## MODIFIED Requirements

### Requirement: Apply module release
The system SHALL apply a set of Kubernetes manifests to a managed namespace using the dynamic client, tracking all created/updated resources with a release label (`tentacular.io/release: <name>`). The input SHALL include the deployment name, namespace, and a list of Kubernetes resource manifests as unstructured JSON/YAML. The ClusterRole SHALL grant `create,update,delete,patch,get,list,watch` verbs on all module resource types including deployments, services, configmaps, secrets, jobs, cronjobs, networkpolicies, and ingresses (networking.k8s.io). The `patch` verb is required for incremental updates (strategic merge patches). The `watch` verb supports status monitoring. The system SHALL validate that the namespace is managed by tentacular and reject the operation if the target namespace is `tentacular-system`.

#### Scenario: Apply new release
- **WHEN** the `wf_apply` tool is called with `namespace: "dev-alice"`, `name: "my-app"`, and a list of manifests
- **THEN** the system creates all resources in the namespace, labels them with `tentacular.io/release: my-app`, and returns the count of created resources

#### Scenario: Update existing release
- **WHEN** the `wf_apply` tool is called with a name that already has resources in the namespace
- **THEN** the system applies the manifests (create or update), labels them, and removes any previously-labeled resources that are no longer in the manifest set

#### Scenario: Reject for unmanaged namespace
- **WHEN** the `wf_apply` tool is called for a namespace without the managed-by label
- **THEN** the system returns an error indicating the namespace is not managed by tentacular

### Requirement: Remove module release
The system SHALL delete all resources labeled with `tentacular.io/release: <name>` in the given namespace. The system SHALL validate the namespace is managed and reject the operation if the target namespace is `tentacular-system`.

#### Scenario: Remove existing release
- **WHEN** the `wf_remove` tool is called with `namespace: "dev-alice"` and `name: "my-app"`
- **THEN** the system deletes all resources with the `tentacular.io/release: my-app` label and returns the count of deleted resources

#### Scenario: Release not found
- **WHEN** the `wf_remove` tool is called with a name that has no matching resources
- **THEN** the system returns a success response with zero resources deleted

### Requirement: Get module release status
The system SHALL list all resources in a namespace labeled with `tentacular.io/release: <name>` and return their kind, name, and readiness status. The system SHALL reject the operation if the target namespace is `tentacular-system`.

#### Scenario: Get status of healthy release
- **WHEN** the `wf_status` tool is called with `namespace: "dev-alice"` and `name: "my-app"` and all resources are ready
- **THEN** the system returns a list of resources with kind, name, and `ready: true` for each

#### Scenario: Get status with unhealthy resources
- **WHEN** the `wf_status` tool is called and some Deployments have unavailable replicas
- **THEN** the system returns resources with `ready: false` for the unhealthy ones, including a reason string
