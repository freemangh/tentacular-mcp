package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	sdkauth "github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/modelcontextprotocol/go-sdk/oauthex"
	"github.com/randybias/tentacular-mcp/pkg/auth"
	"github.com/randybias/tentacular-mcp/pkg/k8s"
	"github.com/randybias/tentacular-mcp/pkg/proxy"
	"github.com/randybias/tentacular-mcp/pkg/scheduler"
	"github.com/randybias/tentacular-mcp/pkg/tools"
	"github.com/randybias/tentacular-mcp/pkg/version"
)

// Server wraps the MCP server with K8s client and auth.
type Server struct {
	mcpServer  *mcp.Server
	client     *k8s.Client
	reconciler *proxy.Reconciler
	scheduler  *scheduler.Scheduler
	token      string
	logger     *slog.Logger
}

// New creates a new MCP server with all tools registered.
func New(client *k8s.Client, reconciler *proxy.Reconciler, sched *scheduler.Scheduler, token string, logger *slog.Logger) (*Server, error) {
	mcpServer := mcp.NewServer(
		&mcp.Implementation{
			Name:    "tentacular-mcp",
			Version: version.Version,
		},
		&mcp.ServerOptions{
			Instructions: "In-cluster MCP server for Kubernetes namespace lifecycle, credential management, workflow introspection, cluster operations, and module proxy management.",
			Logger:       logger,
		},
	)

	s := &Server{
		mcpServer:  mcpServer,
		client:     client,
		reconciler: reconciler,
		scheduler:  sched,
		token:      token,
		logger:     logger,
	}

	s.registerTools()

	return s, nil
}

// Handler returns the HTTP handler with auth middleware and health endpoint.
func (s *Server) Handler() http.Handler {
	mcpHandler := mcp.NewStreamableHTTPHandler(
		func(r *http.Request) *mcp.Server {
			return s.mcpServer
		},
		nil,
	)

	mux := http.NewServeMux()
	mux.Handle("/mcp", mcpHandler)
	mux.HandleFunc("/healthz", s.healthHandler)

	// MCP clients (protocol 2025-11-25+) probe OAuth discovery endpoints
	// before connecting. Serve Protected Resource Metadata (RFC 9728) to
	// tell the client that Bearer token auth via the Authorization header
	// is the expected mechanism. Without this, the client enters an OAuth
	// browser flow instead of using the static token from .mcp.json.
	resourceMetadata := sdkauth.ProtectedResourceMetadataHandler(&oauthex.ProtectedResourceMetadata{
		Resource:               "http://localhost:8080/mcp",
		BearerMethodsSupported: []string{"header"},
	})
	mux.Handle("/.well-known/oauth-protected-resource", resourceMetadata)
	mux.Handle("/.well-known/oauth-protected-resource/mcp", resourceMetadata)
	mux.Handle("/mcp/.well-known/oauth-protected-resource", resourceMetadata)

	// Return JSON 404 for other discovery paths (OIDC, authorization server)
	// so Go's default text/plain 404 doesn't cause JSON parse errors.
	jsonNotFound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"not_found"}`))
	})
	mux.Handle("/.well-known/", jsonNotFound)
	mux.Handle("/mcp/.well-known/", jsonNotFound)
	mux.Handle("/register", jsonNotFound)

	return auth.Middleware(s.token, mux)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// registerTools registers all MCP tools by delegating to the tools package.
func (s *Server) registerTools() {
	tools.RegisterAll(s.mcpServer, s.client, s.reconciler, s.scheduler, s.logger)
}
