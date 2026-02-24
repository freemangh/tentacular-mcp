#!/usr/bin/env bash
# Creates a kind cluster for integration testing.
set -euo pipefail

CLUSTER_NAME="${TENTACULAR_TEST_CLUSTER:-tentacular-mcp-test}"

check_dependencies() {
    local missing=()
    for cmd in kind kubectl; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        echo "ERROR: missing required tools: ${missing[*]}" >&2
        exit 1
    fi
}

create_cluster() {
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        echo "Cluster '${CLUSTER_NAME}' already exists, skipping creation."
        return 0
    fi
    echo "Creating kind cluster '${CLUSTER_NAME}'..."
    kind create cluster --name "${CLUSTER_NAME}" --wait 60s
    echo "Cluster '${CLUSTER_NAME}' created."
}

wait_for_ready() {
    echo "Waiting for cluster nodes to be ready..."
    kubectl --context "kind-${CLUSTER_NAME}" wait \
        --for=condition=Ready nodes --all --timeout=120s
    echo "Cluster nodes are ready."
}

main() {
    check_dependencies
    create_cluster
    wait_for_ready
    echo "Integration test cluster ready: kind-${CLUSTER_NAME}"
}

main "$@"
