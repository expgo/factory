package example

//go:generate ag --dev-plugin=github.com/expgo/factory/annotation --dev-plugin-dir=../

// @Singleton(Init={"aaa", "bbb"})
type MyStruct struct{}

// @Singleton(Init={"aaa", "bbb"}, localVar)
type LocalVarMyStruct struct{}
