package dto

type Success struct {
	User         int    `db:"user"`
	Goal         int    `db:"goal"`
	Accomplished string `db:"accomplished"`
	Amount       int    `db:"amount"`
}
