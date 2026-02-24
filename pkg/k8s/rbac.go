package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	workflowServiceAccount = "tentacular-workflow"
	workflowRoleName       = "tentacular-workflow"
)

// CreateWorkflowServiceAccount creates the tentacular-workflow ServiceAccount
// in the given namespace.
func CreateWorkflowServiceAccount(ctx context.Context, client *Client, namespace string) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workflowServiceAccount,
			Namespace: namespace,
			Labels: map[string]string{
				ManagedByLabel: ManagedByValue,
			},
		},
	}

	_, err := client.Clientset.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("service account %q already exists in namespace %q: %w", workflowServiceAccount, namespace, err)
		}
		return fmt.Errorf("create service account %q in namespace %q: %w", workflowServiceAccount, namespace, err)
	}
	return nil
}

// CreateWorkflowRole creates a Role granting the tentacular workflow permissions
// within the given namespace.
func CreateWorkflowRole(ctx context.Context, client *Client, namespace string) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workflowRoleName,
			Namespace: namespace,
			Labels: map[string]string{
				ManagedByLabel: ManagedByValue,
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments"},
				Verbs:     []string{"create", "update", "delete", "patch", "get", "list", "watch"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"replicasets", "daemonsets", "statefulsets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services", "configmaps", "secrets"},
				Verbs:     []string{"create", "update", "delete", "patch", "get", "list", "watch"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"cronjobs", "jobs"},
				Verbs:     []string{"create", "update", "delete", "patch", "get", "list", "watch"},
			},
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"networkpolicies", "ingresses"},
				Verbs:     []string{"create", "update", "delete", "patch", "get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log", "events"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"serviceaccounts"},
				Verbs:     []string{"get", "list", "watch", "patch", "update"},
			},
		},
	}

	_, err := client.Clientset.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("role %q already exists in namespace %q: %w", workflowRoleName, namespace, err)
		}
		return fmt.Errorf("create role %q in namespace %q: %w", workflowRoleName, namespace, err)
	}
	return nil
}

// CreateWorkflowRoleBinding creates a RoleBinding that binds the workflow Role
// to the workflow ServiceAccount in the given namespace.
func CreateWorkflowRoleBinding(ctx context.Context, client *Client, namespace string) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	rb := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workflowRoleName,
			Namespace: namespace,
			Labels: map[string]string{
				ManagedByLabel: ManagedByValue,
			},
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      workflowServiceAccount,
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     workflowRoleName,
		},
	}

	_, err := client.Clientset.RbacV1().RoleBindings(namespace).Create(ctx, rb, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("rolebinding %q already exists in namespace %q: %w", workflowRoleName, namespace, err)
		}
		return fmt.Errorf("create rolebinding %q in namespace %q: %w", workflowRoleName, namespace, err)
	}
	return nil
}

// DeleteWorkflowServiceAccount deletes the tentacular-workflow ServiceAccount
// from the given namespace. This effectively revokes all tokens issued for it.
func DeleteWorkflowServiceAccount(ctx context.Context, client *Client, namespace string) error {
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	err := client.Clientset.CoreV1().ServiceAccounts(namespace).Delete(ctx, workflowServiceAccount, metav1.DeleteOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("service account %q not found in namespace %q: %w", workflowServiceAccount, namespace, err)
		}
		return fmt.Errorf("delete service account %q in namespace %q: %w", workflowServiceAccount, namespace, err)
	}
	return nil
}

// RecreateWorkflowServiceAccount deletes and recreates the workflow ServiceAccount,
// revoking all existing tokens.
func RecreateWorkflowServiceAccount(ctx context.Context, client *Client, namespace string) error {
	err := client.Clientset.CoreV1().ServiceAccounts(namespace).Delete(
		ctx, workflowServiceAccount, metav1.DeleteOptions{},
	)
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("delete service account %q in namespace %q: %w", workflowServiceAccount, namespace, err)
	}
	return CreateWorkflowServiceAccount(ctx, client, namespace)
}
