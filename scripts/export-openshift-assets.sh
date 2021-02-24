#!/bin/bash

DIRECTORY=`dirname $0`
EXPORT_DIR="${DIRECTORY}/../config/openshift"
EXPORT_PROJECT=$1
PROJECT=${2:-${EXPORT_PROJECT}}

declare -a COMPONENTS=("redis" "trillian-log" "trillian-signer" "trillian-db" "rekor-server")

echo "--- Export Kubernetes resources for OpenShift from ${EXPORT_PROJECT} ---"

mkdir -p ${EXPORT_DIR} 2> /dev/null

for COMPONENT_NAME in "${COMPONENTS[@]}"
do
    echo "Exporting resources for ${COMPONENT_NAME}..."

    SECRET_YAML=${EXPORT_DIR}/${COMPONENT_NAME}-secret.yaml
    SERVICE_YAML=${EXPORT_DIR}/${COMPONENT_NAME}-service.yaml
    ROUTE_YAML=${EXPORT_DIR}/${COMPONENT_NAME}-route.yaml
    CONFIGMAP_YAML=${EXPORT_DIR}/${COMPONENT_NAME}-configmap.yaml
    DEPLOYMENTCONFIG_YAML=${EXPORT_DIR}/${COMPONENT_NAME}-deploymentconfig.yaml
    DEPLOYMENT_YAML=${EXPORT_DIR}/${COMPONENT_NAME}-deployment.yaml

    ## Secret
    oc get secret -n ${EXPORT_PROJECT} -lapp.kubernetes.io/instance=${COMPONENT_NAME} -o yaml --ignore-not-found > ${SECRET_YAML}

    if [ -s ${SECRET_YAML} ]
    then
        yq delete --inplace  ${SECRET_YAML} items[*].metadata.namespace
        yq delete --inplace  ${SECRET_YAML} items[*].metadata.uid
        yq delete --inplace  ${SECRET_YAML} items[*].metadata.selfLink
        yq delete --inplace  ${SECRET_YAML} items[*].metadata.creationTimestamp
        yq delete --inplace  ${SECRET_YAML} items[*].metadata.resourceVersion
        yq delete --inplace  ${SECRET_YAML} items[*].metadata.ownerReferences
        yq delete --inplace  ${SECRET_YAML} items[*].metadata.managedFields
    fi

    ## Service
    oc get service -n ${EXPORT_PROJECT} -lapp.kubernetes.io/instance=${COMPONENT_NAME} -o yaml --ignore-not-found > ${SERVICE_YAML}

    if [ -s ${SERVICE_YAML} ]
    then
        yq delete --inplace  ${SERVICE_YAML} items[*].metadata.namespace
        yq delete --inplace  ${SERVICE_YAML} items[*].metadata.uid
        yq delete --inplace  ${SERVICE_YAML} items[*].metadata.selfLink
        yq delete --inplace  ${SERVICE_YAML} items[*].metadata.creationTimestamp
        yq delete --inplace  ${SERVICE_YAML} items[*].metadata.resourceVersion
        yq delete --inplace  ${SERVICE_YAML} items[*].metadata.ownerReferences
        yq delete --inplace  ${SERVICE_YAML} items[*].metadata.managedFields
        yq delete --inplace  ${SERVICE_YAML} items[*].spec.clusterIP
        
   fi 
    ## Route
    oc get route -n ${EXPORT_PROJECT} -lapp.kubernetes.io/instance=${COMPONENT_NAME} -o yaml --ignore-not-found > ${ROUTE_YAML}

    if [ -s ${ROUTE_YAML} ]
    then
        yq delete --inplace  ${ROUTE_YAML} items[*].metadata.namespace
        yq delete --inplace  ${ROUTE_YAML} items[*].metadata.uid
        yq delete --inplace  ${ROUTE_YAML} items[*].metadata.selfLink
        yq delete --inplace  ${ROUTE_YAML} items[*].metadata.creationTimestamp
        yq delete --inplace  ${ROUTE_YAML} items[*].metadata.resourceVersion
        yq delete --inplace  ${ROUTE_YAML} items[*].metadata.ownerReferences
        yq delete --inplace  ${ROUTE_YAML} items[*].metadata.managedFields
        yq delete --inplace  ${ROUTE_YAML} items[*].spec.host
        yq delete --inplace  ${ROUTE_YAML} items[*].status.ingress[*].conditions[*].lastTransitionTime
        yq delete --inplace  ${ROUTE_YAML} items[*].status.ingress[*].host
        yq delete --inplace  ${ROUTE_YAML} items[*].status.ingress[*].routerCanonicalHostname
        yq delete --inplace  ${ROUTE_YAML} items[*].status.ingress[*].routerName
        yq delete --inplace  ${ROUTE_YAML} items[*].status.ingress[*].wildcardPolicy
