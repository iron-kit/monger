package monger

import (
	"fmt"
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
	Update(selector interface{}, docs interface{}) error
	Upsert(selector interface{}, docs interface{}) (*mgo.ChangeInfo, error)
	UpsertID(id interface{}, data interface{}) (*mgo.ChangeInfo, error)
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
	if len(q.populate) > 0 {
		q.buildPipeQuery()
		res := []map[string]interface{}{}
		err := q.collection.Pipe(q.pipeline).All(&res)
		if err != nil {
			return 0, err
		}

		return len(res), nil
	}

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

func (q *query) one(document interface{}) (err error) {
	q.multiple = false

	if len(q.populate) > 0 {

		err = q.execPopulateQuery(document)
	} else {
		query := q.collection.Find(q.query)
		q.buildQuery(query)

		err = query.One(document)
	}
	return
}

func (q *query) all(documents interface{}) (err error) {
	q.multiple = true

	if len(q.populate) > 0 {
		// use populate mode
		err = q.execPopulateQuery(documents)
	} else {
		query := q.collection.Find(q.query)
		q.buildQuery(query)
		err = query.All(documents)
	}
	// docst := reflect.TypeOf(documents)
	// // docsv := reflect.ValueOf(documents)
	// if docst.Kind() == reflect.Ptr && docst.Elem().Kind() == reflect.Slice {
	// 	if isImplementsDocumenter(docst.Elem().Elem()) {
	// 		if docs, ok := documents.(*[]Documenter); ok {
	// 			for _, doc := range *docs {
	// 				doc.setValue(doc)
	// 			}
	// 		}
	// 	}
	// } else if docst.Kind() == reflect.Slice {
	// 	if isImplementsDocumenter(docst.Elem()) {
	// 		if docs, ok := documents.([]Documenter); ok {
	// 			for _, doc := range docs {
	// 				doc.setValue(doc)
	// 			}
	// 		}
	// 	}
	// }

	return
}

func (q *query) getPopulatePipeline() []bson.M {
	documentStruct := q.mdl.GetDocumentStruct()
	pipeline := []bson.M{}
	populateTree := buildPopulateTree(q.populate)

	fmt.Println(len(documentStruct.StructFields))

	for _, field := range documentStruct.StructFields {
		fmt.Println(field.Name)
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
						// "as":           field.ColumnName + "@ARR",
						"as": field.ColumnName,
					},
				}, bson.M{
					"$unwind": bson.M{
						"path":                       "$" + field.ColumnName,
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
						// "as":           field.ColumnName + "@ARR",
						"as": field.ColumnName,
					},
				}, bson.M{
					"$unwind": bson.M{
						"path":                       "$" + field.ColumnName,
						"preserveNullAndEmptyArrays": true,
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
	// log.Log(pipeline, "pipeline ...")
	// fmt.Println("pipeline: ", pipeline)
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
	} else if resType.Kind() == reflect.Slice {
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
		resMaps := make([]map[string]interface{}, 0, 1)
		err = q.all(&resMaps)

		resultv := reflect.ValueOf(result)
		if resultv.Kind() != reflect.Ptr || resultv.Elem().Kind() != reflect.Slice {
			panic("result argument must be a slice address")
		}
		slicev := resultv.Elem()
		slicev = slicev.Slice(0, slicev.Cap())
		elemt := slicev.Type().Elem()
		i := 0

		for _, resMap := range resMaps {
			// dm := make(map[string]interface{})
			elemp := reflect.New(elemt)
			for k, v := range resMap {
				if strings.HasSuffix(k, "@ARR") {
					fieldNames := strings.Split(k, "@")
					if arr, ok := v.([]interface{}); ok {
						if len(arr) > 0 {
							resMap[fieldNames[0]] = arr[0]
						}
					}
				}
			}
			resByte, _ := bson.Marshal(resMap)
			err = bson.Unmarshal(resByte, elemp.Interface())
			slicev = reflect.Append(slicev, elemp.Elem())
			slicev = slicev.Slice(0, slicev.Cap())
			i++
		}
		resultv.Elem().Set(slicev.Slice(0, i))
	} else {
		resMap := make(map[string]interface{})
		err = q.one(&resMap)

		for k, v := range resMap {
			if strings.HasSuffix(k, "@ARR") {
				fieldNames := strings.Split(k, "@")
				if arr, ok := v.([]interface{}); ok {
					if len(arr) > 0 {
						resMap[fieldNames[0]] = arr[0]
					}
				}
			}
		}
		resByte, _ := bson.Marshal(resMap)
		err = bson.Unmarshal(resByte, result)
	}

	if err == mgo.ErrNotFound {
		return &NotFoundError{NewError("Not found the document")}
	} else if err != nil {
		return err
	}

	// initDocuments(result, q.collection, q.connection)

	return nil
}

func (q *query) insert(args ...interface{}) error {
	// TODO insert relation data
	return q.collection.Insert(args...)
}

func (q *query) Create(document interface{}) error {
	// fmt.Println(document)
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
				if doc, ok := v.(Documenter); ok {
					if err := doc.defaultBeforeCreate(doc, q.mdl); err != nil {
						return err
					}
				}
			}
		}

		err := q.insert(documents...)
		if err != nil {
			return err
		}

		// after create hooker
		if isDocumenter {
			for _, v := range documents {
				if doc, ok := v.(Documenter); ok {
					if err := doc.defaultAfterCreate(doc, q.mdl); err != nil {
						return err
					}
				}
			}
		}
	}

	// before create hooker
	if doc, ok := document.(Documenter); ok {
		if err := doc.defaultBeforeCreate(doc, q.mdl); err != nil {
			return err
		}
	}
	err := q.insert(document)
	if err != nil {
		return err
	}

	// after create hooker
	if doc, ok := document.(Documenter); ok {
		if err := doc.defaultAfterCreate(doc, q.mdl); err != nil {
			return err
		}
	}

	return nil
}

// 处理 Update/Upsert/UpsertID 的包装器
func (q *query) execUpdate(data interface{}, f func(d interface{})) {
	if doc, ok := data.(Documenter); ok {
		doc.defaultBeforeUpdate(doc, q.mdl)
		defer doc.defaultAfterUpdate(doc, q.mdl)
		f(bson.M{"$set": doc})
		return
	}

	if m, ok := data.(map[string]interface{}); ok {
		for _, v := range m {
			if doc, ok := v.(Documenter); ok {
				doc.defaultBeforeUpdate(doc, q.mdl)
				defer doc.defaultAfterUpdate(doc, q.mdl)
			}
		}
		f(data)
		return
	}

	if m, ok := data.(bson.M); ok {
		for _, v := range m {
			if doc, ok := v.(Documenter); ok {
				doc.defaultBeforeUpdate(doc, q.mdl)
				defer doc.defaultAfterUpdate(doc, q.mdl)
			}
		}
		f(data)
		return
	}

	f(data)
	return
}

func (q *query) Update(selector interface{}, docs interface{}) (err error) {

	q.execUpdate(docs, func(d interface{}) {
		err = q.collection.Update(selector, d)
	})

	return
}

func (q *query) Upsert(selector interface{}, docs interface{}) (info *mgo.ChangeInfo, err error) {
	q.execUpdate(docs, func(d interface{}) {
		info, err = q.collection.Upsert(selector, d)
	})

	return
}

func (q *query) UpsertID(id interface{}, docs interface{}) (info *mgo.ChangeInfo, err error) {
	q.execUpdate(docs, func(d interface{}) {
		info, err = q.collection.UpsertId(id, d)
	})

	return
}
