package factory

import (
	"testing"
)

type type1 struct {
	type2 *type2 `wire:"auto"`
}

type type2 struct {
	type1 *type1 `wire:"auto"`
}

func TestAutoWireTypeCircular(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				if r.(error).Error() != "getting *factory.type1, possible circular reference with *factory.type2" {
					t.Errorf("%s", r)
				}
			} else {
				t.Errorf("Expected panic, but no panic occurred")
			}
		}()

		Singleton[type1]()
		Singleton[type2]()

		_ = Find[type1]()
	}()
}

type name1 struct {
	name2 *name2 `wire:"name"`
}

type name2 struct {
	name1 *name1 `wire:"name"`
}

func TestAutoWireNameCircular(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				if r.(error).Error() != "getting *factory.name1, possible circular reference with *factory.name2" {
					t.Errorf("%s", r)
				}
			} else {
				t.Errorf("Expected panic, but no panic occurred")
			}
		}()

		Singleton[name1]().Name("name1")
		Singleton[name2]().Name("name2")

		_ = Find[name1]()
	}()
}

type expr1 struct {
	Name string `value:"${expr2.name}"`
}

type expr2 struct {
	Name string `value:"${expr1.name}"`
}

func TestAutoWireValueCircular(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				if r.(error).Error() != "getting *factory.expr1, possible circular reference with *factory.expr2" {
					t.Errorf("%s", r)
				}
			} else {
				t.Errorf("Expected panic, but no panic occurred")
			}
		}()

		Singleton[expr1]().Name("expr1")
		Singleton[expr2]().Name("expr2")

		_ = Find[expr1]()
	}()
}

type wire1 struct {
	wire2 *wire2
}

func (w *wire1) Init() {
	w.wire2 = New[wire2]()
}

type wire2 struct {
	wire1 *wire1
}

func (w *wire2) Init() {
	w.wire1 = New[wire1]()
}

// TODO fix it
func TestAutoWireInitFuncCircular(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				if r.(error).Error() != "getting *factory.wire1, possible circular reference with *factory.wire2" {
					t.Errorf("%s", r)
				}
			} else {
				t.Errorf("Expected panic, but no panic occurred")
			}
		}()

		_ = New[wire1]()
	}()
}
