package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	inner := NewEnvironment()
	inner.outer = outer
	return inner
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s}
}

func (e *Environment) Set(name string, val Object) {
	e.store[name] = val
}

func (e *Environment) Get(name string) (Object, bool) {
	result, ok := e.store[name]
	if !ok && e.outer != nil {
		result, ok = e.outer.Get(name)
	}
	return result, ok
}
