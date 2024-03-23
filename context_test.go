package factory

import "testing"

func TestContextRange(t *testing.T) {
	Range[DoInf](func(d any) bool {
		println(d.(DoInf).Hello())
		return true
	})

	Range[ExtStruct](func(e any) bool {
		println(e.(*ExtStruct).Hello())
		return true
	})

	RangeInterface[DoInf](func(d DoInf) bool {
		println(d.Hello())
		return true
	})

	RangeStruct[ExtStruct](func(e *ExtStruct) bool {
		println(e.Hello())
		return true
	})
}
