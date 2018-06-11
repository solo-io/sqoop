package exec

import (
	"bytes"
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/dynamic"
	"github.com/vektah/gqlgen/graphql"
	"github.com/vektah/gqlgen/neelance/introspection"
	"github.com/vektah/gqlgen/neelance/query"
	"github.com/vektah/gqlgen/neelance/schema"
)

func NewExecutableSchema(parsedSchema *schema.Schema, resolvers *ExecutableResolverMap) graphql.ExecutableSchema {
	return &executableSchema{schema: parsedSchema, resolvers: resolvers}
}

type executableSchema struct {
	schema    *schema.Schema
	resolvers *ExecutableResolverMap
}

func (e *executableSchema) Schema() *schema.Schema {
	return e.schema
}

func (e *executableSchema) Query(ctx context.Context, op *query.Operation) *graphql.Response {
	ec := executionContext{
		RequestContext: graphql.GetRequestContext(ctx),
		Schema:         e.Schema(),
		resolvers:      e.resolvers,
	}

	buf := ec.RequestMiddleware(ctx, func(ctx context.Context) []byte {
		data := ec._Query(ctx, op.Selections)
		var buf bytes.Buffer
		data.MarshalGQL(&buf)
		return buf.Bytes()
	})

	return &graphql.Response{
		Data:   buf,
		Errors: ec.Errors,
	}
}

func (e *executableSchema) Mutation(ctx context.Context, op *query.Operation) *graphql.Response {
	ec := executionContext{
		RequestContext: graphql.GetRequestContext(ctx),
		Schema:         e.Schema(),
		resolvers:      e.resolvers,
	}

	buf := ec.RequestMiddleware(ctx, func(ctx context.Context) []byte {
		data := ec._Mutation(ctx, op.Selections)
		var buf bytes.Buffer
		data.MarshalGQL(&buf)
		return buf.Bytes()
	})

	return &graphql.Response{
		Data:   buf,
		Errors: ec.Errors,
	}
}

func (e *executableSchema) Subscription(ctx context.Context, op *query.Operation) func() *graphql.Response {
	return graphql.OneShot(graphql.ErrorResponse(ctx, "subscriptions are not supported"))
}

type executionContext struct {
	*graphql.RequestContext
	*schema.Schema

	resolvers *ExecutableResolverMap
}

var queryImplementors = []string{"Query"}

// nolint: gocyclo, errcheck, gas, goconst
func (ec *executionContext) _Query(ctx context.Context, sel []query.Selection) graphql.Marshaler {
	fields := graphql.CollectFields(ec.Doc, sel, queryImplementors, ec.Variables)

	ctx = graphql.WithResolverContext(ctx, &graphql.ResolverContext{
		Object: "Query",
	})

	out := graphql.NewOrderedMap(len(fields))
	for i, field := range fields {
		out.Keys[i] = field.Alias

		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("Query")
		case "__schema":
			out.Values[i] = ec._Query___schema(ctx, field)
		case "__type":
			out.Values[i] = ec._Query___type(ctx, field)
		default:
			val, err := ec.resolveField(ctx, ec.EntryPoints["query"].(*schema.Object), field, nil)
			if err != nil {
				ec.Error(ctx, err)
				out.Values[i] = graphql.Null
				continue
			}
			out.Values[i] = val.Marshaller()
		}
	}

	return out
}

func (ec *executionContext) resolveField(ctx context.Context, objectType *schema.Object, field graphql.CollectedField, parentObject *dynamic.Object) (dynamic.Value, error) {
	val, err := ec.resolvers.Resolve(objectType, field.Name, Params{Parent: parentObject, Args: field.Args})
	if err != nil {
		return nil, errors.Wrapf(err, "executing resolver for field "+strconv.Quote(field.Name))
	}
	// lists and objects need to be recursed into
	switch result := val.(type) {
	case *dynamic.Object:
		val, err := ec.resolveObject(ctx, result.Object, field.Selections, result)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving object "+strconv.Quote(result.Name))
		}
		return val, nil
	case *dynamic.Array:
		for i, item := range result.Data {
			obj, ok := item.(*dynamic.Object)
			if !ok {
				// not an object array, nothing to recurse
				return val, nil
			}
			result.Data[i], err = ec.resolveObject(ctx, obj.Object, field.Selections, obj)
			if err != nil {
				return nil, errors.Wrapf(err, "resolving list item object "+strconv.Quote(obj.Name))
			}
		}
		return result, nil
	}
	return val, nil
}

