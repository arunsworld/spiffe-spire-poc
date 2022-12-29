#!/bin/bash

set -eo pipefail

TRUST_DOMAIN="$1"
CLUSTER_NAME="$2"

DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

endpointAddr=$($DIR/get_service_ip_port.sh spire spire-server-bundle-endpoint)
if [ -z "${endpointAddr}" ]; then
    echo "Endpoint service not ready" 2>&1
    exit 1
fi

bundleContents=$(kubectl exec \
    -n spire \
    -c spire-server \
    statefulset/spire-server -- \
    /opt/spire/bin/spire-server bundle show --format=spiffe) \
trustDomain="$TRUST_DOMAIN" \
resourceName="$CLUSTER_NAME" \
bundleEndpointURL="https://${endpointAddr}" \
endpointSPIFFEID="spiffe://$TRUST_DOMAIN/spire/server" \
    yq eval -n '{
    "apiVersion": "spire.spiffe.io/v1alpha1",
    "kind": "ClusterFederatedTrustDomain",
    "metadata": {
        "name": strenv(resourceName)
    },
    "spec": {
        "trustDomain": strenv(trustDomain),
        "bundleEndpointURL": strenv(bundleEndpointURL),
        "bundleEndpointProfile": {
            "type": "https_spiffe",
            "endpointSPIFFEID": strenv(endpointSPIFFEID)
        },
        "trustDomainBundle": strenv(bundleContents)
    }
}'