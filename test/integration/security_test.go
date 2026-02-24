//go:build integration

package integration_test

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	networkingv1 "k8s.io/api/networking/v1"

	"github.com/randybias/tentacular-mcp/pkg/guard"
	"github.com/randybias/tentacular-mcp/pkg/k8s"
)

func TestIntegration_PSALabelsEnforced(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	nsName := "tnt-int-sec-psa"

	t.Cleanup(func() {
		_ = k8s.DeleteNamespace(context.Background(), client, nsName)
	})

	if err := k8s.CreateNamespace(ctx, client, nsName); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}

	ns, err := k8s.GetNamespace(ctx, client, nsName)
	if err != nil {
		t.Fatalf("GetNamespace: %v", err)
	}

	enforce := ns.Labels["pod-security.kubernetes.io/enforce"]
	if enforce != "restricted" {
		t.Errorf("expected PSA enforce=restricted, got %q", enforce)
	}
}

func TestIntegration_DefaultDenyPolicyCreated(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	nsName := "tnt-int-sec-deny"

	t.Cleanup(func() {
		_ = k8s.DeleteNamespace(context.Background(), client, nsName)
	})

	if err := k8s.CreateNamespace(ctx, client, nsName); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	if err := k8s.CreateDefaultDenyPolicy(ctx, client, nsName); err != nil {
		t.Fatalf("CreateDefaultDenyPolicy: %v", err)
	}

	policy, err := client.Clientset.NetworkingV1().NetworkPolicies(nsName).Get(ctx, "default-deny", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get default-deny policy: %v", err)
	}

	hasIngress := false
	hasEgress := false
	for _, pt := range policy.Spec.PolicyTypes {
		if pt == networkingv1.PolicyTypeIngress {
			hasIngress = true
		}
		if pt == networkingv1.PolicyTypeEgress {
			hasEgress = true
		}
	}
	if !hasIngress {
		t.Error("default-deny missing Ingress policy type")
	}
	if !hasEgress {
		t.Error("default-deny missing Egress policy type")
	}

	if len(policy.Spec.Ingress) != 0 {
		t.Errorf("expected empty ingress rules, got %d", len(policy.Spec.Ingress))
	}
	if len(policy.Spec.Egress) != 0 {
		t.Errorf("expected empty egress rules, got %d", len(policy.Spec.Egress))
	}
}

func TestIntegration_DNSAllowPolicyCreated(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	nsName := "tnt-int-sec-dns"

	t.Cleanup(func() {
		_ = k8s.DeleteNamespace(context.Background(), client, nsName)
	})

	if err := k8s.CreateNamespace(ctx, client, nsName); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	if err := k8s.CreateDNSAllowPolicy(ctx, client, nsName); err != nil {
		t.Fatalf("CreateDNSAllowPolicy: %v", err)
	}

	policy, err := client.Clientset.NetworkingV1().NetworkPolicies(nsName).Get(ctx, "allow-dns", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("get allow-dns policy: %v", err)
	}

	if len(policy.Spec.Egress) == 0 {
		t.Fatal("expected egress rules in allow-dns policy")
	}

	rule := policy.Spec.Egress[0]
	if len(rule.Ports) < 2 {
		t.Fatalf("expected at least 2 ports (UDP+TCP), got %d", len(rule.Ports))
	}

	foundUDP := false
	foundTCP := false
	for _, p := range rule.Ports {
		if p.Port != nil && p.Port.IntValue() == 53 {
			if p.Protocol != nil {
				switch string(*p.Protocol) {
				case "UDP":
					foundUDP = true
				case "TCP":
					foundTCP = true
				}
			}
		}
	}
	if !foundUDP {
		t.Error("allow-dns missing UDP port 53")
	}
	if !foundTCP {
		t.Error("allow-dns missing TCP port 53")
	}
}

func TestIntegration_PreflightAllPass(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()
	nsName := "tnt-int-sec-preflight"

	t.Cleanup(func() {
		_ = k8s.DeleteNamespace(context.Background(), client, nsName)
	})

	// Create fully prepared namespace.
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
	if err := k8s.CreateDefaultDenyPolicy(ctx, client, nsName); err != nil {
		t.Fatalf("CreateDefaultDenyPolicy: %v", err)
	}
	if err := k8s.CreateDNSAllowPolicy(ctx, client, nsName); err != nil {
		t.Fatalf("CreateDNSAllowPolicy: %v", err)
	}
	if err := k8s.CreateResourceQuota(ctx, client, nsName, "medium"); err != nil {
		t.Fatalf("CreateResourceQuota: %v", err)
	}
	if err := k8s.CreateLimitRange(ctx, client, nsName); err != nil {
		t.Fatalf("CreateLimitRange: %v", err)
	}

	results, err := k8s.RunPreflightChecks(ctx, client, nsName)
	if err != nil {
		t.Fatalf("RunPreflightChecks: %v", err)
	}

	for _, r := range results {
		if !r.Passed {
			t.Errorf("preflight check %q failed: %s", r.Name, r.Warning)
		}
		// gVisor check passes with a warning on kind (no runsc runtime).
		if r.Name == "gvisor-runtime" && r.Warning != "" {
			t.Logf("gvisor check warning (expected on kind): %s", r.Warning)
		}
	}
}

func TestIntegration_PreflightMissingNamespace(t *testing.T) {
	client := integrationClient(t)
	ctx := context.Background()

	results, err := k8s.RunPreflightChecks(ctx, client, "tnt-int-sec-nonexistent")
	if err != nil {
		t.Fatalf("RunPreflightChecks: %v", err)
	}

	nsCheckFound := false
	for _, r := range results {
		if r.Name == "namespace-exists" {
			nsCheckFound = true
			if r.Passed {
				t.Error("expected namespace-exists check to fail for nonexistent namespace")
			}
		}
	}
	if !nsCheckFound {
		t.Error("namespace-exists check not found in results")
	}
}

func TestIntegration_GuardRejectsTentacularSystem(t *testing.T) {
	err := guard.CheckNamespace("tentacular-system")
	if err == nil {
		t.Error("expected guard to reject tentacular-system namespace")
	}
}
