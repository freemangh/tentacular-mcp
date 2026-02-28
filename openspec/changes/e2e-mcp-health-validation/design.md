## Context

The MCP server provides wf_health (single workflow) and wf_health_ns (namespace-wide) tools that classify workflow health as Green/Amber/Red. Classification depends on K8s deployment readiness and HTTP health probe results from the engine's /health?detail=1 endpoint. The in-process cron scheduler discovers schedules from deployment annotations and triggers workflows via HTTP. All these components have unit tests with mocked dependencies, but no tests exercise the full tool handler -> K8s client -> HTTP probe chain.

## Goals / Non-Goals

**Goals:**
- Validate wf_health G/A/R classification across all state combinations (pod ready/not ready, health probe success/failure/amber signals)
- Validate wf_health_ns aggregation with mixed deployment states
- Validate cron scheduler lifecycle: register, discover from annotations, trigger via HTTP, remove
- Provide realistic test fixtures for K8s deployment objects with tentacular labels

**Non-Goals:**
- Live cluster testing (all K8s interactions use mock clients)
- Testing the engine's telemetry sink (that belongs to the engine repo)
- Load testing the health probe with many concurrent workflows
- Testing actual cron timing (use manual trigger for determinism)

## Decisions

### 1. Mock strategy: interface-based fakes vs. httptest servers

**Decision**: Use Go httptest.Server for health endpoint mocking and fake K8s clientset for deployment queries. The existing wfHealthProbe package variable enables test injection.

**Rationale**: httptest.Server validates real HTTP client behavior (timeouts, status codes, response parsing). Fake K8s clientset is the standard Go pattern for K8s unit testing.

### 2. Cron scheduler tests use manual trigger, not real time

**Decision**: Test the scheduler's AddSchedule/RemoveSchedule/DiscoverSchedules methods directly, and verify trigger execution by calling the trigger function manually rather than waiting for cron ticks.

**Rationale**: Time-dependent tests are flaky. The cron library itself is well-tested; we only need to verify our integration with it.

### 3. G/A/R classification test matrix

**Decision**: Use table-driven tests with a matrix of (deployment state, health probe response) combinations covering all G/A/R paths.

**Rationale**: Table-driven tests make it clear which state combinations map to which classification. Easy to add new test cases as classification logic evolves.

## Risks / Trade-offs

- [Mock K8s client may drift from real API behavior] -> Use official fake clientset from k8s.io/client-go/kubernetes/fake which tracks real API semantics
- [Health probe timeout behavior hard to test] -> Use httptest.Server with deliberate delay + short test timeout to verify timeout handling
- [Cron scheduler discovery depends on label selectors] -> Test with deployments that have and lack the tentacular managed-by label to verify filtering
