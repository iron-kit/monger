package monger

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Document interface {
	GetID() bson.ObjectId
	SetID(bson.ObjectId)

	SetCreatedAt(time.Time)
	GetCreatedAt() time.Time

	SetUpdatedAt(time.Time)
	GetUpdatedAt() time.Time

	SetDeleted(bool)
	IsDeleted() bool

	GetConnection() Connection
	GetCollection() *mgo.Collection
	SetCollection(*mgo.Collection)
	SetConnection(Connection)
	CollectionName() string
}

type BaseDocument struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt" bson:"updatedAt"`
	Deleted   bool          `json:"-" bson:"deleted"`

	collection *mgo.Collection
	connection Connection
}

func (doc *BaseDocument) CollectionName() string {
	return ""
}

func (doc *BaseDocument) GetID() bson.ObjectId {
	return doc.ID
}

func (doc *BaseDocument) SetID(id bson.ObjectId) {
	doc.ID = id
}

func (doc *BaseDocument) SetCreatedAt(time time.Time) {
	doc.CreatedAt = time
}

func (doc *BaseDocument) GetCreatedAt() time.Time {
	return doc.CreatedAt
}

func (doc *BaseDocument) SetUpdatedAt(time time.Time) {
	doc.UpdatedAt = time
}

func (doc *BaseDocument) GetUpdatedAt() time.Time {
	return doc.UpdatedAt
}

func (doc *BaseDocument) SetDeleted(isDeleted bool) {
	doc.Deleted = isDeleted
}

func (doc *BaseDocument) IsDeleted() bool {
	return doc.Deleted
}

func (doc *BaseDocument) SetCollection(collection *mgo.Collection) {
	doc.collection = collection
}

func (doc *BaseDocument) SetConnection(c Connection) {
	doc.connection = c
}

func (doc *BaseDocument) GetCollection() *mgo.Collection {
	return doc.collection
}

func (doc *BaseDocument) GetConnection() Connection {
	return doc.connection
}
