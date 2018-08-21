package monger

import (
	"gopkg.in/mgo.v2/bson"
	"strings"
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
	Create(document interface{}) error
	// Delete() error
	Remove() error
	RemoveAll() (info *mgo.ChangeInfo, err error)
	// One(document Documenter) error
	// All(documents interface{}) error
}

type query struct {
	mdl        *model
	collection *mgo.Collection
	connection Connection
	query      interface{}
	selector   interface{}
	populate   []string
	sort       []string
	limit      int
	skip       int
	multiple   bool
	pipeline   []bson.M
}

func newQuery(collection *mgo.Collection, connection Connection, q interface{}, multiple bool, mdl *model) Query {
	return &query{
		collection: collection,
		connection: connection,
		query:      q,
		multiple:   multiple,
		mdl:        mdl,
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

func (q *query) Remove() error {
	return q.collection.Remove(q.query)
}

func (q *query) RemoveAll() (info *mgo.ChangeInfo, err error) {
	return q.collection.RemoveAll(q.query)
}

func (q *query) Populate(fields ...string) Query {
	// panic("not implemented")
	populates := []string{}

	for _, v := range fields {
		splitedV := strings.Split(v, ",")
		if len(splitedV) >= 2 {
			for _, v := range splitedV {
				populates = append(populates, v)
			}
		}

		populates = append(populates, v)
	}

	q.populate = populates

	return q
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

func (q *query) one(document interface{}) error {
	q.multiple = false
	// return q.Exec(document)
	if len(q.populate) > 0 {
		// use populate mode
		return q.execPopulateQuery(document)
	}
	query := q.collection.Find(q.query)
	q.buildQuery(query)
	return query.One(document)
}

func (q *query) all(documents interface{}) error {
	q.multiple = true
	if len(q.populate) > 0 {
		// use populate mode
		return q.execPopulateQuery(documents)
	}
	query := q.collection.Find(q.query)
	q.buildQuery(query)
	return query.All(documents)
}

func (q *query) getPopulatePipeline() []bson.M {
	documentStruct := q.mdl.GetDocumentStruct()
	pipeline := []bson.M{}

	// var populateTree map[string]interface{}
	// // populate tree
	// if len(populateTrees) >= 1 {
	// 	populateTree = populateTrees[0]
	// } else {
	populateTree := buildPopulateTree(q.populate)
	// }

	for _, field := range documentStruct.StructFields {
		if field.Relationship != nil {
			rs := field.Relationship
			if _, found := populateTree[strings.ToUpper(field.Name)]; !found {
				// not need populate the relation
				continue
			}
			switch rs.Kind {
			case HasOne:
				pipeline = append(pipeline, bson.M{
					"$lookup": bson.M{
						"from":         rs.CollectionName,
						"localField":   rs.LocalFieldKey,
						"foreignField": rs.ForeignFieldKey,
						"as":           field.ColumnName + "@ARR",
					},
				})
			case HasMany:
				pipeline = append(pipeline, bson.M{
					"$lookup": bson.M{
						"from":         rs.CollectionName,
						"localField":   rs.LocalFieldKey,
						"foreignField": rs.ForeignFieldKey,
						"as":           field.ColumnName,
					},
				})
			case BelongTo:
				pipeline = append(pipeline, bson.M{
					"$lookup": bson.M{
						"from":         rs.CollectionName,
						"localField":   rs.LocalFieldKey,
						"foreignField": rs.ForeignFieldKey,
						"as":           field.ColumnName + "@ARR",
					},
				})
			// TODO BelongsTo
			default:
			}
		}
	}

	return pipeline
}

func (q *query) buildPipeQuery() {
	pipeline := []bson.M{}
	if len(q.populate) > 0 {
		pipes := q.getPopulatePipeline()
		for _, p := range pipes {
			pipeline = append(pipeline, p)
		}
	}
	if q.selector != nil {
		pipeline = append(pipeline, bson.M{"$project": q.selector})
	}
	if q.limit > 0 {
		pipeline = append(pipeline, bson.M{"$limit": q.limit})
	}
	if q.skip > 0 {
		pipeline = append(pipeline, bson.M{"$skip": q.skip})
	}
	if q.query != nil {
		pipeline = append(pipeline, bson.M{"$match": q.query})
	}

	q.pipeline = pipeline
}

func (q *query) execPopulateQuery(result interface{}, populateTrees ...map[string]interface{}) error {
	// pipeline := []bson.M{}
	q.buildPipeQuery()

	if q.multiple {
		return q.collection.Pipe(q.pipeline).All(result)
	}

	return q.collection.Pipe(q.pipeline).One(result)
}

// Exec 处理查询
func (q *query) Exec(result interface{}) error {
	if result == nil {
		panic("The result is required")
	}

	multiple := q.multiple

	resType := reflect.TypeOf(result)

	if resType.Kind() == reflect.Ptr && resType.Elem().Kind() == reflect.Slice {
		if !multiple {
			panic("Want a document but give a slice")
		}
	} else if _, ok := result.(Documenter); ok {
		if multiple {
			panic("Want a slice but give a document")
		}
	} else {
		panic("The resultset must be a Document")
	}

	var err error

	if multiple {
		err = q.all(result)
	} else {
		err = q.one(result)
	}

	if err == mgo.ErrNotFound {
		return &NotFoundError{NewError("Not found the document")}
	} else if err != nil {
		return err
	}

	// initDocuments(result, q.collection, q.connection)

	return nil
}

func (q *query) Create(document interface{}) error {
	doct := reflect.TypeOf(document)
	if doct.Kind() == reflect.Slice {
		elemType := doct.Elem()
		for elemType.Kind() == reflect.Ptr || elemType.Kind() == reflect.Slice {
			elemType = elemType.Elem()
		}

		documents := document.([]interface{})

		isDocumenter := isImplementsDocumenter(elemType)

		// before create hooker
		if isDocumenter {
			for _, v := range documents {
				if hooker, ok := v.(defaultCreateHooker); ok {
					if err := hooker.defaultBeforeCreate(); err != nil {
						return err
					}
				}
			}
		}

		err := q.collection.Insert(documents...)
		if err != nil {
			return err
		}

		// after create hooker
		if isDocumenter {
			for _, v := range documents {
				if doc, ok := v.(Documenter); ok {
					if err := doc.defaultAfterCreate(doc); err != nil {
						return err
					}
				}
			}
		}
	}

	// before create hooker
	if hooker, ok := document.(defaultCreateHooker); ok {
		if err := hooker.defaultBeforeCreate(); err != nil {
			return err
		}
	}
	err := q.collection.Insert(document)
	if err != nil {
		return err
	}

	// after create hooker
	if doc, ok := document.(Documenter); ok {
		if err := doc.defaultAfterCreate(doc); err != nil {
			return err
		}
	}

	return nil
}
