<a name="top"></a>

## Contents
  - [Config](#qloo.api.v1.Config)



<a name="config"></a>
<p align="right"><a href="#top">Top</a></p>




<a name="qloo.api.v1.Config"></a>

### Config
Config is a top-level config object. It is used internally by QLoo as a container for the entire set of config objects.


```yaml
schemas: [{Schema}]
resolver_maps: [{ResolverMap}]

```
| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| schemas | [Schema](schema.md#qloo.api.v1.Schema) | repeated | the set of all schemas defined by the user |
| resolver_maps | [ResolverMap](resolver_map.md#qloo.api.v1.ResolverMap) | repeated | the set of all resolver maps defined by the user |





 

 

 

