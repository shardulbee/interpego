package compiler

import "testing"

func TestDefine(t *testing.T) {
	expected := map[string]Symbol{
		"a": {Name: "a", Scope: GLOBAL_SCOPE, Index: 0},
		"b": {Name: "b", Scope: GLOBAL_SCOPE, Index: 1},
		"c": {Name: "c", Scope: LOCAL_SCOPE, Index: 0},
		"d": {Name: "d", Scope: LOCAL_SCOPE, Index: 1},
		"e": {Name: "e", Scope: LOCAL_SCOPE, Index: 0},
		"f": {Name: "f", Scope: LOCAL_SCOPE, Index: 1},
	}
	global := NewSymbolTable()
	a := global.Define("a")
	if a != expected["a"] {
		t.Errorf("expected a=%+v, got=%+v", expected["a"], a)
	}
	b := global.Define("b")
	if b != expected["b"] {
		t.Errorf("expected b=%+v, got=%+v", expected["b"], b)
	}
	firstLocal := NewNestedSymbolTable(global)
	c := firstLocal.Define("c")
	if c != expected["c"] {
		t.Errorf("expected c=%+v, got=%+v", expected["c"], c)
	}
	d := firstLocal.Define("d")
	if d != expected["d"] {
		t.Errorf("expected d=%+v, got=%+v", expected["d"], d)
	}
	secondLocal := NewNestedSymbolTable(firstLocal)
	e := secondLocal.Define("e")
	if e != expected["e"] {
		t.Errorf("expected e=%+v, got=%+v", expected["e"], e)
	}
	f := secondLocal.Define("f")
	if f != expected["f"] {
		t.Errorf("expected f=%+v, got=%+v", expected["f"], f)
	}
}

func TestResolveGlobal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")
	expected := []Symbol{
		{Name: "a", Scope: GLOBAL_SCOPE, Index: 0},
		{Name: "b", Scope: GLOBAL_SCOPE, Index: 1},
	}
	for _, sym := range expected {
		result, ok := global.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result == sym {
		}
	}
}

func TestResolveLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")
	local := NewNestedSymbolTable(global)
	local.Define("c")
	local.Define("d")
	expected := []Symbol{
		{Name: "a", Scope: GLOBAL_SCOPE, Index: 0},
		{Name: "b", Scope: GLOBAL_SCOPE, Index: 1},
		{Name: "c", Scope: LOCAL_SCOPE, Index: 0},
		{Name: "d", Scope: LOCAL_SCOPE, Index: 1},
	}
	for _, sym := range expected {
		result, ok := local.Resolve(sym.Name)
		if !ok {
			t.Errorf("name %s not resolvable", sym.Name)
			continue
		}
		if result != sym {
			t.Errorf("expected %s to resolve to %+v, got=%+v",
				sym.Name, sym, result)
		}
	}
}

func TestResolveNestedLocal(t *testing.T) {
	global := NewSymbolTable()
	global.Define("a")
	global.Define("b")
	firstLocal := NewNestedSymbolTable(global)
	firstLocal.Define("c")
	firstLocal.Define("d")
	firstLocal.Define("foobar")
	secondLocal := NewNestedSymbolTable(firstLocal)
	secondLocal.Define("e")
	secondLocal.Define("f")
	secondLocal.Define("g")
	secondLocal.Define("foobar")
	tests := []struct {
		table           *SymbolTable
		expectedSymbols []Symbol
	}{
		{
			firstLocal,
			[]Symbol{
				{Name: "a", Scope: GLOBAL_SCOPE, Index: 0},
				{Name: "b", Scope: GLOBAL_SCOPE, Index: 1},
				{Name: "c", Scope: LOCAL_SCOPE, Index: 0},
				{Name: "d", Scope: LOCAL_SCOPE, Index: 1},
				{Name: "foobar", Scope: LOCAL_SCOPE, Index: 2},
			},
		},
		{
			secondLocal,
			[]Symbol{
				{Name: "a", Scope: GLOBAL_SCOPE, Index: 0},
				{Name: "b", Scope: GLOBAL_SCOPE, Index: 1},
				{Name: "e", Scope: LOCAL_SCOPE, Index: 0},
				{Name: "f", Scope: LOCAL_SCOPE, Index: 1},
				{Name: "foobar", Scope: LOCAL_SCOPE, Index: 3},
			},
		},
	}
	for _, tt := range tests {
		for _, sym := range tt.expectedSymbols {
			result, ok := tt.table.Resolve(sym.Name)
			if !ok {
				t.Errorf("name %s not resolvable", sym.Name)
				continue
			}
			if result != sym {
				t.Errorf("expected %s to resolve to %+v, got=%+v",
					sym.Name, sym, result)
			}
		}
	}
}
