package example

import (
	"github.com/expgo/factory"
	"testing"
)

func TestFactoryObject(t *testing.T) {
	my := factory.New[UseFactoryStruct]()
	my.MI.Hello()
}

func TestFactoryMethod(t *testing.T) {
	my := factory.New[UseFuncStruct]()
	my.MI.Hello1()
}
