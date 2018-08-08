package monger

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Model interface {
	Create(document Document) Document
	Find(query ...interface{}) Query
	FindByID(id bson.ObjectId) Query
	FindOne(query ...bson.M) Query
}

type model struct {
	collection *mgo.Collection
	connection Connection
}

func (m *model) Create(document Document) Document {
	document.SetInstance(document)
	document.SetCollection(m.collection)
	document.SetConnection(m.connection)

	return document
}

func (m *model) Find(query ...interface{}) Query {
	var restQuery interface{}
	multiple := true
	if len(query) == 0 {
		restQuery = bson.M{}
	} else if len(query) == 1 {
		restQuery = query[0]
	} else if len(query) == 2 {
		restQuery = query[0]
		if isMultiple, ok := query[1].(bool); ok {
			multiple = isMultiple
		} else {
			panic("[monger] The third params must be bool")
		}
	} else {
		panic("[monger] Too many query params")
	}
	return newQuery(m.collection, m.connection, restQuery, multiple)
}

func (m *model) FindByID(id bson.ObjectId) Query {
	return newQuery(
		m.collection,
		m.connection,
		bson.M{"_id": "id"},
		false,
	)
}

func (m *model) FindOne(query ...bson.M) Query {
	queryLen := len(query)
	if queryLen == 0 {
		return m.Find(bson.M{}, false)
	} else if queryLen == 1 {
		return m.Find(query[0], false)
	} else {
		panic("[monger] Too many query params")
	}
}
