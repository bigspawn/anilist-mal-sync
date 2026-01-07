package verniy

import (
	"fmt"
	"reflect"
	"strings"
)

// formatParamValue formats a single parameter value for the query string.
// It returns the formatted string and a boolean indicating if the value should be included.
func formatParamValue(p interface{}) (string, bool) {
	if p == nil {
		return "", false
	}
	v := reflect.ValueOf(p)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return "", false
		}
		vArr := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			vArr[i] = fmt.Sprintf("%v", v.Index(i))
		}
		return fmt.Sprintf("[%v]", strings.Join(vArr, ",")), true
	case reflect.String:
		if v.Len() == 0 {
			return "", false
		}
		return fmt.Sprintf("%v", p), true
	case reflect.Int:
		if v.Int() == 0 {
			return "", false
		}
		return fmt.Sprintf("%v", p), true
	case reflect.Ptr:
		if v.IsNil() {
			return "", false
		}
		return fmt.Sprintf("%v", v.Elem()), true
	default:
		return fmt.Sprintf("%v", p), true
	}
}

// FieldObject to generate query string.
//
// Example:
//
//	  FieldObject("key", map[string]interface{}{
//		  "name1": "value1",
//		  "name2": []string{"value2", "value3"},
//	  }, "field1", "field2")
//
// Will generate:
//
//	key(name1:value1, name2:[value2,value3]) {
//	  field1
//	  field2
//	}
//
// Or just see usage example in some functions in verniy package.
func FieldObject(key string, params map[string]interface{}, fields ...string) string {
	var paramStr string
	if len(params) > 0 {
		var paramArr []string
		for k, p := range params {
			formattedVal, ok := formatParamValue(p)
			if ok {
				paramArr = append(paramArr, fmt.Sprintf("%v:%v", k, formattedVal))
			}
		}
		if len(paramArr) > 0 {
			paramStr = fmt.Sprintf("(%s)", strings.Join(paramArr, ","))
		}
	}
	var fieldStr string
	if len(fields) > 0 {
		fieldStr = fmt.Sprintf("{%s}", strings.Join(fields, " "))
	}
	return fmt.Sprintf("%s%s%s", key, paramStr, fieldStr)
}

func toQueryString(str string) string {
	if str == "" {
		return str
	}
	return fmt.Sprintf(`"%s"`, str)
}

func toQueryStringArray(strs []string) []string {
	for i := range strs {
		strs[i] = fmt.Sprintf(`"%s"`, strs[i])
	}
	return strs
}
