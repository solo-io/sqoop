---
weight: 3
title: Resolvers
---


# Gloo Resolvers

Gloo resolvers are the primary means of resolving schema fields using Sqoop. This 
document explains the structure of a Gloo resolver and how to write one.

## Request Templates

Let's take a look at the structure of a Gloo resolver (usually written by users as YAML,
stored by Sqoop as Proto).

```yaml
gloo_resolver:
  request_template: '{{ marshal (index .Parent "pet_ids") }}'
  response_template: '{{ "success" }}'
  single_function:
    upstream: petstore
    function: ListPets
```

* `request_template` is optional. If provided, Sqoop will use the provided template
to construct the request body sent to the resolver function.

Request templates follow the conventions of [Go templates](https://golang.org/pkg/text/template/).

Available parameters for use in Request Templates come from the 
[`Params`](https://github.com/solo-io/sqoop/blob/master/pkg/exec/executable_resolvers.go) object.

The Params Object has the following structure (defined in Go):

```go
type Params struct {
	Args   map[string]interface{}
	Parent map[string]interface{}
}
```

`Args` represent arguments that were passed to Sqoop as part of the Client Query.

`Parent` represents the root object the field under query belongs to. `Parent` 
is `nil` for root types (`Query` and `Mutation` type).

The `marshal` function is available for use in Sqoop templates. 
`marshal` will encode any value into JSON.

Here's an example of a Gloo Resolver using multiple destinations, with load balancing:

```yaml
gloo_resolver:
  request_template: '{{ marshal (index .Parent "friend_ids") }}'
  response_template: '{{ marshal (index .Parent "friend_ids") }}'
  multi_function:
    weighted_functions:
    - weight: 1
      function: 
        upstream: petstore-v1
        function: ListPets
    - weight: 1
      function: 
        upstream: petstore-v2
        function: ListPets
```

## Response Templates
Response templates also use Go template syntax. Response templates can refer to 
(sub)fields of the response body, provided that it is JSON-encoded.