# tentacular-mcp

An in-cluster MCP (Model Context Protocol) server for Kubernetes namespace lifecycle, credential management, workflow introspection, and cluster operations. Replaces direct kube-api access from developer workstations with a single authenticated HTTP endpoint backed by scoped RBAC.

## Why

Developer workstations holding cluster-wide admin kubeconfig is a security anti-pattern. tentacular-mcp proxies Kubernetes operations through a controlled ServiceAccount so CLI clients (and any MCP-capable client) interact with the cluster over one authenticated endpoint instead of raw kube-api access.

## Architecture

```
+------------------+        +-------------------------------------+
|  tentacular CLI  |        |   tentacular-system namespace       |
|  (or any MCP     | Bearer |                                     |
|   client)        +------->+  tentacular-mcp Deployment          |
|                  |  :8080 |  +-------------------------------+  |
+------------------+  /mcp  |  | auth.Middleware (Bearer token) |  |
                            |  |   |                            |  |
                            |  | server.Handler (MCP SDK)       |  |
                            |  |   |                            |  |
                            |  | pkg/tools/register.go          |  |
                            |  |   |  guard.CheckNamespace()    |  |
                            |  |   |  unmarshal params          |  |
                            |  |   |  call handler              |  |
                            |  |   |  marshal result            |  |
                            |  |   v                            |  |
                            |  | pkg/tools/*.go (25 tools)      |  |
                            |  |   |                            |  |
                            |  |   v                            |  |
                            |  | pkg/k8s/* (K8s client layer)   |  |
                            |  +---+---------------------------+  |
                            |      |                              |
                            +------+------------------------------+
                                   |
                                   v
                            +------+------------------------------+
                            |  Kubernetes API Server              |
                            +-------------------------------------+
```

### Request Flow

1. HTTP request hits `:8080/mcp` with `Authorization: Bearer <token>`
2. `auth.Middleware` validates the token (rejects with 401 if invalid; bypasses for `/healthz`)
3. MCP SDK `StreamableHTTPHandler` parses the message and routes to the registered tool
4. `register.go` wrapper: unmarshal params, run `guard.CheckNamespace()`, call handler, marshal result
5. Handler calls `pkg/k8s` functions using in-cluster `rest.Config`
6. Result returned as MCP `Content` with `type: "text"` containing JSON

## Prerequisites

- Go 1.25+
- A Kubernetes cluster (kind works for local development)
- `kubectl` configured with cluster access
- Docker (for building the container image)
- `openssl` (for generating the auth token)

## Quick Start

### Build

```bash
# Build the binary
make build

# Build the Docker image
make docker-build
```

### Generate Auth Token

```bash
# Generate a random Bearer token
TOKEN=$(openssl rand -hex 32)

# Patch the auth secret manifest
sed -i "s/REPLACE_ME_WITH_GENERATED_TOKEN/${TOKEN}/" deploy/manifests/auth-secret.yaml
```

### Deploy

```bash
kubectl apply -k deploy/manifests/
```

This creates:
- `tentacular-system` namespace
- ServiceAccount, ClusterRole, and ClusterRoleBinding
- Auth Secret (mounted as a volume)
- Deployment (single replica, distroless container, non-root)
- ClusterIP Service on port 8080

### Connect

```bash
# Port-forward to reach the server from outside the cluster
kubectl port-forward -n tentacular-system svc/tentacular-mcp 8080:8080

# Verify the health endpoint
curl http://localhost:8080/healthz

# Send an MCP initialize request (using any MCP client)
# The server listens on /mcp via Streamable HTTP transport
```

## MCP Tools

25 tools organized across 8 functional groups. All namespace-scoped tools enforce a self-protection guard that rejects operations targeting `tentacular-system`.

### Namespace Lifecycle

| Tool | Description |
|------|-------------|
| `ns_create` | Create a managed namespace with PSA labels, default-deny NetworkPolicy, DNS-allow policy, ResourceQuota, LimitRange, and workflow SA/Role/RoleBinding. Accepts `small`, `medium`, or `large` quota presets. |
| `ns_delete` | Delete a managed namespace and all child resources. |
| `ns_get` | Get namespace details including labels, annotations, quota summary, and limit range. |
| `ns_list` | List all tentacular-managed namespaces. |

