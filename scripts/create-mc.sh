#!/bin/bash

source `dirname $0`/common.sh

SERVICE_CLUSTER_ID=$1
serviceClusterIDCheck
ocmTokenCheck

ocm login --token $OCM_TOKEN --url integration
echo '{"service_cluster_id":'\"$SERVICE_CLUSTER_ID\"'}' | ocm post /api/osd_fleet_mgmt/v1/management_clusters
