package object

import (
	"slug/internal/ast"
	"slug/internal/dec64"
	"strings"
	"testing"
)

func TestStringMapKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.MapKey() != hello2.MapKey() {
		t.Errorf("strings with same content have different map keys")
	}

	if diff1.MapKey() != diff2.MapKey() {
		t.Errorf("strings with same content have different map keys")
	}

	if hello1.MapKey() == diff1.MapKey() {
		t.Errorf("strings with different content have same map keys")
	}
}

func TestBooleanMapKey(t *testing.T) {
	true1 := &Boolean{Value: true}
	true2 := &Boolean{Value: true}
	false1 := &Boolean{Value: false}
	false2 := &Boolean{Value: false}

	if true1.MapKey() != true2.MapKey() {
		t.Errorf("trues do not have same map key")
	}

	if false1.MapKey() != false2.MapKey() {
		t.Errorf("falses do not have same map key")
	}

	if true1.MapKey() == false1.MapKey() {
		t.Errorf("true has same map key as false")
	}
}

func TestIntegerMapKey(t *testing.T) {
	one1 := &Number{Value: dec64.FromInt64(1)}
	one2 := &Number{Value: dec64.FromInt64(1)}
	two1 := &Number{Value: dec64.FromInt64(2)}
	two2 := &Number{Value: dec64.FromInt64(2)}

	if one1.MapKey() != one2.MapKey() {
		t.Errorf("numbers with same content have different map keys")
	}

	if two1.MapKey() != two2.MapKey() {
		t.Errorf("numbers with same content have different map keys")
	}

	if one1.MapKey() == two1.MapKey() {
		t.Errorf("numbers with different content have same map keys, %v : %v", one1, two1)
	}
}

func TestSymbolMapKey(t *testing.T) {
	s1 := InternSymbol("foo")
	s2 := InternSymbol("foo")
	s3 := InternSymbol("bar")

	if s1 != s2 {
		t.Errorf("symbols with same name are not interned")
	}
	if s1.MapKey() != s2.MapKey() {
		t.Errorf("symbols with same name have different map keys")
	}
	if s1.MapKey() == s3.MapKey() {
		t.Errorf("symbols with different names have same map keys")
	}
}

func TestDispatchErrorIncludesCandidates(t *testing.T) {
	fg := &FunctionGroup{
		Functions: map[ast.FSig]Object{
			{Tags: "@map|||", Min: 3, Max: 3, IsVariadic: false}: &Foreign{
				Name: "put",
			},
		},
	}

	_, err := fg.DispatchToFunction("put", nil, nil)
	if err == nil {
		t.Fatal("expected dispatch error, got nil")
	}
	msg := err.Error()
	if !strings.Contains(msg, "Candidates:") {
		t.Fatalf("expected candidates in dispatch error, got %q", msg)
	}
	if !strings.Contains(msg, "args 3") {
		t.Fatalf("expected arity in candidates, got %q", msg)
	}
}

func TestMapBasicOperations(t *testing.T) {
	m := &Map{}

	m.Put(&String{Value: "a"}, &Number{Value: dec64.FromInt(1)})
	m.Put(&String{Value: "b"}, &Number{Value: dec64.FromInt(2)})
	m.Put(InternSymbol("c"), &Number{Value: dec64.FromInt(3)})

	if got := m.Len(); got != 3 {
		t.Fatalf("map length mismatch: got=%d want=3", got)
	}

	pair, ok := m.GetPair((&String{Value: "b"}).MapKey())
	if !ok {
		t.Fatal("expected key 'b' to exist")
	}
	val, ok := pair.Value.(*Number)
	if !ok || !val.Value.Eq(dec64.FromInt(2)) {
		t.Fatalf("unexpected value for 'b': %#v", pair.Value)
	}
}

func TestMapMigratesLegacyPairs(t *testing.T) {
	legacy := map[MapKey]MapPair{}
	k := (&String{Value: "legacy"}).MapKey()
	legacy[k] = MapPair{
		Key:   &String{Value: "legacy"},
		Value: &Number{Value: dec64.FromInt(99)},
	}

	m := &Map{Pairs: legacy}

	if m.Len() != 1 {
		t.Fatalf("expected migrated map length=1, got=%d", m.Len())
	}
	if m.Pairs != nil {
		t.Fatal("expected legacy pairs to be cleared after HAMT migration")
	}
	if pair, ok := m.GetPair(k); !ok || !pair.Value.(*Number).Value.Eq(dec64.FromInt(99)) {
		t.Fatalf("expected migrated value to be preserved, got pair=%#v ok=%v", pair, ok)
	}
}

func TestMapCloneIsIndependent(t *testing.T) {
	orig := (&Map{}).Put(&String{Value: "a"}, &Number{Value: dec64.FromInt(1)})
	clone := orig.Clone()
	clone.Put(&String{Value: "b"}, &Number{Value: dec64.FromInt(2)})

	if orig.Len() != 1 {
		t.Fatalf("original mutated after clone write: len=%d", orig.Len())
	}
	if clone.Len() != 2 {
		t.Fatalf("clone missing mutation: len=%d", clone.Len())
	}
}

func TestMapDeleteKey(t *testing.T) {
	m := (&Map{}).
		Put(&String{Value: "a"}, &Number{Value: dec64.FromInt(1)}).
		Put(&String{Value: "b"}, &Number{Value: dec64.FromInt(2)})

	keyA := (&String{Value: "a"}).MapKey()
	m.DeleteKey(keyA)

	if m.Len() != 1 {
		t.Fatalf("unexpected map size after delete: got=%d want=1", m.Len())
	}
	if _, ok := m.GetPair(keyA); ok {
		t.Fatal("deleted key still present")
	}
}
