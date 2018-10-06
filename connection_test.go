package monger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectOK(t *testing.T) {
	_, err := Connect(
		DBName("monger_test"),
		Hosts([]string{
			"localhost",
		}),
	)

	if err != nil {
		assert.Error(t, err)
	}

	// conn.BatchRegister(
	// 	new(Member),
	// 	new(Profile),
	// )
}
