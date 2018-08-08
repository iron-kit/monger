package monger

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"reflect"
	"strings"
	"time"
)

/*
Connection is the connect manager of mgo.v2 deiver
*/
type Connection interface {
	M(interface{}) Model
	Open() error
	Close()
	CloneSession() *mgo.Session
	getModel(name string) Model
	registerSchemaAndGetModel(schemas Schema) Model
	GetConfig() *Config
}

type connection struct {
	Config     *Config
	Session    *mgo.Session
	modelStore map[string]Model
}

func newConnection(config *Config, session *mgo.Session) Connection {

	return &connection{
		Config:     config,
		Session:    session,
		modelStore: make(map[string]Model),
	}
}

func (conn *connection) GetConfig() *Config {
	return conn.Config
}

func (conn *connection) CloneSession() *mgo.Session {
	return conn.Session.Clone()
}

// Open a database connection
func (conn *connection) Open() error {
	var dialInfo *mgo.DialInfo
	if conn.Config.DialInfo == nil {
		dialInfo = &mgo.DialInfo{
			Addrs:    conn.Config.Hosts,
			Timeout:  3 * time.Second,
			Database: conn.Config.DBName,
			Username: conn.Config.User,
			Password: conn.Config.Password,
			// PoolLimit: conn.Config.PoolLimit,
		}

		// 如果设置了用户设置了连接池，就使用用户设置的连接池，否则使用驱动的缺省值(4096)
		if conn.Config.PoolLimit > 0 {
			dialInfo.PoolLimit = conn.Config.PoolLimit
		}
	} else {
		dialInfo = conn.Config.DialInfo
	}

	session, err := mgo.DialWithInfo(dialInfo)

	if err != nil {
		return err
	}
	session.SetMode(mgo.Monotonic, true)
	conn.Session = session
	return nil
}

// Close a database session
func (conn *connection) Close() {
	if conn.Session != nil {
		conn.Session.Close()
	}
}

func (conn *connection) getModel(name string) Model {
	nameLower := strings.ToLower(name)
	if _, ok := conn.modelStore[nameLower]; ok {
		return conn.modelStore[nameLower]
	}

	panic(fmt.Sprintf("[monger] Schema '%v' is not registered ", nameLower))
}

func (conn *connection) registerSchemaAndGetModel(schema Schema) Model {
	if schema == nil {
		panic("[monger] schema can not be nil")
	}
	collectionName := schema.GetName()
	reflectType := reflect.TypeOf(schema)
	typeName := strings.ToLower(reflectType.Elem().Name())

	if _, ok := conn.modelStore[typeName]; !ok {
		collection := conn.Session.DB("").C(collectionName)
		model := &model{
			collection,
			conn,
		}

		conn.modelStore[typeName] = model
		fmt.Printf("[monger] Type '%v' has registered", typeName)
		return model
	}

	fmt.Printf("[monger] Tried to register type '%v' twice", typeName)
	return conn.modelStore[typeName]
}

func (conn *connection) M(args interface{}) Model {
	if name, ok := args.(string); ok {
		return conn.getModel(name)
	}

	if schema, ok := args.(Schema); ok {
		return conn.registerSchemaAndGetModel(schema)
	}

	return nil
}
