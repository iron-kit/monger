package monger

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type PopulateItem struct {
	Name     string
	Children []*PopulateItem
}

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
	OnlyTrashed() Query
	OffSoftDeletes() Query
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
	Restore() error
	Delete() error
	DeleteAll() (*mgo.ChangeInfo, error)
	ForceDelete() error
	ForceDeleteAll() (*mgo.ChangeInfo, error)
	Skip(skip int) Query
	Limit(limit int) Query
	Sort(fields ...string) Query
	Aggregate([]bson.M) Query
	Pipe(...bson.M) *mgo.Pipe
	Query() Query
}

type query struct {
	withTrashed    bool
	onlyTrashed    bool
	offSoftDeletes bool
	collection     *mgo.Collection
	where          bson.M
	selector       interface{}
	populate       []string
	sort           []string
	limit          int
	skip           int
	pipeline       []bson.M
	multiple       bool
	schemaStruct   *SchemaStruct
}

func (q *query) Query() Query {
	qCopy := *q
	return &qCopy
}

func (q *query) Pipe(pipes ...bson.M) *mgo.Pipe {
	return q.buildPipeQuery(pipes...)
}

func (q *query) OnlyTrashed() Query {
	q.onlyTrashed = true

	return q
}

func (q *query) WithTrashed() Query {
	q.withTrashed = true
	return q
}

