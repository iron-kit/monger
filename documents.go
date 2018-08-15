package monger

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"time"
)

type DocumentHooker interface {
	BeforeSave()
}

type Documenter interface {
	// DocumentManager
	DocumentHooker
	bson.Getter
	bson.Setter
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
}

type documentManager struct {
	document    Documenter
	connection  Connection
	collection  *mgo.Collection
	documentMap map[string]interface{}
}

type Document struct {
	dm        documentManager
	ID        bson.ObjectId `json:"id" monger:"column:_id"`
	CreatedAt time.Time     `json:"createdAt" monger:"column:createdAt"`
	UpdatedAt time.Time     `json:"updatedAt" monger:"column:updatedAt"`
	Deleted   bool          `json:"-" monger:"column:deleted"`
}

func D(doc Documenter, collection *mgo.Collection, connection Connection) Documenter {
	dm := doc.getDocumentManager()
	dm.SetCollection(collection)
	dm.SetConnection(connection)
	dm.SetDocument(doc)
	return doc
}

func initDocuments(documents interface{}, collection *mgo.Collection, connection Connection, bind bool) {
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
				dm = doc.getDocumentManager()
				dm.SetCollection(collection)
				dm.SetConnection(connection)
				dm.SetDocument(doc)

			} else if ele.Type().Kind() == reflect.Struct {
				if doc, ok := ele.Addr().Interface().(Documenter); ok {
					dm = doc.getDocumentManager()
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
		dm = d.getDocumentManager()
		dm.SetCollection(collection)
		dm.SetConnection(connection)
		dm.SetDocument(d)
		// dm.bindDocData()
	} else {
		panic("[Monger] The first param must be Document")
	}
	if bind {
		dm.bindDocData()
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

func (dm *documentManager) bindDocData() {
	// TODO init the document data
	document := dm.document

	doct := reflect.TypeOf(document)
	docv := reflect.ValueOf(document)
	dovk := doct.Kind()

	if dovk == reflect.Ptr {
		docv = docv.Elem()
	}

	structInfo, err := getDocumentStructInfo(doct)

	if err != nil {
		panic(err)
	}

	for _, info := range structInfo.FieldsList {
		// 填充数据
		var field reflect.Value
		// if info.Relate == "" && info.Relate
		if info.Inline == nil {
			field = docv.Field(info.Num)
		} else {
			field = docv.FieldByIndex(info.Inline)
			// continue
		}

		// field.Set()
		if v, ok := dm.documentMap[info.Key]; ok {
			field.Set(reflect.ValueOf(v))
		} else {
			field.Set(reflect.New(field.Type()).Elem())
		}
		// docv.Field()
	}
	return
}

func (dm *documentManager) setDocumentMap(data map[string]interface{}) {
	dm.documentMap = data
}

func (dm *documentManager) SetDocument(document Documenter) {
	dm.document = document
}

func (doc *Document) getDocumentManager() *documentManager {
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

func (doc *Document) executeRelate(dm *documentManager) {

	// hook the relationship function
	// dm.document.RelationShip()
	return

	// for name, rs := range doc.refs {
	// 	rs.
	// }
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
	//
	return collection.Insert(docs...)
}

func (doc *Document) upsertID(id interface{}, docs interface{}) (*mgo.ChangeInfo, error) {
	collection, close := doc.dbCollection()

	defer close()

	return collection.UpsertId(id, docs)
}

func (doc *Document) Save() error {
	dm := doc.dm
	if dm.isNil() {
		panic("[monger] Please init the document")
	}
	now := time.Now()

	var err error

	doc.executeRelate(&dm)

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

		// if v.InlineMap >= 0 {
		// 	m := v.Field(sinfo.InlineMap)
		// 	if m.Len() > 0 {
		// 		for _, k := range m.MapKeys() {
		// 			ks := k.String()
		// 			if _, found := sinfo.FieldsMap[ks]; found {
		// 				panic(fmt.Sprintf("Can't have key %q in inlined map; conflicts with struct field", ks))
		// 			}
		// 			e.addElem(ks, m.MapIndex(k), false)
		// 		}
		// 	}
		// }
		if v.Relate == "" {
			rawMap[v.Key] = val.Interface()
		}
	}

	fmt.Println(rawMap, "Hello")

	return rawMap, nil
}

func (doc *Document) SetBSON(raw bson.Raw) error {
	// panic("not implemented")
	dataMap := make(map[string]interface{})

	raw.Unmarshal(dataMap)
	doc.dm.setDocumentMap(dataMap)

	return nil
}
