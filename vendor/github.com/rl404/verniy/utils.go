package verniy

import (
	"fmt"
	"reflect"
	"strings"
)

// FieldObject to generate query string.
//
// Example:
//
//   FieldObject("key", map[string]interface{}{
// 	  "name1": "value1",
// 	  "name2": []string{"value2", "value3"},
//   }, "field1", "field2")
//
// Will generate:
//
//   key(name1:value1, name2:[value2,value3]) {
//     field1
//     field2
//   }
//
// Or just see usage example in some functions in verniy package.
func FieldObject(key string, params map[string]interface{}, fields ...string) string {
	var paramStr string
	if params != nil || len(params) > 0 {
		var paramArr []string
		for k, p := range params {
			if p == nil {
				continue
			}

			switch reflect.TypeOf(p).Kind() {
			case reflect.Slice, reflect.Array:
				v := reflect.ValueOf(p)
				if v.Len() > 0 {
					vArr := make([]string, v.Len())
					for i := 0; i < v.Len(); i++ {
						vArr[i] = fmt.Sprintf("%v", v.Index(i))
					}
					paramArr = append(paramArr, fmt.Sprintf("%v:[%v]", k, strings.Join(vArr, ",")))
				}
			case reflect.String:
				if reflect.ValueOf(p).Len() != 0 {
					paramArr = append(paramArr, fmt.Sprintf("%v:%v", k, p))
				}
			case reflect.Int:
				if reflect.ValueOf(p).Int() != 0 {
					paramArr = append(paramArr, fmt.Sprintf("%v:%v", k, p))
				}
			case reflect.Ptr:
				if !reflect.ValueOf(p).IsNil() {
					paramArr = append(paramArr, fmt.Sprintf("%v:%v", k, reflect.ValueOf(p).Elem()))
				}
			default:
				paramArr = append(paramArr, fmt.Sprintf("%v:%v", k, p))
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