func (ec *executionContext) resolveObject(ctx context.Context, objectType *schema.Object, sel []query.Selection, parentObject *dynamic.Object) (*dynamic.Object, error) {
	ctx = graphql.WithResolverContext(ctx, &graphql.ResolverContext{
		Object: objectType.TypeName(),
	})

	// get just the fields we need
	// also, resolve nested resolvers if they exist
	implementors := getImplementors(objectType)
	fields := graphql.CollectFields(ec.Doc, sel, implementors, ec.Variables)

	data := dynamic.NewOrderedMap()

	for _, field := range fields {
		switch field.Name {
		// request for the object's typeName
		case "__typename":
			val := &dynamic.String{
				Data: objectType.TypeName(),
			}
			data.Set(field.Name, val)
		default:
			val, err := ec.resolveField(ctx, objectType, field, parentObject)
			if err != nil {
				return nil, errors.Wrapf(err, "resolving field %v", strconv.Quote(field.Name))
			}
			data.Set(field.Name, val)
		}

	}

	return &dynamic.Object{Object: objectType, Data: data}, nil
}

func getImplementors(typ schema.NamedType) []string {
	implementors := []string{typ.TypeName()}
	switch typ := typ.(type) {
	case *schema.Object:
		for _, iface := range typ.Interfaces {
			implementors = append(implementors, iface.TypeName())
		}

	case *schema.Union:
		for _, implementor := range typ.PossibleTypes {
			implementors = append(implementors, implementor.TypeName())
		}

	case *schema.Interface:
		for _, implementor := range typ.PossibleTypes {
			implementors = append(implementors, implementor.TypeName())
		}

	default:
		panic(fmt.Errorf("unknown type %#v", typ))
	}
	return implementors
}

var mutationImplementors = []string{"Mutation"}

// nolint: gocyclo, errcheck, gas, goconst
func (ec *executionContext) _Mutation(ctx context.Context, sel []query.Selection) graphql.Marshaler {
	fields := graphql.CollectFields(ec.Doc, sel, mutationImplementors, ec.Variables)

	ctx = graphql.WithResolverContext(ctx, &graphql.ResolverContext{
		Object: "Mutation",
	})

	out := graphql.NewOrderedMap(len(fields))
	for i, field := range fields {
		out.Keys[i] = field.Alias

		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("Mutation")
		default:
			val, err := ec.resolveField(ctx, ec.EntryPoints["mutation"].(*schema.Object), field, nil)
			if err != nil {
				ec.Error(ctx, err)
				out.Values[i] = graphql.Null
				continue
			}
			out.Values[i] = val.Marshaller()
		}
	}

	return out
}

func (ec *executionContext) _Query___schema(ctx context.Context, field graphql.CollectedField) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "Query"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := ec.introspectSchema()
	if res == nil {
		return graphql.Null
	}
	return ec.___Schema(ctx, field.Selections, res)
}

func (ec *executionContext) _Query___type(ctx context.Context, field graphql.CollectedField) graphql.Marshaler {
	args := map[string]interface{}{}
	var arg0 string
	if tmp, ok := field.Args["name"]; ok {
		var err error
		arg0, err = graphql.UnmarshalString(tmp)
		if err != nil {
			ec.Error(ctx, err)
			return graphql.Null
		}
	}
	args["name"] = arg0
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "Query"
	rctx.Args = args
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := ec.introspectType(args["name"].(string))
	if res == nil {
		return graphql.Null
	}
	return ec.___Type(ctx, field.Selections, res)
}

var __DirectiveImplementors = []string{"__Directive"}

// nolint: gocyclo, errcheck, gas, goconst
func (ec *executionContext) ___Directive(ctx context.Context, sel []query.Selection, obj *introspection.Directive) graphql.Marshaler {
	fields := graphql.CollectFields(ec.Doc, sel, __DirectiveImplementors, ec.Variables)

	out := graphql.NewOrderedMap(len(fields))
	for i, field := range fields {
		out.Keys[i] = field.Alias

		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("__Directive")
		case "name":
			out.Values[i] = ec.___Directive_name(ctx, field, obj)
		case "description":
			out.Values[i] = ec.___Directive_description(ctx, field, obj)
		case "locations":
			out.Values[i] = ec.___Directive_locations(ctx, field, obj)
		case "args":
			out.Values[i] = ec.___Directive_args(ctx, field, obj)
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}

	return out
}

