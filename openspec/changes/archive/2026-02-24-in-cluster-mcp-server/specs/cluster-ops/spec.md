## ADDED Requirements

### Requirement: Run preflight checks
The system SHALL run a series of validation checks for a given namespace and return structured results. Checks SHALL include: API server reachability, namespace existence, RBAC permissions for the workflow ServiceAccount (deployments, services, configmaps, secrets, cronjobs, jobs), and gVisor RuntimeClass availability. Each check result SHALL include name, pass/fail status, and optional warning and remediation text. The system SHALL reject the operation if the target namespace is `tentacular-system`.

#### Scenario: All checks pass
- **WHEN** the `cluster_preflight` tool is called with `namespace: "dev-alice"` and all prerequisites are met
- **THEN** the system returns a list of check results, all with `passed: true`

#### Scenario: Some checks fail
- **WHEN** the `cluster_preflight` tool is called and the namespace does not exist
- **THEN** the system returns check results with `namespace-exists` marked as `passed: false` with a remediation message

#### Scenario: gVisor not installed
- **WHEN** the `cluster_preflight` tool is called and no gVisor RuntimeClass exists
- **THEN** the `gvisor-runtime` check returns `passed: true` with a warning that gVisor is not available

### Requirement: Generate cluster profile
The system SHALL produce a comprehensive snapshot of cluster capabilities including: Kubernetes version, detected distribution (EKS/GKE/AKS/K3s/vanilla), node inventory with OS/arch/capacity/allocatable, RuntimeClass list with gVisor detection, CNI identification with network policy support flags, StorageClass list with default and RWX capability flags, CSI driver list, detected extensions (istio, cert-manager, prometheus, external-secrets, argocd, gateway-api), ingress resources (networking.k8s.io), workload topology (replicasets, daemonsets, statefulsets), storage posture (persistentvolumes, persistentvolumeclaims, volumeattachments), service discovery (endpoints, endpointslices), and namespace-level details (pod security level, quota summary, limit range summary). The ClusterRole SHALL include the `watch` verb on all profiling resources to support event streaming. The system SHALL reject the operation if the target namespace is `tentacular-system`.

#### Scenario: Profile a cluster with namespace context
- **WHEN** the `cluster_profile` tool is called with `namespace: "dev-alice"`
- **THEN** the system returns a ClusterProfile JSON object with all sections populated, including namespace-specific quota and limit range data

#### Scenario: Profile a cluster without namespace
- **WHEN** the `cluster_profile` tool is called without a namespace parameter
- **THEN** the system returns a ClusterProfile JSON object with cluster-wide data and the namespace-specific sections omitted
