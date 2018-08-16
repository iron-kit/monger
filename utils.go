package monger

import (
	"errors"
	"reflect"
	"strings"
	"sync"
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

func parseFieldTag(tags ...string) map[string]string {

	tagMap := map[string]string{}

	for _, tag := range tags {
		// mongerTags := []string{}

		mongerTags := strings.Split(tag, ";")
		if len(mongerTags) == 1 {
			mongerTags = strings.Split(tag, ",")

			for _, t := range mongerTags {
				// if t == "" {
				// 	continue
				// }
				switch t {
				case "inline":
					tagMap["inline"] = "true"
				case "omitempty":
					tagMap["omitempty"] = "true"
				default:
					tagMap["column"] = t
				}
			}
			continue
		}

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

	}
	return tagMap
}

type docStructInfo struct {
	FieldsMap        map[string]docFieldInfo
	FieldsList       []docFieldInfo
	RelateFieldsMap  map[string]docFieldInfo
	RelateFieldsList []docFieldInfo
	InlineMap        int
	Zero             reflect.Value
}

type docFieldInfo struct {
	Key        string
	Num        int
	OmitEmpty  bool
	MinSize    bool
	Relate     string
	RelateType reflect.Type
	RelateZero reflect.Value
	Foreignkey string
	Inline     []int
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
	relateFieldsMap := make(map[string]docFieldInfo)
	relateFieldsList := make([]docFieldInfo, 0, 1)
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
		bsonTag := field.Tag.Get("bson")
		if tag == "-" || bsonTag == "-" {
			continue
		}

		var documenter Documenter
		doct := reflect.TypeOf(&documenter).Elem()
		// guess relationship
		// if field.Type.Implements()
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Slice {
			info.Relate = HasMany
			info.RelateType = field.Type.Elem()
			// info.RelateZero = reflect.New(info.RelateType)
		} else if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Implements(doct) {
			info.Relate = HasOne
			info.RelateType = field.Type.Elem()
			// info.RelateZero = reflect.New(info.RelateType)
		} else if field.Type.Implements(doct) {
			info.Relate = HasOne
			info.RelateType = field.Type
		}

		if info.RelateType != nil {
			info.RelateZero = reflect.New(info.RelateType)
		}
		tags := parseFieldTag(tag, bsonTag)
		inline := false
		for k, v := range tags {
			switch k {
			case "hasOne":
				info.Relate = HasOne
			case "hasMany":
				info.Relate = HasMany
			case "belongTo":
				info.Relate = BelongTo
			case "inline":
				inline = true
				// info.Inline = true
			case "column":
				info.Key = v
			case "foreignkey":
				info.Foreignkey = v
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

		if _, found = relateFieldsMap[info.Key]; found {
			msg := "Duplicated Relate key '" + info.Key + "' in struct " + st.String()
			return nil, errors.New(msg)
		}

		fieldsList = append(fieldsList, info)
		fieldsMap[info.Key] = info
		if info.Relate != "" {
			relateFieldsMap[info.Key] = info
			relateFieldsList = append(relateFieldsList, info)
		}
	}

	sinfo = &docStructInfo{
		FieldsMap:        fieldsMap,
		FieldsList:       fieldsList,
		RelateFieldsList: relateFieldsList,
		RelateFieldsMap:  relateFieldsMap,
		InlineMap:        inlineMap,
		Zero:             reflect.New(st).Elem(),
		// FiefieldsList,
		// relateFieldsMap,
		// relateFieldsList,
		// inlineMap,
		// reflect.New(st).Elem(),
	}

	structMapMutex.Lock()
	structMap[st] = sinfo
	structMapMutex.Unlock()
	return sinfo, nil
}
