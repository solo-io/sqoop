package test

import (
	"github.com/solo-io/qloo/pkg/dynamic"
	"github.com/vektah/gqlgen/neelance/schema"
	"time"
	"github.com/vektah/gqlgen/neelance/common"
	"github.com/pkg/errors"
	"github.com/solo-io/qloo/examples/starwars/imported/starwars"
	"github.com/fatih/structs"
	"github.com/solo-io/qloo/pkg/exec"
)

var LukeSkywalkerParams = exec.Params{
	Parent: LukeSkywalkerObject(),
	Args: map[string]interface{}{
		"acting":     5,
		"best_scene": "cloud city",
	},
}

var LukeSkywalker = starwars.Human{
	CharacterFields: starwars.CharacterFields{
		TypeName:  "Human",
		ID:        "1000",
		Name:      "Luke Skywalker",
		FriendIds: []string{"1002", "1003", "2000", "2001"},
		AppearsIn: []starwars.Episode{starwars.EpisodeNewhope, starwars.EpisodeEmpire, starwars.EpisodeJedi},
	},
	Mass:        77,
	StarshipIds: []string{"3001", "3003"},
}

func LukeSkywalkerObject() *dynamic.Object {
	schemaObj := StarWarsSchema.Types["Human"].(*schema.Object)
	m := structs.Map(LukeSkywalker)
	obj, err := convertValue(schemaObj, m)
	if err != nil {
		panic(err)
	}
	return obj.(*dynamic.Object)
}

// code copied from executable_resolvers.go
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
