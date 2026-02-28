## 1. Tool Implementation

- [ ] 1.1 Create `pkg/tools/wf_health.go` with WfHealthParams, WfHealthResult, WfHealthNsParams, WfHealthNsResult types
- [ ] 1.2 Implement `handleWfHealth` -- get deployment, check ready replicas, probe health endpoint, classify G/A/R
- [ ] 1.3 Implement `handleWfHealthNs` -- list managed deployments with label selector, probe each, aggregate G/A/R summary
- [ ] 1.4 Implement `probeHealthEndpoint` -- HTTP GET with 5s timeout, detail query param support
- [ ] 1.5 Implement `classifyFromDetail` -- substring matching for AMBER signals (last_status failed, in_flight true)

## 2. Tool Registration

- [ ] 2.1 Add `registerWfHealthTools(srv, client)` function in `pkg/tools/wf_health.go`
- [ ] 2.2 Wire `registerWfHealthTools` into `RegisterAll` in `pkg/tools/register.go`

## 3. Tests

- [ ] 3.1 Add unit tests for `classifyFromDetail` in `pkg/tools/wf_health_test.go` -- GREEN, AMBER (failed), AMBER (in-flight)
- [ ] 3.2 Add integration tests for `handleWfHealth` with mock K8s client and HTTP server
- [ ] 3.3 Add integration tests for `handleWfHealthNs` with multiple mock deployments

## 4. Verification

- [ ] 4.1 Run `go test ./pkg/tools/...` -- all pass
- [ ] 4.2 Run `go vet ./...` -- no issues
- [ ] 4.3 Verify wf_health and wf_health_ns appear in MCP tool list
