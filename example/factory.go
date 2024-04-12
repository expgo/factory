package example

import (
	"fmt"
	"reflect"
)

//go:generate ag --dev-plugin=github.com/expgo/factory/annotation

type StructInterface interface {
	Hello()
}

type FactoryStruct struct {
	typeName string
}

func (fs *FactoryStruct) Hello() {
	fmt.Printf("Hello %s", fs.typeName)
}

type UseFactoryStruct struct {
	MI StructInterface `new:""`
}

type MyStructFactory struct{}

// @Factory(params="self")
func (mf MyStructFactory) New1(t any) StructInterface {
	vt := reflect.TypeOf(t)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	return &FactoryStruct{typeName: "struct: " + vt.PkgPath() + "/" + vt.Name()}
}

type FuncInterface interface {
	Hello1()
}

type FuncStruct struct {
	typeName string
}

func (fs *FuncStruct) Hello1() {
	fmt.Printf("Hello %s", fs.typeName)
}

type UseFuncStruct struct {
	MI FuncInterface `new:""`
}

// @Factory(params="self")
func newMyInterface(t any) FuncInterface {
	vt := reflect.TypeOf(t)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	return &FuncStruct{typeName: "func: " + vt.PkgPath() + "/" + vt.Name()}
}
