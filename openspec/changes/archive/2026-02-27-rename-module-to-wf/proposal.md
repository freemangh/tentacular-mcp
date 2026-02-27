## Why

The `module_apply`, `module_remove`, and `module_status` MCP tool names break tentacular's established `wf_` prefix convention used by all other workflow-scoped tools (`wf_pods`, `wf_logs`, `wf_jobs`, `wf_events`). The term "module" also carries Helm baggage that confuses the mental model. Renaming to `wf_apply`, `wf_remove`, `wf_status` unifies the tool namespace and eliminates the Helm association.

## What Changes

- **BREAKING**: Rename MCP tool `module_apply` to `wf_apply`, `module_remove` to `wf_remove`, `module_status` to `wf_status`
- **BREAKING**: Rename the `release` parameter to `name` in all three tools (JSON field changes from `"release"` to `"name"`)
- Rename source file `pkg/tools/module.go` to `pkg/tools/deploy.go` and `pkg/tools/module_test.go` to `pkg/tools/deploy_test.go`
- Rename all Go identifiers from `Module*` prefix to `Workflow*` prefix and `handleModule*` to `handleWorkflow*`
- Rename registration function from `registerModuleTools` to `registerDeployTools` in both `deploy.go` and `register.go`
- Update e2e test tool name strings and parameter keys from `"release"` to `"name"`
- Update README "Module Proxy" section to reflect new tool names
- Internal label `tentacular.io/release` is NOT renamed (it is cluster-internal and not exposed to MCP callers)

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `module-proxy`: Tool names change from `module_*` to `wf_*`; the `release` parameter is renamed to `name` across all three tools. Spec scenarios need updated tool names and parameter keys.

## Impact

- `pkg/tools/module.go` -> `pkg/tools/deploy.go`: All type/function renames
- `pkg/tools/module_test.go` -> `pkg/tools/deploy_test.go`: Test function and helper renames
- `pkg/tools/register.go`: `registerModuleTools` -> `registerDeployTools` call
- `test/integration/e2e_test.go`: Tool name strings and `"release"` -> `"name"` parameter keys
- `README.md`: Module Proxy section renamed and tool table updated
- `openspec/specs/module-proxy/spec.md`: Spec updated to reflect `wf_*` names and `name` parameter
- MCP clients currently calling `module_apply`, `module_remove`, or `module_status` will break and must update their tool calls
