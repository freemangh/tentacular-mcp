package tools

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/randybias/tentacular-mcp/pkg/k8s"
)

func newNsTestClient() *k8s.Client {
	return &k8s.Client{
		Clientset: fake.NewSimpleClientset(),
		Config:    &rest.Config{Host: "https://test-cluster:6443"},
	}
}

func TestNsCreateOrchestration(t *testing.T) {
	client := newNsTestClient()
	ctx := context.Background()

	result, err := handleNsCreate(ctx, client, NsCreateParams{
		Name:        "dev-alice",
		QuotaPreset: "small",
	})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.Name != "dev-alice" {
		t.Errorf("name: got %q, want %q", result.Name, "dev-alice")
	}
	if result.QuotaPreset != "small" {
		t.Errorf("quota_preset: got %q, want %q", result.QuotaPreset, "small")
	}

	// Expect 8 resources created
	expectedCount := 8
	if len(result.ResourcesCreated) != expectedCount {
		t.Errorf("resources_created: got %d, want %d: %v", len(result.ResourcesCreated), expectedCount, result.ResourcesCreated)
	}
}

func TestNsDeleteManagedCheck(t *testing.T) {
	client := newNsTestClient()
	ctx := context.Background()

	// Create a managed namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dev-bob",
			Labels: map[string]string{
				k8s.ManagedByLabel: k8s.ManagedByValue,
			},
		},
	}
	_, err := client.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("setup: create ns: %v", err)
	}

	result, err := handleNsDelete(ctx, client, NsDeleteParams{Name: "dev-bob"})
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !result.Deleted {
		t.Error("expected deleted=true")
	}
}

func TestNsDeleteUnmanagedRejects(t *testing.T) {
	client := newNsTestClient()
	ctx := context.Background()

	// Create an unmanaged namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "unmanaged-ns"},
	}
	_, err := client.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("setup: create ns: %v", err)
	}

	_, err = handleNsDelete(ctx, client, NsDeleteParams{Name: "unmanaged-ns"})
	if err == nil {
		t.Fatal("expected error for unmanaged namespace, got nil")
	}
}

func TestNsList(t *testing.T) {
	client := newNsTestClient()
	ctx := context.Background()

	for _, name := range []string{"managed-1", "managed-2"} {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					k8s.ManagedByLabel: k8s.ManagedByValue,
				},
			},
		}
		_, err := client.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
		if err != nil {
			t.Fatalf("setup: create ns %q: %v", name, err)
		}
	}

	// Create one unmanaged
	unmanaged := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "not-managed"},
	}
	_, err := client.Clientset.CoreV1().Namespaces().Create(ctx, unmanaged, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("setup: create unmanaged ns: %v", err)
	}

	result, err := handleNsList(ctx, client)
	if err != nil {
		t.Fatalf("handleNsList: %v", err)
	}

	if len(result.Namespaces) != 2 {
		t.Errorf("expected 2 managed namespaces, got %d", len(result.Namespaces))
	}
}
