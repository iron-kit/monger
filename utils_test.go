package monger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnakeStringPascal(t *testing.T) {
	s := "UserClubProfile"
	ss := snakeString(s)
	assert.Equal(t, ss, "user_club_profile")
}

func TestSnakeStringCamel(t *testing.T) {
	s := "userClubProfile"

	ss := snakeString(s)

	assert.Equal(t, ss, "user_club_profile")
}
