#!/bin/bash

if [ -z $OCM_TOKEN ];
then
    echo "OCM_TOKEN is not set, please set it and try again"
    exit 1
fi

SERVICE_CLUSTER_ID=$1
if [ -z $SERVICE_CLUSTER_ID ];
then
    echo "SERVICE_CLUSTER_ID is not set, please set it and try again"
    exit 1
fi

ocm login --token $OCM_TOKEN --url integration
echo '{"service_cluster_id":"$SERVICE_CLUSTER_ID"}' | ocm post /api/osd_fleet_mgmt/v1/management_clusters
