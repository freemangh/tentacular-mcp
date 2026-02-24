package guard

import "fmt"

const protectedNamespace = "tentacular-system"

// CheckNamespace returns an error if the given namespace is the protected
// tentacular-system namespace. All tool handlers must call this before
// performing operations.
func CheckNamespace(namespace string) error {
	if namespace == protectedNamespace {
		return fmt.Errorf("operations on namespace %q are not allowed", protectedNamespace)
	}
	return nil
}
