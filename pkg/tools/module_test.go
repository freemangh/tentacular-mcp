package tools

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/randybias/tentacular-mcp/pkg/k8s"
)

// moduleGVRs maps resource name to list kind for the fake dynamic client.
var moduleGVRs = map[schema.GroupVersionResource]string{
	{Group: "apps", Version: "v1", Resource: "deployments"}:                         "DeploymentList",
	{Group: "", Version: "v1", Resource: "services"}:                                "ServiceList",
	{Group: "", Version: "v1", Resource: "configmaps"}:                              "ConfigMapList",
	{Group: "", Version: "v1", Resource: "secrets"}:                                 "SecretList",
	{Group: "batch", Version: "v1", Resource: "jobs"}:                              "JobList",
	{Group: "batch", Version: "v1", Resource: "cronjobs"}:                          "CronJobList",
	{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"}:       "NetworkPolicyList",
}

func moduleScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)
	return scheme
}

func newModuleTestClient() *k8s.Client {
	scheme := moduleScheme()
	dynClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, moduleGVRs)
	staticClient := kubefake.NewSimpleClientset()

	return &k8s.Client{
		Clientset: staticClient,
		Dynamic:   dynClient,
		Config:    &rest.Config{Host: "https://test-cluster:6443"},
	}
}

func newManagedNsClient() *k8s.Client {
	scheme := moduleScheme()
	dynClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, moduleGVRs)
	staticClient := kubefake.NewSimpleClientset()

	return &k8s.Client{
		Clientset: staticClient,
		Dynamic:   dynClient,
		Config:    &rest.Config{Host: "https://test-cluster:6443"},
	}
}

// TestModuleRemoveEmptyRelease verifies module_remove returns 0 deleted for non-existent release.
func TestModuleRemoveEmptyRelease(t *testing.T) {
	client := newModuleTestClient()
	ctx := context.Background()

	result, err := handleModuleRemove(ctx, client, ModuleRemoveParams{
		Namespace: "my-ns",
		Release:   "nonexistent",
	})
	if err != nil {
		t.Fatalf("handleModuleRemove: %v", err)
	}
	if result.Deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", result.Deleted)
	}
}

// TestModuleStatusEmptyRelease verifies module_status returns empty resources for non-existent release.
func TestModuleStatusEmptyRelease(t *testing.T) {
	client := newModuleTestClient()
	ctx := context.Background()

	result, err := handleModuleStatus(ctx, client, ModuleStatusParams{
		Namespace: "my-ns",
		Release:   "nonexistent",
	})
	if err != nil {
		t.Fatalf("handleModuleStatus: %v", err)
	}
	if len(result.Resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(result.Resources))
	}
}

// TestModuleApplyDisallowedKind verifies module_apply rejects manifests with disallowed kinds.
func TestModuleApplyDisallowedKind(t *testing.T) {
	client := newManagedNsClient()
	ctx := context.Background()

	// Create managed namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "managed-ns",
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "tentacular",
			},
		},
	}
	_, err := client.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	disallowedKinds := []string{"ClusterRole", "Namespace", "Node", "PersistentVolume", "ClusterRoleBinding"}
	for _, kind := range disallowedKinds {
		_, err = handleModuleApply(ctx, client, ModuleApplyParams{
			Namespace: "managed-ns",
			Release:   "my-app",
			Manifests: []map[string]interface{}{
				{
					"apiVersion": "v1",
					"kind":       kind,
					"metadata":   map[string]interface{}{"name": "test-resource"},
				},
			},
		})
		if err == nil {
			t.Errorf("expected error for disallowed kind %q, got nil", kind)
		}
	}
}

// TestModuleApplyUnmanagedNamespace verifies module_apply rejects unmanaged namespaces.
func TestModuleApplyUnmanagedNamespace(t *testing.T) {
	client := newManagedNsClient()
	ctx := context.Background()

	// Create unmanaged namespace
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "unmanaged-ns"},
	}
	_, err := client.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err = handleModuleApply(ctx, client, ModuleApplyParams{
		Namespace: "unmanaged-ns",
		Release:   "my-app",
		Manifests: []map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("expected error for unmanaged namespace, got nil")
	}
}
