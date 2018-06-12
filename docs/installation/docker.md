# Installing on Docker

## Installing with `qlooctl`

### What you'll need

 1. [Docker](https://www.docker.com/)
 1. [Docker-Compose](https://docs.docker.com/compose/)
 2. [qlooctl](https://github.com/solo-io/qloo/releases)

## Install

 Run the following command to download the config files necessary for docker-compose to a local directory:

 ```
 glooctl install docker [path]
 ```


`cd` to this directory to begin working with QLoo.


## Running QLoo

Start QLoo by simply running `docker-compose up`

```
cd [folder]

docker-compose up
```

You can check gloo services are running by running `docker ps`

For example,

```
docker ps

CONTAINER ID        IMAGE                             COMMAND                  CREATED             STATUS              PORTS                                                                      NAMES
6aea10e3ac4e        soloio/function-discovery:0.2.5   "/function-discovery…"   49 seconds ago      Up 46 seconds                                                                                  gloo-tutorial_function-discovery_1
d42a9dc94275        soloio/envoy:0.2.27           "envoy -c /config/en…"   49 seconds ago      Up 47 seconds       0.0.0.0:8080->8080/tcp, 0.0.0.0:8443->8443/tcp, 0.0.0.0:19000->19000/tcp   gloo-tutorial_ingress_1
1a37c031adf3        soloio/control-plane:0.2.5        "/control-plane --st…"   49 seconds ago      Up 46 seconds       0.0.0.0:8081->8081/tcp                                                     gloo-tutorial_control-plane_1
```


Everything should be up and running. If this process does not work, please [open an issue](https://github.com/solo-io/gloo/issues/new). We are happy to answer questions on our diligently staffed [Slack channel](https://slack.solo.io)

See [Getting Started on Docker](../getting_started/docker/1.md) to get started creating routes with Gloo.