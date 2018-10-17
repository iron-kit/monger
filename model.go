package monger

import (
	"fmt"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
Model is mongodb actoin model
*/
type Model interface {
	OffSoftDeletes() Query
	UpsertID(id interface{}, data interface{}) (*mgo.ChangeInfo, error)
	Upsert(condition bson.M, data interface{}) (*mgo.ChangeInfo, error)
	Update(condition bson.M, data interface{}) error
	Count(condition ...bson.M) int
	Create(doc interface{}) error
	FindOne(doc interface{}, where ...bson.M) error
	FindAll(doc interface{}, where ...bson.M) error
	FindByID(id bson.ObjectId, doc interface{}) error
	Where(...bson.M) Query
	Select(...bson.M) Query
	Aggregate([]bson.M) Query
	Delete(bson.M) error
	ForceDelete(bson.M) error
	DeleteAll(bson.M) error
	ForceDeleteAll(bson.M) error
	Restore(bson.M) error
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

func (m *model) Restore(condition bson.M) error {
	return m.query().Where(condition).Restore()
}

func (m *model) Delete(condition bson.M) error {
	return m.query().Where(condition).Delete()
}

func (m *model) ForceDelete(condition bson.M) error {
	return m.query().Where(condition).ForceDelete()
}

func (m *model) DeleteAll(condition bson.M) error {
	_, err := m.query().Where(condition).DeleteAll()

	return err
}

func (m *model) ForceDeleteAll(condition bson.M) error {
	_, err := m.query().Where(condition).ForceDeleteAll()

	return err
}

func (m *model) query() Query {
	return newQuery(m.collection, m.getSchemaStruct())
}

func (m *model) getSchemaStruct() *SchemaStruct {
	if m.schemaStruct == nil {
		// m.schemaStruct = getStructInfoOfSchema(m.schema, m.connection)
		m.schemaStruct = GetSchemaStruct(m.schema)
	}

	return m.schemaStruct
}

func (m *model) OffSoftDeletes() Query {
	return m.query().OffSoftDeletes()
}

func (m *model) UpsertID(id interface{}, data interface{}) (*mgo.ChangeInfo, error) {
	return m.query().UpsertID(id, data)
}

func (m *model) Upsert(condition bson.M, data interface{}) (*mgo.ChangeInfo, error) {
	return m.query().Upsert(condition, data)
}

func (m *model) Update(condition bson.M, data interface{}) error {
	fmt.Println(data)
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

func (m *model) Aggregate(pipe []bson.M) Query {
	return m.query().Aggregate(pipe)
}

func (m *model) getCollectionName() string {
	return m.collectionName
}

func newModel(connection *connection, schema Schemer) Model {
	collectionName := getCollectionName(schema)

	collection := connection.Session.DB("").C(collectionName)
	return &model{
		schema: schema,
		// schemaStruct:   getStructInfoOfSchema(schema, connection),
		connection:     connection,
		collection:     collection,
		collectionName: collectionName,
	}
}

func getCollectionName(schema interface{}) string {
	collectionName := ""
	if nameGetter, ok := schema.(SchemaNameGetter); ok {
		collectionName = nameGetter.GetSchemaName()
	} else {
		collectionName = getSchemaTypeName(schema)
	}

	return collectionName
}
