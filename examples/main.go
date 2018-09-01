package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/structs"
	"github.com/iron-kit/monger"
	"gopkg.in/mgo.v2/bson"
	// "gopkg.in/mgo.v2/bson"
)

type Orders struct {
	monger.Document `json:",inline" bson:",inline"`

	OrderName  string      `json:"order_name" bson:"order_name"`
	Price      int         `json:"price" bson:"price"`
	Details    *Details    `json:"details" bson:"details" monger:"hasOne,foreignKey=order_id"`
	OrderItems []OrderItem `json:"items" bson:"order_items" monger:"hasMany,foreignKey=order_id"`
}

type OrderItem struct {
	monger.Document `json:",inline" bson:",inline"`

	ItemName string        `json:"item_name" bson:"item_name"`
	OrderID  bson.ObjectId `json:"-" bson:"order_id"`
}

type Details struct {
	monger.Document `json:",inline" bson:",inline"`
	OrderID         bson.ObjectId `json:"-" bson:"order_id"`
	Name            string        `json:"name" bson:"name"`
	Test            string        `json:"test" bson:"test"`
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
		&Details{},
	)

	Order := connection.M("Orders")
	// orders := []Orders{}
	ord := &Orders{}
	// ord.ID = ""
	ord.OrderName = "你好"

	m := structs.Map(ord)
	fmt.Println("map")
	fmt.Println(m)
	// ord.OrderItems = []OrderItem{
	// 	{
	// 		ItemName: "你好",
	// 	},
	// }

	err := Order.Create(ord)
	fmt.Println(err)
	// orders.OrderName =
	// Order.
	// 	Find().
	// 	Populate("OrderItems", "Details").
	// 	Exec(&orders)

	// fmt.Println(order.Details.OrderID)
	d, _ := json.Marshal(ord)
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