### Credential Management

| Tool | Description |
|------|-------------|
| `cred_issue_token` | Issue a short-lived ServiceAccount token via the TokenRequest API. TTL configurable from 10 to 1440 minutes. |
| `cred_kubeconfig` | Generate a scoped kubeconfig YAML containing a time-limited token, cluster CA, and API server URL. |
| `cred_rotate` | Rotate credentials by recreating the workflow ServiceAccount, invalidating all prior tokens. |

### Workflow Introspection

| Tool | Description |
|------|-------------|
| `wf_pods` | List pods in a namespace with phase, readiness, restart count, images, and age. |
| `wf_logs` | Tail pod logs (snapshot, not streaming). Supports container selection and line count. |
| `wf_events` | List namespace events with type, reason, message, object reference, and count. |
| `wf_jobs` | List Jobs and CronJobs in a namespace with status, schedule, and duration. |

### Cluster Operations

| Tool | Description |
|------|-------------|
| `cluster_preflight` | Run preflight validation checks (API connectivity, namespace access, RBAC, gVisor availability). |
| `cluster_profile` | Generate a full cluster profile: K8s version, nodes, CNI, storage classes, runtime classes, and extensions. |

### gVisor Sandbox

| Tool | Description |
|------|-------------|
| `gvisor_check` | Check if a gVisor RuntimeClass is available in the cluster. |
| `gvisor_apply` | Apply gVisor annotation to a namespace. |
| `gvisor_verify` | Run a verification pod to confirm gVisor sandbox isolation is functional. |

### Module Proxy

| Tool | Description |
|------|-------------|
| `module_apply` | Apply arbitrary Kubernetes manifests as a labeled release using the dynamic client. Tracks resources by release label for garbage collection. |
| `module_remove` | Remove all resources associated with a release label. |
| `module_status` | Check the status of all resources in a release. |

### Cluster Health

| Tool | Description |
|------|-------------|
| `health_nodes` | Query node readiness, capacity, allocatable resources, and conditions. |
| `health_ns_usage` | Report namespace resource utilization vs. quota (CPU, memory, pod count). |
| `health_cluster_summary` | Overall cluster resource summary: total nodes, pods, CPU and memory capacity/requested. |

### Security Audit

| Tool | Description |
|------|-------------|
| `audit_rbac` | Scan namespace RBAC for over-permissioned roles (wildcard verbs, escalation paths). |
| `audit_netpol` | Verify NetworkPolicy coverage: default-deny presence, policy analysis. |
| `audit_psa` | Validate Pod Security Admission labels against the restricted profile. |

## Authentication

All requests to `/mcp` require a `Authorization: Bearer <token>` header. The `/healthz` endpoint is unauthenticated.

The server loads its expected token from a file path configured via the `TOKEN_PATH` environment variable (default: `/etc/tentacular-mcp/token`). In the standard deployment, this is mounted from the `tentacular-mcp-auth` Kubernetes Secret.

### Generating a Token

```bash
# Generate a 32-byte hex token
openssl rand -hex 32
```

Update the Secret in `deploy/manifests/auth-secret.yaml` with the generated value, then redeploy:

```bash
kubectl apply -k deploy/manifests/
```

### Retrieving a Deployed Token

```bash
kubectl get secret tentacular-mcp-auth -n tentacular-system \
  -o jsonpath='{.data.token}' | base64 -d
```

## Deployment

### Kustomize Deploy

```bash
kubectl apply -k deploy/manifests/
```

The kustomization deploys these resources in order:
1. `tentacular-system` Namespace
2. ServiceAccount + ClusterRole + ClusterRoleBinding
3. Auth Secret
4. Deployment (single replica, distroless non-root image)
5. ClusterIP Service (port 8080)

### Verifying the Deployment

```bash
# Check the pod is running
kubectl get pods -n tentacular-system

# Check logs
kubectl logs -n tentacular-system -l app.kubernetes.io/name=tentacular-mcp

# Port-forward and test
kubectl port-forward -n tentacular-system svc/tentacular-mcp 8080:8080 &
curl http://localhost:8080/healthz
```

### Rollback

```bash
kubectl rollout undo deployment/tentacular-mcp -n tentacular-system
```

Or scale to zero:

