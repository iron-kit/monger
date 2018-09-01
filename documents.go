package monger

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"reflect"
	"strings"
	"time"
)

type defaultCreateHooker interface {
	defaultBeforeCreate(document Documenter, mdl *model) error
	defaultAfterCreate(doc Documenter, mdl *model) error
}

type defaultUpdateHooker interface {
	defaultBeforeUpdate(document Documenter, mdl *model) error
	defaultAfterUpdate(doc Documenter, mdl *model) error
}

type CollectionNameGetter interface {
	CollectionName() string
}

type DocumentHooker interface {
	BeforeSave()
}

type Documenter interface {
	GetID() bson.ObjectId
	SetID(bson.ObjectId)

	SetCreatedAt(time.Time)
	GetCreatedAt() time.Time

	SetUpdatedAt(time.Time)
	GetUpdatedAt() time.Time

	SetDeleted(bool)
	IsDeleted() bool

	GetStringID() string

	setValue(value Documenter)
	// defaultBeforeCreate() error
	// defaultBeforeUpdate() error
	defaultCreateHooker
	defaultUpdateHooker
}

func isImplementsDocumenter(t reflect.Type) bool {

	ins := reflect.New(t).Interface()

	if _, ok := ins.(Documenter); ok {
		return true
	}
	return false
}

type documentManager struct {
	document   Documenter
	connection Connection
	collection *mgo.Collection
	// documentMap map[string]interface{}
}

type Document struct {
	value     Documenter
	mdl       *model
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time     `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt time.Time     `json:"updatedAt" bson:"updatedAt,omitempty"`
	Deleted   bool          `json:"-" bson:"deleted"`
	// Upsert    bool          `json:"-" bson:"-"`
}

func getDocumentTypeName(doc Documenter) string {
	reflectType := reflect.TypeOf(doc)
	typeName := strings.ToLower(reflectType.Elem().Name())
	return typeName
}

func (doc *Document) defaultBeforeCreate(document Documenter, mdl *model) error {
	now := time.Now()
	doc.SetID(bson.NewObjectId())
	doc.SetUpdatedAt(now)
	doc.SetCreatedAt(now)
	if doc.value == nil {
		doc.value = document
	}

	if doc.mdl == nil {
		doc.mdl = mdl
	}

	// docStruct := mdl.GetDocumentStruct()
	// for _, f := range docStruct.StructFields {
	// 	fmt.Println(f.Name)
	// }
	// fmt.Println(structs.Map(doc.value))
	// docv := reflect.ValueOf(doc.value)
	// // for {

	// // }
	// if docv.Type().Kind() == reflect.Ptr {
	// 	docv = docv.Elem()
	// }

	// for _, relationField := range docStruct.RelationFields {
	// 	// docv.relationField.Name
	// 	f := docv.FieldByName(relationField.Name)
	// 	if f.CanSet() {
	// 		f.Set(reflect.ValueOf(nil))
	// 		// f.SetPointer(nil)
	// 	}
	// }
	// fmt.Println(doc.value)
	return nil
}

func (doc *Document) defaultAfterCreate(document Documenter, mdl *model) error {
	if doc.value == nil {
		doc.value = document
	}
	if doc.mdl == nil {
		doc.mdl = mdl
	}
	return nil
}

func (doc *Document) defaultBeforeUpdate(document Documenter, mdl *model) error {
	doc.SetUpdatedAt(time.Now())
	if doc.value == nil {
		doc.value = document
	}
	if doc.mdl == nil {
		doc.mdl = mdl
	}
	return nil
}

func (doc *Document) defaultAfterUpdate(document Documenter, mdl *model) error {
	if doc.value == nil {
		doc.value = document
	}
	if doc.mdl == nil {
		doc.mdl = mdl
	}
	return nil
}

func (doc *Document) GetID() bson.ObjectId {
	return doc.ID
}

func (doc *Document) SetID(id bson.ObjectId) {
	doc.ID = id
}

func (doc *Document) SetCreatedAt(time time.Time) {
	doc.CreatedAt = time
}

func (doc *Document) GetCreatedAt() time.Time {
	return doc.CreatedAt
}

func (doc *Document) SetUpdatedAt(time time.Time) {
	doc.UpdatedAt = time
}

func (doc *Document) GetUpdatedAt() time.Time {
	return doc.UpdatedAt
}

func (doc *Document) SetDeleted(isDeleted bool) {
	doc.Deleted = isDeleted
}

func (doc *Document) IsDeleted() bool {
	return doc.Deleted
}

func (doc *Document) BeforeSave() {
	return
}

func (doc *Document) setValue(val Documenter) {
	doc.value = val
}

func (doc *Document) GetValue() Documenter {
	return doc.value
}

func (doc *Document) GetStringID() string {

	return doc.ID.Hex()
}

func (doc *Document) GetBSON() (interface{}, error) {
	mapData := make(map[string]interface{})

	if doc.mdl == nil || doc.value == nil {
		return mapData, &NotInitDocumentError{NewError("You neet init this document before to bson")}
	}
	// fmt.Println(doc.value)
	docStruct := doc.mdl.GetDocumentStruct()
	docv := reflect.ValueOf(doc.value)
	if docv.Type().Kind() == reflect.Ptr {
		docv = docv.Elem()
	}

	// fmt.Println("doc")
	// fmt.Println(docv)

	for _, field := range docStruct.StructFields {
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
	log.Println("monger GetBSON:", mapData)
	return mapData, nil
}
