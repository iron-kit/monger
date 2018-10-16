package monger

import (
	"fmt"
	"reflect"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type SchemaNameGetter interface {
	GetSchemaName() string
}

type Schemer interface {
	Init(value interface{})
	beforeCreate(interface{}) error
	afterCreate() error

	beforeUpdate(interface{}) error
	afterUpdate() error
	IsUpdated() bool
	IsEmpty() bool
}

type Schema struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time     `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt time.Time     `json:"updated_at" bson:"updated_at,omitempty"`
	Deleted   bool          `json:"-" bson:"deleted"`
	isUpdated bool
	value     interface{}
}

func (s *Schema) Init(v interface{}) {
	s.value = v
}

func isImplementsSchemer(t reflect.Type) bool {

	ins := reflect.New(t).Interface()

	if _, ok := ins.(Schemer); ok {
		return true
	}
	return false
}

func getSchemaTypeName(schema interface{}) string {
	reflectType := reflect.TypeOf(schema)
	typeName := reflectType.Elem().Name()
	return snakeString(typeName)
}

func (s *Schema) beforeCreate(value interface{}) error {
	s.isUpdated = false
	now := time.Now()
	s.ID = bson.NewObjectId()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.Deleted = false
	s.value = value
	return nil
}

func (s *Schema) afterCreate() error {
	s.isUpdated = true
	return nil
}

func (s *Schema) beforeUpdate(value interface{}) error {
	s.isUpdated = false
	s.UpdatedAt = time.Now()
	return nil
}

func (s *Schema) afterUpdate() error {
	// panic("not implemented")
	s.isUpdated = true
	return nil
}

func (s *Schema) IsUpdated() bool {
	return s.isUpdated
}

func (s *Schema) IsEmpty() bool {
	if len(s.ID) == 0 {
		return true
	}

	return false
}

func (s *Schema) GetBSON() (interface{}, error) {
	mapData := make(map[string]interface{})

	if s.value == nil {
		return mapData, &NotInitDocumentError{NewError("You neet init this document before to bson")}
	}
	// fmt.Println(doc.value)
	docStruct := GetSchemaStruct(s.value)
	docv := reflect.ValueOf(s.value)
	if docv.Type().Kind() == reflect.Ptr {
		docv = docv.Elem()
	}

	// fmt.Println("doc")
	// fmt.Println(docv)

	for _, field := range docStruct.Fields {
		if field.Relationship != nil {
			// 忽略关联关系字段
			continue
		}

		var val reflect.Value
		if field.IsInline {
			val = docv.FieldByIndex(field.InlineIndex)
			// mapData[field.ColumnName] = val.Interface()
		} else {
			val = docv.Field(field.Index)
			// mapData[field.ColumnName] = val.Interface()
		}
		// valt := val.Type()

		if field.TagMap["OMITEMPTY"] == "true" {
			if isZero(val) {
				continue
			}
		}

		mapData[field.ColumnName] = val.Interface()

	}
	fmt.Println("monger GetBSON:", mapData)
	// log.Println("monger GetBSON:", mapData)
	return mapData, nil
}
