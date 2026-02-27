package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/randybias/tentacular-mcp/pkg/guard"
	"github.com/randybias/tentacular-mcp/pkg/k8s"
)

const releaseLabelKey = "tentacular.io/release"

// allowedKinds is the set of Kubernetes resource kinds module_apply will
// accept. Cluster-scoped and sensitive kinds are not permitted.
var allowedKinds = map[string]bool{
	"Deployment":    true,
	"Service":       true,
	"PersistentVolumeClaim": true,
	"NetworkPolicy": true,
	"ConfigMap":     true,
	"Secret":        true,
	"Job":           true,
	"CronJob":       true,
	"Ingress":       true,
}

// ModuleApplyParams are the parameters for module_apply.
type ModuleApplyParams struct {
	Namespace string                   `json:"namespace" jsonschema:"Target namespace for the module"`
	Release   string                   `json:"release" jsonschema:"Release name for tracking resources"`
	Manifests []map[string]interface{} `json:"manifests" jsonschema:"List of Kubernetes manifest objects to apply"`
}

// ModuleApplyResult is the result of module_apply.
type ModuleApplyResult struct {
	Release   string `json:"release"`
	Namespace string `json:"namespace"`
	Created   int    `json:"created"`
	Updated   int    `json:"updated"`
	Deleted   int    `json:"deleted"`
}

// ModuleRemoveParams are the parameters for module_remove.
type ModuleRemoveParams struct {
	Namespace string `json:"namespace" jsonschema:"Namespace containing the module resources"`
	Release   string `json:"release" jsonschema:"Release name to remove"`
}

// ModuleRemoveResult is the result of module_remove.
type ModuleRemoveResult struct {
	Release   string `json:"release"`
	Namespace string `json:"namespace"`
	Deleted   int    `json:"deleted"`
}

// ModuleStatusParams are the parameters for module_status.
type ModuleStatusParams struct {
	Namespace string `json:"namespace" jsonschema:"Namespace containing the module resources"`
	Release   string `json:"release" jsonschema:"Release name to check status for"`
}

// ModuleResourceStatus is the status of a single resource in a module.
type ModuleResourceStatus struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Ready  bool   `json:"ready"`
	Reason string `json:"reason,omitempty"`
}

// ModuleStatusResult is the result of module_status.
type ModuleStatusResult struct {
	Release   string                 `json:"release"`
	Namespace string                 `json:"namespace"`
	Resources []ModuleResourceStatus `json:"resources"`
}

func registerModuleTools(srv *mcp.Server, client *k8s.Client) {
	mcp.AddTool(srv, &mcp.Tool{
		Name:        "module_apply",
		Description: "Apply a set of Kubernetes manifests as a named release in a namespace. Uses release labels for tracking and garbage collection.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params ModuleApplyParams) (*mcp.CallToolResult, ModuleApplyResult, error) {
		if err := guard.CheckNamespace(params.Namespace); err != nil {
			return nil, ModuleApplyResult{}, err
		}
		result, err := handleModuleApply(ctx, client, params)
		return nil, result, err
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "module_remove",
		Description: "Remove all resources belonging to a named release in a namespace.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params ModuleRemoveParams) (*mcp.CallToolResult, ModuleRemoveResult, error) {
		if err := guard.CheckNamespace(params.Namespace); err != nil {
			return nil, ModuleRemoveResult{}, err
		}
		result, err := handleModuleRemove(ctx, client, params)
		return nil, result, err
	})

	mcp.AddTool(srv, &mcp.Tool{
		Name:        "module_status",
		Description: "Get status of all resources belonging to a named release in a namespace.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, params ModuleStatusParams) (*mcp.CallToolResult, ModuleStatusResult, error) {
		if err := guard.CheckNamespace(params.Namespace); err != nil {
			return nil, ModuleStatusResult{}, err
		}
		result, err := handleModuleStatus(ctx, client, params)
		return nil, result, err
	})
}

