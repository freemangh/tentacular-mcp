## Context

The MCP server runs in the tentacular-system namespace and communicates with workflow pods via K8s Services. Each workflow engine pod exposes GET /health (basic) and GET /health?detail=1 (telemetry snapshot). The engine NetworkPolicy now has an ingress rule allowing TCP 8080 from tentacular-system, enabling direct HTTP probes from the MCP server.

## Goals / Non-Goals

**Goals:**
- wf_health tool checks a single workflow: K8s deployment status + HTTP health probe + G/A/R classification
- wf_health_ns tool scans a namespace: lists tentacular-managed deployments, probes each, returns G/A/R summary
- Classification is deterministic: RED (pod down or health unreachable), AMBER (last execution failed or in-flight), GREEN (healthy)
- Namespace guard validation on all inputs

**Non-Goals:**
- Historical health data or trend analysis
- Push-based health notifications or alerting
- Custom health check thresholds or user-defined classification rules
- Health checks for non-tentacular deployments

## Decisions

### Direct HTTP probe via K8s Service DNS
Probe URL: http://<name>.<namespace>.svc.cluster.local:8080/health. This uses standard K8s DNS resolution and requires no extra service discovery. Alternative: exec into pod and curl -- rejected because it requires pod exec permissions and is slower.

### G/A/R classification via substring matching
The classifyFromDetail function uses simple substring matching on the JSON response body (e.g., checking for "last_status":"failed"). This avoids JSON parsing overhead for a simple health check. Alternative: full JSON unmarshal -- rejected because the response format may evolve and substring matching is more resilient to field additions.

### Namespace-wide scan with configurable limit
wf_health_ns uses a label selector (app.kubernetes.io/managed-by=tentacular) to find deployments, with a default limit of 20. This prevents runaway scans in namespaces with many deployments. The tool reports truncation and total count.

### Registration in RegisterAll
registerWfHealthTools is wired into the existing RegisterAll function pattern, consistent with all other tool registrations.

## Risks / Trade-offs

- **Health probe timeout**: 5-second timeout per probe. In wf_health_ns with 20 workflows, worst case is 100 seconds if all time out sequentially. Mitigation: most probes complete in <100ms; serial probing is acceptable for v1.
- **Substring matching fragility**: If the engine changes its JSON field names, classification breaks silently (defaults to GREEN). Mitigation: the field names are part of the engine's telemetry spec and tested.
- **NetworkPolicy dependency**: Health probes require the engine's MCP ingress rule. If the engine is deployed without it, probes fail and workflows show RED. Mitigation: the ingress rule is unconditional in the engine -- all workflows get it.
