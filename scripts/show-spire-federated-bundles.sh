#!/bin/bash

set -eo pipefail

kubectl exec -t \
    -n spire \
    -c spire-server statefulset/spire-server -- \
        /opt/spire/bin/spire-server bundle list -format spiffe