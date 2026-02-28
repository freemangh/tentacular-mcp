## 1. Test Fixtures

- [ ] 1.1 Create test helper functions for building fake K8s Deployment objects with tentacular managed-by labels and configurable ready replica counts
- [ ] 1.2 Create test helper functions for building httptest.Server instances that return configurable health probe responses (success, amber signals, errors, timeouts)
- [ ] 1.3 Create test helper for mock K8s clientset with pre-loaded deployments for namespace-wide testing

## 2. wf_health E2E Tests

- [ ] 2.1 Add table-driven test for G/A/R classification matrix: GREEN (ready + healthy), RED (not ready), RED (unreachable), AMBER (last failed), AMBER (in-flight)
- [ ] 2.2 Add test for detail=true including raw telemetry snapshot in response
- [ ] 2.3 Add test for detail=false (default) omitting telemetry from response
- [ ] 2.4 Add test for namespace guard rejecting tentacular-system namespace
- [ ] 2.5 Add test for health probe timeout handling (httptest server with deliberate delay)

## 3. wf_health_ns E2E Tests

- [ ] 3.1 Add test for mixed health states: 3 GREEN + 1 AMBER + 1 RED producing correct summary counts
- [ ] 3.2 Add test for all-healthy namespace with correct green count
- [ ] 3.3 Add test for empty namespace returning zero counts and empty workflows list
- [ ] 3.4 Add test for limit parameter capping number of workflows checked
- [ ] 3.5 Add test for label selector filtering excluding non-tentacular deployments

## 4. Cron Scheduler Integration Tests

- [ ] 4.1 Add test for AddSchedule registering a cron entry and verifying entry count
- [ ] 4.2 Add test for RemoveSchedule deregistering an entry
- [ ] 4.3 Add test for duplicate schedule replacing existing entry with new cron expression
- [ ] 4.4 Add test for DiscoverSchedules finding annotated deployments and registering entries
- [ ] 4.5 Add test for DiscoverSchedules skipping deployments without cron annotation
- [ ] 4.6 Add test for DiscoverSchedules removing stale entries when annotation is removed
- [ ] 4.7 Add test for trigger execution sending HTTP POST to workflow /run endpoint

## 5. Verification

- [ ] 5.1 Run `go test ./pkg/tools/...` -- all pass including new E2E tests
- [ ] 5.2 Run `go test ./pkg/scheduler/...` -- all pass including new integration tests
- [ ] 5.3 Run `go vet ./...` -- no issues
