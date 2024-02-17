package dto

import (
	"database/sql"
)

type Goal struct {
	ID        int            `db:"id"`
	Challenge int            `db:"challenge"`
	Typ       byte           `db:"typ"`
	Descript  sql.NullString `db:"descript"`
}
