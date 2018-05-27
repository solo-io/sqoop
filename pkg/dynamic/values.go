package dynamic

import (
	"github.com/vektah/gqlgen/neelance/common"
	"time"
	"github.com/vektah/gqlgen/neelance/schema"
)

type Value interface {
	common.Type
	isValue()
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

type Time struct {
	*schema.Scalar
	Data time.Time
}

func (t *Object) isValue() {}
func (t *Array) isValue() {}
func (t *Int) isValue() {}
func (t *Float) isValue() {}
func (t *String) isValue() {}
func (t *Time) isValue() {}