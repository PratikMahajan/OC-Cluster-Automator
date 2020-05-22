#!/usr/bin/env bash
# openshift-install-download.sh - downloads openshift-install on behalf of the Go program
#
# USAGE
#
#    openshift-install-download.sh -v VERSION -d BIN_DIR
#
# OPTIONS
#
#    -v VERSION      OC Version to Download
#    -d BIN_DIR      Directory where the binaary will be saved
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
while getopts "v:d:" opt; do
    case "$opt" in
	v) version="$OPTARG" ;;
	d) bin_dir="$OPTARG" ;;
	?) die "Unknown option"
    esac
done

if [ -z "$bin_dir" ]; then
    die "-d BIN_DIR option required"
fi

if [ ! -d "$bin_dir" ]; then
    die "-d BIN_DIR directory does not exist"
fi

if [ -z "$version" ]; then
    die "-v VERSION option required"
fi

# download openshift-install binary
if ! wget -O /tmp/openshift-install-linux.tar.gz https://mirror.openshift.com/pub/openshift-v4/clients/\
ocp-dev-preview/latest-"$version"/openshift-install-linux.tar.gz; then
    die "Failed to download openshift-install binary"
fi

# extract openshift-install binary
if ! tar -C "$bin_dir" -xvf /tmp/openshift-install-linux.tar.gz; then
    die "Failed to extract oc binary"
fi

# clear /tmp folder
if ! rm -f /tmp/openshift-install-linux.tar.gz; then
    echo "Failed to delete downloaded binary from /tmp folder"
fi

# give execute permission for openshift-install binary
if ! chmod +x "$bin_dir"/openshift-install; then
    die "Failed to change oc binary permissions"
fi
