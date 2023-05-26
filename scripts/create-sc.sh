#!/bin/bash

source `dirname $0`/common.sh

CLUSTER_ID=$1
clusterIDCheck
ocmTokenCheck

ocm login --token $OCM_TOKEN --url integration
echo '{"region":"eu-central-1", "cloud_provider":"aws"}' | ocm post /api/osd_fleet_mgmt/v1/service_clusters
