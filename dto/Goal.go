package dto

import "database/sql"

type Goal struct {
	ID        int            `db:"id"`
	Challenge int            `db:"challenge"`
	Typ       byte           `db:"typ"`
	Entity    uint16         `db:"entity"`
	Amount    sql.NullInt32  `db:"amount"`
	X         sql.NullInt16  `db:"x"`
	Y         sql.NullInt16  `db:"y"`
	Custom    sql.NullString `db:"custom"`
}
