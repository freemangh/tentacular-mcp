## Why

The MCP server has no way to check workflow runtime health. Operators using the Claude skill need to know if a deployed workflow is healthy, degraded, or down without SSHing into the cluster or reading raw K8s API output. Adding wf_health and wf_health_ns tools enables Green/Amber/Red health classification via direct HTTP probes to workflow engine /health?detail=1 endpoints.

## What Changes

- Add `wf_health` MCP tool: checks a single workflow's health via deployment status + HTTP health probe, returns G/A/R classification with optional execution telemetry
- Add `wf_health_ns` MCP tool: aggregates health across all tentacular deployments in a namespace with configurable limit (default 20), returns G/A/R summary counts
- Wire `registerWfHealthTools` into `RegisterAll` in `register.go`
- Implement G/A/R classification logic: RED = pod not ready or health endpoint unreachable, AMBER = last execution failed or execution in flight, GREEN = default healthy
- Use direct HTTP probes to `http://<name>.<namespace>.svc.cluster.local:8080/health` (enabled by engine NetworkPolicy ingress rule from tentacular-system)

## Capabilities

### New Capabilities
- `wf-health`: Single workflow G/A/R health check via wf_health tool with deployment status and HTTP health probe
- `wf-health-ns`: Namespace-wide workflow health aggregation via wf_health_ns tool with G/A/R summary

### Modified Capabilities
<!-- None -->

## Impact

- `pkg/tools/wf_health.go`: NEW -- wf_health and wf_health_ns tool handlers, health probe, G/A/R classification
- `pkg/tools/wf_health_test.go`: NEW -- unit tests for classification logic and tool handlers
- `pkg/tools/register.go`: Add `registerWfHealthTools(srv, client)` call
- `pkg/k8s/client.go`: May need k8s label constants for managed-by filtering
