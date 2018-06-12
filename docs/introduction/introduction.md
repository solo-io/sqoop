# Introduction

![Overview](high_level_architecture.png "High Level Architecture")


### What is QLoo?

QLoo is a GraphQL Server built on top of [Gloo](https://github.com/solo-io/gloo) and the [Envoy Proxy](https://envoyproxy.io).

QLoo leverages Gloo's function registry and Envoy's advanced HTTP routing features to provide a GraphQL frontend
for REST/gRPC applications and serverless functions. QLoo routes requests to data sources via Envoy, leveraging 
Envoy [HTTP filters](https://www.envoyproxy.io/docs/envoy/latest/api-v2/config/filter/filter.html?highlight=http%20filter) 
for security, load balancing, and more.

QLoo makes HTTP requests through Gloo to invoke service endpoints and serverless functions through Gloo. QLoo users
import their GraphQL Schemas and attach **Gloo functions** to the fields of their schemas. QLoo uses Gloo functions to generate
its own resolvers, allowing users to get a fully-functional GraphQL frontend for their serverless functions and services 
without writing any code. This is why we call QLoo the **Codeless GraphQL Server**.


### Using QLoo

Compared to typical GraphQL implementations, QLoo's configuration API is quite simple. Configuration takes two steps:

0. Importing GraphQL schemas into QLoo (done most easily via `qlooctl`).
0. Attaching **Gloo functions** to schema fields by defining [resolvers](concepts/resolvers.md) in a [ResolverMap](concepts/api_objects.md). 

QLoo will execute GraphQL queries (and mutations) via port `:9090` by default.

