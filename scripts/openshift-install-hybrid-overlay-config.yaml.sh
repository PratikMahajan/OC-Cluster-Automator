#!/usr/bin/env bash
# openshift-install-hybrid-overlay-config.yaml.sh - Echos an hybrid overlay configuration
#
# USAGE
#
#    openshift-install-hybrid-overlay-config.yaml.sh
#

function die() {
    echo "Error: $*" >&2
    exit 1
}


cat <<EOF
apiVersion: operator.openshift.io/v1
kind: Network
metadata:
  creationTimestamp: null
  name: cluster
spec:
  clusterNetwork:
  - cidr: 10.128.0.0/14
    hostPrefix: 23
  externalIP:
    policy: {}
  networkType: OVNKubernetes
  serviceNetwork:
  - 172.30.0.0/16
  defaultNetwork:
    type: OVNKubernetes
    ovnKubernetesConfig:
      hybridOverlayConfig:
        hybridClusterNetwork:
        - cidr: 10.132.0.0/14
          hostPrefix: 23
status: {}
EOF