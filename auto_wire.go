package factory

import (
	"encoding"
	"errors"
	"fmt"
	"github.com/expgo/generic/stream"
	"github.com/expgo/structure"
	"reflect"
	"strings"
)

//go:generate bud

// WireTag is a constant that defines the annotation string used for wire injection in Go code.
const WireTag = "wire"

// ValueTag is a shortcut of wire:"value:"
const ValueTag = "value"

// WireValue is an enum
// @EnumConfig(marshal, values, noComments, noCase)
// @Enum{self, auto, type, name, value}
type WireValue string

type TagValue[T any] struct {
	Tag   T
	Value string
}

func (tv *TagValue[T]) String() string {
	return fmt.Sprintf("%v:%s", tv.Tag, tv.Value)
}

func ParseTagValue(tagValue string, checkAndSet func(tv *TagValue[WireValue])) (tv *TagValue[WireValue], err error) {
	result := &TagValue[WireValue]{}
	values := stream.Must(stream.Of(strings.SplitN(strings.TrimSpace(tagValue), ":", 2)).
		Map(func(s string) (string, error) { return strings.TrimSpace(s), nil }).
		Filter(func(s string) (bool, error) { return len(s) > 0, nil }).ToSlice())

	if len(values) == 0 {
		return nil, errors.New("tag value is empty")
	}

	if unmarshaler, ok := any(&result.Tag).(encoding.TextUnmarshaler); ok {
		if err = unmarshaler.UnmarshalText([]byte(values[0])); err != nil {
			if len(values) == 1 {
				result.Tag = WireValueValue
				result.Value = values[0]
				return result, nil
			} else {
				return nil, err
			}
		} else {
			if len(values) == 2 {
				result.Value = values[1]
			}

			if checkAndSet != nil {
				checkAndSet(result)
			}

			return result, nil
		}
	} else {
		panic("parse type muse implements encoding.TextUnmarshaler")
	}
}

func wireError(structField reflect.StructField, rootValues []reflect.Value, wireRule string) error {
	fieldPath := structure.GetFieldPath(structField, rootValues)
	return errors.New(fmt.Sprintf("The field of 'wire' must be defined as a pointer to an object or an interface. %s, tag value: %s", fieldPath, wireRule))
}

func getExpr(value string) (exprCode string, isExpr bool) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		return strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}"), true
	}
	return value, false
}

func getValueByWireTag(self any, tagValue *TagValue[WireValue], t reflect.Type) (any, error) {
	switch tagValue.Tag {
	case WireValueSelf, WireValueAuto, WireValueType, WireValueName:
		if (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) || t.Kind() == reflect.Interface {
			switch tagValue.Tag {
			case WireValueSelf:
				return self, nil
			case WireValueAuto:
				if len(tagValue.Value) > 0 {
					return _context.getByNameOrType(tagValue.Value, t), nil
				} else {
					return _context.getByType(t), nil
				}
			case WireValueType:
				return _context.getByType(t), nil
			case WireValueName:
				if len(tagValue.Value) > 0 {
					return _context.getByName(tagValue.Value), nil
				}
			}
		} else {
			return nil, errors.New("‘self’, ’auto’, ’type’ and ’name’ tag only used on *struct or interface")
		}
	case WireValueValue:
		if (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) || t.Kind() == reflect.Struct {
			return nil, errors.New("'value' tag can't used on *struct or struct")
		}

		if len(tagValue.Value) > 0 {

			exprCode, isExpr := getExpr(tagValue.Value)
			if isExpr {
				value, err := _context.evalExpr(exprCode)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("Tag value %s expr eval err: %v", tagValue, err))
				}

				return structure.ConvertToType(value, t)
			} else {
				return structure.ConvertToType(exprCode, t)
			}
		}
	}

	return nil, errors.New(fmt.Sprintf("Tag value not supported: %+v", tagValue))
}

func AutoWire(self any) error {
	if self == nil {
		return nil
	}

	vt := reflect.TypeOf(self)
	_context.wiring(vt)
	defer _context.wired(vt)

	return structure.WalkWithTagNames(self, []string{WireTag, ValueTag}, func(fieldValue reflect.Value, structField reflect.StructField, rootValues []reflect.Value, tags map[string]string) (err error) {
		if len(tags) > 1 {
			panic("Only one can exist at a time, either 'wire' or 'value'.")
		}

		var tv *TagValue[WireValue]
		if wireValue, ok := tags[WireTag]; ok {
			tv, err = ParseTagValue(wireValue, func(tv *TagValue[WireValue]) {
				if (tv.Tag == WireValueName && len(tv.Value) == 0) ||
					(tv.Tag == WireValueAuto) {
					tv.Value = structField.Name
				}
			})
		}
		if wireValue, ok := tags[ValueTag]; ok {
			tv = &TagValue[WireValue]{Tag: WireValueValue, Value: wireValue}
		}

		if err != nil {
			panic(err)
		}

		switch tv.Tag {
		case WireValueSelf, WireValueAuto, WireValueType, WireValueName:
			if !fieldValue.IsNil() {
				// field is not nil， skip it
				return nil
			}
		default:
		}

		if wiredValue, err1 := getValueByWireTag(self, tv, structField.Type); err1 == nil {
			// Prefer using the set method
			if structure.SetFieldBySetMethod(fieldValue, wiredValue, structField, rootValues[len(rootValues)-1]) {
				return nil
			}
			return structure.SetField(fieldValue, wiredValue)
		} else {
			return errors.New(fmt.Sprintf("%v on %s", err1, structure.GetFieldPath(structField, rootValues)))
		}
	})
}
