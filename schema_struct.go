package monger

import (
	"go/ast"
	"reflect"
	"sync"
)

var schemaStructsMap sync.Map

type SchemaStruct struct {
	Type           reflect.Type
	Fields         []*SchemaField
	FieldsMap      map[string]*SchemaField
	RelationFields []*SchemaField
}

type SchemaField struct {
	Name               string
	Index              int
	IsIgnored          bool // 是否为隐藏字段 bool
	InlineIndex        []int
	ColumnName         string
	IsInline           bool
	Tag                reflect.StructTag
	TagMap             map[string]string
	Struct             reflect.StructField
	IsForeignKey       bool
	Relationship       *Relationship
	HasRelation        bool
	Zero               reflect.Value
	RelationshipStruct *SchemaStruct
}

func GetSchemaStruct(schema interface{}) *SchemaStruct {
	var schemaStruct SchemaStruct
	var schemaType reflect.Type

	if schema == nil {
		return &schemaStruct
	}

	if s, ok := schema.(reflect.Type); ok {
		schemaType = s
	} else {
		schemaType = reflect.TypeOf(schema)
	}

	for {
		if schemaType.Kind() != reflect.Ptr {
			break
		}

		schemaType = schemaType.Elem()
	}

	if schemaType.Kind() == reflect.Slice {
		schemaType = schemaType.Elem()
	}

	// Documenter must be a struct
	if schemaType.Kind() != reflect.Struct {
		return &schemaStruct
	}

	if v, found := schemaStructsMap.Load(schemaType); found && v != nil {
		return v.(*SchemaStruct)
	}

	schemaStruct = SchemaStruct{}
	schemaStruct.Type = schemaType
	schemaStruct.FieldsMap = make(map[string]*SchemaField)
	schemaStruct.Fields = make([]*SchemaField, 0)
	schemaStruct.RelationFields = make([]*SchemaField, 0)

	// 遍历每个字段，依次处理
	for i := 0; i < schemaType.NumField(); i++ {

		if field := schemaType.Field(i); ast.IsExported(field.Name) {

			tagMap := parseTagConfig(field.Tag)

			schemaField := &SchemaField{
				Struct:      field,
				Name:        field.Name,
				Tag:         field.Tag,
				TagMap:      tagMap,
				Zero:        reflect.New(field.Type).Elem(),
				Index:       i,
				InlineIndex: []int{i},
				IsInline:    false,
			}

			if _, found := tagMap["-"]; found {
				schemaField.IsIgnored = true
			}

			if _, ok := tagMap["DEFAULT"]; ok {
				// schemaField.
				// default value
			}
			if name, ok := tagMap["COLUMN"]; ok {
				schemaField.ColumnName = name
			}

			// the field is inline
			if v, foundInline := tagMap["INLINE"]; foundInline && v == "true" {

				inlineSchemaStruct := GetSchemaStruct(field.Type)

				for _, inlineField := range inlineSchemaStruct.Fields {
					inlineField.IsInline = true
					inlineField.InlineIndex = []int{i, inlineField.Index}
					schemaStruct.Fields = append(schemaStruct.Fields, inlineField)
				}

				continue
			}

			indirectType := field.Type
			for indirectType.Kind() == reflect.Ptr {
				indirectType = indirectType.Elem()
			}

			// if isn't inline execute relations fields
			// if schemaField.IsInline {
			// 	continue
			// }
			// fmt.Println(schemaStruct)
			switch indirectType.Kind() {
			// OneToMany / ManyToMany
			case reflect.Slice:
				schemaField.HasRelation = true
				defer func(field *SchemaField) {
					var (
						localFieldKey   string
						foreignFieldKey string
						elemValue       reflect.Value
						elemType        = field.Struct.Type
					)

					for elemType.Kind() == reflect.Ptr || elemType.Kind() == reflect.Slice {
						elemType = elemType.Elem()
					}

					elemValue = reflect.New(elemType)

					if foreignKey := field.TagMap["FOREIGNKEY"]; foreignKey != "" {
						localFieldKey = "_id"
						foreignFieldKey = foreignKey
					}

					if localField := field.TagMap["LOCALFIELD"]; localField != "" {
						localFieldKey = localField
					}

					if foreignField := field.TagMap["FOREIGNFIELD"]; foreignField != "" {
						foreignFieldKey = foreignField
					}

					if elemType.Kind() == reflect.Struct && isImplementsSchemer(elemType) {
						collectionName := getCollectionName(elemValue.Interface())
						rs := &Relationship{
							RelationType:    elemType,
							From:            collectionName,
							CollectionName:  collectionName,
							As:              field.ColumnName,
							LocalFieldKey:   localFieldKey,
							ForeignFieldKey: foreignFieldKey,
						}
						if _, ok := field.TagMap["HASMANY"]; ok {
							rs.Kind = HasMany
						} else {
							// now just support has many, don't support many to many
							return
						}

						field.Relationship = rs
						v := reflect.New(elemType)
						field.RelationshipStruct = GetSchemaStruct(v.Interface())
						schemaStruct.RelationFields = append(schemaStruct.RelationFields, field)
						// docStruct.RelationFields = append(docStruct.RelationFields, field)
					}
				}(schemaField)
			case reflect.Struct:
				fallthrough
			case reflect.Ptr:
				schemaField.HasRelation = true
				defer func(field *SchemaField) {
					var (
						localFieldKey   string
						foreignFieldKey string
						elemValue       reflect.Value
						elemType        = field.Struct.Type
					)

					for elemType.Kind() == reflect.Ptr {
						elemType = elemType.Elem()
					}

					if !isImplementsSchemer(elemType) {
						return
					}

					elemValue = reflect.New(elemType)
					collectionName := getCollectionName(elemValue.Interface())
					rs := &Relationship{
						RelationType:   elemType,
						From:           collectionName,
						CollectionName: collectionName,
						As:             field.ColumnName,
					}

					if _, ok := field.TagMap["HASONE"]; ok {
						rs.Kind = HasOne

					} else if _, ok := field.TagMap["BELONGTO"]; ok {
						rs.Kind = BelongTo
					}

					if rs.Kind == "" {
						return
					}

					if foreignKey := field.TagMap["FOREIGNKEY"]; foreignKey != "" {
						if rs.Kind == HasOne {
							localFieldKey = "_id"
							foreignFieldKey = foreignKey
						} else {
							localFieldKey = foreignKey
							foreignFieldKey = "_id"
						}
					}

					if localField := field.TagMap["LOCALFIELD"]; localField != "" {
						localFieldKey = localField
					}

					if foreignField := field.TagMap["FOREIGNFIELD"]; foreignField != "" {
						foreignFieldKey = foreignField
					}
					rs.ForeignFieldKey = foreignFieldKey
					rs.LocalFieldKey = localFieldKey
					field.Relationship = rs
					v := reflect.New(elemType)
					field.RelationshipStruct = GetSchemaStruct(v.Interface())
					schemaStruct.RelationFields = append(schemaStruct.RelationFields, field)
					// docStruct.RelationFields = append(docStruct.RelationFields, field)

				}(schemaField)
			default:
				schemaField.HasRelation = false
			}

			// if schemaField.HasRelation {
			// 	v := reflect.New(schemaField.Relationship.RelationType)
			// 	schemaField.RelationshipStruct = GetSchemaStruct(v.Interface())
			// }

			schemaStruct.Fields = append(schemaStruct.Fields, schemaField)
		}
	}

	for _, f := range schemaStruct.Fields {
		schemaStruct.FieldsMap[f.Name] = f
	}

	schemaStructsMap.Store(schemaType, &schemaStruct)
	return &schemaStruct
}

