#!/usr/bin/env bash
# openshift-install-create-config.yaml.sh - Echos an openshift-install install configuration
#
# USAGE
#
#    openshift-install-create-config.yaml.sh -n CLUSTER_NAME -p PLATFORM
#
# ARGUMENTS
#
#    -n CLUSTER_NAME    Name of cluster to create
#    -p PLATFORM        Name of the platform AWS/AZURE
#
# CONFIGURATION
#
#    Environment variables are used to configure the script:
#
#    APP_CLUSTERPULLSECRET    pull-secret
#    APP_SSHKEY               ssh key to configure VMs
#
#?

function die() {
    echo "Error: $*" >&2
    exit 1
}

function bold() {
    echo "$(tput bold)$*$(tput sgr0)"
}

# Options
while getopts "n:p:" opt; do
    case "$opt" in
	n) name="$OPTARG" ;;
	p) platform="$OPTARG" ;;
	?) die "Unknown option"
    esac
done

if [ -z "$name" ]; then
    die "-n NAME option required"
fi

if [ -z "$platform" ]; then
    die "-n PLATFORM option required"
fi
if [[ ! "$platform" =~ ^azure|aws$ ]]; then
    die "-a Platform must be \"aws\" or \"azure\""
fi


if [ -z "$APP_CLUSTERPULLSECRET" ]; then
    die "APP_CLUSTERPULLSECRET not found"
fi
if [ -z "$APP_SSHKEY" ]; then
    die "APP_SSHKEY not found"
fi


case "$platform" in
    aws)

cat <<EOF
apiVersion: v1
baseDomain: devcluster.openshift.com
compute:
- hyperthreading: Enabled
  name: worker
  platform: {}
  replicas: 3
controlPlane:
  hyperthreading: Enabled
  name: master
  platform: {}
  replicas: 3
metadata:
  creationTimestamp: null
  name: "$name"
networking:
  clusterNetwork:
  - cidr: 10.128.0.0/14
    hostPrefix: 23
  machineCIDR: 10.0.0.0/16
  networkType: OVNKubernetes
  serviceNetwork:
  - 172.30.0.0/16
platform:
  aws:
    region: us-east-2
pullSecret: '$(echo $APP_CLUSTERPULLSECRET)'
sshKey: |
  $(echo $APP_SSHKEY)
EOF
	;;
    azure)

cat <<EOF
apiVersion: v1
baseDomain: winc.azure.devcluster.openshift.com
compute:
- architecture: amd64
  hyperthreading: Enabled
  name: worker
  platform: {}
  replicas: 3
controlPlane:
  architecture: amd64
  hyperthreading: Enabled
  name: master
  platform: {}
  replicas: 3
metadata:
  creationTimestamp: null
  name: "$name"
networking:
  clusterNetwork:
  - cidr: 10.128.0.0/14
    hostPrefix: 23
  machineNetwork:
  - cidr: 10.0.0.0/16
  networkType: OVNKubernetes
  serviceNetwork:
  - 172.30.0.0/16
platform:
  azure:
    baseDomainResourceGroupName: os4-common
    region: centralus
publish: External
pullSecret: '$(echo $APP_CLUSTERPULLSECRET)'
sshKey: |
  $(echo $APP_SSHKEY)
EOF
	;;
esac