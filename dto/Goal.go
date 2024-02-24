package dto

type Goal struct {
	ID        int    `db:"id"`
	Challenge int    `db:"challenge"`
	Typ       byte   `db:"typ"`
	Descript  string `db:"descript"`
}
