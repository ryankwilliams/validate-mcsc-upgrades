#!/bin/bash

source `dirname $0`/common.sh

CLUSTER_ID=$1
clusterIDCheck
ocmTokenCheck

ocm login --token $OCM_TOKEN --url integration
cluster_href=`ocm get /api/osd_fleet_mgmt/v1/management_clusters/$CLUSTER_ID | jq -r .cluster_management_reference.href`
ocm get $cluster_href/credentials | jq -r .kubeconfig > $CLUSTER_ID-kubeconfig
