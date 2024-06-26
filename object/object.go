package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"interpego/ast"
	"interpego/code"
)

type ObjectType string

const (
	INTEGER_TYPE           = "INTEGER"
	BOOLEAN_TYPE           = "BOOLEAN"
	NULL_TYPE              = "NULL"
	RETURN_TYPE            = "RETURN"
	ERROR_TYPE             = "ERROR"
	FUNCTION_TYPE          = "FUNCTION"
	STRING_TYPE            = "STRING"
	BUILTIN_TYPE           = "BUILTIN"
	ARRAY_TYPE             = "ARRAY"
	HASH_TYPE              = "HASH"
	COMPILED_FUNCTION_TYPE = "COMPILED_FUNCTION"
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

type Addable interface {
	Add(other Object) (Object, error)
}

func (i *Integer) Add(other Object) (Object, error) {
	otherInt, ok := other.(*Integer)
	if !ok {
		return nil, fmt.Errorf("cannot add non-integer to integer")
	}
	return &Integer{Value: i.Value + otherInt.Value}, nil
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType {
	return STRING_TYPE
}

func (s *String) Inspect() string {
	return s.Value
}

func (s *String) Add(other Object) (Object, error) {
	otherString, ok := other.(*String)
	if !ok {
		return nil, fmt.Errorf("cannot add non-string to string. other=%T (%+v)", other, other)
	}
	return &String{Value: s.Value + otherString.Value}, nil
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

type Null struct{}

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

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType {
	return ERROR_TYPE
}

func (e *Error) Inspect() string {
	return fmt.Sprintf("ERROR: %s", e.Message)
}

type Function struct {
	Env    *Environment
	Params []*ast.Identifier
	Body   *ast.BlockStatement
}

func (fl *Function) Type() ObjectType {
	return FUNCTION_TYPE
}

func (fl *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Params {
		params = append(params, p.String())
	}
	out.WriteString(fmt.Sprintf("fn(%s) { %s }", strings.Join(params, ", "), fl.Body.String()))
	return out.String()
}

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_TYPE }
func (b *Builtin) Inspect() string  { return "builtin function" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_TYPE }
func (a *Array) Inspect() string {
	var out bytes.Buffer

	out.WriteString("[")
	for i, elem := range a.Elements {
		out.WriteString(elem.Inspect())
		if i != len(a.Elements)-1 {
			out.WriteString(", ")
		}
	}

	out.WriteString("]")

	return out.String()
}

type HashKey struct {
	Type  ObjectType
	Value uint64
}

func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

func (s *String) HashKey() HashKey {
	val := fnv.New64a()
	val.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: val.Sum64()}
}

type Hashable interface {
	HashKey() HashKey
}

type HashPair struct {
	Key   Object
	Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_TYPE }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	out.WriteString("({")
	pairs := make([]string, 0, len(h.Pairs))
	for _, hashPair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", hashPair.Key.Inspect(), hashPair.Value.Inspect()))
	}
	out.WriteString(strings.Join(pairs, ", "))

	out.WriteString("})")
	return out.String()
}

type CompiledFunction struct {
	NumLocals     int
	NumParameters int
	Instructions  code.Instructions
}

func (cf *CompiledFunction) Type() ObjectType { return COMPILED_FUNCTION_TYPE }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction: [%p]", cf)
}
