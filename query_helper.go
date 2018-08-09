package monger

import (
	//"fmt"
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
	All(documents interface{}) error
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
	q.multiple = false
	return q.Exec(document)
}

func (q *query) All(documents interface{}) error {
	q.multiple = true
	return q.Exec(documents)
}

// Exec 处理查询
func (q *query) Exec(result interface{}) error {
	if result == nil {
		panic("[Monger] The result is required")
	}

	multiple := false

	resType := reflect.TypeOf(result)

	if resType.Kind() == reflect.Ptr && resType.Elem().Kind() == reflect.Slice {
		multiple = true
		// return true
	} else if d, ok := result.(Document); ok {
		multiple = false
		d.SetCollection(q.collection)
		d.SetConnection(q.connection)
	} else {
		panic("[Monger] The resultset must be a Document")
	}

	// if resType.Kind() != reflect.Ptr {
	// 	// 参数不是指针 且不是切片视为异常
	// 	if resType.Elem().Kind() != reflect.Slice {
	// 		panic("[Monger] The resultset must be a pointer")
	// 	}

	// 	multiple = true
	// 	// panic("[Monger] The resultset must be a pointer")
	// } else {
	// 	if resType.Elem().Kind() == reflect.Slice {
	// 		multiple = true
	// 	} else {
	// 		multiple = false
	// 	}
	// }

	// isListResult := func() bool {
	// 	if resType.Kind() == reflect.Ptr && resType.Elem().Kind() == reflect.Slice {

	// 		return true
	// 	}

	// 	// 向结果集中注入连接参数
	// 	if doc, ok := result.(Document); ok {
	// 		doc.SetCollection(q.collection)
	// 		doc.SetConnection(q.connection)
	// 	} else {
	// 		panic("[Monger] The resultset must be a Document")
	// 	}

	// 	return false
	// }()

	query := q.collection.Find(q.query)
	q.buildQuery(query)

	var err error

	if multiple {
		err = query.All(result)
		resultv := reflect.ValueOf(result)
		slicev := resultv.Elem()

		for i := 0; i < slicev.Len(); i++ {
			ele := slicev.Index(i)
			if doc, ok := ele.Interface().(Document); ok {
				// fmt.Println(ele, "是 Document")
				doc.SetCollection(q.collection)
				doc.SetConnection(q.connection)

			} else if ele.Type().Kind() == reflect.Struct {
				if doc, ok := ele.Addr().Interface().(Document); ok {
					doc.SetCollection(q.collection)
					doc.SetConnection(q.connection)
				} else {
					panic("[Monger] The resultset must be a Document Slice")
				}
			} else {
				panic("[Monger] The resultset must be a Document Slice")
			}
		}
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
