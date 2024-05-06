package factory

import (
	"fmt"
	"github.com/expr-lang/expr"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

type DoInf interface {
	Hello() string
}

var _baseStruct = Getter[BaseStruct](Singleton[BaseStruct]())

type BaseStruct struct {
	self DoInf  `wire:"self"`
	name string `wire:"value:${cfg.Name}"`
	cfg  *Cfg   `wire:"auto"`
}

func (b *BaseStruct) Greet() string {
	return b.self.Hello()
}

func (b *BaseStruct) Hello() string {
	return fmt.Sprintf("Hello(%s): base struct", b.name)
}

var _extStruct = Getter[ExtStruct](Singleton[ExtStruct]())

type ExtStruct struct {
	BaseStruct
}

var _cfg = Getter[Cfg](Singleton[Cfg]().Name("cfg"))

type Cfg struct {
	Name string
}

func (c *Cfg) Init() {
	c.Name = "py"
}

func (e *ExtStruct) Hello() string {
	return fmt.Sprintf("Hello(%s): ext struct", e.name)
}

func TestUpdateSelf(t *testing.T) {
	testCases := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "base struct greet",
			got:  New[BaseStruct]().Greet(),
			want: "Hello(py): base struct",
		},
		{
			name: "ext struct greet",
			got:  New[ExtStruct]().Greet(),
			want: "Hello(py): ext struct",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf("AutoWire = %v, want %v", tc.got, tc.want)
			}
		})
	}
}

func TestExpr(t *testing.T) {
	env := _context.getByNamePanic(getTimeoutContext(Opts.Timeout), "env", nil)

	envMap := map[string]any{}
	envMap["env"] = env

	program, err := expr.Compile("'abc'", expr.Env(envMap), expr.AsAny())
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	output, err := expr.Run(program, env)
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	fmt.Printf("%v\n", reflect.ValueOf(output).Kind())

	//out, err := expr.Eval("env.PWD1 ?? 123", envMap,)
	//if err != nil {
	//	t.Errorf("eval err: %v", err)
	//} else {
	//	t.Errorf("output: %v", reflect.ValueOf(out).Kind())
	//}
}

// copied from github.com/caarlos0/env
type unmarshaler struct {
	time.Duration
}

// TextUnmarshaler implements encoding.TextUnmarshaler.
func (d *unmarshaler) UnmarshalText(data []byte) (err error) {
	if len(data) != 0 {
		d.Duration, err = time.ParseDuration(string(data))
	} else {
		d.Duration = 0
	}
	return err
}

type PointStruct struct {
	NestedNonDefined struct {
		NonDefined struct {
			String string `env:"STR"`
		} `envPrefix:"NONDEFINED_"`
	} `envPrefix:"PRF_"`
}

