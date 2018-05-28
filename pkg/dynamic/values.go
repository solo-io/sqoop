package dynamic

import (
	"github.com/vektah/gqlgen/neelance/common"
	"time"
	"github.com/vektah/gqlgen/neelance/schema"
	"github.com/vektah/gqlgen/graphql"
)

type Value interface {
	common.Type
	Marshaller() graphql.Marshaler
	//TODO:
	//Validate() error
}

// enforce interface
var (
	_ Value = &Object{}
	_ Value = &Array{}
	_ Value = &String{}
	_ Value = &Float{}
	_ Value = &Int{}
	_ Value = &Time{}
)

type Object struct {
	*schema.Object
	Data map[string]Value
}

type Array struct {
	*common.List
	Data []Value
}

type Int struct {
	*schema.Scalar
	Data int
}

type String struct {
	*schema.Scalar
	Data string
}

type Float struct {
	*schema.Scalar
	Data float64
}

type Bool struct {
	*schema.Scalar
	Data bool
}

type Time struct {
	*schema.Scalar
	Data time.Time
}

func (t *Object) Marshaller() graphql.Marshaler {
	fieldMap := graphql.NewOrderedMap(len(t.Data))
	var i int
	for k, v := range t.Data {
		fieldMap.Keys[i] = k
		fieldMap.Values[i] = v.Marshaller()
		i++
	}
	return fieldMap
}
func (t *Array) Marshaller() graphql.Marshaler {
	var array graphql.Array
	for _, val := range t.Data {
		array = append(array, val.Marshaller())
	}
	return array
}
func (t *Int) Marshaller() graphql.Marshaler {
	return graphql.MarshalInt(t.Data)
}
func (t *Float) Marshaller() graphql.Marshaler {
	return graphql.MarshalFloat(t.Data)
}
func (t *String) Marshaller() graphql.Marshaler {
	return graphql.MarshalString(t.Data)
}
func (t *Bool) Marshaller() graphql.Marshaler {
	return graphql.MarshalBoolean(t.Data)
}
func (t *Time) Marshaller() graphql.Marshaler {
	return graphql.MarshalTime(t.Data)
}