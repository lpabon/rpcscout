#!/bin/bash

apps="apple banana cactus"

if [ "$1" = "apply" ] ; then
    kubectl apply -f kong-crds.yml
    for app in $apps ; do
        sed -e "s#@@NS@@#${app}#g" tmpl-app-demo.yml | kubectl apply -f -
        sed -e "s#@@NS@@#${app}#g" tmpl-kong.yml | kubectl apply -f -
    done
elif [ "$1" = "delete" ] ; then
    for app in $apps ; do
        sed -e "s#@@NS@@#${app}#g" tmpl-kong.yml | kubectl delete -f -
    done
    kubectl delete -f kong-crds.yml
else
    kubectl get pods -A
fi