// type SchemaStruct struct {
// 	Type            reflect.Type           // Type of reflect
// 	StructFields    []*SchemaField         // Schema 所有的字段
// 	StructFieldsMap map[string]SchemaField // Schema 字段的字典
// 	RelationFields  []*SchemaField         // 关联字段
// }

// type SchemaField struct {
// 	Name           string // 字段名称
// 	Index          int    // 下标
// 	InlineIndex    []int  // 内联的下标
// 	ColumnName     string // 列名
// 	CollectionName string // 集合名
// 	// IsNormal        bool                // 是否正常
// 	IsIgnored       bool                // 是否为隐藏字段
// 	HasDefaultValue bool                // 是否有默认值
// 	IsInline        bool                // 是否内联
// 	Tag             reflect.StructTag   // 标签
// 	TagMap          map[string]string   // 标签字典
// 	Struct          reflect.StructField // 字段结构
// 	IsForeignKey    bool                // 是否是外键
// 	Relationship    *Relationship       // 关系
// 	HasRelation     bool                // 是否拥有关联关系
// 	Zero            reflect.Value       // 字段零值
// }

// func getStructInfoOfSchema(s Schemer, connection Connection) *SchemaStruct {

// 	return GetSchemaStruct(s, connection)
// }

// func GetSchemaStruct(d interface{}, connection Connection) *SchemaStruct {
// 	var docStruct SchemaStruct
// 	if d == nil {
// 		return &docStruct
// 	}

