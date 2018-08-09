package monger

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type Doc interface {
	GetDocument() Document
	GetCollection() *mgo.Collection
	GetConnection() Connection
	Save() error
	// Update(interface{}) (error, map[string]interface{})
	Validate(...interface{}) (bool, []error)
	DefaultValidate() (bool, []error)
}

type doc struct {
	document   Document
	collection *mgo.Collection
	connection Connection
}

func D(args ...interface{}) Doc {
	if len(args) > 3 {
		panic("[monger] Too many params")
	}

	doc := &doc{}

	for _, v := range args {
		if d, ok := v.(Document); ok {
			doc.document = d
			doc.collection = d.GetCollection()
			doc.connection = d.GetConnection()
			continue
		}

		if coll, ok := v.(*mgo.Collection); ok {
			doc.collection = coll
			continue
		}

		if conn, ok := v.(Connection); ok {
			doc.connection = conn
			continue
		}
	}

	if doc.document == nil {
		panic("[monger] You must be set document")
	}

	return doc
}

func (d *doc) GetDocument() Document {
	return d.document
}

func (d *doc) GetCollection() *mgo.Collection {
	return d.collection
}

func (d *doc) GetConnection() Connection {
	return d.connection
}

func (d *doc) Save() error {
	if d.document == nil || d.collection == nil || d.connection == nil {
		panic("[monger] Please use D(document Document) function to build a saveable Document")
	}
	now := time.Now()
	config := d.connection.GetConfig()

	// TODO Implemented validate document
	session := d.connection.CloneSession()
	defer session.Close()

	collection := session.DB(config.DBName).C(d.collection.Name)

	var err error

	// TODO 处理关联关系

	// 检测 ID 是否已经设置，如果未设置判定为插入, 否则判定为更新。
	if len(d.document.GetID()) == 0 {
		d.document.SetUpdatedAt(now)
		d.document.SetCreatedAt(now)

		d.document.SetID(bson.NewObjectId())

		err = collection.Insert(d.document)

		if err != nil {
			if mgo.IsDup(err) {
				err = &DuplicateDocumentError{NewError("Duplicate Key")}
			}
		}
	} else {
		d.document.SetUpdatedAt(now)
		_, erro := collection.UpsertId(d.document.GetID(), d.document)

		if erro != nil {
			if mgo.IsDup(erro) {
				err = &DuplicateDocumentError{NewError("Duplicate Key")}
			} else {
				err = erro
			}
		}
	}

	return nil
}

func (d *doc) Validate(...interface{}) (bool, []error) {
	return false, nil
}

func (d *doc) DefaultValidate() (bool, []error) {
	return false, nil
}
