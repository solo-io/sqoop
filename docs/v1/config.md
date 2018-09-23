<a name="top"></a>

## Contents
  - [Config](#sqoop.api.v1.Config)



<a name="config"></a>
<p align="right"><a href="#top">Top</a></p>




<a name="sqoop.api.v1.Config"></a>

### Config
Config is a top-level config object. It is used internally by Sqoop as a container for the entire set of config objects.


```yaml
schemas: [{Schema}]
resolver_maps: [{ResolverMap}]

```
| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| schemas | [Schema](schema.md#sqoop.api.v1.Schema) | repeated | the set of all schemas defined by the user |
| resolver_maps | [ResolverMap](resolver_map.md#sqoop.api.v1.ResolverMap) | repeated | the set of all resolver maps defined by the user |





 

 

 

