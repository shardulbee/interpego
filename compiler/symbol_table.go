package compiler

type SymbolScope string

const (
	GLOBAL_SCOPE SymbolScope = "GLOBAL_SCOPE"
	LOCAL_SCOPE              = "LOCAL_SCOPE"
)

type Symbol struct {
	Name  string
	Index int
	Scope SymbolScope
}

type SymbolTable struct {
	outer          *SymbolTable
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{numDefinitions: 0, store: make(map[string]Symbol)}
}

func NewNestedSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{numDefinitions: 0, store: make(map[string]Symbol), outer: outer}
}

func (st *SymbolTable) Define(name string) Symbol {
	var scope SymbolScope
	if st.outer == nil {
		scope = GLOBAL_SCOPE
	} else {
		scope = LOCAL_SCOPE
	}
	newSymbol := Symbol{name, st.numDefinitions, scope}
	st.numDefinitions += 1
	st.store[name] = newSymbol
	return newSymbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	sym, ok := st.store[name]
	if !ok && st.outer != nil {
		return st.outer.Resolve(name)
	}
	return sym, ok
}
