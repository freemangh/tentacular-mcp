## Context

The tentacular-mcp server exposes 25 MCP tools grouped by prefix convention: `ns_*` for namespace lifecycle, `cred_*` for credentials, `wf_*` for workflow introspection, `cluster_*` for cluster ops, `gvisor_*` for sandbox, `health_*` for health, and `audit_*` for security audit. The three deploy/release management tools break this pattern by using `module_*` instead of `wf_*`.

The implementation lives in `pkg/tools/module.go` with tests in `pkg/tools/module_test.go`. Registration happens via `registerModuleTools()` called from `register.go`. The e2e tests in `test/integration/e2e_test.go` call these tools by their MCP string names. The internal release-tracking label `tentacular.io/release` is used for Kubernetes label selectors and garbage collection.

## Goals / Non-Goals

**Goals:**
- Rename MCP tool names from `module_*` to `wf_*` for prefix consistency
- Rename the `release` parameter to `name` for clarity (callers pass a deployment name, not a Helm release)
- Rename Go source files and identifiers to match the new naming
- Update all tests, specs, and documentation
- Keep the internal Kubernetes label `tentacular.io/release` unchanged

**Non-Goals:**
- Changing the behavior of apply, remove, or status operations
- Adding new tools or parameters
- Modifying the allowed resource kinds or garbage collection logic
- Renaming the internal `tentacular.io/release` label (it is cluster-internal metadata, not exposed to MCP callers, and renaming it would require a data migration for existing clusters)

## Decisions

### Decision 1: File rename to deploy.go, not workflow.go

**Choice**: Rename `module.go` to `deploy.go` rather than `workflow.go`.

**Rationale**: `pkg/tools/workflow.go` already exists and contains the `wf_pods`, `wf_logs`, `wf_events`, `wf_jobs` handlers. Naming the new file `workflow.go` would collide. `deploy.go` describes the function of these tools (deploying resources) without conflicting with the existing workflow introspection file.

**Alternative considered**: Merge into `workflow.go`. Rejected because that file is already substantial with four tool handlers, and the deploy tools have a distinct concern (resource lifecycle via dynamic client) versus introspection (read-only pod/event/log queries).

### Decision 2: Rename parameter from "release" to "name"

**Choice**: The JSON parameter changes from `"release"` to `"name"` in all three tools.

**Rationale**: "release" is Helm terminology. The parameter identifies a named deployment tracked by label, not a Helm release. `"name"` is simpler and aligns with how `wf_pods`, `wf_jobs` use `"namespace"` + resource identifiers. The internal label key `tentacular.io/release` stays unchanged since it is never exposed to MCP callers.

**Alternative considered**: Keep `"release"` as the parameter name. Rejected because the whole point of this change is to shed Helm terminology.

### Decision 3: Keep internal label tentacular.io/release unchanged

**Choice**: The Kubernetes label `tentacular.io/release` used for resource tracking and garbage collection is not renamed.

**Rationale**: Renaming the label would require a data migration for any existing clusters with deployed resources. The label is purely internal -- MCP callers never see or reference it. The Go field in result structs that exposes this value will change from `Release` to `Name` with JSON tag `"name"`, but the underlying label selector logic stays the same.

### Decision 4: Single atomic change with no deprecation period

**Choice**: Rename all three tools in one change with no backwards-compatible shim.

**Rationale**: tentacular-mcp is pre-1.0 and has no external consumers beyond the development team. A deprecation period adds complexity for zero benefit. Any MCP client configuration referencing `module_*` tool names needs a one-time update.

## Risks / Trade-offs

- **Risk**: Any existing MCP client configs referencing `module_apply`, `module_remove`, or `module_status` will break immediately. -> **Mitigation**: Pre-1.0 project with no external consumers. The README update documents the new names.
- **Risk**: Parameter rename from `"release"` to `"name"` is a silent breaking change if clients pass the old key. -> **Mitigation**: The MCP SDK will fail to unmarshal the `release` field into the `Name` struct field, producing a clear error. No silent data loss.
- **Trade-off**: The Go struct field `Name` in `WorkflowApplyParams` could collide conceptually with the resource `Name` in `WorkflowStatusResult.Resources`. Acceptable because the contexts are unambiguous (params vs. result, deployment name vs. resource name).
