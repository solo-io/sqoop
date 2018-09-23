#!/usr/bin/env bash

alias sqoopctl=${PWD}/_output/sqoopctl

echo Installing Sqoop
k apply -f ${GOPATH}/src/github.com/solo-io/sqoop/install/kube/install.yaml

echo Deploying Petstore
k apply -f ${GOPATH}/src/github.com/solo-io/gloo/example/petstore/petstore.yaml

sqoopctl schema create -f examples/petstore/petstore.schema.graphql petstore

sqoopctl resolvermap register -u default-petstore-8080 -f findPetById Query pets
sqoopctl resolvermap register -u default-petstore-8080 -f findPetById Query pet --request-template '{{ marshal .Args }}'
sqoopctl resolvermap register -u default-petstore-8080 -f addPet Mutation addPet --request-template '{{ marshal (index .Args "pet") }}'
