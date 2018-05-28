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
	Data *OrderedMap
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

type Null struct{}

func (t *Null) Kind() string   { return "NULL" }
func (t *Null) String() string { return "null" }

func (t *Object) Marshaller() graphql.Marshaler {
	items := t.Data.Items()
	fieldMap := graphql.NewOrderedMap(len(items))
	for i, item := range items {
		fieldMap.Keys[i] = item.Key
		fieldMap.Values[i] = item.Value.Marshaller()
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
func (t *Null) Marshaller() graphql.Marshaler {
	return graphql.Null
}

// preserving order matters
type OrderedMap struct {
	Keys   []string
	Values []Value
}

func NewOrderedMap(len int) *OrderedMap {
	return &OrderedMap{
		Keys:   make([]string, len),
		Values: make([]Value, len),
	}
}

func (m *OrderedMap) Add(key string, value Value) {
	m.Keys = append(m.Keys, key)
	m.Values = append(m.Values, value)
}

func (m *OrderedMap) Get(key string) Value {
	for i, k := range m.Keys {
		if key == k {
			return m.Values[i]
		}
	}
	return nil
}

func (m *OrderedMap) Delete(key string) {
	for i, k := range m.Keys {
		if key == k {
			m.Keys = append(m.Keys[:i], m.Keys[i+1:]...)
			m.Values = append(m.Values[:i], m.Values[i+1:]...)
			return
		}
	}
}

func (m *OrderedMap) Items() []struct {
	Key   string
	Value Value
} {
	var items []struct {
		Key   string;
		Value Value
	}
	for i, k := range m.Keys {
		items = append(items, struct {
			Key   string;
			Value Value
		}{
			Key:   k,
			Value: m.Values[i],
		})
	}
	return items
}
