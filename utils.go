package monger

import (
	"errors"
	"reflect"
	"strings"
	"sync"
)

func parseFieldTags(tag reflect.StructTag) map[string]string {
	tagMap := map[string]string{}
	tagMonger := tag.Get("monger")
	mongerTags := strings.Split(tagMonger, ";")

	for _, t := range mongerTags {
		if t == "" {
			continue
		}
		tgArr := strings.Split(t, ":")
		if len(tgArr) > 1 {
			tagMap[tgArr[0]] = tgArr[1]
		} else {
			tagMap[tgArr[0]] = "true"
		}
	}

	return tagMap
}

func parseFieldTag(tag string) map[string]string {
	tagMap := map[string]string{}
	mongerTags := strings.Split(tag, ";")

	for _, t := range mongerTags {
		if t == "" {
			continue
		}
		tgArr := strings.Split(t, ":")
		if len(tgArr) > 1 {
			tagMap[tgArr[0]] = tgArr[1]
		} else {
			tagMap[tgArr[0]] = "true"
		}
	}

	return tagMap
}

type docStructInfo struct {
	FieldsMap  map[string]docFieldInfo
	FieldsList []docFieldInfo
	InlineMap  int
	Zero       reflect.Value
}

type docFieldInfo struct {
	Key       string
	Num       int
	OmitEmpty bool
	MinSize   bool
	Relate    string
	Inline    []int
}

var structMap = make(map[reflect.Type]*docStructInfo)
var structMapMutex sync.RWMutex

func getDocumentStructInfo(st reflect.Type) (*docStructInfo, error) {
	structMapMutex.RLock()
	sinfo, found := structMap[st]
	structMapMutex.RUnlock()
	if found {
		return sinfo, nil
	}
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	}
	n := st.NumField()
	inlineMap := -1
	fieldsMap := make(map[string]docFieldInfo)
	fieldsList := make([]docFieldInfo, 0, n)
	for i := 0; i != n; i++ {
		field := st.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue // Private field
		}

		info := docFieldInfo{
			Num:    i,
			Relate: "",
		}

		tag := field.Tag.Get("monger")

		if tag == "-" {
			continue
		}

		tags := parseFieldTag(tag)
		inline := false
		for k, v := range tags {
			switch k {
			case "hasOne":
				info.Relate = HasOne
				break
			case "hasMany":
				info.Relate = HasMany
				break
			case "belongTo":
				info.Relate = BelongTo
				break
			case "inline":
				inline = true
				// info.Inline = true
				break
			case "column":
				info.Key = v
				break
			default:
				break
			}
		}

		if inline {
			switch field.Type.Kind() {
			case reflect.Map:
				if inlineMap >= 0 {
					return nil, errors.New("Multiple ,inline maps in struct " + st.String())
				}
				if field.Type.Key() != reflect.TypeOf("") {
					return nil, errors.New("Option ,inline needs a map with string keys in struct " + st.String())
				}
				inlineMap = info.Num
			case reflect.Struct:
				sinfo, err := getDocumentStructInfo(field.Type)
				if err != nil {
					return nil, err
				}
				for _, finfo := range sinfo.FieldsList {
					if _, found := fieldsMap[finfo.Key]; found {
						msg := "Duplicated key '" + finfo.Key + "' in struct " + st.String()
						return nil, errors.New(msg)
					}
					if finfo.Inline == nil {
						finfo.Inline = []int{i, finfo.Num}
					} else {
						finfo.Inline = append([]int{i}, finfo.Inline...)
					}
					fieldsMap[finfo.Key] = finfo
					fieldsList = append(fieldsList, finfo)
				}
			default:
				panic("Option ,inline needs a struct value or map field")
			}
			continue
		}

		if info.Key == "" {
			info.Key = strings.ToLower(field.Name)
		}

		if _, found = fieldsMap[info.Key]; found {
			msg := "Duplicated key '" + info.Key + "' in struct " + st.String()
			return nil, errors.New(msg)
		}

		fieldsList = append(fieldsList, info)
		fieldsMap[info.Key] = info
	}

	sinfo = &docStructInfo{
		fieldsMap,
		fieldsList,
		inlineMap,
		reflect.New(st).Elem(),
	}

	structMapMutex.Lock()
	structMap[st] = sinfo
	structMapMutex.Unlock()
	return sinfo, nil
}
