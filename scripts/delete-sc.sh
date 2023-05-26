#!/bin/bash

source `dirname $0`/common.sh

CLUSTER_ID=$1
clusterIDCheck
ocmTokenCheck

ocm login --token $OCM_TOKEN --url integration
ocm delete /api/osd_fleet_mgmt/v1/service_clusters/$CLUSTER_ID
ocm get /api/osd_fleet_mgmt/v1/service_clusters/$CLUSTER_ID | jq .status
ocm delete /api/osd_fleet_mgmt/v1/service_clusters/$CLUSTER_ID/ack

echo "SD $CLUSTER_ID deletion has initiated. Cluster will be removed shortly."
