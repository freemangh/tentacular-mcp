package k8s

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
)

func newRunTestClient() *Client {
	return &Client{
		Clientset: fake.NewSimpleClientset(),
		Config:    &rest.Config{Host: "https://test:6443"},
	}
}

func managedNsForRun(name string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{ManagedByLabel: ManagedByValue},
		},
	}
}

// TestRunWorkflowPodCreatesWithCorrectLabels verifies the trigger pod is created
// with all required labels before the watch blocks.
func TestRunWorkflowPodCreatesWithCorrectLabels(t *testing.T) {
	client := newRunTestClient()
	bgCtx := context.Background()

	client.Clientset.CoreV1().Namespaces().Create(bgCtx, managedNsForRun("label-ns"), metav1.CreateOptions{})

	runCtx, runCancel := context.WithTimeout(bgCtx, 5*time.Second)
	defer runCancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		RunWorkflowPod(runCtx, client, "label-ns", "my-workflow", json.RawMessage(`{"x":1}`))
	}()

	// Poll for the pod to appear in the fake client
	var foundPod *corev1.Pod
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		pods, _ := client.Clientset.CoreV1().Pods("label-ns").List(bgCtx, metav1.ListOptions{})
		if len(pods.Items) > 0 {
			foundPod = &pods.Items[0]
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	runCancel()
	<-done

	if foundPod == nil {
		t.Fatal("trigger pod was not created in the fake client")
	}

	if foundPod.Labels[ManagedByLabel] != ManagedByValue {
		t.Errorf("missing managed-by label: %v", foundPod.Labels)
	}
	if foundPod.Labels["tentacular/run-target"] != "my-workflow" {
		t.Errorf("missing run-target label: got %q", foundPod.Labels["tentacular/run-target"])
	}
	if foundPod.Labels["tentacular.dev/role"] != "trigger" {
		t.Errorf("missing role label: got %q", foundPod.Labels["tentacular.dev/role"])
	}
	if foundPod.Spec.RestartPolicy != corev1.RestartPolicyNever {
		t.Errorf("expected RestartPolicyNever, got %q", foundPod.Spec.RestartPolicy)
	}
}

// TestRunWorkflowPodSecurityContext verifies that the trigger pod has security context set.
func TestRunWorkflowPodSecurityContext(t *testing.T) {
	client := newRunTestClient()
	bgCtx := context.Background()

	client.Clientset.CoreV1().Namespaces().Create(bgCtx, managedNsForRun("sec-ns"), metav1.CreateOptions{})

	runCtx, runCancel := context.WithTimeout(bgCtx, 5*time.Second)
	defer runCancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		RunWorkflowPod(runCtx, client, "sec-ns", "wf", nil)
	}()

	var foundPod *corev1.Pod
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		pods, _ := client.Clientset.CoreV1().Pods("sec-ns").List(bgCtx, metav1.ListOptions{})
		if len(pods.Items) > 0 {
			foundPod = &pods.Items[0]
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	runCancel()
	<-done

	if foundPod == nil {
		t.Fatal("trigger pod was not created in the fake client")
	}

	if foundPod.Spec.SecurityContext == nil {
		t.Error("pod security context should not be nil")
	}
	if len(foundPod.Spec.Containers) == 0 {
		t.Fatal("expected at least one container")
	}
	csc := foundPod.Spec.Containers[0].SecurityContext
	if csc == nil {
		t.Error("container security context should not be nil")
		return
	}
	if csc.AllowPrivilegeEscalation == nil || *csc.AllowPrivilegeEscalation {
		t.Error("AllowPrivilegeEscalation should be false")
	}
	if csc.RunAsNonRoot == nil || !*csc.RunAsNonRoot {
		t.Error("RunAsNonRoot should be true")
	}
}

// TestRunWorkflowPodContextCancellation verifies that cancelling the context
// causes RunWorkflowPod to return promptly.
func TestRunWorkflowPodContextCancellation(t *testing.T) {
	client := newRunTestClient()
	bgCtx := context.Background()

	client.Clientset.CoreV1().Namespaces().Create(bgCtx, managedNsForRun("cancel-ns"), metav1.CreateOptions{})

	runCtx, runCancel := context.WithTimeout(bgCtx, 5*time.Second)

	done := make(chan error, 1)
	go func() {
		_, _, err := RunWorkflowPod(runCtx, client, "cancel-ns", "wf", nil)
		done <- err
	}()

	// Cancel context after the pod is created
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		pods, _ := client.Clientset.CoreV1().Pods("cancel-ns").List(bgCtx, metav1.ListOptions{})
		if len(pods.Items) > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	runCancel()

	select {
	case <-done:
		// returned promptly -- good
	case <-time.After(3 * time.Second):
		t.Error("RunWorkflowPod did not return after context cancellation")
	}
}
