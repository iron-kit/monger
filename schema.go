package monger

import (
	"reflect"
	"time"

	"gopkg.in/mgo.v2/bson"
)

type SchemaNameGetter interface {
	GetSchemaName() string
}

type Schemer interface {
	beforeCreate() error
	afterCreate() error

	beforeUpdate() error
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

func (s *Schema) beforeCreate() error {
	s.isUpdated = false
	now := time.Now()
	s.ID = bson.NewObjectId()
	s.CreatedAt = now
	s.UpdatedAt = now
	s.Deleted = false
	return nil
}

func (s *Schema) afterCreate() error {
	s.isUpdated = true
	return nil
}

func (s *Schema) beforeUpdate() error {
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
