#!/bin/bash

NAMESPACE=gloo-system

helm repo add coreos https://s3-eu-west-1.amazonaws.com/coreos-charts/stable/
helm install coreos/prometheus-operator --name prometheus-operator --namespace $NAMESPACE
cat ../bootstrap.yaml | sed -e "s/{{ .Namespace }}/$NAMESPACE/" | kubectl create -f -
helm install ../gloo --namespace $NAMESPACE -n gloo-demo 
