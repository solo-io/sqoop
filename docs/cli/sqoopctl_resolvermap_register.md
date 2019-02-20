---
title: "sqoopctl resolvermap register"
weight: 5
---
## sqoopctl resolvermap register

Register a resolver for a field in your Schema

### Synopsis

Sets the resolver for a field in your schema. TypeName.FieldName will always be resolved using this resolver 
Resolvers must be defined in yaml format. 
See the documentation at https://sqoop.solo.io/v1/resolver_map/#sqoop.api.v1.Resolver for the API specification for Sqoop Resolvers

```
sqoopctl resolvermap register TypeName FieldName -f resolver.yaml [-s schema-name] [flags]
```

### Options

```
  -g, --function string            function to use as resolver
  -h, --help                       help for register
  -b, --request-template string    template to use for the request body
  -r, --response-template string   template to use for the response body
  -s, --schema string              name of the schema to connect this resolver to. this is required if more than one schema contains a definition for the type name.
  -u, --upstream string            upstream where the function lives
```

### Options inherited from parent commands

```
  -f, --file string     file to be read or written to
  -i, --interactive     interactive mode
  -o, --output string   output format: (yaml, json, table)
```

### SEE ALSO

* [sqoopctl resolvermap](../sqoopctl_resolvermap)	 - 