func (ec *executionContext) ___Directive_name(ctx context.Context, field graphql.CollectedField, obj *introspection.Directive) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Directive"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Name()
	return graphql.MarshalString(res)
}

func (ec *executionContext) ___Directive_description(ctx context.Context, field graphql.CollectedField, obj *introspection.Directive) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Directive"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Description()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

func (ec *executionContext) ___Directive_locations(ctx context.Context, field graphql.CollectedField, obj *introspection.Directive) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Directive"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Locations()
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			return graphql.MarshalString(res[idx1])
		}())
	}
	return arr1
}

func (ec *executionContext) ___Directive_args(ctx context.Context, field graphql.CollectedField, obj *introspection.Directive) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Directive"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Args()
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___InputValue(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

var __EnumValueImplementors = []string{"__EnumValue"}

// nolint: gocyclo, errcheck, gas, goconst
func (ec *executionContext) ___EnumValue(ctx context.Context, sel []query.Selection, obj *introspection.EnumValue) graphql.Marshaler {
	fields := graphql.CollectFields(ec.Doc, sel, __EnumValueImplementors, ec.Variables)

	out := graphql.NewOrderedMap(len(fields))
	for i, field := range fields {
		out.Keys[i] = field.Alias

		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("__EnumValue")
		case "name":
			out.Values[i] = ec.___EnumValue_name(ctx, field, obj)
		case "description":
			out.Values[i] = ec.___EnumValue_description(ctx, field, obj)
		case "isDeprecated":
			out.Values[i] = ec.___EnumValue_isDeprecated(ctx, field, obj)
		case "deprecationReason":
			out.Values[i] = ec.___EnumValue_deprecationReason(ctx, field, obj)
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}

	return out
}

func (ec *executionContext) ___EnumValue_name(ctx context.Context, field graphql.CollectedField, obj *introspection.EnumValue) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__EnumValue"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Name()
	return graphql.MarshalString(res)
}

func (ec *executionContext) ___EnumValue_description(ctx context.Context, field graphql.CollectedField, obj *introspection.EnumValue) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__EnumValue"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Description()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

func (ec *executionContext) ___EnumValue_isDeprecated(ctx context.Context, field graphql.CollectedField, obj *introspection.EnumValue) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__EnumValue"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.IsDeprecated()
	return graphql.MarshalBoolean(res)
}

func (ec *executionContext) ___EnumValue_deprecationReason(ctx context.Context, field graphql.CollectedField, obj *introspection.EnumValue) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__EnumValue"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.DeprecationReason()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

var __FieldImplementors = []string{"__Field"}

// nolint: gocyclo, errcheck, gas, goconst
func (ec *executionContext) ___Field(ctx context.Context, sel []query.Selection, obj *introspection.Field) graphql.Marshaler {
	fields := graphql.CollectFields(ec.Doc, sel, __FieldImplementors, ec.Variables)

	out := graphql.NewOrderedMap(len(fields))
	for i, field := range fields {
		out.Keys[i] = field.Alias

		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("__Field")
		case "name":
			out.Values[i] = ec.___Field_name(ctx, field, obj)
		case "description":
			out.Values[i] = ec.___Field_description(ctx, field, obj)
		case "args":
			out.Values[i] = ec.___Field_args(ctx, field, obj)
		case "type":
			out.Values[i] = ec.___Field_type(ctx, field, obj)
		case "isDeprecated":
			out.Values[i] = ec.___Field_isDeprecated(ctx, field, obj)
		case "deprecationReason":
			out.Values[i] = ec.___Field_deprecationReason(ctx, field, obj)
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}

	return out
}

func (ec *executionContext) ___Field_name(ctx context.Context, field graphql.CollectedField, obj *introspection.Field) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Field"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Name()
	return graphql.MarshalString(res)
}

func (ec *executionContext) ___Field_description(ctx context.Context, field graphql.CollectedField, obj *introspection.Field) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Field"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Description()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

func (ec *executionContext) ___Field_args(ctx context.Context, field graphql.CollectedField, obj *introspection.Field) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Field"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Args()
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___InputValue(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

func (ec *executionContext) ___Field_type(ctx context.Context, field graphql.CollectedField, obj *introspection.Field) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Field"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Type()
	if res == nil {
		return graphql.Null
	}
	return ec.___Type(ctx, field.Selections, res)
}

