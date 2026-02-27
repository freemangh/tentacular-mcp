## MODIFIED Requirements

### Requirement: List pods in namespace
The system SHALL list all pods in a given namespace, returning pod name, phase, ready condition, restart count, container images, and age. The ClusterRole SHALL include the `watch` verb on pods and events to support future event streaming for `wf_pods` and `wf_events`. The system SHALL reject the operation if the target namespace is in the protected set (`tentacular-system`, `kube-system`, `kube-public`, `kube-node-lease`, `default`).

#### Scenario: List pods with running workloads
- **WHEN** the `wf_pods` tool is called with `namespace: "dev-alice"` and pods exist
- **THEN** the system returns a list of pods with name, phase, ready status, restart count, images, and creation timestamp

#### Scenario: No pods in namespace
- **WHEN** the `wf_pods` tool is called with `namespace: "dev-alice"` and no pods exist
- **THEN** the system returns an empty list

### Requirement: Tail pod logs
The system SHALL retrieve the most recent log lines from a specified container in a pod. The caller MAY specify `tail_lines` (default 100) and an optional `container` name. If the pod has a single container, the container parameter MAY be omitted. The system SHALL reject the operation if the target namespace is in the protected set.

#### Scenario: Tail logs from single-container pod
- **WHEN** the `wf_logs` tool is called with `namespace: "dev-alice"`, `pod: "web-abc123"`, and `tail_lines: 50`
- **THEN** the system returns the last 50 log lines from the pod's only container

#### Scenario: Tail logs from specific container in multi-container pod
- **WHEN** the `wf_logs` tool is called with `namespace: "dev-alice"`, `pod: "web-abc123"`, `container: "sidecar"`, and `tail_lines: 20`
- **THEN** the system returns the last 20 log lines from the `sidecar` container

#### Scenario: Pod not found
- **WHEN** the `wf_logs` tool is called with a pod name that does not exist
- **THEN** the system returns an error indicating the pod was not found

### Requirement: List namespace events
The system SHALL list recent events in a namespace, returning type (Normal/Warning), reason, message, involved object, count, and last timestamp. Results SHALL be sorted by last timestamp descending and limited to the most recent 100 events by default. The system SHALL reject the operation if the target namespace is in the protected set.

#### Scenario: List events with warnings present
- **WHEN** the `wf_events` tool is called with `namespace: "dev-alice"`
- **THEN** the system returns events sorted by last timestamp descending, each with type, reason, message, involved object reference, count, and timestamp

#### Scenario: No events in namespace
- **WHEN** the `wf_events` tool is called and no events exist in the namespace
- **THEN** the system returns an empty list

### Requirement: List jobs and cronjobs
The system SHALL list all Jobs and CronJobs in a namespace. For Jobs, return name, status (active/succeeded/failed), start time, completion time, and duration. For CronJobs, return name, schedule, last schedule time, active count, and suspend status. The system SHALL reject the operation if the target namespace is in the protected set.

#### Scenario: List jobs and cronjobs
- **WHEN** the `wf_jobs` tool is called with `namespace: "dev-alice"`
- **THEN** the system returns separate lists of Jobs and CronJobs with their respective status fields

#### Scenario: No jobs or cronjobs
- **WHEN** the `wf_jobs` tool is called and no Jobs or CronJobs exist
- **THEN** the system returns empty lists for both Jobs and CronJobs
