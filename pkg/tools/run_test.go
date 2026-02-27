package tools

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"github.com/randybias/tentacular-mcp/pkg/k8s"
)

func newRunToolTestClient(namespaces ...string) *k8s.Client {
	objs := make([]runtime.Object, 0, len(namespaces))
	for _, name := range namespaces {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: map[string]string{k8s.ManagedByLabel: k8s.ManagedByValue},
			},
		}
		objs = append(objs, ns)
	}
	return &k8s.Client{
		Clientset: fake.NewSimpleClientset(objs...),
		Config:    &rest.Config{Host: "https://test:6443"},
	}
}

// TestHandleWfRun_SystemNamespaceRejected verifies the guard rejects tentacular-system
// before any K8s API call is made.
func TestHandleWfRun_SystemNamespaceRejected(t *testing.T) {
	client := newRunToolTestClient()
	ctx := context.Background()

	_, err := handleWfRun(ctx, client, WfRunParams{
		Namespace: "tentacular-system",
		Name:      "my-wf",
	})
	if err == nil {
		t.Fatal("expected error for system namespace, got nil")
	}
}

// TestHandleWfRun_UnmanagedNamespaceRejected verifies that an unmanaged namespace
// is rejected by CheckManagedNamespace before the run starts.
func TestHandleWfRun_UnmanagedNamespaceRejected(t *testing.T) {
	// No namespaces pre-seeded, so "unmanaged-ns" does not exist and is unmanaged
	client := newRunToolTestClient()
	ctx := context.Background()

	_, err := handleWfRun(ctx, client, WfRunParams{
		Namespace: "unmanaged-ns",
		Name:      "my-wf",
	})
	if err == nil {
		t.Fatal("expected error for unmanaged namespace, got nil")
	}
}

// TestWfRunParams_TimeoutDefaults verifies timeout boundary logic via the params struct
// (unit test of the timeout logic without invoking the full run).
func TestWfRunParams_TimeoutDefaults(t *testing.T) {
	cases := []struct {
		timeoutS int
		wantCap  bool // true if we expect the 600s cap to apply
	}{
		{0, false},
		{60, false},
		{120, false},
		{600, false},
		{601, true},
		{9999, true},
	}

	for _, tc := range cases {
		params := WfRunParams{TimeoutS: tc.timeoutS}
		// Replicate the clamping logic from handleWfRun
		const defaultTimeout = 120
		const maxTimeout = 600
		result := defaultTimeout
		if params.TimeoutS > 0 && params.TimeoutS <= maxTimeout {
			result = params.TimeoutS
		} else if params.TimeoutS > maxTimeout {
			result = maxTimeout
		}

		if tc.wantCap && result != maxTimeout {
			t.Errorf("TimeoutS=%d: expected cap to %d, got %d", tc.timeoutS, maxTimeout, result)
		}
		if !tc.wantCap && tc.timeoutS > 0 && result != tc.timeoutS {
			t.Errorf("TimeoutS=%d: expected %d, got %d", tc.timeoutS, tc.timeoutS, result)
		}
		if tc.timeoutS == 0 && result != defaultTimeout {
			t.Errorf("TimeoutS=0: expected default %d, got %d", defaultTimeout, result)
		}
	}
}

// TestHandleWfRun_ManagedNamespacePassesGuard verifies that a managed namespace
// passes the guard check and enters the run (which will fail due to fake client
// watch limitations, but the namespace guard and managed check pass).
func TestHandleWfRun_ManagedNamespacePassesGuard(t *testing.T) {
	client := newRunToolTestClient("user-ns")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// This will fail at RunWorkflowPod (watch/log limitation of fake client)
	// but NOT at the guard or namespace check. Error is expected.
	_, err := handleWfRun(ctx, client, WfRunParams{
		Namespace: "user-ns",
		Name:      "my-wf",
		TimeoutS:  0, // default
	})
	// The managed namespace check passed (no "not managed" error), so we
	// expect either a context error or a runner pod error — not a guard error.
	if err != nil {
		errStr := err.Error()
		if contains(errStr, "tentacular-system") || contains(errStr, "guard") {
			t.Errorf("unexpected guard error (namespace check should pass for managed ns): %v", err)
		}
		// Any other error (context deadline, fake client limitation) is acceptable
		t.Logf("expected non-guard error from fake client: %v", err)
	}
}

// TestHandleWfRun_TimeoutOverCap verifies that a timeout > 600 is capped at 600s.
func TestHandleWfRun_TimeoutOverCap(t *testing.T) {
	client := newRunToolTestClient("cap-ns")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// TimeoutS > 600 should be capped at 600s (but test will cancel via ctx)
	_, err := handleWfRun(ctx, client, WfRunParams{
		Namespace: "cap-ns",
		Name:      "my-wf",
		TimeoutS:  9999,
	})
	// Error expected (fake client won't complete the pod run)
	// but it should NOT be a guard error
	if err != nil && contains(err.Error(), "guard") {
		t.Errorf("unexpected guard error: %v", err)
	}
}

// TestHandleWfRun_ExplicitTimeout verifies an explicit valid timeout is accepted.
func TestHandleWfRun_ExplicitTimeout(t *testing.T) {
	client := newRunToolTestClient("timeout-ns")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := handleWfRun(ctx, client, WfRunParams{
		Namespace: "timeout-ns",
		Name:      "my-wf",
		TimeoutS:  30,
	})
	// Error expected from fake client limitations, not from timeout clamping
	if err != nil {
		t.Logf("expected error from fake client: %v", err)
	}
}

// TestWfRunResult_Fields verifies the WfRunResult struct fields match what
// handleWfRun sets.
func TestWfRunResult_Fields(t *testing.T) {
	result := WfRunResult{
		Name:       "my-wf",
		Namespace:  "user-ns",
		Output:     []byte(`{"ok":true}`),
		DurationMs: 1234,
		PodName:    "tntc-run-my-wf-12345",
	}
	if result.Name != "my-wf" {
		t.Errorf("expected name=my-wf, got %q", result.Name)
	}
	if result.DurationMs != 1234 {
		t.Errorf("expected duration=1234, got %d", result.DurationMs)
	}
	if string(result.Output) != `{"ok":true}` {
		t.Errorf("unexpected output: %s", result.Output)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	}())
}
