package compiler

type SymbolScope string

const (
	GLOBAL_SCOPE SymbolScope = "GLOBAL_SCOPE"
)

type Symbol struct {
	Name  string
	Index int
	Scope SymbolScope
}

type SymbolTable struct {
	store          map[string]Symbol
	numDefinitions int
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{numDefinitions: 0, store: make(map[string]Symbol)}
}

func (st *SymbolTable) Define(name string) Symbol {
	newSymbol := Symbol{name, st.numDefinitions, GLOBAL_SCOPE}
	st.numDefinitions += 1
	st.store[name] = newSymbol
	return newSymbol
}

func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	sym, ok := st.store[name]
	return sym, ok
}
