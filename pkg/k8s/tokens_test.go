package k8s_test

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/randybias/tentacular-mcp/pkg/k8s"
)

func TestGenerateKubeconfig_ContainsRequiredFields(t *testing.T) {
	kubeconfig, err := k8s.GenerateKubeconfig(
		"https://cluster.example.com:6443",
		"fake-ca-cert-data",
		"my-bearer-token",
		"my-namespace",
	)
	if err != nil {
		t.Fatalf("GenerateKubeconfig: %v", err)
	}

	checks := []string{
		"apiVersion: v1",
		"kind: Config",
		"https://cluster.example.com:6443",
		"my-bearer-token",
		"my-namespace",
		"tentacular-workflow",
		"tentacular",
	}
	for _, want := range checks {
		if !strings.Contains(kubeconfig, want) {
			t.Errorf("kubeconfig missing expected content %q", want)
		}
	}
}

func TestGenerateKubeconfig_CADataIsBase64(t *testing.T) {
	caRaw := "my-ca-cert"
	kubeconfig, err := k8s.GenerateKubeconfig(
		"https://cluster.example.com:6443",
		caRaw,
		"token",
		"ns",
	)
	if err != nil {
		t.Fatalf("GenerateKubeconfig: %v", err)
	}

	expected := base64.StdEncoding.EncodeToString([]byte(caRaw))
	if !strings.Contains(kubeconfig, expected) {
		t.Errorf("expected base64-encoded CA %q in kubeconfig, got:\n%s", expected, kubeconfig)
	}
}

func TestGenerateKubeconfig_HasCurrentContext(t *testing.T) {
	kubeconfig, err := k8s.GenerateKubeconfig("https://s", "ca", "tok", "ns")
	if err != nil {
		t.Fatalf("GenerateKubeconfig: %v", err)
	}
	if !strings.Contains(kubeconfig, "current-context: tentacular") {
		t.Error("kubeconfig missing current-context field")
	}
}