func (q *query) OffSoftDeletes() Query {
	q.offSoftDeletes = true

	return q
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

func (q *query) usePipeline() bool {
	if len(q.pipeline) > 0 || len(q.populate) > 0 {
		return true
	}

	return false
}

func (q *query) Count() int {

	if q.usePipeline() {
		appendPipes := []bson.M{
			{
				"$group": bson.M{
					"_id": "null",
					"count": bson.M{
						"$sum": 1,
					},
				},
			},
			{
				"$project": bson.M{"_id": 0},
			},
		}
		result := struct {
			Total int `bson:"count"`
		}{}

		q.buildPipeQuery(appendPipes...).One(&result)

		return result.Total
		// q.pipeline = append(q.pipeline, )
	}

	c, err := q.collection.Find(q.where).Count()
	if err != nil {
		return 0
	}
	return c
}

func (q *query) Restore() error {
	if q.offSoftDeletes {
		return nil
	}

	_, err := q.collection.Upsert(q.where, bson.M{"$set": bson.M{
		"deleted": false,
	}})

	return err
}

func (q *query) Delete() error {
	if !q.offSoftDeletes {
		return q.collection.Update(q.where, bson.M{"$set": bson.M{"deleted": true}})
	}
	return q.ForceDelete()
}

func (q *query) DeleteAll() (info *mgo.ChangeInfo, err error) {
	if !q.offSoftDeletes {
		return q.collection.Upsert(q.where, bson.M{"$set": bson.M{
			"deleted": true,
		}})
	}

	return q.ForceDeleteAll()
}

func (q *query) ForceDelete() error {
	return q.collection.Remove(q.where)
}

func (q *query) ForceDeleteAll() (*mgo.ChangeInfo, error) {
	return q.collection.RemoveAll(q.where)
}

func (q *query) Populate(fields ...string) Query {
	if q.populate == nil {
		q.populate = make([]string, 0)
	}
	q.populate = append(q.populate, fields...)
	return q
}

func executeWhere(in interface{}, condition bson.M) {
	if w, ok := in.(bson.M); ok {
		for key, val := range condition {
			switch v := val.(type) {
			case string:
				if bson.IsObjectIdHex(v) {
					w[key] = bson.ObjectIdHex(v)
				} else {
					w[key] = v
				}
			case bson.M:
				executeWhere(w[key], v)
			default:
				w[key] = val
			}
		}
	}
}

func (q *query) Where(condition bson.M) Query {

	// panic("not implemented")
	if q.where == nil {
		q.where = make(bson.M)
	}

	executeWhere(q.where, condition)

	if !q.offSoftDeletes {
		if !q.withTrashed {
			q.where["deleted"] = false
		}

		if q.onlyTrashed {
			q.where["deleted"] = true
		}
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

func (q *query) Aggregate(pipe []bson.M) Query {
	if len(q.pipeline) > 0 {
		q.pipeline = append(q.pipeline, pipe...)

		return q
	}

	q.pipeline = pipe
	return q
	// return q.buildPipeQuery()
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

func getPopulateTree(populate []string) []*PopulateItem {
	cache := make(map[string]*PopulateItem)
	for _, p := range populate {
		k := strings.ToUpper(p)
		karr := strings.Split(p, ".")
		if _, ok := cache[k]; !ok {
			cache[k] = &PopulateItem{
				Name:     karr[len(karr)-1],
				Children: make([]*PopulateItem, 0),
			}
		}

		// strings.Split()
	}

	items := make([]*PopulateItem, 0)

	for _, p := range populate {
		k := strings.ToUpper(p)
		karr := strings.Split(k, ".")

		if len(karr) == 1 {
			items = append(items, cache[karr[0]])
		}

		if len(karr) > 1 {
			parentK := karr[:len(karr)-1]
			if i, ok := cache[strings.Join(parentK, ".")]; ok {
				i.Children = append(i.Children, cache[k])
			}
		}
	}

	return items
}

func getRelationLookup(populateItems []*PopulateItem, schemaStruct *SchemaStruct) []bson.M {
	// populate := make([]string, 0)

	pipelines := make([]bson.M, 0)
	for index, item := range populateItems {
		// populate = append(populate, item.Name)
		// if
		if field, ok := schemaStruct.FieldsMap[item.Name]; ok && field.HasRelation {
			rs := field.Relationship

			localFieldKey := fmt.Sprintf("refLocalFieldKey_%s%d", rs.CollectionName, index)
			childPipeline := []bson.M{
				{
					"$match": bson.M{
						"$expr": bson.M{
							"$eq": []string{"$" + rs.ForeignFieldKey, "$$" + localFieldKey},
						},
					},
				},
			}

			if len(item.Children) > 0 && field.RelationshipStruct != nil {
				// fmt.Println(item.Children[0], "children")
				// fmt.Println(field.RelationshipStruct, "struct")
				pipes := getRelationLookup(item.Children, field.RelationshipStruct)
				// fmt.Println(pipes, "pipes")
				childPipeline = append(childPipeline, pipes...)
			}

			pipeline := bson.M{
				"$lookup": bson.M{
					"from":     rs.From,
					"let":      bson.M{localFieldKey: "$" + rs.LocalFieldKey},
					"pipeline": childPipeline,
					"as":       rs.As,
				},
			}

			// fmt.Println(pipeline)
			pipelines = append(pipelines, pipeline)
			switch rs.Kind {
			case HasOne:
				fallthrough
			case BelongTo:
				// defer func() {
				pipelines = append(pipelines, bson.M{
					"$unwind": bson.M{
						"path": "$" + rs.As,
						"preserveNullAndEmptyArrays": true,
					},
				})
				// }()
			default:
				break
			}

		}

	}

	fmt.Println(pipelines)
	return pipelines
	// for _, p := range populate {

	// }
	// rs := field.Relationship
	// localFieldKey := "refLocalFieldKey_" + rs.CollectionName
	// // sLocalFieldKey := "$" + localFieldKey

	// childPipeline := []bson.M{
	// 	{
	// 		"$match": bson.M{"$expr": bson.M{
	// 				"$eq": []string{"$" + rs.ForeignFieldKey, "$$" + localFieldKey},
	// 			},
	// 		},
	// 	},
	// }

	// // if field.RelationshipStruct.

	// // if field.RelationshipStruct != nil {
	// // 	childPipeline = append(childPipeline, getRelationLookup(field.RelationshipStruct))
	// // }

	// pipeline := bson.M{
	// 	"$lookup": {
	// 		"from": rs.From,
	// 		"let": bson.M{localFieldKey: "$" + rs.LocalFieldKey},
	// 		"pipeline": childPipeline,
	// 		"as": rs.As,
	// 	},
	// }
}

func (q *query) getPopulatePipeline() []bson.M {
	populateTree := getPopulateTree(q.populate)

	return getRelationLookup(populateTree, q.schemaStruct)

	// documentStruct := q.documentStruct

	// populate = []string{"User", "User.Profile", "User.School"}
	// q.populate

	// schemaStruct := q.schemaStruct
	// pipeline := make([]bson.M, 0)

	// for _, item := range populateTree {
	// 	if field, ok := schemaStruct.FieldsMap[item.Name]; ok && field.HasRelation {
	// 		rs := field.Relationship

	// 		switch rs.Kind {
	// 		case HasOne:
	// 			fallthrough
	// 		case BelongTo:
	// 			pipeline
	// 			// pipeline = append(pipeline, bson.M{
	// 			// 	"$lookup":
	// 			// })
	// 		}
	// 	}
	// }

	// for _, field := range schemaStruct.StructFields {
	// 	if field.Relationship != nil {
	// 		rs := field.Relationship
	// 		switch rs.Kind {
	// 		case HasOne:
	// 			pipeline = append(pipeline, bson.M{
	// 				"$lookup": bson.M{
	// 					"from":         rs.CollectionName,
	// 					"localField":   rs.LocalFieldKey,
	// 					"foreignField": rs.ForeignFieldKey,
	// 					"as":           field.ColumnName,
	// 				},
	// 			}, bson.M{
	// 				"$unwind": bson.M{
	// 					"path": "$" + field.ColumnName,
	// 					"preserveNullAndEmptyArrays": true,
	// 				},
	// 			})
	// 		case HasMany:
	// 			pipeline = append(pipeline, bson.M{
	// 				"$lookup": bson.M{
	// 					"from":         rs.CollectionName,
	// 					"localField":   rs.LocalFieldKey,
	// 					"foreignField": rs.ForeignFieldKey,
	// 					"as":           field.ColumnName,
	// 				},
	// 			})
	// 		case BelongTo:
	// 			pipeline = append(pipeline, bson.M{
	// 				"$lookup": bson.M{
	// 					"from":         rs.CollectionName,
	// 					"localField":   rs.LocalFieldKey,
	// 					"foreignField": rs.ForeignFieldKey,
	// 					"as":           field.ColumnName,
	// 				},
	// 			}, bson.M{
	// 				"$unwind": bson.M{
	// 					"path": "$" + field.ColumnName,
	// 					"preserveNullAndEmptyArrays": true,
	// 				},
	// 			})
	// 			// TODO
	// 		case BelongsTo:
	// 		default:
	// 		}
	// 	}
	// }
	// TODO impl
	// return make([]bson.M, 0)
	// return pipeline
}

func (q *query) buildPipeQuery(appendPipes ...bson.M) *mgo.Pipe {
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

	if len(appendPipes) > 0 {
		q.pipeline = append(q.pipeline, appendPipes...)
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

		if q.usePipeline() {
			return q.execPipeMuli(result)
		}

		return q.execMuli(result)
	}

	if q.usePipeline() {
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
		d.beforeCreate(doc)
		q.collection.Insert(doc)
		d.afterCreate()
		return nil
	}

	return &InvalidParamsError{NewError("Document must be schemer")}
}

func (q *query) execUpdate(data interface{}, f func(d interface{})) {
	// datat := reflect.TypeOf(data)
	datav := reflect.ValueOf(data)
	for datav.Kind() == reflect.Ptr {
		datav = datav.Elem()
	}

	if isImplementsSchemer(datav.Type()) {

		if reflect.TypeOf(data).Kind() == reflect.Struct {
			panic("the doc must be a pointer")
		}

		if doc, ok := data.(Schemer); ok {
			doc.beforeUpdate(data)
			defer doc.afterUpdate()
			f(bson.M{"$set": doc})
		}
	}

	switch datav.Kind() {
	case reflect.Map:
		vv := datav.Interface()
		mapData := vv.(bson.M)
		foundSet := false
		now := time.Now()
		for k, val := range mapData {
			if k == "$set" {
				foundSet = true

				if d, ok := val.(bson.M); ok {
					d["updated_at"] = now
				}

				if d, ok := val.(map[string]interface{}); ok {
					d["updated_at"] = now
				}

				if d, ok := val.(Schemer); ok {
					d.beforeUpdate(data)

					defer d.afterUpdate()
				}
			}
		}

		if !foundSet {
			mapData["updated_at"] = now
		}

		f(data)

	case reflect.Struct:
		f(bson.M{"$set": data})
	default:
		f(data)
	}

	// defer func() {
	// 	fmt.Println(data, "data")
	// }()
}

func (q *query) Update(condition bson.M, doc interface{}) (err error) {
	// panic("not implemented")
	cond := bson.M{}
	executeWhere(cond, condition)
	q.execUpdate(doc, func(d interface{}) {
		err = q.collection.Update(cond, d)
	})

	return
}

func (q *query) Upsert(condition bson.M, docs interface{}) (changeInfo *mgo.ChangeInfo, err error) {
	cond := bson.M{}
	executeWhere(cond, condition)
	q.execUpdate(docs, func(d interface{}) {
		changeInfo, err = q.collection.Upsert(condition, d)
	})

	return
}

func (q *query) UpsertID(id interface{}, docs interface{}) (changeInfo *mgo.ChangeInfo, err error) {
	// cond := bson.M{}
	if ids, ok := id.(string); ok {
		if bson.IsObjectIdHex(ids) {
			id = bson.ObjectIdHex(ids)
		}
	}
	// executeWhere(cond, condition)
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
