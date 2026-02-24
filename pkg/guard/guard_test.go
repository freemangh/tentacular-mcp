package guard_test

import (
	"testing"

	"github.com/randybias/tentacular-mcp/pkg/guard"
)

func TestCheckNamespace_ProtectedRejected(t *testing.T) {
	err := guard.CheckNamespace("tentacular-system")
	if err == nil {
		t.Error("expected error for tentacular-system, got nil")
	}
}

func TestCheckNamespace_OtherNamespacePasses(t *testing.T) {
	cases := []string{
		"default",
		"production",
		"kube-system",
		"my-workflow-ns",
		"tentacular-user",
		"",
	}
	for _, ns := range cases {
		if err := guard.CheckNamespace(ns); err != nil {
			t.Errorf("CheckNamespace(%q) returned unexpected error: %v", ns, err)
		}
	}
}
