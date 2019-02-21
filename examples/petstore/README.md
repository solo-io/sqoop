# Getting Started on Kubernetes

### What you'll need
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [`sqoopctl`](https://github.com/solo-io/sqoop)
- [`glooctl`](https://github.com/solo-io/gloo): (OPTIONAL) to see how Sqoop is interacting with the underlying system
- Kubernetes v1.8+ deployed somewhere. [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) is a great way to get a cluster up quickly.

This tutorial will install sqoop into the namespace `gloo-system` by default, this is configurable from the `sqoopctl` cli.

### Steps

#### Deploy Sqoop and Gloo

        sqoopctl install kube


####  Deploy the Pet Store

        kubectl apply -n gloo-system -f petstore.yaml

#### OPTIONAL: View the petstore functions using `glooctl`:

        glooctl upstream get
        
        +--------------------------------+------------+----------+-------------+
        |              NAME              |    TYPE    |  STATUS  |  FUNCTION   |
        +--------------------------------+------------+----------+-------------+
        | gloo-system-petstore-8080	     | kubernetes | Accepted | addPet      |
        |                                |            |          | deletePet   |
        |                                |            |          | findPetById |
        |                                |            |          | findPets    |
        +--------------------------------+------------+----------+-------------+

The upstream we want to see is `gloo-system-petstore-8080`. The functions `addPet`, `deletePet`, `findPetById`, and `findPets`
will become the resolvers for our GraphQL schema.  

##### Alternatively: find the upstreams using `kubectl`
```bash
kubectl get upstreams -n gloo-system

NAME                                                    AGE
gloo-system-gloo-9977                              1h
gloo-system-petstore-8080                          1h
gloo-system-sqoop-9090                             1h
```

The upstream we are interested in is the petstore, so we run the following to find the functions:

```bash
kubectl get upstreams -n gloo-system gloo-system-petstore-8080 -oyaml

apiVersion: gloo.solo.io/v1
kind: Upstream
metadata:
  labels:
    discovered_by: kubernetesplugin
    service: petstore
  name: gloo-system-petstore-8080
  namespace: gloo-system
spec:
  upstreamSpec:
    kube:
      selector:
        app: petstore
      serviceName: petstore
      serviceNamespace: gloo-system
      servicePort: 8080
      serviceSpec:
        rest:
          swaggerInfo:
            urlå: http://petstore.gloo-system.svc.cluster.local:8080/swagger.json
          transformations:
            addPet:
              body:
                text: '{"id": {{ default(id, "") }},"name": "{{ default(name, "")}}","tag":
                  "{{ default(tag, "")}}"}'
              headers:
                :method:
                  text: POST
                :path:
                  text: /api/pets
                content-type:
                  text: application/json
            deletePet:
              headers:
                :method:
                  text: DELETE
                :path:
                  text: /api/pets/{{ default(id, "") }}
                content-type:
                  text: application/json
            findPetById:
              body: {}
              headers:
                :method:
                  text: GET
                :path:
                  text: /api/pets/{{ default(id, "") }}
                content-length:
                  text: "0"
                content-type: {}
                transfer-encoding: {}
            findPets:
              body: {}
              headers:
                :method:
                  text: GET
                :path:
                  text: /api/pets?tags={{default(tags, "")}}&limit={{default(limit,
                    "")}}
                content-length:
                  text: "0"
                content-type: {}
                transfer-encoding: {}

```


#### Create a GraphQL Schema

An example schema is located in `petstore.schema.graphql`

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
}

input InputPet{
    id: ID!
    name: String!
    tag: String
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
  namespace: gloo-system
  resource_version: "573676"
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
sqoopctl resolvermap register -u gloo-system-petstore-8080 -f findPetById Query pets
# register a resolver for Query.pet
sqoopctl resolvermap register -u gloo-system-petstore-8080 -f findPetById Query pet
# register a resolver for Mutation.addPet
# the request template tells Sqoop to use the Variable "pet" as an argument 
sqoopctl resolvermap register -u gloo-system-petstore-8080 -f addPet Mutation addPet --request-template '{{ marshal (index .Args "pet") }}'
```

That's it! Now we should have a functioning GraphQL frontend for our REST service.

#### Visit the Playground

Visit the exposed address of the `sqoop` service in your browser.

If you're running in minkube, you can get this address with the command

```bash
echo http://$(minikube ip):$(kubectl get svc sqoop -n gloo-system -o 'jsonpath={.spec.ports[?(@.name=="http")].nodePort}')

http://192.168.39.47:30935/
```

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