// resolveGVR derives the GroupVersionResource from apiVersion and kind using the discovery client.
func resolveGVR(ctx context.Context, client *k8s.Client, apiVersion, kind string) (schema.GroupVersionResource, error) {
	_, resourceLists, err := client.Clientset.Discovery().ServerGroupsAndResources()
	if err != nil && resourceLists == nil {
		return schema.GroupVersionResource{}, fmt.Errorf("discovery failed: %w", err)
	}

	for _, rl := range resourceLists {
		if rl.GroupVersion != apiVersion {
			continue
		}
		for _, r := range rl.APIResources {
			if r.Kind == kind {
				gv, err := schema.ParseGroupVersion(apiVersion)
				if err != nil {
					return schema.GroupVersionResource{}, fmt.Errorf("parse group version %q: %w", apiVersion, err)
				}
				return schema.GroupVersionResource{
					Group:    gv.Group,
					Version:  gv.Version,
					Resource: r.Name,
				}, nil
			}
		}
	}

	return schema.GroupVersionResource{}, fmt.Errorf("no resource found for apiVersion=%q kind=%q", apiVersion, kind)
}

// resourceKey returns a unique identifier for a resource.
func resourceKey(gvr schema.GroupVersionResource, name string) string {
	return fmt.Sprintf("%s/%s/%s", gvr.Group, gvr.Resource, name)
}

func handleModuleApply(ctx context.Context, client *k8s.Client, params ModuleApplyParams) (ModuleApplyResult, error) {
	if err := k8s.CheckManagedNamespace(ctx, client, params.Namespace); err != nil {
		return ModuleApplyResult{}, err
	}

	created, updated, deleted := 0, 0, 0
	appliedKeys := make(map[string]bool)

	for _, manifest := range params.Manifests {
		obj := &unstructured.Unstructured{Object: manifest}

		apiVersion := obj.GetAPIVersion()
		kind := obj.GetKind()
		if apiVersion == "" || kind == "" {
			return ModuleApplyResult{}, fmt.Errorf("manifest missing apiVersion or kind")
		}

		if !allowedKinds[kind] {
			return ModuleApplyResult{}, fmt.Errorf("kind %q is not permitted in module manifests; allowed kinds: Deployment, Service, PersistentVolumeClaim, NetworkPolicy, ConfigMap, Secret, Job, CronJob, Ingress", kind)
		}

		gvr, err := resolveGVR(ctx, client, apiVersion, kind)
		if err != nil {
			return ModuleApplyResult{}, fmt.Errorf("resolve GVR for %s/%s: %w", apiVersion, kind, err)
		}

		// Set namespace and release label
		obj.SetNamespace(params.Namespace)
		labels := obj.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		labels[releaseLabelKey] = params.Release
		obj.SetLabels(labels)

		name := obj.GetName()
		if name == "" {
			return ModuleApplyResult{}, fmt.Errorf("manifest of kind %s is missing a name", kind)
		}

		key := resourceKey(gvr, name)
		appliedKeys[key] = true

		// Try to get existing resource
		existing, err := client.Dynamic.Resource(gvr).Namespace(params.Namespace).Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			// Create
			_, createErr := client.Dynamic.Resource(gvr).Namespace(params.Namespace).Create(ctx, obj, metav1.CreateOptions{})
			if createErr != nil {
				return ModuleApplyResult{}, fmt.Errorf("create %s/%s: %w", kind, name, createErr)
			}
			created++
		} else if err != nil {
			return ModuleApplyResult{}, fmt.Errorf("get %s/%s: %w", kind, name, err)
		} else {
			// Update: preserve resource version
			obj.SetResourceVersion(existing.GetResourceVersion())
			_, updateErr := client.Dynamic.Resource(gvr).Namespace(params.Namespace).Update(ctx, obj, metav1.UpdateOptions{})
			if updateErr != nil {
				return ModuleApplyResult{}, fmt.Errorf("update %s/%s: %w", kind, name, updateErr)
			}
			updated++
		}
	}

	// Garbage collect: delete previously-labeled resources not in new manifest set
	// We need to list resources across known types that may carry the release label
	knownGVRs := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1", Resource: "cronjobs"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	}

	labelSelector := fmt.Sprintf("%s=%s", releaseLabelKey, params.Release)
	for _, gvr := range knownGVRs {
		list, err := client.Dynamic.Resource(gvr).Namespace(params.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			continue // skip GVRs that don't exist or are not accessible
		}
		for _, item := range list.Items {
			key := resourceKey(gvr, item.GetName())
			if !appliedKeys[key] {
				err := client.Dynamic.Resource(gvr).Namespace(params.Namespace).Delete(ctx, item.GetName(), metav1.DeleteOptions{})
				if err != nil {
					continue // best-effort GC
				}
				deleted++
			}
		}
	}

	return ModuleApplyResult{
		Release:   params.Release,
		Namespace: params.Namespace,
		Created:   created,
		Updated:   updated,
		Deleted:   deleted,
	}, nil
}

