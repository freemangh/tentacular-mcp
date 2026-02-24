package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterProfile contains a snapshot of cluster capabilities and configuration.
type ClusterProfile struct {
	GeneratedAt    time.Time          `json:"generatedAt"`
	K8sVersion     string             `json:"k8sVersion"`
	Distribution   string             `json:"distribution"`
	Nodes          []NodeInfo         `json:"nodes"`
	RuntimeClasses []RuntimeClassInfo `json:"runtimeClasses"`
	GVisor         bool               `json:"gvisor"`
	CNI            CNIInfo            `json:"cni"`
	StorageClasses []StorageClassInfo `json:"storageClasses"`
	CSIDrivers     []string           `json:"csiDrivers"`
	Extensions     map[string]bool    `json:"extensions"`
	Namespace      string             `json:"namespace"`
	Quota          *QuotaSummary      `json:"quota,omitempty"`
	LimitRange     *LimitRangeSummary `json:"limitRange,omitempty"`
	PodSecurity    string             `json:"podSecurity"`
}

// NodeInfo describes a single cluster node.
type NodeInfo struct {
	Name             string            `json:"name"`
	Ready            bool              `json:"ready"`
	OS               string            `json:"os"`
	Arch             string            `json:"arch"`
	KubeletVersion   string            `json:"kubeletVersion"`
	KernelVersion    string            `json:"kernelVersion"`
	ContainerRuntime string            `json:"containerRuntime"`
	Allocatable      map[string]string `json:"allocatable"`
	Capacity         map[string]string `json:"capacity"`
}

// RuntimeClassInfo describes a RuntimeClass.
type RuntimeClassInfo struct {
	Name    string `json:"name"`
	Handler string `json:"handler"`
}

// CNIInfo describes the detected CNI plugin.
type CNIInfo struct {
	Name                   string `json:"name"`
	NetworkPolicySupported bool   `json:"networkPolicySupported"`
	EgressSupported        bool   `json:"egressSupported"`
}

// StorageClassInfo describes a StorageClass.
type StorageClassInfo struct {
	Name        string `json:"name"`
	Provisioner string `json:"provisioner"`
	Default     bool   `json:"default"`
	RWXCapable  bool   `json:"rwxCapable"`
}

// QuotaSummary is a simplified view of a ResourceQuota.
type QuotaSummary struct {
	CPULimit string `json:"cpuLimit"`
	MemLimit string `json:"memLimit"`
	PodLimit int    `json:"podLimit"`
}

// LimitRangeSummary is a simplified view of a LimitRange.
type LimitRangeSummary struct {
	DefaultCPURequest string `json:"defaultCPURequest"`
	DefaultMemRequest string `json:"defaultMemRequest"`
	DefaultCPULimit   string `json:"defaultCPULimit"`
	DefaultMemLimit   string `json:"defaultMemLimit"`
}

// ProfileCluster performs a full scan of the cluster and returns a ClusterProfile.
func ProfileCluster(ctx context.Context, client *Client, namespace string) (*ClusterProfile, error) {
	ctx, cancel := context.WithTimeout(ctx, Timeout*2)
	defer cancel()

	profile := &ClusterProfile{
		GeneratedAt:    time.Now().UTC(),
		Namespace:      namespace,
		Extensions:     make(map[string]bool),
		Nodes:          []NodeInfo{},
		RuntimeClasses: []RuntimeClassInfo{},
		StorageClasses: []StorageClassInfo{},
		CSIDrivers:     []string{},
	}

	if err := profileVersion(ctx, client, profile); err != nil {
		return nil, err
	}
	if err := profileNodes(ctx, client, profile); err != nil {
		return nil, err
	}
	if err := profileRuntimeClasses(ctx, client, profile); err != nil {
		return nil, err
	}
	if err := profileCNI(ctx, client, profile); err != nil {
		return nil, err
	}
	if err := profileStorageClasses(ctx, client, profile); err != nil {
		return nil, err
	}
	if err := profileCSIDrivers(ctx, client, profile); err != nil {
		return nil, err
	}
	if err := profileExtensions(ctx, client, profile); err != nil {
		return nil, err
	}
	if err := profileNamespaceDetails(ctx, client, profile, namespace); err != nil {
		return nil, err
	}

	return profile, nil
}

