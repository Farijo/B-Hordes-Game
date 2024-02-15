package dto

import (
	"time"
)

type Success struct {
	User         int       `db:"user"`
	Goal         int       `db:"goal"`
	Accomplished time.Time `db:"accomplished"`
	Amount       int       `db:"amount"`
}
