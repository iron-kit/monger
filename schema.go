package monger

type Schema interface {
	GetName() string
	Document() Document
}

type schema struct {
	name     string
	document Document
}

func (self *schema) GetName() string {
	return self.name
}

func (self *schema) Document() Document {
	return self.document
}

func NewSchema(collectionName string, doc Document) Schema {
	return &schema{
		collectionName,
		doc,
	}
}
