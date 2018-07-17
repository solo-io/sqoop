# Installing on Docker

## Installing with `qlooctl`

#### What you'll need

 1. [Docker](https://www.docker.com/)
 1. [Docker-Compose](https://docs.docker.com/compose/)
 2. [qlooctl](https://github.com/solo-io/qloo/releases)

#### Install

 Run the following command to download the config files necessary for docker-compose to a local directory:

```bash
glooctl install docker [path]
```


`cd` to this directory to begin working with QLoo.


#### Running QLoo

Start QLoo by simply running `docker-compose up`

```bash
cd [folder]

docker-compose up
```
## Installing without `qlooctl`

#### What you'll need

 1. [Docker](https://www.docker.com/)
 1. [Docker-Compose](https://docs.docker.com/compose/))

#### Install

 `git clone` the QLoo repository:
 
```bash
git clone https://github.com/solo-io/qloo.git
``` 
 
 Initialize the storage directories and `qlooctl` configuration for QLoo:
 
```bash
cd qloo/install/docker-compose
./prepare-config-directories.sh
```

#### Running QLoo

Start QLoo by simply running `docker-compose up`

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
708445e2825c        soloio/qloo:0.1.1                  "/qloo --storage.typ…"   7 seconds ago       Up 4 seconds        0.0.0.0:9090->9090/tcp                             docker-compose_qloo_1
90645fff651e        soloio/control-plane:0.4.0        "/control-plane --st…"   31 hours ago        Up 5 seconds        0.0.0.0:8081->8081/tcp                             docker-compose_control-plane_1
36bb7f23808c        soloio/envoy:0.4.0                "envoy -c /config/en…"   31 hours ago        Up 6 seconds        0.0.0.0:8080->8080/tcp, 0.0.0.0:19000->19000/tcp   docker-compose_proxy_1
7747741da1df        soloio/function-discovery:0.4.0   "/function-discovery…"   31 hours ago        Up 6 seconds                                                           docker-compose_function-discovery_1```
```

Everything should be up and running. If this process does not work, please [open an issue](https://github.com/solo-io/gloo/issues/new). We are happy to answer questions on our diligently staffed [Slack channel](https://slack.solo.io)

See [Getting Started on Docker](../getting_started/docker/1.md) to get started creating your first GraphQL endpoint with QLoo.