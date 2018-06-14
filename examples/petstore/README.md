# Getting Started on Kubernetes

### What you'll need
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [`qlooctl`](https://github.com/solo-io/qloo)
- [`glooctl`](https://github.com/solo-io/gloo): (OPTIONAL) to see how QLoo is interacting with the underlying system
- Kubernetes v1.8+ deployed somewhere. [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) is a great way to get a cluster up quickly.



### Steps

#### Deploy QLoo and Gloo

        qlooctl install kube


####  Deploy the Pet Store

        kubectl apply \
          -f https://raw.githubusercontent.com/solo-io/gloo/master/example/petstore/petstore.yaml

#### OPTIONAL: View the petstore functions using `glooctl`:

        glooctl upstream get
        
        +--------------------------------+------------+----------+-------------+
        |              NAME              |    TYPE    |  STATUS  |  FUNCTION   |
        +--------------------------------+------------+----------+-------------+
        | default-petstore-8080          | kubernetes | Accepted | addPet      |
        |                                |            |          | deletePet   |
        |                                |            |          | findPetById |
        |                                |            |          | findPets    |
        | gloo-system-control-plane-8081 | kubernetes | Accepted |             |
        | gloo-system-ingress-8080       | kubernetes | Accepted |             |
        | gloo-system-ingress-8443       | kubernetes | Accepted |             |
        +--------------------------------+------------+----------+-------------+

The upstream we want to see is `default-petstore-8080`. The functions `addPet`, `deletePet`, `findPetById`, and `findPets`
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

Upload the schema to QLoo using `qlooctl`:

```bash
qlooctl schema create petstore -f petstore.graphql
```


#### OPTIONAL: View the Generated Resolvers

A QLoo [**ResolverMap**](https://qloo.solo.io/v1/resolver_map/) will have been generated
for the new schema.

Take a look at its structure:

```bash
qlooctl resolvermap get petstore-resolvers -o yaml

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
      status: {}
  Query:
    fields:
      pet: {}
      pets: {}
```

The empty `{}`'s are QLoo [**Resolver**](https://qloo.solo.io/v1/resolver_map/#qloo.api.v1.Resolver)
objects, waiting to be filled in. QLoo supports a variety of Resolver types (and supports extension to its
resolution system). In this tutorial, we will create Gloo resolvers, which allow you to connect schema fields
to REST APIs, serverless functions and other Gloo functions. 
 
#### Register some Resolvers

Let's use `qlooctl` to register some resolvers.

```bash
# register findPetById for Query.pets (specifying no arguments)
qlooctl resolvermap register -u default-petstore-8080 -f findPetById Query pets
# register a resolver for Query.pet
qlooctl resolvermap register -u default-petstore-8080 -f findPetById Query pet
# register a resolver for Mutation.addPet
# the request template tells QLoo to use the Variable "pet" as an argument 
qlooctl resolvermap register -u default-petstore-8080 -f addPet Mutation addPet --request-template '{{ marshal (index .Args "pet") }}'
```

That's it! Now we should have a functioning GraphQL frontend for our REST service.

#### Visit the Playground

Visit the exposed address of the `qloo` service in your browser.

If you're running in minkube, you can get this address with the command

```bash
echo http://$(minikube ip):$(kubectl get svc qloo -n gloo-system -o 'jsonpath={.spec.ports[?(@.name=="http")].nodePort}')

http://192.168.39.47:30935/
```

You should see a landing page for QLoo which contains a link to the GraphQL Playground for our
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
