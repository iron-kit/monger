package monger

import (
	"reflect"
	"strings"
	"time"
)

var typeTime = reflect.TypeOf(time.Time{})

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return len(v.String()) == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Slice:
		return v.Len() == 0
	case reflect.Map:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Struct:
		vt := v.Type()
		if vt == typeTime {
			return v.Interface().(time.Time).IsZero()
		}
		for i := 0; i < v.NumField(); i++ {
			if vt.Field(i).PkgPath != "" && !vt.Field(i).Anonymous {
				continue // Private field
			}
			if !isZero(v.Field(i)) {
				return false
			}
		}
		return true
	}
	return false
}

func parseTagConfig(tags reflect.StructTag) map[string]string {
	conf := map[string]string{}
	for index, str := range []string{tags.Get("bson"), tags.Get("monger")} {
		// bson
		if index == 0 {
			tags := strings.Split(str, ",")

			for _, tag := range tags {
				// if t == "" {
				// 	continue
				// }
				switch tag {
				case "inline":
					conf["INLINE"] = "true"
				case "omitempty":
					conf["OMITEMPTY"] = "true"
				default:
					conf["COLUMN"] = tag
				}
			}

			continue
		}
		tags := strings.Split(str, ",")
		for _, value := range tags {
			v := strings.Split(value, "=")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if k == "COLUMN" {
				continue
			}
			if len(v) >= 2 {
				conf[k] = strings.Join(v[1:], "=")
			} else {
				conf[k] = k
			}
		}
	}

	return conf
}

func buildPopulateTree(populate []string) map[string]interface{} {
	tree := make(map[string]interface{})
	for _, popStr := range populate {
		pop := strings.Split(popStr, ".")
		k := strings.TrimSpace(strings.ToUpper(pop[0]))
		if len(pop) >= 2 {
			tree[k] = buildPopulateTree([]string{strings.Join(pop[1:], ".")})
		} else {
			tree[k] = k
		}
	}

	return tree
}

func snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}

func camelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}
