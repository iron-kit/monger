package monger

import (
	"gopkg.in/mgo.v2"
	"reflect"
)

/*
Query is the mongodb query help tools,

For Example:
	users := []*models.User{}
	UserModel.Find(bson.M{"id": "12131312"}).Populate("Messages").Skip(0).Limit(12).Exec(&users)

*/
type Query interface {
	Select(selector interface{}) Query
	Sort(fields ...string) Query
	Limit(limit int) Query
	Skip(skip int) Query
	Count() (n int, err error)
	Populate(fields ...string) Query
	Exec(result interface{}) error
	One(document Document) error
	All(documents []Document) error
}

type query struct {
	collection *mgo.Collection
	connection Connection
	query      interface{}
	selector   interface{}
	populate   []string
	sort       []string
	limit      int
	skip       int
	multiple   bool
}

func newQuery(collection *mgo.Collection, connection Connection, q interface{}, multiple bool) Query {
	return &query{
		collection: collection,
		connection: connection,
		query:      q,
		multiple:   multiple,
	}
}

func (q *query) Select(selector interface{}) Query {
	q.selector = selector

	return q
}

func (q *query) Sort(fields ...string) Query {
	q.sort = append(q.sort, fields...)
	return q
}

func (q *query) Limit(limit int) Query {
	q.limit = limit
	return q
}

func (q *query) Skip(skip int) Query {
	q.skip = skip

	return q
}

func (q *query) Count() (n int, err error) {
	return q.collection.Find(q.query).Count()
}

func (q *query) Populate(fields ...string) Query {
	panic("not implemented")
}

func (q *query) buildQuery(query *mgo.Query) {
	if q.selector != nil {
		query.Select(q.selector)
	}

	if q.skip > 0 {
		query.Skip(q.skip)
	}

	if q.limit > 0 {
		query.Limit(q.limit)
	}

	if len(q.sort) > 0 {
		query.Sort(q.sort...)
	}

}

func (q *query) One(document Document) error {
	return q.Exec(document)
}

func (q *query) All(documents []Document) error {
	return q.Exec(documents)
}

// Exec 处理查询
func (q *query) Exec(result interface{}) error {
	if result == nil {
		panic("[Monger] The result is required")
	}

	resType := reflect.TypeOf(result)

	isListResult := func() bool {
		if resType.Kind() == reflect.Ptr && resType.Elem().Kind() == reflect.Slice {
			// 向结果集中注入连接参数
			if documents, ok := result.([]Document); ok {
				for _, doc := range documents {
					doc.SetInstance(doc)
					doc.SetCollection(q.collection)
					doc.SetConnection(q.connection)
				}
			} else {
				panic("[Monger] Every one of resultset must be Document")
			}
			return true
		}
		// 向结果集中注入连接参数
		if doc, ok := result.(Document); ok {
			doc.SetInstance(doc)
			doc.SetCollection(q.collection)
			doc.SetConnection(q.connection)
		} else {
			panic("[Monger] The resultset must be a Document")
		}
		return false
	}()

	query := q.collection.Find(q.query)
	q.buildQuery(query)

	var err error

	if isListResult {
		err = query.All(result)
	} else {
		err = query.One(result)
	}

	if err == mgo.ErrNotFound {
		return &NotFoundError{NewError("Not found the document")}
	} else if err != nil {
		return err
	}

	return nil
}
