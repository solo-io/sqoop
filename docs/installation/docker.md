---
weight: 1
title: Docker
---


# Installing on Docker

## Installing with `sqoopctl`

#### What you'll need

 1. [Docker](https://www.docker.com/)
 1. [Docker-Compose](https://docs.docker.com/compose/)
 2. [sqoopctl](https://github.com/solo-io/sqoop/releases)

#### Install

 Run the following command to download the config files necessary for docker-compose to a local directory:

```bash
glooctl install docker [path]
```


`cd` to this directory to begin working with Sqoop.


#### Running Sqoop

Start Sqoop by simply running `docker-compose up`

```bash
cd [folder]

docker-compose up
```
## Installing without `sqoopctl`

#### What you'll need

 1. [Docker](https://www.docker.com/)
 1. [Docker-Compose](https://docs.docker.com/compose/))

#### Install

 `git clone` the Sqoop repository:
 
```bash
git clone https://github.com/solo-io/sqoop.git
``` 
 
 Initialize the storage directories and `sqoopctl` configuration for Sqoop:
 
```bash
cd sqoop/install/docker-compose
./prepare-config-directories.sh
```

#### Running Sqoop

Start Sqoop by simply running `docker-compose up`

```bash
cd [folder]

docker-compose up
```

## Verify the installation

You can check gloo services are running by running `docker ps`

For example,

```bash
docker ps

CONTAINER ID        IMAGE                              COMMAND                  CREATED             STATUS              PORTS                                              NAMES
708445e2825c        soloio/sqoop:0.1.1                  "/sqoop --storage.typ…"   7 seconds ago       Up 4 seconds        0.0.0.0:9090->9090/tcp                             docker-compose_sqoop_1
90645fff651e        soloio/control-plane:0.4.4        "/control-plane --st…"   31 hours ago        Up 5 seconds        0.0.0.0:8081->8081/tcp                             docker-compose_control-plane_1
36bb7f23808c        soloio/envoy:0.4.4                "envoy -c /config/en…"   31 hours ago        Up 6 seconds        0.0.0.0:8080->8080/tcp, 0.0.0.0:19000->19000/tcp   docker-compose_proxy_1
7747741da1df        soloio/function-discovery:0.4.4   "/function-discovery…"   31 hours ago        Up 6 seconds                                                           docker-compose_function-discovery_1```
```

Everything should be up and running. If this process does not work, please [open an issue](https://github.com/solo-io/gloo/issues/new). We are happy to answer questions on our diligently staffed [Slack channel](https://slack.solo.io)

See [Getting Started on Docker](../getting_started/docker/1.md) to get started creating your first GraphQL endpoint with Sqoop.