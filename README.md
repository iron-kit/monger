# monger

A MongoDB ODM For Golang

## Todos

* 实现 Schema 的 One to One 关联模型
* 实现 Schema 的 One to Many 关联模型
* 实现 Schema 的 Many to Many 关联模型
* 实现 Model 更多的查询方法
* 实现 Populate

... 更多想法思考中

## 特性

* 像 mongoose 一样的 Schema 定义
* 像 mongoose 一样的模型查询
* 支持 Populate
* 总之就是想让你在golang里面用到 mongoose （因为mongoose确实做得很好）

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

### Create A Schema

```golang

type Member struct {
  monger.BaseDocument `json:",inline" bson:",inline"`
  
  Name string `json:"name,omitempty" bson:"name, omitempty"`
  Password string `json:"-" bson:"password"`
}

MemberSchema := monger.NewSchema("collection_name", &Member{})

```

### Create A Model

```golang
connection.M(MemberSchema)
```

### Get A Model

```golane
MemberModel := connection.M("Member")

// Create member document
member := MemberModel.Create(&Member{})

// save to db
member.Save()

```

## 感谢

> 这里一些接口设计的方法，一些编程哲学是通过学习其他项目领悟到的。在此感谢下面的项目。

* [go-micro](https://github.com/micro/go-micro) golang 的微服务框架
* [mongodm](https://github.com/zebresel-com/mongodm) 同样也是一个mongodb的ODM库
* [mongoose](http://mongoosejs.com/)