func handleModuleRemove(ctx context.Context, client *k8s.Client, params ModuleRemoveParams) (ModuleRemoveResult, error) {
	if err := k8s.CheckManagedNamespace(ctx, client, params.Namespace); err != nil {
		return ModuleRemoveResult{}, err
	}
	knownGVRs := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1", Resource: "cronjobs"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	}

	labelSelector := fmt.Sprintf("%s=%s", releaseLabelKey, params.Release)
	deleted := 0

	for _, gvr := range knownGVRs {
		list, err := client.Dynamic.Resource(gvr).Namespace(params.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			continue
		}
		for _, item := range list.Items {
			err := client.Dynamic.Resource(gvr).Namespace(params.Namespace).Delete(ctx, item.GetName(), metav1.DeleteOptions{})
			if err != nil {
				continue
			}
			deleted++
		}
	}

	return ModuleRemoveResult{
		Release:   params.Release,
		Namespace: params.Namespace,
		Deleted:   deleted,
	}, nil
}

func handleModuleStatus(ctx context.Context, client *k8s.Client, params ModuleStatusParams) (ModuleStatusResult, error) {
	if err := k8s.CheckManagedNamespace(ctx, client, params.Namespace); err != nil {
		return ModuleStatusResult{}, err
	}
	knownGVRs := []schema.GroupVersionResource{
		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1", Resource: "cronjobs"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
		{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	}

	labelSelector := fmt.Sprintf("%s=%s", releaseLabelKey, params.Release)
	resources := []ModuleResourceStatus{}

	for _, gvr := range knownGVRs {
		list, err := client.Dynamic.Resource(gvr).Namespace(params.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
		if err != nil {
			continue
		}
		for _, item := range list.Items {
			ready, reason := resourceReadiness(item, gvr.Resource)
			resources = append(resources, ModuleResourceStatus{
				Kind:   strings.ToTitle(gvr.Resource[:1]) + gvr.Resource[1:],
				Name:   item.GetName(),
				Ready:  ready,
				Reason: reason,
			})
		}
	}

	return ModuleStatusResult{
		Release:   params.Release,
		Namespace: params.Namespace,
		Resources: resources,
	}, nil
}

// resourceReadiness determines readiness from an unstructured resource.
func resourceReadiness(obj unstructured.Unstructured, resource string) (bool, string) {
	switch resource {
	case "deployments":
		readyReplicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "readyReplicas")
		replicas, _, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas")
		if replicas == 0 {
			replicas = 1
		}
		if readyReplicas >= replicas {
			return true, ""
		}
		return false, fmt.Sprintf("%d/%d replicas ready", readyReplicas, replicas)
	case "jobs":
		conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
		for _, c := range conditions {
			cond, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if cond["type"] == "Complete" && cond["status"] == "True" {
				return true, ""
			}
			if cond["type"] == "Failed" && cond["status"] == "True" {
				return false, "job failed"
			}
		}
		return false, "job in progress"
	default:
		// Services, ConfigMaps, Secrets, NetworkPolicies, CronJobs: presence = ready
		return true, ""
	}
}
