package command

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

func ternaryString(cond bool, whenTrue, whenFalse string) string {
	if cond {
		return whenTrue
	}
	return whenFalse
}

func ternaryAny(cond bool, whenTrue, whenFalse any) any {
	if cond {
		return whenTrue
	}
	return whenFalse
}

func mustJSON(v any) string {
	payload, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("json marshal failed: %v", err))
	}
	return string(payload)
}

func jsJoin(v any, sep string) string {
	slice, ok := toSlice(v)
	if !ok {
		panic(fmt.Sprintf("join on non-array %T", v))
	}
	parts := make([]string, len(slice))
	for i := range slice {
		parts[i] = jsToString(slice[i])
	}
	return strings.Join(parts, sep)
}

func jsToString(v any) string {
	v = unwrapPointers(v)
	switch x := v.(type) {
	case nil:
		return "null"
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case float64:
		return formatJSNumber(x)
	case float32:
		return formatJSNumber(float64(x))
	case int:
		return strconv.FormatInt(int64(x), 10)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case json.Number:
		return x.String()
	}

	if arr, ok := toSlice(v); ok {
		parts := make([]string, len(arr))
		for i := range arr {
			parts[i] = jsToString(arr[i])
		}
		return strings.Join(parts, ",")
	}

	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() == reflect.Map {
		return "[object Object]"
	}

	return fmt.Sprint(v)
}

func jsTruthy(v any) bool {
	v = unwrapPointers(v)
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case float64:
		return x != 0 && !math.IsNaN(x)
	case float32:
		f := float64(x)
		return f != 0 && !math.IsNaN(f)
	case int:
		return x != 0
	case int8:
		return x != 0
	case int16:
		return x != 0
	case int32:
		return x != 0
	case int64:
		return x != 0
	case uint:
		return x != 0
	case uint8:
		return x != 0
	case uint16:
		return x != 0
	case uint32:
		return x != 0
	case uint64:
		return x != 0
	default:
		return true
	}
}

func jsTypeOf(v any) string {
	v = unwrapPointers(v)
	switch v.(type) {
	case nil:
		// JS quirk: typeof null === "object"
		return "object"
	case bool:
		return "boolean"
	case string:
		return "string"
	case float64, float32, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "number"
	default:
		return "object"
	}
}

func jsLooseEqual(left, right any) bool {
	left = unwrapPointers(left)
	right = unwrapPointers(right)

	if lnum, ok := toNumber(left); ok {
		if rnum, ok := toNumber(right); ok {
			return !math.IsNaN(lnum) && !math.IsNaN(rnum) && lnum == rnum
		}
	}

	switch l := left.(type) {
	case string:
		r, ok := right.(string)
		return ok && l == r
	case bool:
		r, ok := right.(bool)
		return ok && l == r
	case nil:
		return right == nil
	default:
		return false
	}
}

func unwrapPointers(v any) any {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil
	}
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	return rv.Interface()
}

func toSlice(v any) ([]any, bool) {
	v = unwrapPointers(v)
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil, false
	}
	switch rv.Kind() {
	case reflect.Array, reflect.Slice:
	default:
		return nil, false
	}
	out := make([]any, rv.Len())
	for i := range out {
		out[i] = rv.Index(i).Interface()
	}
	return out, true
}

func toNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func formatJSNumber(v float64) string {
	switch {
	case math.IsNaN(v):
		return "NaN"
	case math.IsInf(v, 1):
		return "Infinity"
	case math.IsInf(v, -1):
		return "-Infinity"
	case v == 0:
		return "0"
	default:
		return strconv.FormatFloat(v, 'f', -1, 64)
	}
}
