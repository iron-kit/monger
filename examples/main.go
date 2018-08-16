package main

import (
	"github.com/iron-kit/monger"
	"gopkg.in/mgo.v2/bson"
	// "gopkg.in/mgo.v2/bson"
)

type User struct {
	monger.Document `json:",inline" bson:",inline"`

	Username string `json:"username,omitempty" bson:"username"`
	Password string `json:"password,omitempty" bson:"password"`
	Mobile   string `json:"mobile,omitempty" bson:"mobile"`
	Email    string `json:"email,omitempty" bson:"email"`

	Other struct {
		Name string
		Fire bool
	} `bson:"other"`

	Profile *Profile `json:"profile,omitempty" bson:"profile" monger:"hasOne;foreignkey:UserID"`
	// Messages []Message `json:"messages,omitempty" monger:"relate:hasMany;"`
}

type Profile struct {
	monger.Document `json:",inline" bson:",inline"`
	UserID          bson.ObjectId `json:"-" bson:"user_id"`
	User            *User         `json:",inline" bson:"user" monger:"belongTo;foreignkey:UserID"`
	Name            string
}

type Message struct {
	monger.Document `json:",inline" bson:",inline" monger:"inline"`
	UserID          bson.ObjectId `json:"-" bson:"user_id"`
	Body            string        `json:"body"`
}

// 设置 CollectionName
func (user *User) CollectionName() string {
	return "User"
}

func (p *Profile) CollectionName() string {
	return "Profile"
}

func main() {
	connection, err := monger.Connect(
		monger.DBName("unite"),
		monger.Hosts([]string{
			"localhost",
		}),
	)

	if err != nil {
		panic(err.Error())
	}
	connection.BatchRegister(
		&User{},
		&Profile{},
	)

	ProfileModel := connection.M("Profile")

	profile := &Profile{}
	ProfileModel.Doc(profile)
	profile.Name = "周金顺"
	profile.User = &User{
		Username: "zjs",
		Password: "123456",
	}

	profile.Save()
	// register User model
	// UserModel := connection.M(&User{})
	// ProfileModel := connection.M(&Profile{})
	// u := &User{}
	// UserModel.FindByID(bson.ObjectIdHex("5b739c6016a44bd1ee8cd8a8")).One(u)

	// u.Mobile = "18627894264"
	// u.Other.Fire = true
	// u.Other.Name = "test4asdfasfd"
	// u.Profile = &Profile{
	// 	Name: "HelloWorld2asdfasfsf",
	// }
	// // UserModel.Upsert(bson.M{
	// // 	"_id": u.ID,
	// // }, u)

	// p := &Profile{
	// 	Name: "HelloWorldBetterhsdfasf",
	// }
	// ProfileModel.Doc(p)

	// // ProfileModel.FindOne(bson.M{"user_id": u.ID}).One(p)
	// p.Upsert = true
	// ProfileModel.Upsert(bson.M{
	// 	"user_id": u.ID,
	// }, p)
	// u.Save()

	// ProfileModel.Upsert(bson.M{
	// 	"user_id": u.ID,
	// }, bson.M{
	// 	"$set": bson.M{
	// 		"name": "Test",
	// 	},
	// })

	// // u.Set("ppp", "Test ppp")
	// // a := &B{
	// // 	// A.V:   map[string]interface{}{},
	// // 	// Hello: "World",
	// // 	A: A{
	// // 		V:     map[string]interface{}{},
	// // 		Hello: "World",
	// // 	},
	// // 	Test: "better",
	// // }

	// // a.Set("World", "Hello")
	// u.Save()
	// UserModel.Insert(u)
}
