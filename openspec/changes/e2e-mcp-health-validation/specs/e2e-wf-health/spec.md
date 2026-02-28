## ADDED Requirements

### Requirement: G/A/R classification correctness
The test suite SHALL verify that wf_health produces the correct Green/Amber/Red classification for all meaningful combinations of deployment state and health probe response.

#### Scenario: GREEN when pod ready and health probe succeeds
- **WHEN** the deployment has ready replicas > 0 and the health probe returns HTTP 200 with no amber signals
- **THEN** wf_health SHALL return status "GREEN" with pod_ready true

#### Scenario: RED when pod not ready
- **WHEN** the deployment has ready replicas == 0
- **THEN** wf_health SHALL return status "RED" with reason indicating pod not ready

#### Scenario: RED when health probe unreachable
- **WHEN** the deployment has ready replicas > 0 but the health probe returns a connection error or timeout
- **THEN** wf_health SHALL return status "RED" with reason indicating health endpoint unreachable

#### Scenario: AMBER when last execution failed
- **WHEN** the deployment has ready replicas > 0 and the health probe detail response contains last_status "failed"
- **THEN** wf_health SHALL return status "AMBER" with reason indicating last execution failure

#### Scenario: AMBER when execution in flight
- **WHEN** the deployment has ready replicas > 0 and the health probe detail response contains in_flight true
- **THEN** wf_health SHALL return status "AMBER" with reason indicating execution in progress

### Requirement: wf_health detail mode
The test suite SHALL verify that wf_health passes the detail flag to the health probe and includes the raw detail in the response.

#### Scenario: Detail mode includes telemetry snapshot
- **WHEN** wf_health is called with detail=true and the health probe returns a telemetry snapshot
- **THEN** the result SHALL include the raw detail string from the health endpoint

#### Scenario: Non-detail mode omits telemetry
- **WHEN** wf_health is called with detail=false (default)
- **THEN** the result SHALL have an empty detail field

### Requirement: Namespace guard enforcement
The test suite SHALL verify that wf_health rejects requests targeting the tentacular-system namespace.

#### Scenario: Reject system namespace
- **WHEN** wf_health is called with namespace "tentacular-system"
- **THEN** the tool SHALL return an error indicating the namespace is not allowed
