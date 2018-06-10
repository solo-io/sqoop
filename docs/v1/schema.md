<a name="top"></a>

## Contents
  - [Schema](#qloo.api.v1.Schema)



<a name="schema"></a>
<p align="right"><a href="#top">Top</a></p>




<a name="qloo.api.v1.Schema"></a>

### Schema
The Schema object wraps the user&#39;s GraphQL Schema, which is stored as an inline string.
The Schema Object contains a Status field which is used by QLoo to validate the user&#39;s input schema.


```yaml
name: string
resolver_map: string
inline_schema: string
status: {gloo.api.v1.Status}
metadata: {gloo.api.v1.Metadata}

```
| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of the schema. Schema names must be unique

Schema Names must be unique and follow the following syntax rules: One or more lowercase rfc1035/rfc1123 labels separated by &#39;.&#39; with a maximum length of 253 characters. |
| resolver_map | string |  | name of the resolver map to use to resolve this schema. if the user leaves this empty, QLoo will generate the skeleton of a resolver map for the user |
| inline_schema | string |  | inline the entire graphql schema as a string here |
| status | [gloo.api.v1.Status](schema.md#gloo.api.v1.Status) |  | Status indicates the validation status of the role resource. Status is read-only by clients, and set by gloo during validation |
| metadata | [gloo.api.v1.Metadata](schema.md#gloo.api.v1.Metadata) |  | Metadata contains the resource metadata for the role |





 

 

 

