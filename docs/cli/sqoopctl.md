---
title: "sqoopctl"
weight: 5
---
## sqoopctl

Interact with Sqoop's storage API from the command line

### Synopsis

As Sqoop features a storage-based API, direct communication with the Sqoop server is not necessary. sqoopctl simplifies the administration of Sqoop by providing an easy way to create, read, update, and delete Sqoop storage objects.

The primary concerns of sqoopctl are Schemas and ResolverMaps. Schemas contain your GraphQL schema; ResolverMaps define how your schema fields are resolved.

Start by creating a schema using sqoopctl schema create --from-file <path/to/your/graphql/schema>

```
sqoopctl [flags]
```

### Options

```
  -f, --file string     file to be read or written to
  -h, --help            help for sqoopctl
  -i, --interactive     interactive mode
  -o, --output string   output format: (yaml, json, table)
```

### SEE ALSO

* [sqoopctl install](../sqoopctl_install)	 - install gloo on different platforms
* [sqoopctl resolvermap](../sqoopctl_resolvermap)	 - 
* [sqoopctl schema](../sqoopctl_schema)	 - 
* [sqoopctl uninstall](../sqoopctl_uninstall)	 - uninstall gloo