type WireValueExprStructTest struct {
	String          string    `wire1:"value:${env.STRING}"`
	StringPtr       *string   `wire1:"value:${env.STRING}"`
	Strings         []string  `wire:"value:${env.STRINGS}"`
	StringPtrs      []*string `wire1:"value:${env.STRINGS}"`
	StringPtrInited *string   `wire1:"value:${env.STRINGInited}"`

	Bool     bool    `wire1:"value:${env.BOOL}"`
	BoolPtr  *bool   `wire1:"value:${env.BOOL}"`
	Bools    []bool  `wire1:"value:${env.BOOLS}"`
	BoolPtrs []*bool `wire1:"value:${env.BOOLS}"`

	Int     int    `wire1:"value:${env.INT}"`
	IntPtr  *int   `wire1:"value:${env.INT}"`
	Ints    []int  `wire1:"value:${env.INTS}"`
	IntPtrs []*int `wire1:"value:${env.INTS}"`

	Int8     int8    `wire1:"value:${env.INT8}"`
	Int8Ptr  *int8   `wire1:"value:${env.INT8}"`
	Int8s    []int8  `wire1:"value:${env.INT8S}"`
	Int8Ptrs []*int8 `wire1:"value:${env.INT8S}"`

	Int16     int16    `wire1:"value:${env.INT16}"`
	Int16Ptr  *int16   `wire1:"value:${env.INT16}"`
	Int16s    []int16  `wire1:"value:${env.INT16S}"`
	Int16Ptrs []*int16 `wire1:"value:${env.INT16S}"`

	Int32     int32    `wire1:"value:${env.INT32}"`
	Int32Ptr  *int32   `wire1:"value:${env.INT32}"`
	Int32s    []int32  `wire1:"value:${env.INT32S}"`
	Int32Ptrs []*int32 `wire1:"value:${env.INT32S}"`

	Int64     int64    `wire1:"value:${env.INT64}"`
	Int64Ptr  *int64   `wire1:"value:${env.INT64}"`
	Int64s    []int64  `wire1:"value:${env.INT64S}"`
	Int64Ptrs []*int64 `wire1:"value:${env.INT64S}"`

	Uint     uint    `wire1:"value:${env.UINT}"`
	UintPtr  *uint   `wire1:"value:${env.UINT}"`
	Uints    []uint  `wire1:"value:${env.UINTS}"`
	UintPtrs []*uint `wire1:"value:${env.UINTS}"`

	Uint8     uint8    `wire1:"value:${env.UINT8}"`
	Uint8Ptr  *uint8   `wire1:"value:${env.UINT8}"`
	Uint8s    []uint8  `wire1:"value:${env.UINT8S}"`
	Uint8Ptrs []*uint8 `wire1:"value:${env.UINT8S}"`

	Uint16     uint16    `wire1:"value:${env.UINT16}"`
	Uint16Ptr  *uint16   `wire1:"value:${env.UINT16}"`
	Uint16s    []uint16  `wire1:"value:${env.UINT16S}"`
	Uint16Ptrs []*uint16 `wire1:"value:${env.UINT16S}"`

	Uint32     uint32    `wire1:"value:${env.UINT32}"`
	Uint32Ptr  *uint32   `wire1:"value:${env.UINT32}"`
	Uint32s    []uint32  `wire1:"value:${env.UINT32S}"`
	Uint32Ptrs []*uint32 `wire1:"value:${env.UINT32S}"`

	Uint64     uint64    `wire1:"value:${env.UINT64}"`
	Uint64Ptr  *uint64   `wire1:"value:${env.UINT64}"`
	Uint64s    []uint64  `wire1:"value:${env.UINT64S}"`
	Uint64Ptrs []*uint64 `wire1:"value:${env.UINT64S}"`

	Float32     float32    `wire1:"value:${env.FLOAT32}"`
	Float32Ptr  *float32   `wire1:"value:${env.FLOAT32}"`
	Float32s    []float32  `wire1:"value:${env.FLOAT32S}"`
	Float32Ptrs []*float32 `wire1:"value:${env.FLOAT32S}"`

	Float64     float64    `wire1:"value:${env.FLOAT64}"`
	Float64Ptr  *float64   `wire1:"value:${env.FLOAT64}"`
	Float64s    []float64  `wire1:"value:${env.FLOAT64S}"`
	Float64Ptrs []*float64 `wire1:"value:${env.FLOAT64S}"`

	Duration     time.Duration    `wire1:"value:${env.DURATION}"`
	Durations    []time.Duration  `wire1:"value:${env.DURATIONS}"`
	DurationPtr  *time.Duration   `wire1:"value:${env.DURATION}"`
	DurationPtrs []*time.Duration `wire1:"value:${env.DURATIONS}"`

	Unmarshaler     unmarshaler    `wire1:"value:${env.UNMARSHALER}"`
	UnmarshalerPtr  *unmarshaler   `wire1:"value:${env.UNMARSHALER}"`
	Unmarshalers    []unmarshaler  `wire1:"value:${env.UNMARSHALERS}"`
	UnmarshalerPtrs []*unmarshaler `wire1:"value:${env.UNMARSHALERS}"`

	URL     url.URL    `wire1:"value:${env.URL}"`
	URLPtr  *url.URL   `wire1:"value:${env.URL}"`
	URLs    []url.URL  `wire1:"value:${env.URLS}"`
	URLPtrs []*url.URL `wire1:"value:${env.URLS}"`

	StringWithdefault string `wire1:"value:${env.DATABASE_URL}" envDefault:"postgres://localhost:5432/db"`

	CustomSeparator []string `wire1:"value:${env.SEPSTRINGS}" envSeparator:":"`

	NonDefined struct {
		String string `wire1:"value:${env.NONDEFINED_STR}"`
	}

	NestedNonDefined struct {
		NonDefined struct {
			String string `wire1:"value:${env.STR}"`
		} `envPrefix:"NONDEFINED_"`
	} `envPrefix:"PRF_"`

	NotAnEnv   string
	unexported string `wire1:"value:${env.FOO}"`

	PointStruct
	PS         PointStruct
	PPS        *PointStruct
	PPSNotInit *PointStruct

	SlicePointStruct   []PointStruct
	SlicePPointStruct  []*PointStruct
	PSlicePointStruct  *[]PointStruct
	PSlicePPointStruct *[]PointStruct
}

