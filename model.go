package monger

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

/*

 */
type Model interface {
	UpsertID(id interface{}, data interface{}) (*mgo.ChangeInfo, error)
	Insert(docs interface{}) error
	Update(selector interface{}, docs interface{}) error
	Upsert(selector interface{}, docs interface{}) (*mgo.ChangeInfo, error)
	Doc(documents interface{})
	Count(q ...interface{}) int
	Create(docs interface{})
	Find(query ...interface{}) Query
	FindByID(id bson.ObjectId) Query
	// FindByIDAndDelete(id bson.ObjectId) Query
	// FindByIDAndRemove(id bson.ObjectId) Query
	// FindByIDAndUpdate(id bson.ObjectId) Query

	FindOne(query ...bson.M) Query
	// FindOneAndDelete() Query
	// FindOneAndRemove() Query
	// FindOneAndUpdate() Query

	// Remove() Query
	// Update() Query
	// UpdateMany() Query
	Where(...interface{}) Query

	// DeleteOne() Query
	// Delete() Query
}

type model struct {
	collection *mgo.Collection
	connection Connection
}

func (m *model) query(q ...interface{}) Query {

	isMultiple := true
	resQuery := bson.M{}

	for _, v := range q {
		if multiple, ok := v.(bool); ok {
			// resQuery.
			isMultiple = multiple
		}

		if resq, ok := v.(bson.M); ok {
			resQuery = resq
		}
	}

	return newQuery(
		m.collection,
		m.connection,
		resQuery,
		isMultiple,
	)
}

func (m *model) Create(docs interface{}) {
	m.Doc(docs)
}

func (m *model) Doc(documents interface{}) {
	initDocuments(documents, m.collection, m.connection)
}

// func (m *model) Doc(doc Document)

func (m *model) Where(q ...interface{}) Query {
	q = append(q, true)
	return m.query(q...)
}

func (m *model) Count(q ...interface{}) int {
	q = append(q, true)
	query := m.query(q...)
	log.Println(m.collection.Name)
	c, err := query.Count()
	if err != nil {
		// log.Output(0, err.Error())
		log.Println(err.Error())
		return 0
	}

	return c
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
		bson.M{"_id": id},
		false,
	)
}

func (m *model) dbCollection() (*mgo.Collection, func()) {
	config := m.connection.GetConfig()

	// TODO Implemented validate document
	session := m.connection.CloneSession()
	// defer session.Close()

	closeFunc := func() {
		session.Close()
	}

	collection := session.DB(config.DBName).C(m.collection.Name)

	return collection, closeFunc
}

func (m *model) Insert(docs interface{}) error {
	collection, close := m.dbCollection()

	defer close()
	initDocumentData(docs, true)
	return collection.Insert(docs)
}

// func (m *model) getCanUpdateDoc(docs interface{}) interface{} {
// 	docst := reflect.TypeOf(docs)
// 	docsv := reflect.ValueOf(docs)
// 	if docst.Kind() == reflect.Ptr && docst.Elem().Kind() == reflect.Slice {
// 		panic("[Monger] Can't update a slice")
// 	}
// 	var documenter Documenter
// 	t := docst
// 	if docst.Kind() == reflect.Ptr {
// 		t = docst.Elem()
// 	}

// 	if docst.Implements(reflect.TypeOf(&documenter)) {
// 		bm := bson.M{}
// 		structInfo, err := getDocumentStructInfo(reflect.TypeOf(docs))

// 		for _, info := range structInfo.FieldsList {
// 			var val reflect.Value

// 			if info.Inline == nil {
// 				val = docsv.Field(info.Num)
// 			} else {
// 				val = docsv.FieldByIndex(info.Inline)
// 			}

// 		}
// 		// n := docsv.NumField()
// 		// for i := 0; i < n; i++ {
// 		// 	field := docsv.Field(i)

// 		// }
// 	} else {
// 		return docs
// 	}
// }

func (m *model) Update(selector interface{}, docs interface{}) error {
	collection, close := m.dbCollection()

	defer close()

	initDocumentData(docs, false)
	return collection.Update(selector, docs)
}

func (m *model) Upsert(selector interface{}, docs interface{}) (*mgo.ChangeInfo, error) {
	collection, close := m.dbCollection()

	defer close()

	initDocumentData(docs, false)
	return collection.Upsert(selector, docs)
}

func (m *model) UpsertID(id interface{}, data interface{}) (*mgo.ChangeInfo, error) {
	collection, close := m.dbCollection()

	defer close()

	initDocumentData(data, false)
	return collection.UpsertId(id, data)
}

// func (m *model) FindByIDAndDelete(id bson.ObjectId) Query {
// 	// m.FindByID(id).Remove()

// }

// func (m *model) FindByIDAndRemove(id bson.ObjectId) Query {

// }

// func (m *model) FindByIDAndUpdate(id bson.ObjectId) Query {

// }

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
