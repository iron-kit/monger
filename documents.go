package monger

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strings"
	// "reflect"
	// "strings"
	"time"
)

type defaultCreateHooker interface {
	defaultBeforeCreate() error
	defaultAfterCreate(doc Documenter) error
}

type defaultUpdateHooker interface {
	defaultBeforeUpdate() error
	defaultAfterUpdate(doc Documenter) error
}

type CollectionNameGetter interface {
	CollectionName() string
}

type DocumentHooker interface {
	BeforeSave()
}

type Documenter interface {
	// DocumentManager
	// DocumentHooker
	// bson.Getter
	// bson.Setter
	GetID() bson.ObjectId
	SetID(bson.ObjectId)

	SetCreatedAt(time.Time)
	GetCreatedAt() time.Time

	SetUpdatedAt(time.Time)
	GetUpdatedAt() time.Time

	SetDeleted(bool)
	IsDeleted() bool

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

func (doc *Document) defaultBeforeCreate() error {
	now := time.Now()
	doc.SetID(bson.NewObjectId())
	doc.SetUpdatedAt(now)
	doc.SetCreatedAt(now)
	return nil
}

func (doc *Document) defaultAfterCreate(document Documenter) error {
	doc.value = document
	return nil
}

func (doc *Document) defaultBeforeUpdate() error {
	return nil
}

func (doc *Document) defaultAfterUpdate(document Documenter) error {
	doc.value = document
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
