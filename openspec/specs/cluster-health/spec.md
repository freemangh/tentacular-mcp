# cluster-health Specification

## Purpose
TBD - created by archiving change in-cluster-mcp-server. Update Purpose after archive.
## Requirements
### Requirement: Report node health
The system SHALL return the health status of all cluster nodes including name, ready condition, CPU/memory capacity, CPU/memory allocatable, kubelet version, and any node conditions that are not healthy (e.g., MemoryPressure, DiskPressure, PIDPressure).

#### Scenario: All nodes healthy
- **WHEN** the `health_nodes` tool is called and all nodes are Ready
- **THEN** the system returns a list of nodes with `ready: true` and their capacity/allocatable resources

#### Scenario: Node with pressure condition
- **WHEN** the `health_nodes` tool is called and a node has MemoryPressure=True
- **THEN** the system returns that node with `ready: true` (or false if NotReady) and includes the pressure condition in a `conditions` list

### Requirement: Report namespace resource usage
The system SHALL return resource utilization for a given namespace, comparing actual usage against the ResourceQuota limits. The response SHALL include CPU and memory used vs. limit, pod count vs. limit, and a utilization percentage for each resource. The system SHALL reject the operation if the target namespace is `tentacular-system`.

#### Scenario: Namespace with quota and active workloads
- **WHEN** the `health_ns_usage` tool is called with `namespace: "dev-alice"` and the namespace has a quota and running pods
- **THEN** the system returns used CPU, used memory, pod count, their respective quota limits, and utilization percentages

#### Scenario: Namespace without quota
- **WHEN** the `health_ns_usage` tool is called for a namespace without a ResourceQuota
- **THEN** the system returns usage figures with limits shown as "unlimited"

### Requirement: Report cluster resource summary
The system SHALL return an aggregate cluster resource summary including total/allocatable/requested CPU and memory across all nodes, total node count, ready node count, and total pod count across all namespaces.

#### Scenario: Multi-node cluster summary
- **WHEN** the `health_cluster_summary` tool is called on a 3-node cluster
- **THEN** the system returns aggregated capacity, allocatable, and requested totals for CPU and memory, node count (3 total, N ready), and total pod count

