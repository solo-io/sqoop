---
title: "sqoopctl install kube"
weight: 5
---
## sqoopctl install kube

install sqoop on kubernetes

### Synopsis

requires kubectl to be installed

```
sqoopctl install kube [flags]
```

### Options

```
  -d, --dry-run            Dump the raw installation yaml instead of applying it to kubernetes
  -h, --help               help for kube
  -n, --namespace string   which namespace to install sqoop into (default "gloo-system")
```

### Options inherited from parent commands

```
  -f, --file string     file to be read or written to
  -i, --interactive     interactive mode
  -o, --output string   output format: (yaml, json, table)
```

### SEE ALSO

* [sqoopctl install](../sqoopctl_install)	 - install sqoop on different platforms (into gloo-system namespace by default)

