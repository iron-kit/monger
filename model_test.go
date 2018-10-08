package monger

import (
	"fmt"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"github.com/stretchr/testify/assert"
)

type Member struct {
	Schema   `json:",inline" bson:",inline"`
	Username string   `json:"username,omitempty" bson:"username,omitempty"`
	Password string   `json:"password,omitempty" bson:"password,omitempty"`
	Profile  *Profile `json:"profile,omitempty" bson:"profile,omitempty" monger:"hasOne,foreignKey=user_id"`
}

type Profile struct {
	Schema   `json:",inline" bson:",inline"`
	Avatar   string        `json:"avatar,omitempty" bson:"avatar,omitempty"`
	Nickname string        `json:"nickname,omitempty" bson:"nickname,omitempty"`
	UserID   bson.ObjectId `json:"user_id,omitempty" bson:"user_id,omitempty"`
}

// func TestModelRegisterOK() {

// }

func TestModelRegsiterOK(t *testing.T) {
	conn, _ := Connect(
		Hosts([]string{"localhost"}),
		DBName("monger_test"),
	)

	conn.BatchRegister(
		new(Member),
		new(Profile),
	)

	MemberModel := conn.M("Member")
	ProfileModel := conn.M("Profile")

	assert.NotNil(t, MemberModel)
	assert.NotNil(t, ProfileModel)
}

func TestCreateFuncOK(t *testing.T) {
	connect := conn()
	MemberModel := connect.M("Member")
	ProfileModel := connect.M("Profile")

	member := &Member{
		Username: "alixezz",
		Password: "123456",
	}

	err := MemberModel.Create(member)

	profile := &Profile{
		Avatar:   "Hello",
		Nickname: "nova",
		UserID:   member.ID,
	}

	err2 := ProfileModel.Create(profile)

	assert.NoError(t, err)
	assert.NoError(t, err2)
}

func TestPopulateFuncOK(t *testing.T) {
	connect := conn()
	MemberModel := connect.M("Member")
	// ProfileModel := connect.M("Profile")
	member := new(Member)
	query := MemberModel.Where(bson.M{
		"_id": bson.ObjectIdHex("5bb86b3c16a44b4c69e667f9"),
	}).Populate("Profile")

	err := query.FindOne(member)

	fmt.Println(member)

	if err != nil {
		assert.Error(t, err)
	}

	assert.Equal(t, member.ID.Hex(), "5bb86b3c16a44b4c69e667f9")
}

func conn() Connection {
	conn, _ := Connect(
		Hosts([]string{"localhost"}),
		DBName("monger_test"),
	)

	conn.BatchRegister(
		new(Member),
		new(Profile),
	)

	return conn
}
