#!/bin/bash

ocmTokenCheck() {
    if [ -z $OCM_TOKEN ];
    then
        echo "OCM_TOKEN is not set, please set it and try again"
        exit 1
    fi
}

serviceClusterIDCheck() {
    if [ -z $SERVICE_CLUSTER_ID ];
    then
        echo "SERVICE_CLUSTER_ID is not set, please set it and try again"
        exit 1
    fi
}

clusterIDCheck() {
    if [ -z $CLUSTER_ID ];
    then
        echo "CLUSTER_ID is not set, please set it and try again"
        exit 1
    fi
}
