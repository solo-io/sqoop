#!/bin/bash
minikube start --extra-config=apiserver.Authorization.Mode=RBAC --memory 4096
sleep 5
kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default