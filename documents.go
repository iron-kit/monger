package monger

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"strings"
	"time"
)

type DocumentHooker interface {
	BeforeSave()
}

type Documenter interface {
	// DocumentManager
	DocumentHooker
	bson.Getter
	// bson.Setter
	GetID() bson.ObjectId
	SetID(bson.ObjectId)

	SetCreatedAt(time.Time)
	GetCreatedAt() time.Time

	SetUpdatedAt(time.Time)
	GetUpdatedAt() time.Time

	SetDeleted(bool)
	IsDeleted() bool
	CollectionName() string

	Save() error
	Validate(...interface{}) (bool, []error)
	DefaultValidate() (bool, []error)

	getDocumentManager() *documentManager
	dM() *documentManager
}

type documentManager struct {
	document   Documenter
	connection Connection
	collection *mgo.Collection
	// documentMap map[string]interface{}
}

type Document struct {
	dm        documentManager
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time     `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt time.Time     `json:"updatedAt" bson:"updatedAt,omitempty"`
	Deleted   bool          `json:"-" bson:"deleted"`
	Upsert    bool          `json:"-" bson:"-"`
}

func D(doc Documenter, collection *mgo.Collection, connection Connection) Documenter {

	dm := doc.dM()
	dm.SetCollection(collection)
	dm.SetConnection(connection)
	dm.SetDocument(doc)
	return doc
}

func initDocumentData(docs interface{}, isCreate bool) {
	if docs == nil {
		return
	}

	now := time.Now()

	docst := reflect.TypeOf(docs)
	if docst.Kind() == reflect.Ptr && docst.Elem().Kind() == reflect.Slice {
		docsv := reflect.ValueOf(docs)
		slicev := docsv.Elem()
		for i := 0; i < slicev.Len(); i++ {
			slicevItem := slicev.Index(i)
			if doc, ok := slicevItem.Interface().(Documenter); ok {
				if isCreate {
					doc.SetCreatedAt(now)
					doc.SetID(bson.NewObjectId())
				}
				doc.SetUpdatedAt(now)

			} else if slicevItem.Type().Kind() == reflect.Struct {
				if doc, ok := slicevItem.Addr().Interface().(Documenter); ok {
					doc.SetUpdatedAt(now)
					if isCreate {
						doc.SetCreatedAt(now)
						doc.SetID(bson.NewObjectId())
					}
				} else {
					panic("[Monger] The first param must be []Document Slice")
				}
			} else {
				panic("[Monger] The first param must be []Document Slice")
			}
		}
	} else if d, ok := docs.(Documenter); ok {
		// not slice
		if isCreate {
			d.SetCreatedAt(now)
			d.SetID(bson.NewObjectId())
		}
		d.SetUpdatedAt(now)
	} else {
		// panic("[Monger] The first param must be Document")
		return
	}
}

func initDocuments(documents interface{}, collection *mgo.Collection, connection Connection) {
	if documents == nil {
		return
	}
	// isSlice := false

	resType := reflect.TypeOf(documents)
	var dm *documentManager
	if resType.Kind() == reflect.Ptr && resType.Elem().Kind() == reflect.Slice {

		// is slice
		resultv := reflect.ValueOf(documents)
		slicev := resultv.Elem()

		for i := 0; i < slicev.Len(); i++ {
			ele := slicev.Index(i)
			if doc, ok := ele.Interface().(Documenter); ok {
				// fmt.Println(ele, "是 Document")
				dm = doc.dM()
				dm.SetCollection(collection)
				dm.SetConnection(connection)
				dm.SetDocument(doc)

			} else if ele.Type().Kind() == reflect.Struct {
				if doc, ok := ele.Addr().Interface().(Documenter); ok {
					dm = doc.dM()
					dm.SetCollection(collection)
					dm.SetConnection(connection)
					dm.SetDocument(doc)
					// dm.bindDocData()
				} else {
					panic("[Monger] The first param must be []Document Slice")
				}
			} else {
				panic("[Monger] The first param must be []Document Slice")
			}
		}
	} else if d, ok := documents.(Documenter); ok {
		// not slice
		dm = d.dM()
		dm.SetCollection(collection)
		dm.SetConnection(connection)
		dm.SetDocument(d)
		// dm.bindDocData()
	} else {
		panic("[Monger] The first param must be Document")
	}
	return
}

func (dm *documentManager) SetConnection(c Connection) {
	dm.connection = c
}

func (dm *documentManager) GetCollection() *mgo.Collection {
	return dm.collection
}

func (dm *documentManager) SetCollection(collection *mgo.Collection) {
	dm.collection = collection
}

func (dm *documentManager) GetConnection() Connection {
	return dm.connection
}

func (dm *documentManager) isNil() bool {
	if dm.document == nil || dm.collection == nil || dm.connection == nil {
		// panic("[monger] Please use D(document Document) function to build a saveable Document")
		return true
	}

	return false
}

func (dm *documentManager) SetDocument(document Documenter) {
	dm.document = document
}

func (doc *Document) dM() *documentManager {
	return &doc.dm
}

func (doc *Document) getDocumentManager() *documentManager {
	if doc.dm.isNil() {
		panic("[Monger] Please init the document")
	}
	return &doc.dm
}

func (doc *Document) CollectionName() string {
	return ""
}

func (doc *Document) GetID() bson.ObjectId {
	return doc.ID
}

func (doc *Document) SetID(id bson.ObjectId) {
	doc.ID = id
}

func (doc *Document) SetCreatedAt(time time.Time) {
	doc.CreatedAt = time
}

func (doc *Document) GetCreatedAt() time.Time {
	return doc.CreatedAt
}

func (doc *Document) SetUpdatedAt(time time.Time) {
	doc.UpdatedAt = time
}

func (doc *Document) GetUpdatedAt() time.Time {
	return doc.UpdatedAt
}

func (doc *Document) SetDeleted(isDeleted bool) {
	doc.Deleted = isDeleted
}

func (doc *Document) IsDeleted() bool {
	return doc.Deleted
}

func (doc *Document) BeforeSave() {
	return
}

// func (doc *Document) executeHasOne()

func (doc *Document) executeRelate(dm *documentManager) {
	doct := reflect.TypeOf(dm.document)
	docv := reflect.ValueOf(dm.document)
	if doct.Kind() == reflect.Ptr {
		docv = docv.Elem()
	}
	// fmt.Println(dm.document)
	structInfo, err := getDocumentStructInfo(doct)

	if err != nil {
		panic(err)
	}
	// fmt.Println(structInfo.FieldsList)
	for _, fieldInfo := range structInfo.RelateFieldsList {
		var fieldt reflect.Type
		if doct.Kind() == reflect.Ptr {
			fieldt = doct.Elem()
		} else {
			fieldt = doct
		}

		field := docv.Field(fieldInfo.Num)
		if field.Kind() != reflect.Ptr {
			// field = field.Elem()
			break
		}
		if field.IsNil() {
			break
		}
		fieldv := field.Elem()

		modelName := fieldInfo.RelateType.Elem().Name()
		mdl := dm.connection.M(modelName)
		mdl.Doc(field.Interface())

		switch fieldInfo.Relate {
		case HasOne:
			fieldID := fieldv.FieldByName("ID").Interface().(bson.ObjectId)
			foreignkeyField := fieldv.FieldByName(fieldInfo.Foreignkey)
			foreignkeyField.Set(reflect.ValueOf(dm.document.GetID()))
			bsonM := bson.M{}
			foreignkeyFieldt, _ := fieldt.Field(fieldInfo.Num).Type.Elem().FieldByName(fieldInfo.Foreignkey)
			tags := parseFieldTag(foreignkeyFieldt.Tag.Get("bson"))
			colName := strings.ToLower(foreignkeyFieldt.Name)
			if v, ok := tags["column"]; ok && v != "" {
				colName = v
			}

			bsonM[colName] = dm.document.GetID()

			if mdl.Count(bsonM) > 0 {
				field.FieldByName("Upsert").SetBool(true)
				mdl.Upsert(bsonM, field.Interface())
			} else if len(fieldID) == 0 {
				field.FieldByName("Upsert").SetBool(false)
				mdl.Insert(field.Interface())
			} else {
				field.FieldByName("Upsert").SetBool(true)
				mdl.UpsertID(fieldID, field.Interface())
			}
		case BelongTo:

			foreignkeyField := docv.FieldByName(fieldInfo.Foreignkey)
			id := foreignkeyField.Interface().(bson.ObjectId)
			bsonM := bson.M{
				"_id": foreignkeyField.Interface(),
			}
			if mdl.Count(bsonM) > 0 {
				field.FieldByName("Upsert").SetBool(true)
				mdl.Upsert(bsonM, field.Interface())
			} else if len(id) == 0 {
				if field.Kind() == reflect.Ptr {
					field.Elem().FieldByName("Upsert").SetBool(false)
				} else {
					field.FieldByName("Upsert").SetBool(false)
				}

				mdl.Insert(field.Interface())
				doc := field.Interface().(Documenter)
				// docv.FieldByName()
				foreignkeyField.Set(reflect.ValueOf(doc.GetID()))
			} else {
				field.FieldByName("Upsert").SetBool(true)
				mdl.UpsertID(id, field.Interface())
			}
		default:
		}
	}
	return
}

func (doc *Document) dbCollection() (*mgo.Collection, func()) {
	dm := doc.dm
	config := dm.connection.GetConfig()

	// TODO Implemented validate document
	session := dm.connection.CloneSession()

	collection := session.DB(config.DBName).C(dm.collection.Name)

	return collection, func() {
		session.Close()
	}
}

func (doc *Document) insert(docs ...interface{}) error {
	dm := doc.dm
	config := dm.connection.GetConfig()

	// TODO Implemented validate document
	session := dm.connection.CloneSession()
	defer session.Close()

	collection := session.DB(config.DBName).C(dm.collection.Name)
	// TODO 处理关联关系的插入
	doc.executeRelate(doc.getDocumentManager())

	return collection.Insert(docs...)
}

func (doc *Document) upsertID(id interface{}, docs interface{}) (*mgo.ChangeInfo, error) {
	collection, close := doc.dbCollection()

	defer close()
	doc.executeRelate(doc.getDocumentManager())
	return collection.UpsertId(id, docs)
}

func (doc *Document) Save() error {
	dm := doc.dm
	if dm.isNil() {
		panic("[Monger] Please init the document")
	}
	now := time.Now()

	var err error

	// doc.executeRelate(&dm)

	// 检测 ID 是否已经设置，如果未设置判定为插入, 否则判定为更新。
	if len(dm.document.GetID()) == 0 {
		dm.document.SetUpdatedAt(now)
		dm.document.SetCreatedAt(now)

		dm.document.SetID(bson.NewObjectId())

		err = doc.insert(dm.document)

		if err != nil {
			if mgo.IsDup(err) {
				err = &DuplicateDocumentError{NewError("Duplicate Key")}
			}
		}
	} else {
		dm.document.SetUpdatedAt(now)
		_, erro := doc.upsertID(dm.document.GetID(), dm.document)

		if erro != nil {
			if mgo.IsDup(erro) {
				err = &DuplicateDocumentError{NewError("Duplicate Key")}
			} else {
				err = erro
			}
		}
	}

	return nil
}

func (doc *Document) Validate(...interface{}) (bool, []error) {
	return false, nil
}

func (doc *Document) DefaultValidate() (bool, []error) {
	return false, nil
}

func (doc *Document) GetBSON() (interface{}, error) {
	if doc.dm.isNil() {
		panic("[Monger] Please init the document")
	}
	dm := doc.getDocumentManager()
	rawMap := make(map[string]interface{})
	// typeOfDocument := reflect.TypeOf(dm.document)
	// valueOfDocument := reflect.ValueOf(dm.document)
	// docElem := typeOfDocument.Elem()
	documentStructValue := reflect.ValueOf(dm.document)
	structv := documentStructValue
	structInfo, err := getDocumentStructInfo(reflect.TypeOf(dm.document))

	if err != nil {
		return nil, err
	}

	for _, v := range structInfo.FieldsList {
		var val reflect.Value
		if documentStructValue.Type().Kind() == reflect.Ptr {
			structv = documentStructValue.Elem()
		}

		if v.Inline == nil {
			val = structv.Field(v.Num)
		} else {
			val = structv.FieldByIndex(v.Inline)
		}

		if v.Relate == "" && !isZero(val) {
			rawMap[v.Key] = val.Interface()
		}
	}

	if doc.Upsert {
		bm := bson.M{
			"$set": rawMap,
		}
		doc.Upsert = false
		return bm, nil
	}

	return rawMap, nil
}
