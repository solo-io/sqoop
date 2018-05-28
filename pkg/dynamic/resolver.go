package dynamic

import (
	"github.com/vektah/gqlgen/neelance/schema"
	"github.com/vektah/gqlgen/neelance/common"
	"github.com/pkg/errors"
	"github.com/solo-io/gloo/pkg/log"
	"encoding/json"
	"time"
	"strconv"
)

// store all the user resolvers
type ResolverMap struct {
	// resolvers for all named types
	Types map[schema.NamedType]*TypeResolver
}

type TypeResolver struct {
	// resolve each field of the type
	Fields map[string]*FieldResolver
}

type FieldResolver struct {
	// type the field resolves to
	Type common.Type

	// how to resolve this field. should return Type
	ResolverFunc ResolverFunc
}

type ResolverFunc func(params Params) (Value, error)
type RawResolver func(params Params) ([]byte, error)

type Params struct {
	Source map[string]interface{}
	Args   map[string]interface{}
}

func (p Params) Arg(name string) interface{} {
	if len(p.Args) == 0 {
		return nil
	}
	return p.Args[name]
}

func NewResolverMap(sch *schema.Schema) *ResolverMap {
	typeMap := make(map[schema.NamedType]*TypeResolver)
	for _, t := range sch.Types {
		if metaType(t.TypeName()) {
			continue
		}
		fields := make(map[string]*FieldResolver)
		switch t := t.(type) {
		case *schema.Object:
			for _, f := range t.Fields {
				inputKey := t.Name + "." + f.Name
				log.Printf("initializing resolver: %v", inputKey)
				fields[f.Name] = &FieldResolver{Type: f.Type, ResolverFunc: nil}
			}
		}
		if len(fields) == 0 {
			continue
		}
		typeMap[t] = &TypeResolver{Fields: fields}
	}
	return &ResolverMap{
		Types: typeMap,
	}
}

func (rm *ResolverMap) RegisterResolver(typ schema.NamedType, field string, rawResolver RawResolver) error {
	fieldResolver, err := rm.getFieldResolver(typ, field)
	if err != nil {
		return err
	}
	fieldResolver.ResolverFunc = func(params Params) (Value, error) {
		data, err := rawResolver(params)
		if err != nil {
			return nil, errors.Wrap(err, "calling raw resolver")
		}
		switch fieldType := fieldResolver.Type.(type) {
		case *schema.Object:
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
		}
		return nil, errors.Errorf("unable to resolve field type %v", fieldResolver.Type)
	}
	return nil
}

// TODO: support custom scalars
func scalarFromBytes(scalar *schema.Scalar, raw string) (Value, error) {
	switch scalar.TypeName() {
	case "Int":
		v, err := strconv.Atoi(raw)
		if err != nil {
			return nil, errors.Wrap(err, "converting bytes to int")
		}
		return &Int{Scalar: scalar, Data: v}, nil
	case "Float":
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, errors.Wrap(err, "converting bytes to float")
		}
		return &Float{Scalar: scalar, Data: v}, nil
	case "String", "ID":
		return &String{Scalar: scalar, Data: raw}, nil
	case "Boolean":
		v, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, errors.Wrap(err, "converting bytes to float")
		}
		return &Bool{Scalar: scalar, Data: v}, nil
	default:
		return nil, errors.Errorf("custom scalars unsupported: %v", scalar.TypeName())
	}
}

func convertValue(typ common.Type, rawValue interface{}) (Value, error) {
	// TODO: be careful about these nil returns
	if rawValue == nil {
		return &Null{}, nil
	}
	switch typ := typ.(type) {
	case *schema.Object:
		// rawValue must be map[string]interface{}
		rawObj, ok := rawValue.(map[string]interface{})
		if !ok {
			// TODO: filter data out of logs (could be sensitive)
			return nil, errors.Errorf("raw value %v was not type *schema.Object", rawValue)
		}
		obj := NewOrderedMap(len(typ.Fields))
		// convert each interface{} type to Value type
		for _, field := range typ.Fields {
			// set each field of the *Object to be a
			// value wrapper around the raw object's value for the field
			convertedValue, err := convertValue(field.Type, rawObj[field.Name])
			if err != nil {
				return nil, errors.Wrapf(err, "converting object field %v", field.Name)
			}
			obj.Add(field.Name, convertedValue)
		}
		return &Object{Data: obj, Object: typ}, nil
	case *common.List:
		// rawValue must be map[string]interface{}
		rawList, ok := rawValue.([]interface{})
		if !ok {
			// TODO: filter data out of logs (could be sensitive)
			return nil, errors.Errorf("raw value %v was not type *common.List", rawValue)
		}
		var array []Value
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
		return &Array{Data: array, List: typ}, nil
	case *schema.Scalar:
		switch data := rawValue.(type) {
		case int:
			return &Int{Data: data, Scalar: typ}, nil
		case string:
			return &String{Data: data, Scalar: typ}, nil
		case float32:
			return &Float{Data: float64(data), Scalar: typ}, nil
		case float64:
			return &Float{Data: data, Scalar: typ}, nil
		case bool:
			return &Bool{Data: data, Scalar: typ}, nil
		case time.Time:
			return &Time{Data: data, Scalar: typ}, nil
		default:
			// TODO: sanitize logs/error messages
			return nil, errors.Errorf("unknown return type %v", data)
		}
	}
	return nil, errors.Errorf("unknown or unsupported type %v", typ.String())
}

func (rm *ResolverMap) Resolve(typ schema.NamedType, field string, params Params) (Value, error) {
	fieldResolver, err := rm.getFieldResolver(typ, field)
	if err != nil {
		return nil, errors.Wrap(err, "resolver lookup")
	}
	if fieldResolver.ResolverFunc == nil {
		return nil, errors.Wrapf(err, "resolver for %v.%v has not been registered", typ.String(), field)
	}
	data, err := fieldResolver.ResolverFunc(params)
	if err != nil {
		return nil, errors.Wrapf(err, "failed executing resolver for %v.%v", typ.String(), field)
	}
	result, err := convertResult(typ, data)
	if err != nil {
		return nil, errors.Wrap(err, "converting interface{} to result")
	}
	return result, nil
}

func convertResult(typ common.Type, data interface{}) (Value, error) {
	var result Value
	switch typ := typ.(type) {
	case *schema.Object:
		obj, ok := data.(*OrderedMap)
		if !ok {
			return nil, errors.Errorf("resolver did not return expected type *OrderedMap: %v", data)
		}
		result = &Object{
			Object: typ,
			Data:   obj,
		}
	case *common.List:
		items, ok := data.([]interface{})
		if !ok {
			return nil, errors.Errorf("resolver did not return expected type []interface{}: %v", data)
		}
		var list []Value
		for _, item := range items {
			val, err := convertResult(typ.OfType, item)
			if err != nil {
				return nil, errors.Wrap(err, "converting array element into result")
			}
			list = append(list, val)
		}
		result = &Array{
			List: typ,
			Data: list,
		}
	}
	return result, nil
}

func (rm *ResolverMap) getFieldResolver(typ schema.NamedType, field string) (*FieldResolver, error) {
	typeResolver, ok := rm.Types[typ]
	if !ok {
		return nil, errors.Errorf("type %v unknown", typ.TypeName())
	}
	fieldResolver, ok := typeResolver.Fields[field]
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

func metaType(typeName string) bool {
	for _, mt := range metaTypes {
		if typeName == mt {
			return true
		}
	}
	return false
}
