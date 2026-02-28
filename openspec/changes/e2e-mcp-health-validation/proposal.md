## Why

The wf_health and wf_health_ns MCP tools, the in-process cron scheduler, and the wf_run tool have unit tests with mocked HTTP probes and K8s clients, but no end-to-end tests validate the full chain: MCP tool call -> K8s deployment lookup -> HTTP health probe -> G/A/R classification. Without E2E validation, the health monitoring workflow that operators depend on could silently break when any component in the chain changes.

## What Changes

- Add E2E test suite for wf_health tool: verify G/A/R classification logic with realistic mock deployment states and HTTP health endpoint responses
- Add E2E test suite for wf_health_ns tool: verify namespace-wide aggregation with multiple deployments in mixed health states
- Add integration tests for the cron scheduler: verify schedule registration, discovery from K8s annotations, and trigger execution via HTTP
- Add integration tests for wf_run tool: verify workflow trigger via HTTP POST and result parsing
- Add test fixtures for mock K8s deployments with tentacular labels and health endpoint responses

## Capabilities

### New Capabilities
- `e2e-wf-health`: End-to-end test suite for wf_health tool covering G/A/R classification with deployment status and HTTP health probe combinations
- `e2e-wf-health-ns`: End-to-end test suite for wf_health_ns tool covering namespace-wide health aggregation with mixed deployment states
- `e2e-cron-scheduler`: Integration test suite for the in-process cron scheduler covering schedule CRUD, annotation discovery, and trigger execution

### Modified Capabilities
<!-- None -->

## Impact

- `pkg/tools/wf_health_test.go`: Extend with E2E tests covering full tool handler flow with realistic mock scenarios
- `pkg/tools/run_test.go`: Extend with integration tests for wf_run HTTP trigger chain
- `pkg/scheduler/scheduler_test.go`: Extend with integration tests for schedule lifecycle and annotation-based discovery
- `pkg/scheduler/discover_test.go`: Extend with tests for deployment annotation scanning
- Test fixtures: Mock K8s deployment objects and HTTP health endpoint responses for various G/A/R states
