package factory

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"testing"
)

type poolTest struct {
	poolName string
	poolSize int
}

func TestGet(t *testing.T) {
	t.Run("TestGetForPoolTestType", func(t *testing.T) {
		t.Parallel()
		str := Get[poolTest]()
		if str == nil {
			t.Error("Expected non-nil value")
		}
	})

	t.Run("TestGetForCustomStruct", func(t *testing.T) {
		t.Parallel()
		type customStruct struct {
			name string
			age  int
		}
		cs := Get[customStruct]()
		if cs == nil {
			t.Error("Expected non-nil value")
		}
	})
}

type MyTestType struct {
	name string
}

func TestPut(t *testing.T) {
	tests := []struct {
		name  string
		input *MyTestType
		want  *MyTestType
	}{
		{
			name:  "NilInput",
			input: nil,
			want:  &MyTestType{},
		},
		{
			name:  "ValidInput",
			input: &MyTestType{name: "Test"},
			want:  &MyTestType{name: "Test"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// cleanup pool
			_poolCache = map[reflect.Type]*sync.Pool{}

			Put(test.input)

			got := Get[MyTestType]()

			assert.Equal(t, test.want, got)
		})
	}
}
