package factory

import (
	"context"
	"encoding"
	"errors"
	"fmt"
	"github.com/expgo/structure"
	"reflect"
	"strings"
	"time"
)

//go:generate ag

// Tag is an enum
// @Enum{wire, value, new}
type Tag string

// WireValue is an enum
// @EnumConfig(marshal, values, noComments, noCase)
// @Enum{self, auto, type, name, value}
type WireValue string

type TagWithValue struct {
	Tag   WireValue
	Value string
}

func (tv *TagWithValue) String() string {
	return fmt.Sprintf("%v:%s", tv.Tag, tv.Value)
}

func ParseTagValue(tagValue string, checkAndSet func(tv *TagWithValue)) (tv *TagWithValue, err error) {
	result := &TagWithValue{}

	var values []string
	for _, v := range strings.SplitN(strings.TrimSpace(tagValue), ":", 2) {
		if s := strings.TrimSpace(v); len(s) > 0 {
			values = append(values, s)
		}
	}

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
	return fmt.Errorf("the field of 'wire' must be defined as a pointer to an object or an interface. %s, tag value: %s", fieldPath, wireRule)
}

func getExpr(value string) (exprCode string, isExpr bool) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		return strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}"), true
	}
	return value, false
}

func getValueByWireTag(ctx context.Context, self any, tagValue *TagWithValue, t reflect.Type) (any, error) {
	switch tagValue.Tag {
	case WireValueSelf, WireValueAuto, WireValueType, WireValueName:
		if (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) || t.Kind() == reflect.Interface {
			switch tagValue.Tag {
			case WireValueSelf:
				return self, nil
			case WireValueAuto:
				if len(tagValue.Value) > 0 {
					return _context.getByNameOrType(ctx, tagValue.Value, t), nil
				} else {
					return _context.getByType(ctx, t), nil
				}
			case WireValueType:
				return _context.getByType(ctx, t), nil
			case WireValueName:
				if len(tagValue.Value) > 0 {
					return _context.getByNamePanic(ctx, tagValue.Value, t), nil
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
				value, err := _context.evalExpr(ctx, exprCode)
				if err != nil {
					return nil, fmt.Errorf("tag value %s expr eval err: %v", tagValue, err)
				}

				return structure.ConvertToType(value, t)
			} else {
				return structure.ConvertToType(exprCode, t)
			}
		}
	}

	return nil, fmt.Errorf("tag value not supported: %+v", tagValue)
}

func AutoWire(self any) error {
	return AutoWireTimeout(self, Opts.Timeout)
}

func AutoWireTimeout(self any, timeout time.Duration) error {
	return autoWireContext(getTimeoutContext(timeout), self)
}

func autoWireContext(ctx context.Context, self any) error {
	if self == nil {
		return nil
	}

	return structure.WalkWithTagNames(self, []string{TagWire.Name(), TagValue.Name(), TagNew.Name()}, func(fieldValue reflect.Value, structField reflect.StructField, rootValues []reflect.Value, tags map[string]string) (err error) {
		if len(tags) > 1 {
			panic("Only one can exist at a time, either 'wire' or 'value'.")
		}

		if newValue, ok := tags[TagNew.Name()]; ok {
			// new， create by factory
			factoriesLock.RLock()
			f, loaded := factories[structField.Type]
			factoriesLock.RUnlock()

			if !loaded {
				return fmt.Errorf("can't get factory type of %s", structField.Type.String())
			}
			newValue = strings.TrimSpace(newValue)
			if len(newValue) > 0 {
				return callFactory(ctx, f, self, fieldValue, structField, strings.Split(newValue, ","))
			} else {
				return callFactory(ctx, f, self, fieldValue, structField, nil)
			}
		}

		var tv *TagWithValue
		if wireValue, ok := tags[TagWire.Name()]; ok {
			tv, err = ParseTagValue(wireValue, func(tv *TagWithValue) {
				if (tv.Tag == WireValueName && len(tv.Value) == 0) ||
					(tv.Tag == WireValueAuto) {
					tv.Value = structField.Name
				}
			})
		}
		if wireValue, ok := tags[TagValue.Name()]; ok {
			tv = &TagWithValue{Tag: WireValueValue, Value: wireValue}
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

		if wiredValue, err1 := getValueByWireTag(ctx, self, tv, structField.Type); err1 == nil {
			// Prefer using the set method
			if structure.SetFieldBySetMethod(fieldValue, wiredValue, structField, rootValues[len(rootValues)-1]) {
				return nil
			}
			return structure.SetField(fieldValue, wiredValue)
		} else {
			return fmt.Errorf("%v on %s", err1, structure.GetFieldPath(structField, rootValues))
		}
	})
}
