package monger

import (
	"reflect"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
Query is the mongodb query help tools,

For Example:
	users := make([]*models.User, 0)
	UserModel.
		Where(bson.M{"_id": "12131312"}).
		Populate("Messages").
		Skip(0).
		Limit(12).
		FindAll(&users)

	user := new(models.User)
	UserModel.
		Where(bson.M{"_id": "qweqeqweqweqweqwe"}).
		FindOne(user)
*/
type Query interface {
	// Unscoped() Query
	Collection() *mgo.Collection
	Select(selector bson.M) Query
	Where(condition bson.M) Query
	FindOne(interface{}) error
	FindAll(interface{}) error
	Count() int
	Populate(fields ...string) Query
	exec(interface{}) error
	Create(interface{}) error
	Update(condition bson.M, docs interface{}) error
	Upsert(condition bson.M, docs interface{}) (*mgo.ChangeInfo, error)
	UpsertID(id interface{}, docs interface{}) (*mgo.ChangeInfo, error)
	Remove() error
	RemoveAll() (*mgo.ChangeInfo, error)
	Skip(skip int) Query
	Limit(limit int) Query
	Sort(fields ...string) Query
	Aggregate([]bson.M) *mgo.Pipe
}

type query struct {
	collection   *mgo.Collection
	where        bson.M
	selector     interface{}
	populate     []string
	sort         []string
	limit        int
	skip         int
	pipeline     []bson.M
	multiple     bool
	schemaStruct *SchemaStruct
}

func (q *query) Collection() *mgo.Collection {
	return q.collection
}

func (q *query) Select(selector bson.M) Query {
	q.selector = selector
	return q
}

func (q *query) Sort(fields ...string) Query {
	if q.sort == nil {
		q.sort = make([]string, 0)
	}
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

func (q *query) Count() int {
	c, err := q.collection.Find(q.where).Count()
	if err != nil {
		return 0
	}
	return c
}

func (q *query) Remove() error {
	return q.collection.Remove(q.where)
}

func (q *query) RemoveAll() (info *mgo.ChangeInfo, err error) {
	return q.collection.RemoveAll(q.where)
}

func (q *query) Populate(fields ...string) Query {
	if q.populate == nil {
		q.populate = make([]string, 0)
	}
	q.populate = append(q.populate, fields...)
	return q
}

func (q *query) Where(condition bson.M) Query {
	// panic("not implemented")
	if q.where == nil {
		q.where = make(bson.M)
	}

	for key, val := range condition {
		q.where[key] = val
	}

	return q
}

func (q *query) FindOne(result interface{}) error {
	// panic("not implemented")
	q.multiple = false
	return q.exec(result)
}

func (q *query) FindAll(result interface{}) error {
	// panic("not implemented")
	q.multiple = true
	return q.exec(result)
}

func (q *query) Aggregate(pipe []bson.M) *mgo.Pipe {
	q.pipeline = pipe

	return q.buildPipeQuery()
}

func (q *query) buildQuery() *mgo.Query {
	query := q.collection.Find(q.where)

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

	return query
}

func (q *query) getPopulatePipeline() []bson.M {
	// documentStruct := q.documentStruct

	schemaStruct := q.schemaStruct
	pipeline := make([]bson.M, 0)

	for _, field := range schemaStruct.StructFields {
		if field.Relationship != nil {
			rs := field.Relationship
			switch rs.Kind {
			case HasOne:
				pipeline = append(pipeline, bson.M{
					"$lookup": bson.M{
						"from":         rs.CollectionName,
						"localField":   rs.LocalFieldKey,
						"foreignField": rs.ForeignFieldKey,
						"as":           field.ColumnName,
					},
				}, bson.M{
					"$unwind": bson.M{
						"path": "$" + field.ColumnName,
						"preserveNullAndEmptyArrays": true,
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
						"as":           field.ColumnName,
					},
				}, bson.M{
					"$unwind": bson.M{
						"path": "$" + field.ColumnName,
						"preserveNullAndEmptyArrays": true,
					},
				})
				// TODO
			case BelongsTo:
			default:
			}
		}
	}
	// TODO impl
	// return make([]bson.M, 0)
	return pipeline
}

func (q *query) buildPipeQuery() *mgo.Pipe {
	pipeline := make([]bson.M, 0)

	if len(q.populate) > 0 {
		pipes := q.getPopulatePipeline()
		pipeline = append(pipeline, pipes...)
	}

	if q.selector != nil {
		pipeline = append(pipeline, bson.M{"$project": q.selector})
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

	if q.where != nil {
		pipeline = append(pipeline, bson.M{"$match": q.where})
	}

	if q.pipeline == nil {
		q.pipeline = pipeline
	} else {
		q.pipeline = append(q.pipeline, pipeline...)
	}

	return q.collection.Pipe(q.pipeline)
}

func (q *query) execPipeMuli(results interface{}) error {
	return q.buildPipeQuery().All(results)
}

func (q *query) execPipeOne(result interface{}) error {
	return q.buildPipeQuery().One(result)
}

func (q *query) execMuli(result interface{}) error {
	return q.buildQuery().All(result)
}

func (q *query) execOne(result interface{}) error {
	return q.buildQuery().One(result)
}

func (q *query) exec(result interface{}) error {
	if result == nil {
		return &InvalidParamsError{NewError("The result is required")}
	}
	multiple := q.multiple
	// resultv := reflect.ValueOf(result)
	resultType := reflect.TypeOf(result)

	for {
		resultType = resultType.Elem()
		if resultType.Kind() != reflect.Ptr {
			break
		}
	}

	if multiple {
		// check is result a slice
		if resultType.Kind() != reflect.Slice {
			return &InvalidParamsError{NewError("The result must be a slice")}
		}

		if len(q.populate) > 0 {
			return q.execPipeMuli(result)
		}

		return q.execMuli(result)
	}

	if len(q.populate) > 0 {
		return q.execPipeOne(result)
	}

	return q.execOne(result)
}

func (q *query) Create(doc interface{}) error {
	doct := reflect.TypeOf(doc)

	for {
		if doct.Kind() != reflect.Ptr {
			break
		}
		doct = doct.Elem()
	}
	if doct.Kind() == reflect.Slice {

		// TODO batch create
		return nil
	}
	if d, ok := doc.(Schemer); ok {
		d.beforeCreate()
		q.collection.Insert(doc)
		d.afterCreate()
		return nil
	}

	return &InvalidParamsError{NewError("Document must be schemer")}
}

func (q *query) execUpdate(data interface{}, f func(d interface{})) {
	if doc, ok := data.(Schemer); ok {
		doc.beforeUpdate()
		defer doc.afterUpdate()

		f(bson.M{"$set": doc})

		return
	}

	// datat := reflect.TypeOf(data)

	datav := reflect.ValueOf(data)

	// datav.SetMapIn
	for {
		if datav.Kind() != reflect.Ptr {
			break
		}
		datav = datav.Elem()
	}
	// reflect.
	t := reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf(new(interface{})))

	switch datav.Kind() {
	case t.Kind():
		mapData := datav.Interface().(map[string]interface{})
		foundSet := false
		now := time.Now()
		for k, val := range mapData {
			if k == "$set" {
				foundSet = true

				if d, ok := val.(map[string]interface{}); ok {
					d["updated_at"] = now
				}

				if d, ok := val.(Schemer); ok {
					d.beforeUpdate()

					defer d.afterUpdate()
				}
			}
		}

		if !foundSet {
			mapData["updated_at"] = now
		}
		f(data)
	default:
		f(data)
	}
}

func (q *query) Update(condition bson.M, doc interface{}) (err error) {
	// panic("not implemented")
	q.execUpdate(doc, func(d interface{}) {
		err = q.collection.Update(condition, d)
	})

	return
}

func (q *query) Upsert(condition bson.M, docs interface{}) (changeInfo *mgo.ChangeInfo, err error) {
	q.execUpdate(docs, func(d interface{}) {
		changeInfo, err = q.collection.Upsert(condition, d)
	})

	return
}

func (q *query) UpsertID(id interface{}, docs interface{}) (changeInfo *mgo.ChangeInfo, err error) {
	q.execUpdate(docs, func(d interface{}) {
		changeInfo, err = q.collection.UpsertId(id, d)
	})

	return
}

func newQuery(coll *mgo.Collection, sinfo *SchemaStruct) Query {
	return &query{
		collection:   coll,
		schemaStruct: sinfo,
	}
}