// 	reflectValue := reflect.ValueOf(d)
// 	reflectType := reflectValue.Type()

// 	for {
// 		if reflectType.Kind() != reflect.Ptr {
// 			break
// 		}

// 		reflectType = reflectType.Elem()
// 	}

// 	if reflectType.Kind() == reflect.Slice {
// 		reflectType = reflectType.Elem()
// 	}

// 	// if reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
// 	// 	reflectType = reflectType.Elem()
// 	// }

// 	// Documenter must be a struct
// 	if reflectType.Kind() != reflect.Struct {
// 		return &docStruct
// 	}

// 	if v, found := docStructsMap.Load(reflectType); found && v != nil {
// 		return v.(*SchemaStruct)
// 	}

// 	docStruct.Type = reflectType

// 	for i := 0; i < reflectType.NumField(); i++ {

// 		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {

// 			field := &SchemaField{
// 				Struct:      fieldStruct,
// 				Name:        fieldStruct.Name,
// 				Tag:         fieldStruct.Tag,
// 				TagMap:      parseTagConfig(fieldStruct.Tag),
// 				Zero:        reflect.New(fieldStruct.Type).Elem(),
// 				Index:       i,
// 				InlineIndex: []int{i},
// 				IsInline:    false,
// 			}

// 			// hidden
// 			if _, found := field.TagMap["-"]; found {
// 				field.IsIgnored = true
// 			} else if v, foundInline := field.TagMap["INLINE"]; foundInline && v == "true" {
// 				// the field is inline
// 				inlineFieldStruct := GetSchemaStruct(reflect.New(fieldStruct.Type).Interface(), connection)

// 				for _, inlineField := range inlineFieldStruct.StructFields {
// 					inlineField.IsInline = true
// 					// inlineField.Index = []int{i, field.Index[0]}
// 					inlineField.InlineIndex = []int{i, inlineField.Index}
// 					docStruct.StructFields = append(docStruct.StructFields, inlineField)
// 					if inlineField.Relationship != nil {
// 						docStruct.RelationFields = append(docStruct.RelationFields, inlineField)
// 					}
// 				}
// 				continue
// 			} else {
// 				if _, ok := field.TagMap["DEFAULT"]; ok {
// 					field.HasDefaultValue = true
// 				}

// 				if name, ok := field.TagMap["COLUMN"]; ok {
// 					field.ColumnName = name
// 				}

// 				indirectType := fieldStruct.Type
// 				for indirectType.Kind() == reflect.Ptr {
// 					indirectType = indirectType.Elem()
// 				}

// 				fieldValue := reflect.New(indirectType).Interface()
// 				if _, isTime := fieldValue.(*time.Time); isTime {
// 					field.HasRelation = false
// 				} else if fieldStruct.Type.Kind() == reflect.Struct {
// 					field.HasRelation = false
// 				} else {

