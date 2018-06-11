# Quick minikube instructions:

On linux, the easiest way to start is with the kvm2 driver:
```
minikube stop; \
minikube start --vm-driver=kvm2  \
        --feature-gates=AdvancedAudit=true \
        --extra-config=apiserver.Authorization.Mode=RBAC \
        --extra-config=apiserver.Audit.LogOptions.Path=/hosthome/$USER/.minikube/logs/audit.log \
        --extra-config=apiserver.Audit.MaxAge=30 \
        --extra-config=apiserver.Audit.MaxSize=100 \
        --extra-config=apiserver.Audit.MaxBackups=5 \
        --extra-config=apiserver.Audit.PolicyFile=/etc/kubernetes/addons/audit-policy.yaml && \
kubectl create clusterrolebinding permissive-binding \
         --clusterrole=cluster-admin \
         --user=admin \
         --user=kubelet \
         --group=system:serviceaccounts


minikube stop; minikube start --vm-driver=kvm2          --feature-gates=AdvancedAudit=true         --extra-config=apiserver.Authorization.Mode=RBAC         --extra-config=apiserver.Audit.LogOptions.Path=/hosthome/$USER/.minikube/logs/audit.log         --extra-config=apiserver.Audit.PolicyFile=/etc/kubernetes/addons/audit-policy.yaml && kubectl create clusterrolebinding permissive-binding \
          --clusterrole=cluster-admin \
          --user=admin \
          --user=kubelet \
          --group=system:serviceaccounts
```

Then just install gloo and example:
```
kubectl apply -f install.yaml
kubectl apply -f example-gloo.yaml
```
To access:
```
export GATEWAY_URL=http://$(minikube ip):$(kubectl get svc qloo -n gloo-system -o 'jsonpath={.spec.ports[?(@.name=="http")].nodePort}')
```


And open in your browser:
```
xdg-open $GATEWAY_URL
```
