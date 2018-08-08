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

	Get(field string) interface{}
	Save() error
	Update(interface{}) (error, map[string]interface{})
	Validate(...interface{}) (bool, []error)
	DefaultValidate() (bool, []error)

	SetInstance(Document)
	SetCollection(*mgo.Collection)
	SetConnection(Connection)
	// Instance() Document
	// Collection() *mgo.Collation
	// Connection() Connection
}

type BaseDocument struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time     `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt" bson:"updatedAt"`
	Deleted   bool          `json:"-" bson:"deleted"`

	instance   Document
	connection Connection
	collection *mgo.Collection
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

func (self *BaseDocument) Get(field string) interface{} {
	panic("not implemented")
}

func (doc *BaseDocument) Save() error {
	if doc.instance == nil || doc.collection == nil || doc.connection == nil {
		panic("[monger] Please use Model.Create(document Document) function to build a saveable Document")
	}
	now := time.Now()
	config := doc.connection.GetConfig()

	// TODO Implemented validate document
	session := doc.connection.CloneSession()
	defer session.Close()

	collection := session.DB(config.DBName).C(doc.collection.Name)

	var err error

	// TODO 处理关联关系

	// 检测 ID 是否已经设置，如果未设置判定为插入, 否则判定为更新。
	if len(doc.ID) == 0 {
		doc.SetUpdatedAt(now)
		doc.SetCreatedAt(now)

		doc.SetID(bson.NewObjectId())

		err = collection.Insert(doc.instance)

		if err != nil {
			if mgo.IsDup(err) {
				err = &DuplicateDocumentError{NewError("Duplicate Key")}
			}
		}
	} else {
		doc.SetUpdatedAt(now)
		_, erro := collection.UpsertId(doc.ID, doc.instance)

		if erro != nil {
			if mgo.IsDup(erro) {
				err = &DuplicateDocumentError{NewError("Duplicate Key")}
			} else {
				err = erro
			}
		}
	}

	return err
}

func (doc *BaseDocument) Update(interface{}) (error, map[string]interface{}) {
	panic("not implemented")
}

func (doc *BaseDocument) Validate(...interface{}) (bool, []error) {

	// TODO Implemented BaseDocument.Validate function
	panic("not implemented")
}

func (doc *BaseDocument) DefaultValidate() (bool, []error) {
	// TODO Implemented BaseDocument.DefaultValidate function
	panic("not implemented")
}

func (doc *BaseDocument) SetInstance(d Document) {
	doc.instance = d
}

func (doc *BaseDocument) SetCollection(collection *mgo.Collection) {
	doc.collection = collection
}

func (doc *BaseDocument) SetConnection(c Connection) {
	doc.connection = c
}
