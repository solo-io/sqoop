### What you'll need
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [`qlooctl`](https://github.com/solo-io/qloo)
- [`glooctl`](https://github.com/solo-io/gloo): (OPTIONAL) to see how QLoo is interacting with the underlying system
- Kubernetes v1.8+ deployed somewhere. [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) is a great way to get a cluster up quickly.



<br/>

### Steps

1. QLoo and Gloo deployed and running on Kubernetes:

        qlooctl install kube

 
1. Next, deploy the Pet Store app to kubernetes:

        kubectl apply \
          -f https://raw.githubusercontent.com/solo-io/gloo/master/example/petstore/petstore.yaml

1. OPTIONAL: Verify the petstore service and its functions were discovered by Gloo, using `glooctl`:

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

    The upstream we want to see is `default-petstore-8080`. Digging a little deeper,
    we can verify that Gloo's function discovery populated our upstream with 
    the available rest endpoints it implements. Note: the upstream was created in 
    the `gloo-system` namespace rather than `default` because it was created by a
    discovery service. Upstreams and virtualservices do not need to live in the `gloo-system`
    namespace to be processed by Gloo.

    <br/>

1. The Petstore implements a Swagger-based REST API. Let's create a GraphQL Schema that contains some queries we
can run using the Petstore as our data source.

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


1. Let's now use `glooctl` to create a route for this upstream.

        glooctl route create \
          --path-exact /petstore/list \
          --upstream default-petstore-8080 \
          --prefix-rewrite /api/pets

    We need the `--prefix-rewrite` flag so Envoy knows to change the path on the outgoing request
    to the path our petstore expects. 

    With `glooctl`, we can see that a virtual service was created with our route:

        glooctl virtualservice get -o yaml
        
        metadata:
          namespace: gloo-system
          resource_version: "3052"
        name: default
        routes:
        - request_matcher:
            path_exact: /petstore/list
          single_destination:
            upstream:
              name: default-petstore-8080
        status:
          state: Accepted

1. Let's test the route `/petstore/list` using `curl`:

        export GATEWAY_URL=http://$(minikube ip):$(kubectl get svc ingress -n gloo-system -o 'jsonpath={.spec.ports[?(@.name=="http")].nodePort}')

        curl ${GATEWAY_URL}/petstore/list
        
        [{"id":1,"name":"Dog","status":"available"},{"id":2,"name":"Cat","status":"pending"}]
        
        
Great! our gateway is up and running. Let's make things a bit more sophisticated in the next section with [Function Routing](2.md).