#        sed -i "s/${EXPORT_PROJECT}/${PROJECT}/g" ${ROUTE_YAML}
    fi
    ## ConfigMap
    oc get configmap -n ${EXPORT_PROJECT} -lapp.kubernetes.io/instance=${COMPONENT_NAME} -o yaml --ignore-not-found > ${CONFIGMAP_YAML}

    if [ -s ${CONFIGMAP_YAML} ]
    then
        yq delete --inplace  ${CONFIGMAP_YAML} items[*].metadata.namespace
        yq delete --inplace  ${CONFIGMAP_YAML} items[*].metadata.uid
        yq delete --inplace  ${CONFIGMAP_YAML} items[*].metadata.selfLink
        yq delete --inplace  ${CONFIGMAP_YAML} items[*].metadata.creationTimestamp
        yq delete --inplace  ${CONFIGMAP_YAML} items[*].metadata.resourceVersion
        yq delete --inplace  ${CONFIGMAP_YAML} items[*].metadata.managedFields
    fi

    ## Deployment Config
    oc get deploymentconfig -n ${EXPORT_PROJECT} -lapp.kubernetes.io/instance=${COMPONENT_NAME} -o yaml --ignore-not-found > ${DEPLOYMENTCONFIG_YAML}

    if [ -s ${DEPLOYMENTCONFIG_YAML} ]
    then
        yq delete --inplace  ${DEPLOYMENTCONFIG_YAML} items[*].metadata.namespace
        yq delete --inplace  ${DEPLOYMENTCONFIG_YAML} items[*].metadata.uid
        yq delete --inplace  ${DEPLOYMENTCONFIG_YAML} items[*].metadata.selfLink
        yq delete --inplace  ${DEPLOYMENTCONFIG_YAML} items[*].metadata.creationTimestamp
        yq delete --inplace  ${DEPLOYMENTCONFIG_YAML} items[*].metadata.resourceVersion
        yq delete --inplace  ${DEPLOYMENTCONFIG_YAML} items[*].metadata.generation
        yq delete --inplace  ${DEPLOYMENTCONFIG_YAML} items[*].metadata.managedFields
        yq delete --inplace  ${DEPLOYMENTCONFIG_YAML} items[*].status
        yq write --inplace ${DEPLOYMENTCONFIG_YAML} items[*].spec.triggers null

        sed -i "s/  envFrom:/- envFrom:/g"  ${DEPLOYMENTCONFIG_YAML}
        grep '\- envFrom:' ${DEPLOYMENTCONFIG_YAML} &> /dev/null || sed -i "s/  image:/- image:/g"  ${DEPLOYMENTCONFIG_YAML}
        sed -i "/^.*: \[\]$/d"  ${DEPLOYMENTCONFIG_YAML}
        sed -i "s/triggers: .*/triggers: []/g"  ${DEPLOYMENTCONFIG_YAML}
        # Don't rename tcp to http
        # sed -i "s/8080-tcp/http/g" ${DEPLOYMENTCONFIG_YAML}
        # Don't use local image registry
        # sed -i "s/image: .*$/image: image-registry.openshift-image-registry.svc:5000\/${PROJECT}\/${COMPONENT_NAME}:latest/g" ${DEPLOYMENTCONFIG_YAML}
    fi

    ## Deployment
    oc get deployment -n ${EXPORT_PROJECT} -lapp.kubernetes.io/instance=${COMPONENT_NAME} -o yaml --ignore-not-found > ${DEPLOYMENT_YAML}

    if [ -s ${DEPLOYMENT_YAML} ]
    then
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.namespace'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.namespace'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.uid'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.selfLink'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.creationTimestamp'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.resourceVersion'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.generation'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.managedFields'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.annotations.[deployment.kubernetes.io/revision]'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].metadata.annotations.[image.openshift.io/triggers]'
        yq delete --inplace  ${DEPLOYMENT_YAML} 'items[*].status'

        sed -i "s/  envFrom:/- envFrom:/g"  ${DEPLOYMENT_YAML}
        # grep '\- envFrom:' ${DEPLOYMENT_YAML} &> /dev/null || sed -i "s/  image:/- image:/g"  ${DEPLOYMENT_YAML}
        sed -i "/^.*: \[\]$/d"  ${DEPLOYMENT_YAML}
        # sed -i "s/8080-tcp/http/g" ${DEPLOYMENT_YAML}
        # Don't use local image registry
        # sed -i "s/image: .*$/image: image-registry.openshift-image-registry.svc:5000\/${PROJECT}\/${COMPONENT_NAME}:latest/g" ${DEPLOYMENT_YAML}
    fi
done

echo "--- Kubernetes resources has been exported! ---"