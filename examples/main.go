package main

import (
	"encoding/json"
	"fmt"
	"github.com/iron-kit/monger"
	"gopkg.in/mgo.v2/bson"
	// "gopkg.in/mgo.v2/bson"
)

type Orders struct {
	monger.Document `json:",inline" bson:",inline"`

	OrderName  string      `json:"order_name" bson:"order_name"`
	Price      int         `json:"price" bson:"price"`
	OrderItems []OrderItem `json:"items" bson:"order_items" monger:"hasMany,foreignKey=order_id"`
}

type OrderItem struct {
	monger.Document `json:",inline" bson:",inline"`

	ItemName string        `json:"item_name" bson:"item_name"`
	OrderID  bson.ObjectId `json:"-" bson:"order_id"`
}

func main() {
	connection, _ := monger.Connect(
		monger.DBName("test"),
		monger.Hosts([]string{
			"localhost",
		}),
	)

	connection.BatchRegister(
		&Orders{},
		&OrderItem{},
	)

	Order := connection.M("Orders")
	order := &Orders{}
	Order.
		FindOne(bson.M{"_id": bson.ObjectIdHex("5b7beebf16a44b2dbb0bd78d")}).
		Populate("OrderItems").
		Exec(order)
	d, _ := json.Marshal(order)
	fmt.Println(string(d))
	// OrderItemModel := connection.M("OrderItem")
	// orders := []Orders{}

	// for i := 0; i < 10; i++ {
	// 	o := Orders{
	// 		OrderName: fmt.Sprintf("订单: %d", i),
	// 		Price:     i,
	// 	}
	// 	Order.Create(&o)

	// 	oi := OrderItem{
	// 		ItemName: "项目名",
	// 		OrderID:  o.ID,
	// 	}
	// 	err := OrderItemModel.Create(&oi)
	// 	fmt.Println(err)
	// }

}
