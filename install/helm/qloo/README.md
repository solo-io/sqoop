# Helm chart for QLoo

This Helm chart can be used to deploy QLoo and Gloo to Kubernetes.

To use monitoring and open tracing, please ensure you have at least
4GB RAM assigned to the minikube VM. Using Prometheus requires minikube
to be started in RBAC mode.

    minikube start --extra-config=apiserver.Authorization.Mode=RBAC --memory 4096
    kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