// 					switch fieldStruct.Type.Kind() {
// 					// OneToMany / ManyToMany
// 					case reflect.Slice:
// 						field.HasRelation = true
// 						f := func(field *SchemaField) {
// 							var (
// 								localFieldKey   string
// 								foreignFieldKey string
// 								elemType        = field.Struct.Type
// 							)

// 							for elemType.Kind() == reflect.Ptr || elemType.Kind() == reflect.Slice {
// 								elemType = elemType.Elem()
// 							}

// 							if foreignKey := field.TagMap["FOREIGNKEY"]; foreignKey != "" {
// 								localFieldKey = "_id"
// 								foreignFieldKey = foreignKey
// 							}

// 							if localField := field.TagMap["LOCALFIELD"]; localField != "" {
// 								localFieldKey = localField
// 							}

// 							if foreignField := field.TagMap["FOREIGNFIELD"]; foreignField != "" {
// 								foreignFieldKey = foreignField
// 							}

// 							if elemType.Kind() == reflect.Struct && isImplementsSchemer(elemType) {
// 								relationMdl := connection.M(elemType.Name())
// 								rs := &Relationship{
// 									ModelName:       elemType.Name(),
// 									RelationModel:   relationMdl,
// 									LocalFieldKey:   localFieldKey,
// 									ForeignFieldKey: foreignFieldKey,
// 								}

// 								rs.CollectionName = relationMdl.getCollectionName()
// 								if _, ok := field.TagMap["HASMANY"]; ok {
// 									rs.Kind = HasMany
// 								} else {
// 									// now just support has many, don't support many to many
// 									return
// 								}

// 								field.Relationship = rs
// 								docStruct.RelationFields = append(docStruct.RelationFields, field)
// 							}
// 						}
// 						defer f(field)
// 					case reflect.Struct:
// 						fallthrough
// 					case reflect.Ptr:
// 						field.HasRelation = true
// 						f := func(field *SchemaField) {
// 							var (
// 								localFieldKey   string
// 								foreignFieldKey string
// 								elemType        = field.Struct.Type
// 							)

// 							for elemType.Kind() == reflect.Ptr {
// 								elemType = elemType.Elem()
// 							}

// 							if !isImplementsSchemer(elemType) {
// 								return
// 							}

// 							relationMdl := connection.M(elemType.Name())
// 							rs := &Relationship{
// 								ModelName:     elemType.Name(),
// 								RelationModel: relationMdl,
// 							}

// 							rs.CollectionName = relationMdl.getCollectionName()

// 							if _, ok := field.TagMap["HASONE"]; ok {
// 								rs.Kind = HasOne

// 							} else if _, ok := field.TagMap["BELONGTO"]; ok {
// 								rs.Kind = BelongTo
// 							}

// 							if rs.Kind == "" {
// 								return
// 							}

// 							if foreignKey := field.TagMap["FOREIGNKEY"]; foreignKey != "" {
// 								if rs.Kind == HasOne {
// 									localFieldKey = "_id"
// 									foreignFieldKey = foreignKey
// 								} else {
// 									localFieldKey = foreignKey
// 									foreignFieldKey = "_id"
// 								}
// 							}

// 							if localField := field.TagMap["LOCALFIELD"]; localField != "" {
// 								localFieldKey = localField
// 							}

// 							if foreignField := field.TagMap["FOREIGNFIELD"]; foreignField != "" {
// 								foreignFieldKey = foreignField
// 							}
// 							rs.ForeignFieldKey = foreignFieldKey
// 							rs.LocalFieldKey = localFieldKey
// 							field.Relationship = rs
// 							docStruct.RelationFields = append(docStruct.RelationFields, field)
// 						}

// 						f(field)
// 					default:
// 						field.HasRelation = true
// 					}
// 				}
// 			}
// 			docStruct.StructFields = append(docStruct.StructFields, field)
// 		}
// 	}

// 	docStructsMap.Store(reflectType, &docStruct)
// 	return &docStruct
// }
