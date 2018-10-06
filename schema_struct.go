package monger

import (
	"go/ast"
	"reflect"
	"sync"
	"time"
)

type SchemaStruct struct {
	Type            reflect.Type           // Type of reflect
	StructFields    []*SchemaField         // Schema 所有的字段
	StructFieldsMap map[string]SchemaField // Schema 字段的字典
	RelationFields  []*SchemaField         // 关联字段
}

type SchemaField struct {
	Name           string // 字段名称
	Index          int    // 下标
	InlineIndex    []int  // 内联的下标
	ColumnName     string // 列名
	CollectionName string // 集合名
	// IsNormal        bool                // 是否正常
	IsIgnored       bool                // 是否为隐藏字段
	HasDefaultValue bool                // 是否有默认值
	IsInline        bool                // 是否内联
	Tag             reflect.StructTag   // 标签
	TagMap          map[string]string   // 标签字典
	Struct          reflect.StructField // 字段结构
	IsForeignKey    bool                // 是否是外键
	Relationship    *Relationship       // 关系
	HasRelation     bool                // 是否拥有关联关系
	Zero            reflect.Value       // 字段零值
}

var docStructsMap sync.Map

func getStructInfoOfSchema(s Schemer, connection Connection) *SchemaStruct {

	return GetSchemaStruct(s, connection)
}

func GetSchemaStruct(d interface{}, connection Connection) *SchemaStruct {
	var docStruct SchemaStruct
	if d == nil {
		return &docStruct
	}

	reflectValue := reflect.ValueOf(d)
	reflectType := reflectValue.Type()

	for {
		if reflectType.Kind() != reflect.Ptr {
			break
		}

		reflectType = reflectType.Elem()
	}

	if reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}

	// if reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
	// 	reflectType = reflectType.Elem()
	// }

	// Documenter must be a struct
	if reflectType.Kind() != reflect.Struct {
		return &docStruct
	}

	if v, found := docStructsMap.Load(reflectType); found && v != nil {
		return v.(*SchemaStruct)
	}

	docStruct.Type = reflectType

	for i := 0; i < reflectType.NumField(); i++ {

		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {

			field := &SchemaField{
				Struct:      fieldStruct,
				Name:        fieldStruct.Name,
				Tag:         fieldStruct.Tag,
				TagMap:      parseTagConfig(fieldStruct.Tag),
				Zero:        reflect.New(fieldStruct.Type).Elem(),
				Index:       i,
				InlineIndex: []int{i},
				IsInline:    false,
			}

			// hidden
			if _, found := field.TagMap["-"]; found {
				field.IsIgnored = true
			} else if v, foundInline := field.TagMap["INLINE"]; foundInline && v == "true" {
				// the field is inline
				inlineFieldStruct := GetSchemaStruct(reflect.New(fieldStruct.Type).Interface(), connection)

				for _, inlineField := range inlineFieldStruct.StructFields {
					inlineField.IsInline = true
					// inlineField.Index = []int{i, field.Index[0]}
					inlineField.InlineIndex = []int{i, inlineField.Index}
					docStruct.StructFields = append(docStruct.StructFields, inlineField)
					if inlineField.Relationship != nil {
						docStruct.RelationFields = append(docStruct.RelationFields, inlineField)
					}
				}
				continue
			} else {
				if _, ok := field.TagMap["DEFAULT"]; ok {
					field.HasDefaultValue = true
				}

				if name, ok := field.TagMap["COLUMN"]; ok {
					field.ColumnName = name
				}

				indirectType := fieldStruct.Type
				for indirectType.Kind() == reflect.Ptr {
					indirectType = indirectType.Elem()
				}

				fieldValue := reflect.New(indirectType).Interface()
				if _, isTime := fieldValue.(*time.Time); isTime {
					field.HasRelation = false
				} else if fieldStruct.Type.Kind() == reflect.Struct {
					field.HasRelation = false
				} else {

					switch fieldStruct.Type.Kind() {
					// OneToMany / ManyToMany
					case reflect.Slice:
						field.HasRelation = true
						f := func(field *SchemaField) {
							var (
								localFieldKey   string
								foreignFieldKey string
								elemType        = field.Struct.Type
							)

							for elemType.Kind() == reflect.Ptr || elemType.Kind() == reflect.Slice {
								elemType = elemType.Elem()
							}

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
								relationMdl := connection.M(elemType.Name())
								rs := &Relationship{
									ModelName:       elemType.Name(),
									RelationModel:   relationMdl,
									LocalFieldKey:   localFieldKey,
									ForeignFieldKey: foreignFieldKey,
								}

								rs.CollectionName = relationMdl.getCollectionName()
								if _, ok := field.TagMap["HASMANY"]; ok {
									rs.Kind = HasMany
								} else {
									// now just support has many, don't support many to many
									return
								}

								field.Relationship = rs
								docStruct.RelationFields = append(docStruct.RelationFields, field)
							}
						}
						defer f(field)
					case reflect.Struct:
						fallthrough
					case reflect.Ptr:
						field.HasRelation = true
						f := func(field *SchemaField) {
							var (
								localFieldKey   string
								foreignFieldKey string
								elemType        = field.Struct.Type
							)

							for elemType.Kind() == reflect.Ptr {
								elemType = elemType.Elem()
							}

							if !isImplementsSchemer(elemType) {
								return
							}

							relationMdl := connection.M(elemType.Name())
							rs := &Relationship{
								ModelName:     elemType.Name(),
								RelationModel: relationMdl,
							}

							rs.CollectionName = relationMdl.getCollectionName()

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
							docStruct.RelationFields = append(docStruct.RelationFields, field)
						}

						f(field)
					default:
						field.HasRelation = true
					}
				}
			}
			docStruct.StructFields = append(docStruct.StructFields, field)
		}
	}

	docStructsMap.Store(reflectType, &docStruct)
	return &docStruct
}
