---
title: "sqoopctl schema"
weight: 5
---
## sqoopctl schema

interacting with sqoop schema resources

### Synopsis

interacting with sqoop schema resources

### Options

```
  -h, --help               help for schema
      --name string        name of the resource to read or write
  -n, --namespace string   namespace for reading or writing resources (default "gloo-system")
```

### Options inherited from parent commands

```
  -f, --file string     file to be read or written to
  -i, --interactive     interactive mode
  -o, --output string   output format: (yaml, json, table)
```

### SEE ALSO

* [sqoopctl](../sqoopctl)	 - Interact with Sqoop's storage API from the command line. 
For more information, visit https://sqoop.solo.io.
* [sqoopctl schema create](../sqoopctl_schema_create)	 - upload a schema to Sqoop from a local GraphQL Schema file
* [sqoopctl schema delete](../sqoopctl_schema_delete)	 - delete a schema by its name
* [sqoopctl schema update](../sqoopctl_schema_update)	 - upload a schema to Sqoop from a local GraphQL Schema file

