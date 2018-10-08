package monger

import "reflect"

const (
	HasOne    string = "HAS_ONE"
	HasMany   string = "HAS_MANY"
	BelongTo  string = "BELONG_TO"
	BelongsTo string = "BELONGS_TO"
)

type Relationship struct {
	Kind string
	// RelationModel   Model
	// ModelName       string
	// From            string
	// Schema
	RelationType    reflect.Type
	From            string
	As              string
	CollectionName  string
	LocalFieldKey   string
	ForeignFieldKey string

	// // 无用的
	// ForeignCollectionNames            []string
	// ForeignFieldNames                 []string
	// AssociationForeignFieldNames      []string
	// AssociationForeignCollectionNames []string
}
