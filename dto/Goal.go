package dto

import "database/sql"

type Goal struct {
	ID        int            `db:"id"`
	Challenge int            `db:"challenge"`
	Typ       byte           `db:"typ"`
	Entity    uint16         `db:"entity"`
	Amount    sql.NullInt32  `db:"amount"`
	X         sql.NullByte   `db:"x"`
	Y         sql.NullByte   `db:"y"`
	Custom    sql.NullString `db:"custom"`
}
