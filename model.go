package monger

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
Model is mongodb actoin model
*/
type Model interface {
	UpsertID(id bson.ObjectId, data interface{}) (*mgo.ChangeInfo, error)
	Upsert(condition bson.M, data interface{}) (*mgo.ChangeInfo, error)
	Update(condition bson.M, data interface{}) error
	Count(condition ...bson.M) int
	Create(doc interface{}) error
	FindOne(doc interface{}, where ...bson.M) error
	FindAll(doc interface{}, where ...bson.M) error
	FindByID(id bson.ObjectId, doc interface{}) error
	Where(...bson.M) Query
	Select(...bson.M) Query
	Aggregate([]bson.M) *mgo.Pipe
	getCollectionName() string
	getSchemaStruct() *SchemaStruct
}

type model struct {
	schema         Schemer
	schemaStruct   *SchemaStruct
	collection     *mgo.Collection
	connection     Connection
	collectionName string
}

func (m *model) query() Query {
	return newQuery(m.collection, m.getSchemaStruct())
}

func (m *model) getSchemaStruct() *SchemaStruct {
	if m.schemaStruct == nil {
		m.schemaStruct = getStructInfoOfSchema(m.schema, m.connection)
	}
	return m.schemaStruct
}

func (m *model) UpsertID(id bson.ObjectId, data interface{}) (*mgo.ChangeInfo, error) {
	return m.query().UpsertID(id, data)
}

func (m *model) Upsert(condition bson.M, data interface{}) (*mgo.ChangeInfo, error) {
	return m.query().Upsert(condition, data)
}

func (m *model) Update(condition bson.M, data interface{}) error {

	return m.query().Update(condition, data)
}

func (m *model) Count(condition ...bson.M) int {
	q := m.query()
	if len(condition) > 0 {
		q.Where(condition[0])
	}
	return q.Count()
}

func (m *model) Create(doc interface{}) error {
	// panic("not implemented")
	return m.query().Create(doc)
}

func (m *model) FindOne(doc interface{}, where ...bson.M) error {
	q := m.query()
	if len(where) > 0 {
		q.Where(where[0])
	}
	return q.FindOne(doc)
}

func (m *model) FindAll(doc interface{}, where ...bson.M) error {
	q := m.query()

	if len(where) > 0 {
		q.Where(where[0])
	}

	return q.FindAll(doc)
}

func (m *model) FindByID(id bson.ObjectId, doc interface{}) error {
	return m.query().Where(bson.M{"_id": id}).FindOne(doc)
}

func (m *model) Where(where ...bson.M) Query {

	if len(where) > 0 {
		return m.query().Where(where[0])
	}
	return m.query().Where(nil)
}

func (m *model) Select(selector ...bson.M) Query {
	if len(selector) > 0 {
		return m.query().Select(selector[0])
	}

	return m.query().Select(nil)
}

func (m *model) Aggregate(pipe []bson.M) *mgo.Pipe {
	return m.query().Aggregate(pipe)
}

func (m *model) getCollectionName() string {
	return m.collectionName
}

func newModel(connection *connection, schema Schemer) Model {
	collectionName := ""
	if nameGetter, ok := schema.(SchemaNameGetter); ok {
		collectionName = nameGetter.GetSchemaName()
	} else {
		collectionName = getSchemaTypeName(schema)
	}

	collection := connection.Session.DB("").C(collectionName)
	return &model{
		schema: schema,
		// schemaStruct:   getStructInfoOfSchema(schema, connection),
		connection:     connection,
		collection:     collection,
		collectionName: collectionName,
	}
}
