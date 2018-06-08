#!/bin/bash

NAMESPACE=gloo-system
helm delete --purge gloo-demo
cat ../bootstrap.yaml | sed -e "s/{{ .Namespace }}/$NAMESPACE/" | kubectl delete -f -
helm delete --purge prometheus-operator

# total cleanup
kubectl delete crd servicemonitors.monitoring.coreos.com
kubectl delete crd prometheuses.monitoring.coreos.com
kubectl delete crd alertmanagers.monitoring.coreos.com
