package factory

import (
	"errors"
	"testing"
)

func init() {
	Singleton[testRepo]()
}

type testRepo struct {
}

func (tp *testRepo) Hello() {
	panic(errors.New("test repo hello called"))
}

type testStruct struct{}

func (ts *testStruct) TestStruct() {
	panic(errors.New("constructor called"))
}

func (ts *testStruct) Init() {
	panic(errors.New("init called"))
}

func (ts *testStruct) InitWithReturn() error {
	return nil
}

func (ts testStruct) MyInit(tp *testRepo) {
	if tp != nil {
		tp.Hello()
	}
}

func (ts *testStruct) MyErrorInit(tp testRepo) {
	tp.Hello()
}

func (ts *testStruct) nonExistentMethodName() {

}

func TestNewWithOption(t *testing.T) {
	test_cases := []struct {
		name          string
		option        *Option
		expectedPanic error
		expectError   bool
	}{
		{
			name:          "OptionWithInitMethod",
			option:        NewOption().UseConstructor(true),
			expectedPanic: errors.New("constructor called"),
			expectError:   true,
		},
		{
			name:          "OptionWithNonexistentMethod",
			option:        NewOption().InitMethodName("NonExistentMethodName"),
			expectedPanic: errors.New(""),
			expectError:   false,
		},
		{
			name:          "OptionWithInitWithReturn",
			option:        &Option{useConstructor: false, initMethodName: "InitWithReturn"},
			expectedPanic: errors.New("init method 'InitWithReturn' must not have return values"),
			expectError:   true,
		},
		{
			name:          "OptionWithMyInitMethodWithWareObjParams",
			option:        &Option{useConstructor: false, initMethodName: "MyInit"},
			expectedPanic: errors.New("test repo hello called"),
			expectError:   true,
		},
		{
			name:          "OptionWithMyInitMethodWithWareObjErrorParams",
			option:        &Option{useConstructor: false, initMethodName: "MyErrorInit"},
			expectedPanic: errors.New("create testStruct error: method MyErrorInit's 1 argument must be a struct point or an interface"),
			expectError:   true,
		},
	}

	for _, tt := range test_cases {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.expectError {
					t.Errorf("panic = %v, expectError %v", r, tt.expectError)
					return
				}
				if tt.expectError && r != nil {
					if r.(error).Error() != tt.expectedPanic.Error() {
						t.Errorf("Got panic = %v, want %v", r, tt.expectedPanic)
					}
				}
			}()
			NewWithOption[testStruct](tt.option)
		})
	}
}
