---
weight: 2
title: Docker
---

# Getting started on Docker

#### What you'll need

 1. [Docker](https://www.docker.com/)
 1. [Docker-Compose](https://docs.docker.com/compose/)
 1. [sqoopctl](https://github.com/solo-io/sqoop/releases)
 1. [glooctl](https://github.com/solo-io/glooctl/releases) (optional)



### Steps

#### Deploy Sqoop and Gloo

    sqoopctl install docker sqoop-docker
    cd ./sqoop-docker
    docker-compose up

or

    git clone https://github.com/solo-io/sqoop
    cd sqoop/install/docker-compose
    ./prepare-config-directories
    docker-compose up


####  Deploy the Pet Store

    docker run -d -p 1234:8080 soloio/petstore-example:latest

#### Create a Gloo upstream for the petstore

  * using `glooctl`:
  
```bash
cat << EOF | glooctl upstream create -f -
name: petstore
type: static
spec:
  hosts:
  # gateway ip for the docker network
  - addr: $(docker inspect sqoop-docker_default -f '{{ (index .IPAM.Config 0).Gateway }}')
    port: 1234
EOF
```




  * writing directly to disk

```bash
cat > ./_gloo_config/upstreams/petstore.yaml << EOF 
name: petstore
type: static
spec:
  hosts:
  # gateway ip for the docker network
  - addr: $(docker inspect sqoop-docker_default -f '{{ (index .IPAM.Config 0).Gateway }}')
    port: 1234
EOF
```


#### OPTIONAL: View the petstore functions using `glooctl`:

        glooctl upstream get
        
        +----------+---------+--------+-------------+
        |   NAME   |  TYPE   | STATUS |  FUNCTION   |
        +----------+---------+--------+-------------+
        | petstore | static |        | addPet      |
        |          |         |        | deletePet   |
        |          |         |        | findPetById |
        |          |         |        | findPets    |
        +----------+---------+--------+-------------+

The upstream we want to see is `petstore`. The functions `addPet`, `deletePet`, `findPetById`, and `findPets`
will become the resolvers for our GraphQL schema.  


#### Create a GraphQL Schema

Copy and paste the following schema into `petstore.graphql` (or wherever you like):

```graphql
# The query type, represents all of the entry points into our object graph
type Query {
    pets: [Pet]
    pet(id: Int!): Pet
}

type Mutation {
    addPet(pet: InputPet!): Pet
}

type Pet{
    id: ID!
    name: String!
    status: Status!
}

input InputPet{
    id: ID!
    name: String!
    tag: String
}

enum Status {
    pending
    available
}
```   

#### Upload the Schema

Upload the schema to Sqoop using `sqoopctl`:

```bash
sqoopctl schema create petstore -f petstore.graphql
```


#### OPTIONAL: View the Generated Resolvers

A Sqoop [**ResolverMap**](https://sqoop.solo.io/v1/resolver_map/) will have been generated
for the new schema.

Take a look at its structure:

```bash
sqoopctl resolvermap get petstore-resolvers -o yaml

metadata:
  resource_version: "1"
name: petstore-resolvers
status:
  state: Accepted
types:
  Mutation:
    fields:
      addPet: {}
  Pet:
    fields:
      id: {}
      name: {}
      status: {}
  Query:
    fields:
      pet: {}
      pets: {}
```

The empty `{}`'s are Sqoop [**Resolver**](https://sqoop.solo.io/v1/resolver_map/#sqoop.api.v1.Resolver)
objects, waiting to be filled in. Sqoop supports a variety of Resolver types (and supports extension to its
resolution system). In this tutorial, we will create Gloo resolvers, which allow you to connect schema fields
to REST APIs, serverless functions and other Gloo functions. 
 
#### Register some Resolvers

Let's use `sqoopctl` to register some resolvers.

```bash
# register findPetById for Query.pets (specifying no arguments)
sqoopctl resolvermap register -u petstore -f findPetById Query pets
# register a resolver for Query.pet
sqoopctl resolvermap register -u petstore -f findPetById Query pet
# register a resolver for Mutation.addPet
# the request template tells Sqoop to use the Variable "pet" as an argument 
sqoopctl resolvermap register -u petstore -f addPet Mutation addPet --request-template '{{ marshal (index .Args "pet") }}'
```

*Note*: if you get a `permission denied` error, run
```bash
sudo chown -R $USER _gloo_config
sudo chgrp -R $USER _gloo_config
``` 

That's it! Now we should have a functioning GraphQL frontend for our REST service.

#### Visit the Playground

Visit the Sqoop UI from your browser: http://localhost:9090/

You should see a landing page for Sqoop which contains a link to the GraphQL Playground for our
Pet Store. Visit it and try out some queries!

examples:

```graphql
{
  pet(id:1 ) {
    name
  }
}
```

&darr;

```json
{
  "data": {
    "pet": {
      "name": "Dog"
    }
  }
}
```

```graphql
{
  pets {
    name
  }
}
```

&darr;

```json
{
  "data": {
    "pets": [
      {
        "name": "Dog"
      },
      {
        "name": "Cat"
      }
    ]
  }
}
```
```graphql
mutation($pet: InputPet!) {
  addPet(pet: $pet) {
    id
    name
  }
}
```
with input variable
````json
{
  "pet":{
    "id":3,
    "name": "monkey"
  }
}
````

&darr;

```json
{
  "data": {
    "addPet": {
      "name": "monkey"
    }
  }
}
```
