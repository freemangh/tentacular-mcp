## ADDED Requirements

### Requirement: wf_health tool registration
The MCP server SHALL register a wf_health tool that accepts namespace, name, and optional detail parameters. The tool SHALL validate the namespace via guard.CheckNamespace before proceeding.

#### Scenario: Tool is registered
- **WHEN** the MCP server starts
- **THEN** the wf_health tool SHALL be available with name "wf_health" and description mentioning G/A/R health status

#### Scenario: Invalid namespace rejected
- **WHEN** wf_health is called with a namespace that fails guard.CheckNamespace
- **THEN** the tool SHALL return an error without querying the cluster

### Requirement: wf_health G/A/R classification
The wf_health tool SHALL classify workflow health as Green, Amber, or Red based on deployment status and health endpoint probe.

#### Scenario: RED when pod not ready
- **WHEN** the deployment has 0 ready replicas
- **THEN** the status SHALL be "red" with reason indicating replica count

#### Scenario: RED when health endpoint unreachable
- **WHEN** the deployment has ready replicas but the health HTTP probe fails
- **THEN** the status SHALL be "red" with reason indicating the endpoint is unreachable

#### Scenario: AMBER when last execution failed
- **WHEN** the health probe response body contains "last_status":"failed"
- **THEN** the status SHALL be "amber" with reason "last execution failed"

#### Scenario: AMBER when execution in flight
- **WHEN** the health probe response body contains "in_flight":true
- **THEN** the status SHALL be "amber" with reason "execution in flight"

#### Scenario: GREEN when healthy
- **WHEN** the deployment has ready replicas and the health probe succeeds with no AMBER signals
- **THEN** the status SHALL be "green" with empty reason

### Requirement: wf_health detail mode
The wf_health tool SHALL support an optional detail parameter. When detail is true, the health probe SHALL request /health?detail=1 and include the telemetry snapshot in the result.

#### Scenario: Detail mode returns telemetry
- **WHEN** wf_health is called with detail=true
- **THEN** the result SHALL include the full health endpoint response in the detail field

#### Scenario: Non-detail mode omits telemetry
- **WHEN** wf_health is called with detail=false or omitted
- **THEN** the result SHALL have an empty detail field

### Requirement: wf_health HTTP probe
The tool SHALL probe the workflow health endpoint at http://<name>.<namespace>.svc.cluster.local:8080/health with a 5-second timeout.

#### Scenario: Successful probe
- **WHEN** the health endpoint returns HTTP 2xx
- **THEN** the response body SHALL be returned as a string for classification

#### Scenario: Non-2xx response
- **WHEN** the health endpoint returns a non-2xx status code
- **THEN** the probe SHALL return an error
