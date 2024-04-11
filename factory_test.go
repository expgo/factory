package factory

import (
	"fmt"
	"reflect"
	"testing"
)

type MyInterface interface {
	Hello()
}

type MyInstance struct {
	typeName string
}

func (mi *MyInstance) Hello() {
	fmt.Printf("Hello %s", mi.typeName)
}

type MyStruct struct {
	MI MyInterface `new:"self"`
}

type MyFactory struct{}

func (mf *MyFactory) New(t any) MyInterface {
	vt := reflect.TypeOf(t)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	return &MyInstance{typeName: "struct: " + vt.PkgPath() + "/" + vt.Name()}
}

func newMyInterface(t any) MyInterface {
	vt := reflect.TypeOf(t)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	return &MyInstance{typeName: "func: " + vt.PkgPath() + "/" + vt.Name()}
}

func TestFactoryObject(t *testing.T) {
	Factory[MyInterface](&MyFactory{})

	my := New[MyStruct]()
	my.MI.Hello()
}

func TestFactoryMethod(t *testing.T) {
	Factory[MyInterface](newMyInterface)

	my := New[MyStruct]()
	my.MI.Hello()
}
