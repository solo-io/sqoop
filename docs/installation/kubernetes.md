---
weight: 1
title: Kubernetes
---


# Installing on Kubernetes

## Installing with `kubectl`

#### What you'll need

1. Kubernetes v1.8+ or higher deployed. We recommend using [minikube](https://kubernetes.io/docs/getting-started-guides/minikube/) to get a demo cluster up quickly.
1. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed on your local machine.

Once your Kubernetes cluster is up and running, run the following command to deploy Sqoop and Gloo to the `gloo-system` namespace:

```bash
# install Sqoop
kubectl apply -f \
    https://raw.githubusercontent.com/solo-io/sqoop/master/install/manifest/sqoop.yaml
```


## Installing with `sqoopctl`

#### What you'll need

1. Kubernetes v1.8+ or higher deployed. We recommend using [minikube](https://kubernetes.io/docs/getting-started-guides/minikube/) to get a demo cluster up quickly.
1. [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed on your local machine.
1. [`sqoopctl`](https://github.com/solo-io/sqoop/releases/) installed on your local machine.

Once your Kubernetes cluster is up and running, run the following command to deploy Sqoop and Gloo to the `gloo-system` namespace:

```bash
sqoopctl install kube 
```

## Confirming the installation

Check that the Gloo pods and services have been created:

```bash
$ kubectl get all -n gloo-system

NAME                                 READY   STATUS    RESTARTS   AGE
pod/discovery-7f6865dc44-n4qrf       1/1     Running   0          2m5s
pod/gateway-66c549fdf6-s5kl2         1/1     Running   0          2m5s
pod/gateway-proxy-5c4df77bc6-n7v9r   1/1     Running   0          2m5s
pod/gloo-69f67879cb-6qcrv            1/1     Running   0          2m5s
pod/sqoop-855dc98dfd-bx97f           2/2     Running   0          2m5s

NAME                    TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                      AGE
service/gateway-proxy   LoadBalancer   10.101.11.56    <pending>     80:30045/TCP,443:31283/TCP   2m5s
service/gloo            ClusterIP      10.98.204.164   <none>        9977/TCP                     2m5s
service/sqoop           LoadBalancer   10.99.235.90    <pending>     9095:30772/TCP               2m5s

NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/discovery       1/1     1            1           2m5s
deployment.apps/gateway         1/1     1            1           2m5s
deployment.apps/gateway-proxy   1/1     1            1           2m5s
deployment.apps/gloo            1/1     1            1           2m5s
deployment.apps/sqoop           1/1     1            1           2m5s

NAME                                       DESIRED   CURRENT   READY   AGE
replicaset.apps/discovery-7f6865dc44       1         1         1       2m5s
replicaset.apps/gateway-66c549fdf6         1         1         1       2m5s
replicaset.apps/gateway-proxy-5c4df77bc6   1         1         1       2m5s
replicaset.apps/gloo-69f67879cb            1         1         1       2m5s
replicaset.apps/sqoop-855dc98dfd           1         1         1       2m5s
```

Everything should be up and running. If this process does not work, please [open an issue](https://github.com/solo-io/sqoop/issues/new). We are happy to answer
questions on our [diligently staffed Slack channel](https://slack.solo.io/).

See [Getting Started on Kubernetes](../getting_started/kubernetes.md) to get started creating your first GraphQL endpoint with Sqoop.
