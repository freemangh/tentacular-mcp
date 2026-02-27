package tools

import (
	"log/slog"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/randybias/tentacular-mcp/pkg/k8s"
	"github.com/randybias/tentacular-mcp/pkg/proxy"
)

// RegisterAll registers all MCP tools with the given server.
// Logger is reserved for future structured logging in tool handlers.
func RegisterAll(srv *mcp.Server, client *k8s.Client, reconciler *proxy.Reconciler, logger *slog.Logger) {
	_ = logger // reserved for future use
	registerNamespaceTools(srv, client)
	registerCredentialTools(srv, client)
	registerWorkflowTools(srv, client)
	registerRunTools(srv, client)
	registerDiscoverTools(srv, client)
	registerClusterOpsTools(srv, client)
	registerGVisorTools(srv, client)
	registerDeployTools(srv, client)
	registerHealthTools(srv, client)
	registerAuditTools(srv, client)
	registerProxyTools(srv, reconciler)
}
