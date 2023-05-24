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
cluster_href=`ocm get /api/osd_fleet_mgmt/v1/service_clusters/$CLUSTER_ID | jq -r .cluster_management_reference.href`
ocm get $cluster_href/credentials | jq -r .kubeconfig > $CLUSTER_ID-kubeconfig
