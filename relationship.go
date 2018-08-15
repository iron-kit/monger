package monger

import (
	"reflect"
)

const (
	HasOne   string = "HAS_ONE"
	HasMany  string = "HAS_MANY"
	BelongTo string = "BELONG_TO"
)

type RelationShip struct {
	// 虚拟字段名称
	fieldName string
	// 关联对象的值
	value interface{}

	relateKind string
	ref        reflect.Type
	refValue   reflect.Value

	// refDocument
	// // 关联结构
	// documentType reflect.Type
}

func newRelationShipOptions(rs *RelationShip, opts ...RelationShipOption) {
	for _, v := range opts {
		v(rs)
	}

	if rs.ref == nil {
		panic("[Monger] Please set RelateRef")
	}

	if rs.fieldName == "" {
		panic("[Monger] Please set RelateFieldName")
	}
}

func (ref *RelationShip) GetValue() interface{} {
	return ref.value
}

type RelationShipOption func(*RelationShip)

func RelateFieldName(name string) RelationShipOption {
	return func(r *RelationShip) {
		r.fieldName = name
	}
}

func RelateValue(val interface{}) RelationShipOption {
	return func(r *RelationShip) {
		r.value = val
	}
}

func RelateRef(doc Documenter) RelationShipOption {
	return func(r *RelationShip) {
		r.ref = reflect.TypeOf(doc)
		r.refValue = reflect.ValueOf(doc)
	}
}