```bash
kubectl scale deployment/tentacular-mcp -n tentacular-system --replicas=0
```

No persistent state to clean up -- all state lives in Kubernetes objects.

## Development

### Building

```bash
make build         # Build binary to bin/tentacular-mcp
make docker-build  # Build Docker image
make lint          # Run golangci-lint and go vet
make clean         # Remove build artifacts
```

### Testing

Tests are organized in 4 tiers:

| Tier | Command | Requirements |
|------|---------|-------------|
| Unit | `make test-unit` | No cluster needed |
| Integration | `make test-integration` | kind cluster (auto-provisioned) |
| E2E | `make test-e2e` | Production k0s cluster; set `TENTACULAR_E2E_KUBECONFIG` |
| All | `make test-all` | Runs all tiers sequentially |

```bash
# Unit tests only (default)
make test

# Integration tests (sets up and tears down a kind cluster)
make test-integration

# E2E tests (requires a real cluster)
TENTACULAR_E2E_KUBECONFIG=/path/to/kubeconfig make test-e2e
```

### Project Structure

```
cmd/tentacular-mcp/main.go   Entry point with graceful shutdown
pkg/auth/                     Bearer token middleware
pkg/guard/                    Self-protection namespace guard
pkg/k8s/                      Kubernetes client and operations
pkg/server/                   MCP server setup and HTTP handler
pkg/tools/                    24 MCP tool handlers (one file per group)
deploy/manifests/             Kustomize-based deployment manifests
test/integration/             Integration tests (kind cluster)
test/e2e/                     E2E tests (production cluster)
```

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `LISTEN_ADDR` | `:8080` | Address and port the HTTP server binds to |
| `TOKEN_PATH` | `/etc/tentacular-mcp/token` | File path to the Bearer auth token |

## Security Model

### Pod Security Admission (PSA)

All namespaces created by `ns_create` are labeled with the `restricted` PSA profile:
- `pod-security.kubernetes.io/enforce: restricted`
- `pod-security.kubernetes.io/enforce-version: latest`

### Network Policies

Every created namespace gets:
- A **default-deny** NetworkPolicy blocking all ingress and egress
- A **DNS-allow** NetworkPolicy permitting UDP/TCP egress on port 53 to kube-system/kube-dns

### RBAC Scoping

The server's ClusterRole is scoped to exactly the verbs and resources needed by the 25 tools. It is significantly narrower than `cluster-admin`. Key constraints:
- Read-only access to nodes, storage classes, runtime classes, CRDs
- Create/delete for pods (gVisor verification only)
- Namespaced CRUD for resources managed by tool handlers
- `selfsubjectaccessreviews` for preflight RBAC validation

### Self-Protection

`guard.CheckNamespace()` runs before every namespace-scoped tool. It rejects any operation targeting the `tentacular-system` namespace, preventing the server from modifying its own deployment.

### Container Security

The Deployment runs with:
- `runAsNonRoot: true` (UID 65534)
- `readOnlyRootFilesystem: true`
- `allowPrivilegeEscalation: false`
- All capabilities dropped
- `RuntimeDefault` seccomp profile
- Distroless base image (`gcr.io/distroless/static-debian12:nonroot`)

## CLI Integration

The tentacular CLI (`tntc`) can optionally delegate cluster operations to this MCP server using the `--mcp-url` flag or the `mcp_url` config field. When set, the CLI communicates through the MCP server instead of direct kube-api access.

For full details, see [docs/cli-integration.md](docs/cli-integration.md).

### Port-Forward Access

```bash
kubectl port-forward -n tentacular-system svc/tentacular-mcp 8080:8080 &
tntc cluster check --mcp-url http://localhost:8080
```

### In-Cluster Access

```bash
tntc cluster check --mcp-url http://tentacular-mcp.tentacular-system.svc:8080
```

## Contributing

1. Follow the existing code patterns -- tool handlers are standalone functions that take `*k8s.Client` and return structured results
2. Add new tools in `pkg/tools/` following the one-file-per-group convention
3. Register tools through `pkg/tools/register.go` -- the wrapper handles JSON unmarshaling, guard checks, and MCP protocol concerns
4. Write unit tests alongside your code; add integration tests for K8s interactions
5. Run `make lint` before submitting changes
6. Use conventional commits for all commit messages

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
