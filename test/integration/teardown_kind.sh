#!/usr/bin/env bash
# Destroys the kind cluster created for integration testing.
set -euo pipefail

CLUSTER_NAME="${TENTACULAR_TEST_CLUSTER:-tentacular-mcp-test}"

check_dependencies() {
    if ! command -v kind &>/dev/null; then
        echo "ERROR: 'kind' not found in PATH" >&2
        exit 1
    fi
}

destroy_cluster() {
    if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        echo "Cluster '${CLUSTER_NAME}' does not exist, nothing to teardown."
        return 0
    fi
    echo "Deleting kind cluster '${CLUSTER_NAME}'..."
    kind delete cluster --name "${CLUSTER_NAME}"
    echo "Cluster '${CLUSTER_NAME}' deleted."
}

main() {
    check_dependencies
    destroy_cluster
}

main "$@"
