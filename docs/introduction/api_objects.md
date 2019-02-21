---
weight: 3
title: Api Objects
---


# ResolverMaps

### Storage-Based API

Sqoop, like [Gloo](https://gloo.solo.io), features a storage-based API. Inspired by Kubernetes, Sqoop's API is accessed 
by applications and users by reading and writing API objects to a storage layer Sqoop is configured (at boot-time) to monitor
for changes. Currently supported storage backends are [Kubernetes CRDs](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/), 
[Consul Key-Value Pairs](https://www.consul.io/), or Sqoop's local filesystem. 


### API Objects

Sqoop's API Objects take two forms:

0. [Schemas](../../v1/schema.md)
    * Schemas are made up of three pieces of information:
      - A name for the schema. This can be anything, but must be uniquee
      - An inline string containing the entire [GraphQL Schema](https://graphql.org/learn/schema/)
      - The name of a ResolverMap object which contains Sqoop-specific instructions
      on how to resolve the fields of the schema. 
      - If the user leaves this empty,
      Sqoop will attempt to generate an empty ResolverMap skeleton for the user, 
      which the user can edit using `sqoopctl`.

    * GraphQL Schemas can be uploaded to Sqoop using `sqoopctl`

1. [ResolverMaps](../../v1/resolver_map.md)
    * ResolverMaps represent a mapping between the fields in a GraphQL schema 
    and the [Resolvers](resolvers.md) that Sqoop will use to resolve them.
    
    * Resolvers define the action Sqoop will perform when executing a GraphQL Query. Sqoop leverages
    [Gloo's function registry](https://gloo.solo.io/introduction/concepts/#Functions) to generate resolvers, 
    allowing users to define GraphQL resolvers using configuration rather than code. 
    Read more about GraphQL Resolvers here: https://graphql.org/learn/execution/   
    
    * An example ResolverMap might look like the following:
                       
            name: starwars-resolvers
            types:
            Droid:
              fields:
                appearsIn:
                  template_resolver:
                    inline_template: '{{ index .Parent "appears_in" }}}'
                friends:
                  gloo_resolver:
                    request_template: '{{ marshal (index .Parent "friend_ids") }}'
                    single_function:
                      function: GetCharacters
                      upstream: starwars-rest
            Human:
              fields:
                appearsIn:
                  template_resolver:
                    inline_template: '{{ index .Parent "appears_in" }}}'
                friends:
                  gloo_resolver:
                    request_template: '{{ marshal (index .Parent "friend_ids") }}'
                    single_function:
                      function: GetCharacters
                      upstream: starwars-rest
            Query:
              fields:
                droid:
                  gloo_resolver:
                    request_template: '{"id": {{ index .Args "id" }}}'
                    single_function:
                      function: GetCharacter
                      upstream: starwars-rest
                hero:
                  gloo_resolver:
                    single_function:
                      function: GetHero
                      upstream: starwars-rest
                human:
                  gloo_resolver:
                    request_template: '{"id": {{ index .Args "id" }}}'
                    single_function:
                      function: GetCharacter
                      upstream: starwars-rest
                      
    Here we have defined resolvers for the fields `Query.droid`,
    `Query.hero`, `Query.human`, `Human.friends`, `Human.appearsIn`, 
    `Droid.friends`, and `Droid.appearsIn`.
    
