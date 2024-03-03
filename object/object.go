package object

import "fmt"

type ObjectType string

const (
	INTEGER_TYPE = "INTEGER"
	BOOLEAN_TYPE = "BOOLEAN"
	NULL_TYPE    = "NULL"
	RETURN_TYPE  = "RETURN"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType {
	return INTEGER_TYPE
}
func (i *Integer) Inspect() string {
	return fmt.Sprintf("%d", i.Value)
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType {
	return BOOLEAN_TYPE
}
func (b *Boolean) Inspect() string {
	return fmt.Sprintf("%t", b.Value)
}

type Null struct {
}

func (n *Null) Type() ObjectType {
	return NULL_TYPE
}
func (n *Null) Inspect() string {
	return fmt.Sprintf("null")
}

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType {
	return RETURN_TYPE
}
func (rv *ReturnValue) Inspect() string {
	return rv.Value.Inspect()
}
