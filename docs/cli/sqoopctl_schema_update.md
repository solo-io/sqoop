---
title: "sqoopctl schema update"
weight: 5
---
## sqoopctl schema update

upload a schema to Sqoop from a local GraphQL Schema file

### Synopsis

upload a schema to Sqoop from a local GraphQL Schema file

```
sqoopctl schema update NAME -f <path/to/your/graphql/schema> [flags]
```

### Options

```
  -h, --help   help for update
```

### Options inherited from parent commands

```
  -f, --file string        file to be read or written to
  -i, --interactive        interactive mode
      --name string        name of the resource to read or write
  -n, --namespace string   namespace for reading or writing resources (default "gloo-system")
  -o, --output string      output format: (yaml, json, table)
```

### SEE ALSO

* [sqoopctl schema](../sqoopctl_schema)	 - 