func profileVersion(ctx context.Context, client *Client, profile *ClusterProfile) error {
	info, err := client.Clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("get server version: %w", err)
	}
	profile.K8sVersion = info.GitVersion
	return nil
}

func profileNodes(ctx context.Context, client *Client, profile *ClusterProfile) error {
	nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list nodes: %w", err)
	}

	for _, n := range nodes.Items {
		ready := false
		for _, c := range n.Status.Conditions {
			if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}

		alloc := make(map[string]string)
		for k, v := range n.Status.Allocatable {
			alloc[string(k)] = v.String()
		}
		cap := make(map[string]string)
		for k, v := range n.Status.Capacity {
			cap[string(k)] = v.String()
		}

		profile.Nodes = append(profile.Nodes, NodeInfo{
			Name:             n.Name,
			Ready:            ready,
			OS:               n.Status.NodeInfo.OperatingSystem,
			Arch:             n.Status.NodeInfo.Architecture,
			KubeletVersion:   n.Status.NodeInfo.KubeletVersion,
			KernelVersion:    n.Status.NodeInfo.KernelVersion,
			ContainerRuntime: n.Status.NodeInfo.ContainerRuntimeVersion,
			Allocatable:      alloc,
			Capacity:         cap,
		})
	}

	profile.Distribution = detectDistribution(nodes.Items)
	return nil
}

func detectDistribution(nodes []corev1.Node) string {
	for _, n := range nodes {
		labels := n.Labels
		if _, ok := labels["eks.amazonaws.com/nodegroup"]; ok {
			return "EKS"
		}
		if _, ok := labels["cloud.google.com/gke-nodepool"]; ok {
			return "GKE"
		}
		if _, ok := labels["kubernetes.azure.com/agentpool"]; ok {
			return "AKS"
		}
		if it, ok := labels["node.kubernetes.io/instance-type"]; ok && strings.Contains(strings.ToLower(it), "k3s") {
			return "K3s"
		}
	}
	return "vanilla"
}

func profileRuntimeClasses(ctx context.Context, client *Client, profile *ClusterProfile) error {
	rcs, err := client.Clientset.NodeV1().RuntimeClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list runtime classes: %w", err)
	}

	for _, rc := range rcs.Items {
		profile.RuntimeClasses = append(profile.RuntimeClasses, RuntimeClassInfo{
			Name:    rc.Name,
			Handler: rc.Handler,
		})
		if strings.Contains(strings.ToLower(rc.Handler), "gvisor") || strings.Contains(strings.ToLower(rc.Handler), "runsc") {
			profile.GVisor = true
		}
	}
	return nil
}

func profileCNI(ctx context.Context, client *Client, profile *ClusterProfile) error {
	pods, err := client.Clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list kube-system pods: %w", err)
	}

	cniMap := map[string]CNIInfo{
		"calico":  {Name: "calico", NetworkPolicySupported: true, EgressSupported: true},
		"cilium":  {Name: "cilium", NetworkPolicySupported: true, EgressSupported: true},
		"weave":   {Name: "weave", NetworkPolicySupported: true, EgressSupported: true},
		"flannel": {Name: "flannel", NetworkPolicySupported: false, EgressSupported: false},
		"kindnet": {Name: "kindnet", NetworkPolicySupported: false, EgressSupported: false},
	}

	for _, pod := range pods.Items {
		podName := strings.ToLower(pod.Name)
		for prefix, info := range cniMap {
			if strings.Contains(podName, prefix) {
				profile.CNI = info
				return nil
			}
		}
	}

	profile.CNI = CNIInfo{Name: "unknown", NetworkPolicySupported: false, EgressSupported: false}
	return nil
}