func (ec *executionContext) ___Field_isDeprecated(ctx context.Context, field graphql.CollectedField, obj *introspection.Field) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Field"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.IsDeprecated()
	return graphql.MarshalBoolean(res)
}

func (ec *executionContext) ___Field_deprecationReason(ctx context.Context, field graphql.CollectedField, obj *introspection.Field) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Field"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.DeprecationReason()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

var __InputValueImplementors = []string{"__InputValue"}

// nolint: gocyclo, errcheck, gas, goconst
func (ec *executionContext) ___InputValue(ctx context.Context, sel []query.Selection, obj *introspection.InputValue) graphql.Marshaler {
	fields := graphql.CollectFields(ec.Doc, sel, __InputValueImplementors, ec.Variables)

	out := graphql.NewOrderedMap(len(fields))
	for i, field := range fields {
		out.Keys[i] = field.Alias

		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("__InputValue")
		case "name":
			out.Values[i] = ec.___InputValue_name(ctx, field, obj)
		case "description":
			out.Values[i] = ec.___InputValue_description(ctx, field, obj)
		case "type":
			out.Values[i] = ec.___InputValue_type(ctx, field, obj)
		case "defaultValue":
			out.Values[i] = ec.___InputValue_defaultValue(ctx, field, obj)
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}

	return out
}

func (ec *executionContext) ___InputValue_name(ctx context.Context, field graphql.CollectedField, obj *introspection.InputValue) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__InputValue"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Name()
	return graphql.MarshalString(res)
}

func (ec *executionContext) ___InputValue_description(ctx context.Context, field graphql.CollectedField, obj *introspection.InputValue) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__InputValue"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Description()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

func (ec *executionContext) ___InputValue_type(ctx context.Context, field graphql.CollectedField, obj *introspection.InputValue) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__InputValue"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Type()
	if res == nil {
		return graphql.Null
	}
	return ec.___Type(ctx, field.Selections, res)
}

func (ec *executionContext) ___InputValue_defaultValue(ctx context.Context, field graphql.CollectedField, obj *introspection.InputValue) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__InputValue"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.DefaultValue()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

var __SchemaImplementors = []string{"__Schema"}

// nolint: gocyclo, errcheck, gas, goconst
func (ec *executionContext) ___Schema(ctx context.Context, sel []query.Selection, obj *introspection.Schema) graphql.Marshaler {
	fields := graphql.CollectFields(ec.Doc, sel, __SchemaImplementors, ec.Variables)

	out := graphql.NewOrderedMap(len(fields))
	for i, field := range fields {
		out.Keys[i] = field.Alias

		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("__Schema")
		case "types":
			out.Values[i] = ec.___Schema_types(ctx, field, obj)
		case "queryType":
			out.Values[i] = ec.___Schema_queryType(ctx, field, obj)
		case "mutationType":
			out.Values[i] = ec.___Schema_mutationType(ctx, field, obj)
		case "subscriptionType":
			out.Values[i] = ec.___Schema_subscriptionType(ctx, field, obj)
		case "directives":
			out.Values[i] = ec.___Schema_directives(ctx, field, obj)
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}

	return out
}

func (ec *executionContext) ___Schema_types(ctx context.Context, field graphql.CollectedField, obj *introspection.Schema) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Schema"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Types()
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___Type(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

func (ec *executionContext) ___Schema_queryType(ctx context.Context, field graphql.CollectedField, obj *introspection.Schema) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Schema"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.QueryType()
	if res == nil {
		return graphql.Null
	}
	return ec.___Type(ctx, field.Selections, res)
}

func (ec *executionContext) ___Schema_mutationType(ctx context.Context, field graphql.CollectedField, obj *introspection.Schema) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Schema"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.MutationType()
	if res == nil {
		return graphql.Null
	}
	return ec.___Type(ctx, field.Selections, res)
}

func (ec *executionContext) ___Schema_subscriptionType(ctx context.Context, field graphql.CollectedField, obj *introspection.Schema) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Schema"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.SubscriptionType()
	if res == nil {
		return graphql.Null
	}
	return ec.___Type(ctx, field.Selections, res)
}

