# Gloo Helm Charts

If using Minikube make sure RBAC is enabled. If you enable monitoring or open tracing, please ensure you have allocated at least 4GB of RAM to the minikube VM.

    minikube start --extra-config=apiserver.Authorization.Mode=RBAC --memory 4096
    kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default

hack/deploy.sh deploys default instance of Gloo and hack/teardown.sh uninstalls it.