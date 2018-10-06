# monger

A MongoDB ODM For Golang

## Todos

* Many To Many Relate
* Deep Populate

... More idea is thinking

## Features

* Schema define
* Use tag define ralation
* Populate query

## Usage

### Add monger to your project

```text
go get -u github.com/iron-kit/monger
```

### Import

```golang
import "github.com/iron-kit/monger"
```

### Create Database Connection

```golang
connection, err := monger.Connect(
  monger.DBName('your_database_name'),
  monger.Hosts([]string{
    "127.0.0.1",
    "196.1.1.2",
  }),
  monger.User("your_mongodb_user"),
  monger.Password("your_mongodb_password"),
)

if err != nil {
  panic(fmt.Sprintln("Database connection error: %v", err))
}

```

### Define A Schema

```golang

type Member struct {
  monger.Schema `json:",inline" bson:",inline"`
  
  Name string `json:"name,omitempty" bson:"name,omitempty"`
  Email string `json:"email,omitempty" bson:"email,omitempty"`
  Password string `json:"-" bson:"password"`

}

// Default collection of this schema is member
// you can custom use hook function
func (m *Member) GetSchemaName() string {

  // this schema's collection is Member
  return "Memebr"
}

```

### HasOne RelationShip

```golang

type Member struct {
  monger.Schema `json:",inline" bson:",inline"`
  
  Mobile string `json:"mobile,omitempty" bson:"mobile,omitempty"`
  Email string `json:"email,omitempty" bson:"email,omitempty"`
  Password string `json:"-" bson:"password"`

  Profile *Profile `json:"profile,omitempty" bson:"profile" monger:"hasOne,foreignKey=member_id"`
}

type Profile struct {
  monger.Schema `json:",inline" bson:",inline"`

  FirstName string `json:"firstname,omitempty" bson:"firstname,omitempty"`
  LastName string `json:"lastname,omitempty" bson:"lastname,omitempty"`
  MemberID string `json:"-" bson:"member_id"`
}
```

### BelongTo RelationShip

```golang

type Member struct {
  monger.Document    `json:",inline" bson:",inline"`
  
  Mobile   string    `json:"mobile,omitempty" bson:"mobile,omitempty"`
  Email    string    `json:"email,omitempty" bson:"email,omitempty"`
  Password string    `json:"-" bson:"password"`
  Profile  *Profile  `json:"profile,omitempty" bson:"profile" monger:"hasOne;foreignkey:MemberID"`
}

type Profile struct {
  monger.Document   `json:",inline" bson:",inline"`

  FirstName string  `json:"firstname,omitempty" bson:"firstname,omitempty"`
  LastName  string  `json:"lastname,omitempty" bson:"lastname,omitempty"`
  MemberID  string  `json:"-" bson:"member_id"`
  Member *Member `json:"member,omitempty" bson:"member,omitempty" monger:"belongTo,foreignKey=member_id"`
}
```

### Use Model

```golang

// register a model
connection.M(&Member{})

// batch register
connection.BatchRegister(
  &Member{},
  &Profile{},
)

// get a model
MemberModel := connection.M("Member")
ProfileModel := connection.M("Profile")

// initial a document

member := &Member{
  Mobile: "13423453456",
}
member.Password = "123456"

// update a new document
member.ID = bson.ObjectIdHex("id")
member.Password = "1234567"

MemberModel.Create(member)

```

## Thanks

> 这里一些接口设计的方法，一些编程哲学是通过学习其他项目领悟到的。在此感谢下面的项目。

* [go-micro](https://github.com/micro/go-micro) golang 的微服务框架
* [mongodm](https://github.com/zebresel-com/mongodm) 同样也是一个mongodb的ODM库
* [mongoose](http://mongoosejs.com/) mongodb odm library of nodejs