func Te1stAutoWireEnv(t *testing.T) {
	tos := func(v interface{}) string {
		return fmt.Sprintf("%v", v)
	}

	toss := func(v ...interface{}) string {
		ss := []string{}
		for _, s := range v {
			ss = append(ss, tos(s))
		}
		return strings.Join(ss, ",")
	}

	str1 := "str1"
	str2 := "str2"
	t.Setenv("STRING", str1)
	t.Setenv("STRINGS", toss(str1, str2))

	bool1 := true
	bool2 := false
	t.Setenv("BOOL", tos(bool1))
	t.Setenv("BOOLS", toss(bool1, bool2))

	int1 := -1
	int2 := 2
	t.Setenv("INT", tos(int1))
	t.Setenv("INTS", toss(int1, int2))

	var int81 int8 = -2
	var int82 int8 = 5
	t.Setenv("INT8", tos(int81))
	t.Setenv("INT8S", toss(int81, int82))

	var int161 int16 = -24
	var int162 int16 = 15
	t.Setenv("INT16", tos(int161))
	t.Setenv("INT16S", toss(int161, int162))

	var int321 int32 = -14
	var int322 int32 = 154
	t.Setenv("INT32", tos(int321))
	t.Setenv("INT32S", toss(int321, int322))

	var int641 int64 = -12
	var int642 int64 = 150
	t.Setenv("INT64", tos(int641))
	t.Setenv("INT64S", toss(int641, int642))

	var uint1 uint = 1
	var uint2 uint = 2
	t.Setenv("UINT", tos(uint1))
	t.Setenv("UINTS", toss(uint1, uint2))

	var uint81 uint8 = 15
	var uint82 uint8 = 51
	t.Setenv("UINT8", tos(uint81))
	t.Setenv("UINT8S", toss(uint81, uint82))

	var uint161 uint16 = 532
	var uint162 uint16 = 123
	t.Setenv("UINT16", tos(uint161))
	t.Setenv("UINT16S", toss(uint161, uint162))

	var uint321 uint32 = 93
	var uint322 uint32 = 14
	t.Setenv("UINT32", tos(uint321))
	t.Setenv("UINT32S", toss(uint321, uint322))

	var uint641 uint64 = 5
	var uint642 uint64 = 43
	t.Setenv("UINT64", tos(uint641))
	t.Setenv("UINT64S", toss(uint641, uint642))

	var float321 float32 = 9.3
	var float322 float32 = 1.1
	t.Setenv("FLOAT32", tos(float321))
	t.Setenv("FLOAT32S", toss(float321, float322))

	float641 := 1.53
	float642 := 0.5
	t.Setenv("FLOAT64", tos(float641))
	t.Setenv("FLOAT64S", toss(float641, float642))

	duration1 := time.Second
	duration2 := time.Second * 4
	t.Setenv("DURATION", tos(duration1))
	t.Setenv("DURATIONS", toss(duration1, duration2))

	unmarshaler1 := unmarshaler{time.Minute}
	unmarshaler2 := unmarshaler{time.Millisecond * 1232}
	t.Setenv("UNMARSHALER", tos(unmarshaler1.Duration))
	t.Setenv("UNMARSHALERS", toss(unmarshaler1.Duration, unmarshaler2.Duration))

	url1 := "https://goreleaser.com"
	url2 := "https://caarlos0.dev"
	t.Setenv("URL", tos(url1))
	t.Setenv("URLS", toss(url1, url2))

	t.Setenv("SEPSTRINGS", strings.Join([]string{str1, str2}, ":"))

	nonDefinedStr := "nonDefinedStr"
	t.Setenv("NONDEFINED_STR", nonDefinedStr)
	t.Setenv("PRF_NONDEFINED_STR", nonDefinedStr)

	sample := WireValueExprStructTest{}
	//isNoErr(t, factory.AutoWire(&sample))

	//isEqual(t, str1, sample.String)
	//isEqual(t, &str1, sample.StringPtr)
	isEqual(t, str1, sample.Strings[0])
	isEqual(t, str2, sample.Strings[1])
	isEqual(t, &str1, sample.StringPtrs[0])
	isEqual(t, &str2, sample.StringPtrs[1])

	isEqual(t, bool1, sample.Bool)
	isEqual(t, &bool1, sample.BoolPtr)
	isEqual(t, bool1, sample.Bools[0])
	isEqual(t, bool2, sample.Bools[1])
	isEqual(t, &bool1, sample.BoolPtrs[0])
	isEqual(t, &bool2, sample.BoolPtrs[1])

	isEqual(t, int1, sample.Int)
	isEqual(t, &int1, sample.IntPtr)
	isEqual(t, int1, sample.Ints[0])
	isEqual(t, int2, sample.Ints[1])
	isEqual(t, &int1, sample.IntPtrs[0])
	isEqual(t, &int2, sample.IntPtrs[1])

	isEqual(t, int81, sample.Int8)
	isEqual(t, &int81, sample.Int8Ptr)
	isEqual(t, int81, sample.Int8s[0])
	isEqual(t, int82, sample.Int8s[1])
	isEqual(t, &int81, sample.Int8Ptrs[0])
	isEqual(t, &int82, sample.Int8Ptrs[1])

	isEqual(t, int161, sample.Int16)
	isEqual(t, &int161, sample.Int16Ptr)
	isEqual(t, int161, sample.Int16s[0])
	isEqual(t, int162, sample.Int16s[1])
	isEqual(t, &int161, sample.Int16Ptrs[0])
	isEqual(t, &int162, sample.Int16Ptrs[1])

	isEqual(t, int321, sample.Int32)
	isEqual(t, &int321, sample.Int32Ptr)
	isEqual(t, int321, sample.Int32s[0])
	isEqual(t, int322, sample.Int32s[1])
	isEqual(t, &int321, sample.Int32Ptrs[0])
	isEqual(t, &int322, sample.Int32Ptrs[1])

	isEqual(t, int641, sample.Int64)
	isEqual(t, &int641, sample.Int64Ptr)
	isEqual(t, int641, sample.Int64s[0])
	isEqual(t, int642, sample.Int64s[1])
	isEqual(t, &int641, sample.Int64Ptrs[0])
	isEqual(t, &int642, sample.Int64Ptrs[1])

	isEqual(t, uint1, sample.Uint)
	isEqual(t, &uint1, sample.UintPtr)
	isEqual(t, uint1, sample.Uints[0])
	isEqual(t, uint2, sample.Uints[1])
	isEqual(t, &uint1, sample.UintPtrs[0])
	isEqual(t, &uint2, sample.UintPtrs[1])

	isEqual(t, uint81, sample.Uint8)
	isEqual(t, &uint81, sample.Uint8Ptr)
	isEqual(t, uint81, sample.Uint8s[0])
	isEqual(t, uint82, sample.Uint8s[1])
	isEqual(t, &uint81, sample.Uint8Ptrs[0])
	isEqual(t, &uint82, sample.Uint8Ptrs[1])

	isEqual(t, uint161, sample.Uint16)
	isEqual(t, &uint161, sample.Uint16Ptr)
	isEqual(t, uint161, sample.Uint16s[0])
	isEqual(t, uint162, sample.Uint16s[1])
	isEqual(t, &uint161, sample.Uint16Ptrs[0])
	isEqual(t, &uint162, sample.Uint16Ptrs[1])

	isEqual(t, uint321, sample.Uint32)
	isEqual(t, &uint321, sample.Uint32Ptr)
	isEqual(t, uint321, sample.Uint32s[0])
	isEqual(t, uint322, sample.Uint32s[1])
	isEqual(t, &uint321, sample.Uint32Ptrs[0])
	isEqual(t, &uint322, sample.Uint32Ptrs[1])

	isEqual(t, uint641, sample.Uint64)
	isEqual(t, &uint641, sample.Uint64Ptr)
	isEqual(t, uint641, sample.Uint64s[0])
	isEqual(t, uint642, sample.Uint64s[1])
	isEqual(t, &uint641, sample.Uint64Ptrs[0])
	isEqual(t, &uint642, sample.Uint64Ptrs[1])

	isEqual(t, float321, sample.Float32)
	isEqual(t, &float321, sample.Float32Ptr)
	isEqual(t, float321, sample.Float32s[0])
	isEqual(t, float322, sample.Float32s[1])
	isEqual(t, &float321, sample.Float32Ptrs[0])

	isEqual(t, float641, sample.Float64)
	isEqual(t, &float641, sample.Float64Ptr)
	isEqual(t, float641, sample.Float64s[0])
	isEqual(t, float642, sample.Float64s[1])
	isEqual(t, &float641, sample.Float64Ptrs[0])
	isEqual(t, &float642, sample.Float64Ptrs[1])

	isEqual(t, duration1, sample.Duration)
	isEqual(t, &duration1, sample.DurationPtr)
	isEqual(t, duration1, sample.Durations[0])
	isEqual(t, duration2, sample.Durations[1])
	isEqual(t, &duration1, sample.DurationPtrs[0])
	isEqual(t, &duration2, sample.DurationPtrs[1])

	isEqual(t, unmarshaler1, sample.Unmarshaler)
	isEqual(t, &unmarshaler1, sample.UnmarshalerPtr)
	isEqual(t, unmarshaler1, sample.Unmarshalers[0])
	isEqual(t, unmarshaler2, sample.Unmarshalers[1])
	isEqual(t, &unmarshaler1, sample.UnmarshalerPtrs[0])
	isEqual(t, &unmarshaler2, sample.UnmarshalerPtrs[1])

	isEqual(t, url1, sample.URL.String())
	isEqual(t, url1, sample.URLPtr.String())
	isEqual(t, url1, sample.URLs[0].String())
	isEqual(t, url2, sample.URLs[1].String())
	isEqual(t, url1, sample.URLPtrs[0].String())
	isEqual(t, url2, sample.URLPtrs[1].String())

	isEqual(t, "postgres://localhost:5432/db", sample.StringWithdefault)
	isEqual(t, nonDefinedStr, sample.NonDefined.String)
	isEqual(t, nonDefinedStr, sample.NestedNonDefined.NonDefined.String)

	isEqual(t, str1, sample.CustomSeparator[0])
	isEqual(t, str2, sample.CustomSeparator[1])

	isEqual(t, sample.NotAnEnv, "")

	isEqual(t, sample.unexported, "")
}

// copied from https://github.com/matryer/is
func isNil(object interface{}) bool {
	if object == nil {
		return true
	}
	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}
	return false
}

// copied from https://github.com/matryer/is
func areEqual(a, b interface{}) bool {
	if isNil(a) && isNil(b) {
		return true
	}
	if isNil(a) || isNil(b) {
		return false
	}
	if reflect.DeepEqual(a, b) {
		return true
	}
	aValue := reflect.ValueOf(a)
	bValue := reflect.ValueOf(b)
	return aValue == bValue
}

func isEqual(tb testing.TB, a, b interface{}) {
	tb.Helper()

	if areEqual(a, b) {
		return
	}

	tb.Fatalf("expected %#v (type %T) == %#v (type %T)", a, a, b, b)
}

func isNoErr(tb testing.TB, err error) {
	tb.Helper()

	if err != nil {
		tb.Fatalf("unexpected error: %v", err)
	}
}
