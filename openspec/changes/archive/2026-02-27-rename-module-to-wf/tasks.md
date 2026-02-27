## 1. Rename Source Files

- [x] 1.1 Rename `pkg/tools/module.go` to `pkg/tools/deploy.go` (use `git mv` to preserve history)
- [x] 1.2 Rename `pkg/tools/module_test.go` to `pkg/tools/deploy_test.go` (use `git mv` to preserve history)

## 2. Rename Go Identifiers in deploy.go

- [x] 2.1 Rename `registerModuleTools` to `registerDeployTools`
- [x] 2.2 Rename param structs: `ModuleApplyParams` to `WorkflowApplyParams`, `ModuleRemoveParams` to `WorkflowRemoveParams`, `ModuleStatusParams` to `WorkflowStatusParams`
- [x] 2.3 Rename result structs: `ModuleApplyResult` to `WorkflowApplyResult`, `ModuleRemoveResult` to `WorkflowRemoveResult`, `ModuleStatusResult` to `WorkflowStatusResult`, `ModuleResourceStatus` to `WorkflowResourceStatus`
- [x] 2.4 Rename handler functions: `handleModuleApply` to `handleWorkflowApply`, `handleModuleRemove` to `handleWorkflowRemove`, `handleModuleStatus` to `handleWorkflowStatus`
- [x] 2.5 Rename the `Release` field to `Name` in `WorkflowApplyParams`, `WorkflowRemoveParams`, and `WorkflowStatusParams` structs; update JSON tags from `json:"release"` to `json:"name"`
- [x] 2.6 Rename the `Release` field to `Name` in `WorkflowApplyResult`, `WorkflowRemoveResult`, and `WorkflowStatusResult` structs; update JSON tags from `json:"release"` to `json:"name"`
- [x] 2.7 Update all references to `params.Release` to `params.Name` inside the three handler functions
- [x] 2.8 Update all references to `Release:` in result struct literals to `Name:`
- [x] 2.9 Change MCP tool name strings from `"module_apply"` to `"wf_apply"`, `"module_remove"` to `"wf_remove"`, `"module_status"` to `"wf_status"` in the `registerDeployTools` function
- [x] 2.10 Update MCP tool description strings to replace "release" with "name" and "module" with "workflow" where appropriate (e.g., "Apply a set of Kubernetes manifests as a named deployment" instead of "as a named release")
- [x] 2.11 Update the `allowedKinds` error message in `handleWorkflowApply` to say "workflow manifests" instead of "module manifests"
- [x] 2.12 Update the comment on `allowedKinds` from "module_apply" to "wf_apply"
- [x] 2.13 Keep `releaseLabelKey = "tentacular.io/release"` constant unchanged

## 3. Rename Go Identifiers in deploy_test.go

- [x] 3.1 Rename `moduleGVRs` variable to `deployGVRs` and update all references
- [x] 3.2 Rename `moduleScheme` function to `deployScheme` and update all call sites
- [x] 3.3 Rename `newModuleTestClient` to `newDeployTestClient` and update all call sites
- [x] 3.4 Rename test functions: `TestModuleRemoveEmptyRelease` to `TestWorkflowRemoveEmptyName`, `TestModuleStatusEmptyRelease` to `TestWorkflowStatusEmptyName`, `TestModuleApplyDisallowedKind` to `TestWorkflowApplyDisallowedKind`, `TestModuleApplyUnmanagedNamespace` to `TestWorkflowApplyUnmanagedNamespace`
- [x] 3.5 Update all `handleModule*` calls to `handleWorkflow*` and all `Module*Params` to `Workflow*Params` inside the test functions
- [x] 3.6 Update `Release:` field references to `Name:` in test param struct literals
- [x] 3.7 Update test comments to replace "module_*" references with "wf_*"

## 4. Update register.go

- [x] 4.1 Change `registerModuleTools(srv, client)` call to `registerDeployTools(srv, client)` in `RegisterAll`

## 5. Update E2E Tests

- [x] 5.1 In `test/integration/e2e_test.go`, rename `TestE2E_ModuleApplyStatusRemove` to `TestE2E_WorkflowApplyStatusRemove`
- [x] 5.2 Change all `callTool` name arguments from `"module_apply"` to `"wf_apply"`, `"module_status"` to `"wf_status"`, `"module_remove"` to `"wf_remove"`
- [x] 5.3 Change all `"release"` parameter keys to `"name"` in the `map[string]any` arguments to `callTool`
- [x] 5.4 Update JSON unmarshal struct fields from `Release string` with `json:"release"` to `Name string` with `json:"name"`
- [x] 5.5 Update all assertion references from `applyResult.Release` to `applyResult.Name`, `statusResult.Release` to `statusResult.Name`, `removeResult.Release` to `removeResult.Name`
- [x] 5.6 Update test log messages and error messages to say "wf_apply", "wf_status", "wf_remove" instead of "module_*"
- [x] 5.7 Rename the namespace variable from `"tnt-e2e-module"` to `"tnt-e2e-deploy"` for consistency

## 6. Update README.md

- [x] 6.1 Rename the "Module Proxy" section heading to "Workflow Deploy" (or "Deploy Lifecycle")
- [x] 6.2 Update the tool table: `module_apply` to `wf_apply`, `module_remove` to `wf_remove`, `module_status` to `wf_status`
- [x] 6.3 Update tool descriptions to replace "release" with "name" where it refers to the parameter
- [x] 6.4 Update the tool count if it appears in any summary text (should remain 25 -- no tools added or removed)

## 7. Verify

- [x] 7.1 Run `go build ./...` and confirm zero compilation errors
- [x] 7.2 Run `go test ./pkg/tools/...` and confirm all unit tests pass
- [x] 7.3 Run `go vet ./...` and confirm no warnings
- [x] 7.4 Grep the entire codebase for any remaining references to `module_apply`, `module_remove`, `module_status`, `ModuleApply`, `ModuleRemove`, `ModuleStatus`, `handleModule`, `registerModuleTools` and confirm none remain (except in openspec archive and changelog)