func (ec *executionContext) ___Schema_directives(ctx context.Context, field graphql.CollectedField, obj *introspection.Schema) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Schema"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Directives()
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___Directive(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

var __TypeImplementors = []string{"__Type"}

// nolint: gocyclo, errcheck, gas, goconst
func (ec *executionContext) ___Type(ctx context.Context, sel []query.Selection, obj *introspection.Type) graphql.Marshaler {
	fields := graphql.CollectFields(ec.Doc, sel, __TypeImplementors, ec.Variables)

	out := graphql.NewOrderedMap(len(fields))
	for i, field := range fields {
		out.Keys[i] = field.Alias

		switch field.Name {
		case "__typename":
			out.Values[i] = graphql.MarshalString("__Type")
		case "kind":
			out.Values[i] = ec.___Type_kind(ctx, field, obj)
		case "name":
			out.Values[i] = ec.___Type_name(ctx, field, obj)
		case "description":
			out.Values[i] = ec.___Type_description(ctx, field, obj)
		case "fields":
			out.Values[i] = ec.___Type_fields(ctx, field, obj)
		case "interfaces":
			out.Values[i] = ec.___Type_interfaces(ctx, field, obj)
		case "possibleTypes":
			out.Values[i] = ec.___Type_possibleTypes(ctx, field, obj)
		case "enumValues":
			out.Values[i] = ec.___Type_enumValues(ctx, field, obj)
		case "inputFields":
			out.Values[i] = ec.___Type_inputFields(ctx, field, obj)
		case "ofType":
			out.Values[i] = ec.___Type_ofType(ctx, field, obj)
		default:
			panic("unknown field " + strconv.Quote(field.Name))
		}
	}

	return out
}

func (ec *executionContext) ___Type_kind(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Kind()
	return graphql.MarshalString(res)
}

func (ec *executionContext) ___Type_name(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Name()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

func (ec *executionContext) ___Type_description(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Description()
	if res == nil {
		return graphql.Null
	}
	return graphql.MarshalString(*res)
}

func (ec *executionContext) ___Type_fields(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	args := map[string]interface{}{}
	var arg0 bool
	if tmp, ok := field.Args["includeDeprecated"]; ok {
		var err error
		arg0, err = graphql.UnmarshalBoolean(tmp)
		if err != nil {
			ec.Error(ctx, err)
			return graphql.Null
		}
	}
	args["includeDeprecated"] = arg0
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = args
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Fields(args["includeDeprecated"].(bool))
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___Field(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

func (ec *executionContext) ___Type_interfaces(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.Interfaces()
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___Type(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

func (ec *executionContext) ___Type_possibleTypes(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.PossibleTypes()
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___Type(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

func (ec *executionContext) ___Type_enumValues(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	args := map[string]interface{}{}
	var arg0 bool
	if tmp, ok := field.Args["includeDeprecated"]; ok {
		var err error
		arg0, err = graphql.UnmarshalBoolean(tmp)
		if err != nil {
			ec.Error(ctx, err)
			return graphql.Null
		}
	}
	args["includeDeprecated"] = arg0
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = args
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.EnumValues(args["includeDeprecated"].(bool))
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___EnumValue(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

func (ec *executionContext) ___Type_inputFields(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.InputFields()
	arr1 := graphql.Array{}
	for idx1 := range res {
		arr1 = append(arr1, func() graphql.Marshaler {
			rctx := graphql.GetResolverContext(ctx)
			rctx.PushIndex(idx1)
			defer rctx.Pop()
			if res[idx1] == nil {
				return graphql.Null
			}
			return ec.___InputValue(ctx, field.Selections, res[idx1])
		}())
	}
	return arr1
}

func (ec *executionContext) ___Type_ofType(ctx context.Context, field graphql.CollectedField, obj *introspection.Type) graphql.Marshaler {
	rctx := graphql.GetResolverContext(ctx)
	rctx.Object = "__Type"
	rctx.Args = nil
	rctx.Field = field
	rctx.PushField(field.Alias)
	defer rctx.Pop()
	res := obj.OfType()
	if res == nil {
		return graphql.Null
	}
	return ec.___Type(ctx, field.Selections, res)
}

func (ec *executionContext) introspectSchema() *introspection.Schema {
	return introspection.WrapSchema(ec.Schema)
}

func (ec *executionContext) introspectType(name string) *introspection.Type {
	t := ec.Schema.Resolve(name)
	if t == nil {
		return nil
	}
	return introspection.WrapType(t)
}