func profileStorageClasses(ctx context.Context, client *Client, profile *ClusterProfile) error {
	scs, err := client.Clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list storage classes: %w", err)
	}

	rwxProvisioners := []string{"nfs", "efs", "azurefile", "cephfs", "gluster"}

	for _, sc := range scs.Items {
		isDefault := false
		if sc.Annotations != nil {
			if sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true" {
				isDefault = true
			}
		}

		rwx := false
		provLower := strings.ToLower(sc.Provisioner)
		for _, kw := range rwxProvisioners {
			if strings.Contains(provLower, kw) {
				rwx = true
				break
			}
		}

		profile.StorageClasses = append(profile.StorageClasses, StorageClassInfo{
			Name:        sc.Name,
			Provisioner: sc.Provisioner,
			Default:     isDefault,
			RWXCapable:  rwx,
		})
	}
	return nil
}

func profileCSIDrivers(ctx context.Context, client *Client, profile *ClusterProfile) error {
	drivers, err := client.Clientset.StorageV1().CSIDrivers().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list CSI drivers: %w", err)
	}

	for _, d := range drivers.Items {
		profile.CSIDrivers = append(profile.CSIDrivers, d.Name)
	}
	return nil
}

func profileExtensions(ctx context.Context, client *Client, profile *ClusterProfile) error {
	_, resources, err := client.Clientset.Discovery().ServerGroupsAndResources()
	if err != nil {
		// Some groups may fail but we can still check what we got.
		if resources == nil {
			return fmt.Errorf("discover API resources: %w", err)
		}
	}

	extensionCRDs := map[string]string{
		"istio":            "networking.istio.io",
		"cert-manager":     "cert-manager.io",
		"prometheus":       "monitoring.coreos.com",
		"external-secrets": "external-secrets.io",
		"argocd":           "argoproj.io",
		"gateway-api":      "gateway.networking.k8s.io",
	}

	groupSet := make(map[string]bool)
	for _, rl := range resources {
		parts := strings.Split(rl.GroupVersion, "/")
		if len(parts) > 0 {
			groupSet[parts[0]] = true
		}
	}

	for name, group := range extensionCRDs {
		profile.Extensions[name] = groupSet[group]
	}
	return nil
}

func profileNamespaceDetails(ctx context.Context, client *Client, profile *ClusterProfile, namespace string) error {
	if namespace == "" {
		return nil
	}

	ns, err := client.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get namespace %q: %w", namespace, err)
	}

	if ns.Labels != nil {
		profile.PodSecurity = ns.Labels["pod-security.kubernetes.io/enforce"]
	}

	quotas, err := client.Clientset.CoreV1().ResourceQuotas(namespace).List(ctx, metav1.ListOptions{})
	if err == nil && len(quotas.Items) > 0 {
		q := quotas.Items[0]
		qs := &QuotaSummary{}
		if v, ok := q.Spec.Hard[corev1.ResourceLimitsCPU]; ok {
			qs.CPULimit = v.String()
		}
		if v, ok := q.Spec.Hard[corev1.ResourceLimitsMemory]; ok {
			qs.MemLimit = v.String()
		}
		if v, ok := q.Spec.Hard[corev1.ResourcePods]; ok {
			qs.PodLimit = int(v.Value())
		}
		profile.Quota = qs
	}

	lrs, err := client.Clientset.CoreV1().LimitRanges(namespace).List(ctx, metav1.ListOptions{})
	if err == nil && len(lrs.Items) > 0 {
		lr := lrs.Items[0]
		for _, item := range lr.Spec.Limits {
			if item.Type == corev1.LimitTypeContainer {
				lrs := &LimitRangeSummary{}
				if v, ok := item.DefaultRequest[corev1.ResourceCPU]; ok {
					lrs.DefaultCPURequest = v.String()
				}
				if v, ok := item.DefaultRequest[corev1.ResourceMemory]; ok {
					lrs.DefaultMemRequest = v.String()
				}
				if v, ok := item.Default[corev1.ResourceCPU]; ok {
					lrs.DefaultCPULimit = v.String()
				}
				if v, ok := item.Default[corev1.ResourceMemory]; ok {
					lrs.DefaultMemLimit = v.String()
				}
				profile.LimitRange = lrs
				break
			}
		}
	}

	return nil
}
