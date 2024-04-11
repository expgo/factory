package example

import "reflect"

//go:generate ag --dev-plugin=github.com/expgo/factory/annotation --dev-plugin-dir=../

type Interface interface {
	Hello() string
}

type MyInterface struct {
	Type string
}

func (mi *MyInterface) Hello() string {
	return "Hello " + mi.Type
}

// @Factory
func Create(self any) Interface {
	vt := reflect.TypeOf(self)
	if vt.Kind() == reflect.Ptr {
		vt = vt.Elem()
	}

	return &MyInterface{Type: vt.PkgPath() + "/" + vt.Name()}
}
