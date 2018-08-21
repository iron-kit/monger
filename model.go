package monger

import (
	"go/ast"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"reflect"
	"sync"
	"time"
)

/*

 */
type Model interface {
	getCollectionName() string
	UpsertID(id interface{}, data interface{}) (*mgo.ChangeInfo, error)
	Insert(docs interface{}) error
	Update(selector interface{}, docs interface{}) error
	Upsert(selector interface{}, docs interface{}) (*mgo.ChangeInfo, error)
	Doc(documents interface{})
	Count(q ...interface{}) int
	Create(docs interface{}) error
	Find(query ...interface{}) Query
	FindByID(id bson.ObjectId) Query
	Where(...interface{}) Query
	FindOne(query ...bson.M) Query

	// TODO
	// FindByIDAndDelete(id bson.ObjectId) Query
	// FindByIDAndRemove(id bson.ObjectId) Query
	// FindByIDAndUpdate(id bson.ObjectId) Query

	// FindOneAndDelete() Query
	// FindOneAndRemove() Query
	// FindOneAndUpdate() Query
	// Remove() Query
	// Update() Query
	// UpdateMany() Query

	// DeleteOne() Query
	// Delete() Query
}

type model struct {
	collection     *mgo.Collection
	connection     Connection
	documentSchema Documenter
	collectionName string
}

type DocumentStruct struct {
	Type            reflect.Type
	StructFields    []*DocumentField
	StructFieldsMap map[string]DocumentField
}

type DocumentField struct {
	Name            string
	ColumnName      string
	CollectionName  string
	IsNormal        bool
	IsIgnored       bool
	HasDefaultValue bool
	// IsInline        bool
	Tag          reflect.StructTag
	TagMap       map[string]string
	Struct       reflect.StructField
	IsForeignKey bool
	Relationship *Relationship
}

func newModel(connection *connection, document Documenter) Model {
	if document == nil {
		panic("Document can not be nil")
	}

	collectionName := ""
	if nameGetter, ok := document.(CollectionNameGetter); ok {
		collectionName = nameGetter.CollectionName()
	} else {
		collectionName = getDocumentTypeName(document)
	}

	// if _, ok := conn.modelStore[typeName]; !ok {
	collection := connection.Session.DB("").C(collectionName)
	mdl := &model{
		collection,
		connection,
		document,
		collectionName,
	}

	// conn.modelStore[typeName] = mdl
	// fmt.Printf("Type '%v' has registered \r\n", typeName)

	return mdl
	// }
}

func (m *model) getCollectionName() string {
	return m.collectionName
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
		m,
	)
}

func (m *model) Create(docs interface{}) error {
	m.Doc(docs)
	// insert document
	q := m.query()
	return q.Create(docs)
}

func (m *model) Populate(populates ...string) Query {
	// for _, str := populates

	query := newQuery(
		m.collection,
		m.connection,
		bson.M{},
		false,
		m,
	)

	query.Populate(populates...)

	return query
}

func (m *model) Doc(documents interface{}) {
	// initDocuments(documents, m.collection, m.connection)
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
	return newQuery(m.collection, m.connection, restQuery, multiple, m)
}

