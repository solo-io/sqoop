#!/usr/bin/env bash

alias qlooctl=${PWD}/_output/qlooctl

echo Installing Gloo
k apply -f ${GOPATH}/src/github.com/solo-io/gloo/install/kube/install.yaml

echo Installing QLoo
k apply -f ${GOPATH}/src/github.com/solo-io/qloo/install/kube/install.yaml

echo Deploying Petstore
k apply -f ${GOPATH}/src/github.com/solo-io/gloo/example/petstore/petstore.yaml

qlooctl schema create -f examples/petstore/petstore.schema.graphql petstore

qlooctl resolvermap register -u default-petstore-8080 -f findPetById Query pets
qlooctl resolvermap register -u default-petstore-8080 -f findPetById Query pet --request-template '{{ marshal .Args }}'
qlooctl resolvermap register -u default-petstore-8080 -f addPet Mutation addPet --request-template '{{ marshal (index .Args "pet") }}'
