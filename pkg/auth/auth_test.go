package auth_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/randybias/tentacular-mcp/pkg/auth"
)

const testToken = "super-secret-token"

// okHandler is a trivial HTTP handler that returns 200 OK.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
})

func makeRequest(t *testing.T, handler http.Handler, method, path, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func TestMiddleware_ValidToken(t *testing.T) {
	h := auth.Middleware(testToken, okHandler)
	rr := makeRequest(t, h, http.MethodGet, "/some/path", "Bearer "+testToken)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestMiddleware_MissingToken(t *testing.T) {
	h := auth.Middleware(testToken, okHandler)
	rr := makeRequest(t, h, http.MethodGet, "/some/path", "")
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	h := auth.Middleware(testToken, okHandler)
	rr := makeRequest(t, h, http.MethodGet, "/some/path", "Bearer wrong-token")
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestMiddleware_MissingBearerPrefix(t *testing.T) {
	h := auth.Middleware(testToken, okHandler)
	rr := makeRequest(t, h, http.MethodGet, "/some/path", testToken)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing Bearer prefix, got %d", rr.Code)
	}
}

func TestMiddleware_HealthzBypassesAuth(t *testing.T) {
	h := auth.Middleware(testToken, okHandler)
	// No Authorization header on /healthz should still succeed.
	rr := makeRequest(t, h, http.MethodGet, "/healthz", "")
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for /healthz without auth, got %d", rr.Code)
	}
}

func TestLoadToken_Success(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "token")
	if err := os.WriteFile(tokenFile, []byte("  mytoken\n"), 0600); err != nil {
		t.Fatal(err)
	}
	tok, err := auth.LoadToken(tokenFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "mytoken" {
		t.Errorf("expected trimmed token 'mytoken', got %q", tok)
	}
}

func TestLoadToken_FileMissing(t *testing.T) {
	_, err := auth.LoadToken("/nonexistent/path/token")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestLoadToken_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	tokenFile := filepath.Join(dir, "empty-token")
	if err := os.WriteFile(tokenFile, []byte("   \n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err := auth.LoadToken(tokenFile)
	if err == nil {
		t.Error("expected error for empty token file, got nil")
	}
}
