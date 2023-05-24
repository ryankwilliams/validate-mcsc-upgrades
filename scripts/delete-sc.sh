#!/bin/bash

if [ -z $OCM_TOKEN ];
then
    echo "OCM_TOKEN is not set, please set it and try again"
    exit 1
fi

CLUSTER_ID=$1
if [ -z $CLUSTER_ID ];
then
    echo "CLUSTER_ID is not set, please set it and try again"
    exit 1
fi

ocm login --token $OCM_TOKEN --url integration
ocm delete /api/osd_fleet_mgmt/v1/service_clusters/$CLUSTER_ID
ocm get /api/osd_fleet_mgmt/v1/service_clusters/$CLUSTER_ID | jq .status
ocm delete /api/osd_fleet_mgmt/v1/service_clusters/$CLUSTER_ID/ack

echo "SD $CLUSTER_ID deletion has initiated. Cluster will be removed shortly."
