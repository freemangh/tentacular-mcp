//go:build integration

package integration_test

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/randybias/tentacular-mcp/pkg/k8s"
)

func TestIntegration_IssueToken(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	nsName := "tnt-int-cred-token"

	t.Cleanup(func() {
		_ = k8s.DeleteNamespace(context.Background(), client, nsName)
	})

	if err := k8s.CreateNamespace(ctx, client, nsName); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	if err := k8s.CreateWorkflowServiceAccount(ctx, client, nsName); err != nil {
		t.Fatalf("CreateWorkflowServiceAccount: %v", err)
	}

	token, err := k8s.IssueToken(ctx, client, nsName, 10)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestIntegration_TokenAuthentication(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	nsName := "tnt-int-cred-auth"

	t.Cleanup(func() {
		_ = k8s.DeleteNamespace(context.Background(), client, nsName)
	})

	if err := k8s.CreateNamespace(ctx, client, nsName); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	if err := k8s.CreateWorkflowServiceAccount(ctx, client, nsName); err != nil {
		t.Fatalf("CreateWorkflowServiceAccount: %v", err)
	}
	if err := k8s.CreateWorkflowRole(ctx, client, nsName); err != nil {
		t.Fatalf("CreateWorkflowRole: %v", err)
	}
	if err := k8s.CreateWorkflowRoleBinding(ctx, client, nsName); err != nil {
		t.Fatalf("CreateWorkflowRoleBinding: %v", err)
	}

	token, err := k8s.IssueToken(ctx, client, nsName, 10)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}

	// Create a new K8s client using the issued token.
	tokenCfg := rest.CopyConfig(client.Config)
	tokenCfg.BearerToken = token
	// Clear cert-based auth so we only use the token.
	tokenCfg.CertData = nil
	tokenCfg.CertFile = ""
	tokenCfg.KeyData = nil
	tokenCfg.KeyFile = ""

	tokenCS, err := kubernetes.NewForConfig(tokenCfg)
	if err != nil {
		t.Fatalf("NewForConfig with token: %v", err)
	}

	// The token-based client should be able to list pods in the target namespace.
	pods, err := tokenCS.CoreV1().Pods(nsName).List(ctx, metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list pods with token client: %v", err)
	}
	// No pods expected yet, but the call should succeed (not 403).
	_ = pods
}

func TestIntegration_GenerateKubeconfig(t *testing.T) {
	kc, err := k8s.GenerateKubeconfig(
		"https://127.0.0.1:6443",
		"test-ca-cert",
		"test-token-value",
		"test-namespace",
	)
	if err != nil {
		t.Fatalf("GenerateKubeconfig: %v", err)
	}

	if !strings.Contains(kc, "apiVersion: v1") {
		t.Error("kubeconfig missing apiVersion")
	}
	if !strings.Contains(kc, "kind: Config") {
		t.Error("kubeconfig missing kind")
	}
	if !strings.Contains(kc, "https://127.0.0.1:6443") {
		t.Error("kubeconfig missing cluster URL")
	}
	if !strings.Contains(kc, "test-namespace") {
		t.Error("kubeconfig missing namespace")
	}
	if !strings.Contains(kc, "test-token-value") {
		t.Error("kubeconfig missing token")
	}
}

func TestIntegration_RotateServiceAccount(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	nsName := "tnt-int-cred-rotate"

	t.Cleanup(func() {
		_ = k8s.DeleteNamespace(context.Background(), client, nsName)
	})

	if err := k8s.CreateNamespace(ctx, client, nsName); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	if err := k8s.CreateWorkflowServiceAccount(ctx, client, nsName); err != nil {
		t.Fatalf("CreateWorkflowServiceAccount: %v", err)
	}

	// Issue a token before rotation.
	_, err := k8s.IssueToken(ctx, client, nsName, 10)
	if err != nil {
		t.Fatalf("IssueToken before rotate: %v", err)
	}

	// Rotate (recreate) the service account.
	if err := k8s.RecreateWorkflowServiceAccount(ctx, client, nsName); err != nil {
		t.Fatalf("RecreateWorkflowServiceAccount: %v", err)
	}

	// Verify the new SA exists.
	sa, err := client.Clientset.CoreV1().ServiceAccounts(nsName).Get(ctx, "tentacular-workflow", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get SA after rotate: %v", err)
	}
	if sa.Name != "tentacular-workflow" {
		t.Errorf("unexpected SA name: %s", sa.Name)
	}
}
