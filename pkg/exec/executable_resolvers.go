package exec

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/solo-io/qloo/pkg/dynamic"
	"github.com/vektah/gqlgen/neelance/common"
	"github.com/vektah/gqlgen/neelance/schema"
)

// store all the user resolvers
type ExecutableResolverMap struct {
	// resolvers for all named types
	types map[schema.NamedType]*typeResolver
}

type typeResolver struct {
	// resolve each field of the type
	fields map[string]*fieldResolver
}

type fieldResolver struct {
	// type the field resolves to
	typ common.Type

	// how to resolve this field. should return Type
	resolverFunc RawResolver
}

type RawResolver func(params Params) ([]byte, error)

type Params struct {
	Parent *dynamic.Object
	Args   map[string]interface{}
}

func (p Params) Arg(name string) interface{} {
	if len(p.Args) == 0 {
		return nil
	}
	return p.Args[name]
}

func NewExecutableResolvers(sch *schema.Schema, generateResolver func(string, string) (RawResolver, error)) (*ExecutableResolverMap, error) {
	typeMap := make(map[schema.NamedType]*typeResolver)
	for _, namedType := range sch.Types {
		if MetaType(namedType.TypeName()) {
			continue
		}
		fields := make(map[string]*fieldResolver)
		switch typ := namedType.(type) {
		case *schema.Object:
			for _, field := range typ.Fields {
				rawResolver, err := generateResolver(typ.Name, field.Name)
				if err != nil {
					return nil, errors.Wrapf(err, "generating resolver for %v.%v", typ.Name, field.Name)
				}
				fields[field.Name] = &fieldResolver{typ: field.Type, resolverFunc: rawResolver}
			}
		}
		if len(fields) == 0 {
			continue
		}
		typeMap[namedType] = &typeResolver{fields: fields}
	}
	return &ExecutableResolverMap{
		types: typeMap,
	}, nil
}

func toValue(data []byte, typ common.Type) (dynamic.Value, error) {
	switch fieldType := typ.(type) {
	case *schema.Object, *schema.Interface:
		var rawResult map[string]interface{}
		if err := json.Unmarshal(data, &rawResult); err != nil {
			return nil, errors.Wrap(err, "parsing response as json")
		}
		return convertValue(fieldType, rawResult)
	case *common.List:
		var rawResult []interface{}
		if err := json.Unmarshal(data, &rawResult); err != nil {
			return nil, errors.Wrap(err, "parsing response as json")
		}
		return convertValue(fieldType, rawResult)
	case *schema.Scalar:
		return scalarFromBytes(fieldType, string(data))
	case *common.NonNull:
		return toValue(data, fieldType.OfType)
	}
	return nil, errors.Errorf("unable to resolve field type %v", typ)
}

// TODO: support custom scalars
func scalarFromBytes(scalar *schema.Scalar, raw string) (dynamic.Value, error) {
	switch scalar.TypeName() {
	case "Int":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, errors.Wrap(err, "converting bytes to int")
		}
		return &dynamic.Int{Scalar: scalar, Data: v}, nil
	case "Float":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, errors.Wrap(err, "converting bytes to float")
		}
		return &dynamic.Float{Scalar: scalar, Data: v}, nil
	case "String", "ID":
		return &dynamic.String{Scalar: scalar, Data: raw}, nil
	case "Boolean":
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, errors.Wrap(err, "converting bytes to float")
		}
		return &dynamic.Bool{Scalar: scalar, Data: v}, nil
	default:
		return nil, errors.Errorf("custom scalars unsupported: %v", scalar.TypeName())
	}
}

