package monger

import (
	"log"
)

type logger struct{}

func (l *logger) Output(calldepth int, s string) error {
	// fmt.Printf(s)

	return log.Output(calldepth, s)
}
