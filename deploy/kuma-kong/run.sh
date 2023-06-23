#!/bin/bash

apps="apple banana cactus"
count=30000

if [ "$1" = "apply" ] ; then
    kubectl apply -f kong-crds.yml
    for app in $apps ; do
        sed -e "s#@@NS@@#${app}#g" \
            tmpl-app-demo.yml | kubectl apply -f -
        sed -e "s#@@NS@@#${app}#g" \
            -e "s#@@NODEPORT@@#${count}#g" \
            tmpl-kong.yml | kubectl apply -f -
        count=$(( count + 10 ))
    done
elif [ "$1" = "delete" ] ; then
    for app in $apps ; do
        sed -e "s#@@NS@@#${app}#g" tmpl-app-demo.yml | kubectl delete -f -
        sed -e "s#@@NS@@#${app}#g" \
            -e "s#@@NODEPORT@@#${count}#g" \
            tmpl-kong.yml | kubectl delete -f -
    done
    kubectl delete -f kong-crds.yml
else
    echo "Pods"
    echo "===="
    kubectl get pods -A

    echo ""
    echo "Services"
    echo "========"
    kubectl get svc -A
fi