func (m *model) FindByID(id bson.ObjectId) Query {
	return newQuery(
		m.collection,
		m.connection,
		bson.M{"_id": id},
		false,
		m,
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
	// initDocumentData(docs, true)
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

	// initDocumentData(docs, false)
	return collection.Update(selector, docs)
}

func (m *model) Upsert(selector interface{}, docs interface{}) (*mgo.ChangeInfo, error) {
	collection, close := m.dbCollection()

	defer close()

	// initDocumentData(docs, false)
	return collection.Upsert(selector, docs)
}

func (m *model) UpsertID(id interface{}, data interface{}) (*mgo.ChangeInfo, error) {
	collection, close := m.dbCollection()

	defer close()

	// initDocumentData(data, false)
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

var docStructsMap sync.Map

func (m *model) GetDocumentStruct() *DocumentStruct {
	var docStruct DocumentStruct
	if m.documentSchema == nil {
		return &docStruct
	}

	reflectValue := reflect.ValueOf(m.documentSchema)
	reflectType := reflectValue.Type()

	if reflectType.Kind() == reflect.Slice || reflectType.Kind() == reflect.Ptr {
		reflectType = reflectType.Elem()
	}

	// Documenter first must be a struct
	if reflectType.Kind() != reflect.Struct {
		return &docStruct
	}

	if v, found := docStructsMap.Load(reflectType); found && v != nil {
		return v.(*DocumentStruct)
	}

	docStruct.Type = reflectType

	for i := 0; i < reflectType.NumField(); i++ {

		if fieldStruct := reflectType.Field(i); ast.IsExported(fieldStruct.Name) {
			field := &DocumentField{
				Struct: fieldStruct,
				Name:   fieldStruct.Name,
				Tag:    fieldStruct.Tag,
				TagMap: parseTagConfig(fieldStruct.Tag),
			}

			if _, found := field.TagMap["-"]; found {
				field.IsIgnored = true
			} else {
				if _, ok := field.TagMap["DEFAULT"]; ok {
					field.HasDefaultValue = true
				}

				if name, ok := field.TagMap["COLUMN"]; ok {
					field.ColumnName = name
				}

				indirectType := fieldStruct.Type
				for indirectType.Kind() == reflect.Ptr {
					indirectType = indirectType.Elem()
				}

				fieldValue := reflect.New(indirectType).Interface()
				if _, isTime := fieldValue.(*time.Time); isTime {
					field.IsNormal = true
				} else if indirectType.Kind() == reflect.Struct {
					field.IsNormal = true
				} else {

					switch indirectType.Kind() {
					case reflect.Slice:

						f := func(field *DocumentField) {
							var (
								localFieldKey   string
								foreignFieldKey string
								elemType        = field.Struct.Type
							)

							for elemType.Kind() == reflect.Ptr || elemType.Kind() == reflect.Slice {
								elemType = elemType.Elem()
							}

							if foreignKey := field.TagMap["FOREIGNKEY"]; foreignKey != "" {
								localFieldKey = "_id"
								foreignFieldKey = foreignKey
							}

							if localField := field.TagMap["LOCALFIELD"]; localField != "" {
								localFieldKey = localField
							}

							if foreignField := field.TagMap["FOREIGNFIELD"]; foreignField != "" {
								foreignFieldKey = foreignField
							}

							if elemType.Kind() == reflect.Struct && isImplementsDocumenter(elemType) {
								relationMdl := m.connection.M(elemType.Name())
								rs := &Relationship{
									ModelName:       elemType.Name(),
									RelationModel:   relationMdl,
									LocalFieldKey:   localFieldKey,
									ForeignFieldKey: foreignFieldKey,
								}

								rs.CollectionName = relationMdl.getCollectionName()
								if _, ok := field.TagMap["HASMANY"]; ok {
									rs.Kind = HasMany
								} else {
									// now just support has many, don't support many to many
									return
								}

								field.Relationship = rs

							}
						}
						defer f(field)
					case reflect.Struct:
						fallthrough
					case reflect.Ptr:
						defer func(field *DocumentField) {
							var (
								localFieldKey   string
								foreignFieldKey string
								elemType        = field.Struct.Type
							)

							for elemType.Kind() == reflect.Ptr {
								elemType = elemType.Elem()
							}

							if !isImplementsDocumenter(elemType) {
								return
							}
							relationMdl := m.connection.M(elemType.Name())
							rs := &Relationship{
								ModelName:     elemType.Name(),
								RelationModel: relationMdl,
							}

							rs.CollectionName = relationMdl.getCollectionName()

							if _, ok := field.TagMap["HASONE"]; ok {
								rs.Kind = HasOne
							} else if _, ok := field.TagMap["BELONGTO"]; ok {
								rs.Kind = BelongTo
								return
							}

							if rs.Kind == "" {
								return
							}

							if foreignKey := field.TagMap["FOREIGNKEY"]; foreignKey != "" {
								if rs.Kind == HasOne {
									localFieldKey = "_id"
									foreignFieldKey = foreignKey
								} else {
									localFieldKey = foreignKey
									foreignFieldKey = "_id"
								}
							}

							if localField := field.TagMap["LOCALFIELD"]; localField != "" {
								localFieldKey = localField
							}

							if foreignField := field.TagMap["FOREIGNFIELD"]; foreignField != "" {
								foreignFieldKey = foreignField
							}
							rs.ForeignFieldKey = foreignFieldKey
							rs.LocalFieldKey = localFieldKey
							field.Relationship = rs
						}(field)
					default:
						field.IsNormal = true
					}
				}
			}

			docStruct.StructFields = append(docStruct.StructFields, field)
		}
	}

	docStructsMap.Store(reflectType, &docStruct)
	return &docStruct
}