func convertValue(typ common.Type, rawValue interface{}) (dynamic.Value, error) {
	// TODO: be careful about these nil returns
	if rawValue == nil {
		return &dynamic.Null{}, nil
	}
	switch typ := typ.(type) {
	case *schema.Interface:
		concreteType, err := determineType(typ, rawValue)
		if err != nil {
			// TODO: sanitize
			return nil, errors.Wrapf(err, "determining concrete type of interface %v", rawValue)
		}
		return convertValue(concreteType, rawValue)
	case *schema.Object:
		// rawValue must be map[string]interface{}
		rawObj, ok := rawValue.(map[string]interface{})
		if !ok {
			// TODO: sanitize
			return nil, errors.Errorf("raw value %v was not type *schema.Object", rawValue)
		}
		obj := dynamic.NewOrderedMap()
		// convert each interface{} type to Value type
		for _, field := range typ.Fields {
			// set each field of the *Object to be a
			// value wrapper around the raw object's value for the field
			convertedValue, err := convertValue(field.Type, rawObj[field.Name])
			if err != nil {
				return nil, errors.Wrapf(err, "converting object field %v", field.Name)
			}
			obj.Set(field.Name, convertedValue)
			// so we can pass extra data down
			delete(rawObj, field.Name)
		}
		for extraField, val := range rawObj {
			obj.Set(extraField, &dynamic.InternalOnly{Data: val})
		}
		return &dynamic.Object{Data: obj, Object: typ}, nil
	case *common.List:
		// rawValue must be map[string]interface{}
		rawList, ok := rawValue.([]interface{})
		if !ok {
			// TODO: filter data out of logs (could be sensitive)
			return nil, errors.Errorf("raw value %v was not type *common.List", rawValue)
		}
		var array []dynamic.Value
		// convert each interface{} type to Value type
		for _, rawElement := range rawList {
			// set each field of the *Object to be a
			// value wrapper around the raw object's value for the field
			convertedValue, err := convertValue(typ.OfType, rawElement)
			if err != nil {
				return nil, errors.Wrapf(err, "converting array element")
			}
			array = append(array, convertedValue)
		}
		return &dynamic.Array{Data: array, List: typ}, nil
	case *common.NonNull:
		return convertValue(typ.OfType, rawValue)
	case *schema.Scalar:
		switch data := rawValue.(type) {
		case int:
			return &dynamic.Int{Data: data, Scalar: typ}, nil
		case string:
			return &dynamic.String{Data: data, Scalar: typ}, nil
		case float32:
			return &dynamic.Float{Data: float64(data), Scalar: typ}, nil
		case float64:
			return &dynamic.Float{Data: data, Scalar: typ}, nil
		case bool:
			return &dynamic.Bool{Data: data, Scalar: typ}, nil
		case time.Time:
			return &dynamic.Time{Data: data, Scalar: typ}, nil
		default:
			// TODO: sanitize logs/error messages
			return nil, errors.Errorf("unknown return type %v", data)
		}
	case *schema.Enum:
		data, ok := rawValue.(string)
		if !ok {
			return nil, errors.Errorf("expected string type for enum, got %v", rawValue)
		}
		return &dynamic.Enum{Data: data, Enum: typ}, nil
	}
	return nil, errors.Errorf("unknown or unsupported type %v", typ.String())
}

func determineType(iface *schema.Interface, rawValue interface{}) (*schema.Object, error) {
	// rawValue must be map[string]interface{}
	rawObj, ok := rawValue.(map[string]interface{})
	if !ok {
		// TODO: sanitize
		return nil, errors.Errorf("raw value %v was not type *schema.Object", rawValue)
	}
	objType := rawObj["__typename"]
	if objType == nil {
		// TODO: sanitize
		return nil, errors.Errorf("object implements interface %v but does not contain field __typename, "+
			"cannot determine object type", iface.Name)
	}
	objTypeName, ok := objType.(string)
	if !ok {
		// TODO: sanitize
		return nil, errors.Errorf("__typename must be a string")
	}
	for _, possibleType := range iface.PossibleTypes {
		if possibleType.Name == objTypeName {
			return possibleType, nil
		}
	}
	return nil, errors.Errorf("%v does not implement %v", objTypeName, iface.Name)
}

func (rm *ExecutableResolverMap) Resolve(typ schema.NamedType, field string, params Params) (dynamic.Value, error) {
	fieldResolver, err := rm.getFieldResolver(typ, field)
	if err != nil {
		return nil, errors.Wrap(err, "resolver lookup")
	}
	if fieldResolver.resolverFunc == nil {
		// no resolver func? look in the parent for the field
		if params.Parent != nil {
			if fieldValue := params.Parent.Data.Get(field); fieldValue != nil && fieldValue.Kind() != "NULL" {
				return fieldValue, nil
			}
		}
		return nil, errors.Errorf("resolver for %v.%v has not been registered", typ.String(), field)
	}
	data, err := fieldResolver.resolverFunc(params)
	if err != nil {
		return nil, errors.Wrapf(err, "failed executing resolver for %v.%v", typ.String(), field)
	}
	return toValue(data, fieldResolver.typ)
}

func (rm *ExecutableResolverMap) getFieldResolver(typ schema.NamedType, field string) (*fieldResolver, error) {
	typeResolver, ok := rm.types[typ]
	if !ok {
		return nil, errors.Errorf("type %v unknown", typ.TypeName())
	}
	fieldResolver, ok := typeResolver.fields[field]
	if !ok {
		return nil, errors.Errorf("type %v does not contain field %v", typ.TypeName(), field)
	}
	return fieldResolver, nil
}

var metaTypes = []string{
	"Map",
	"Float",
	"ID",
	"Int",
	"Boolean",
	"String",
	"__Type",
	"__TypeKind",
	"__Directive",
	"__EnumValue",
	"__Schema",
	"__InputValue",
	"__DirectiveLocation",
	"__Field",
}

func MetaType(typeName string) bool {
	for _, mt := range metaTypes {
		if typeName == mt {
			return true
		}
	}
	return false
}
