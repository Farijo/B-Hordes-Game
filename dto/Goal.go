package dto

import (
	"database/sql"
)

type Goal struct {
	ID        int            `db:"id"`
	Challenge int            `db:"challenge"`
	Typ       int            `db:"typ"`
	Descript  sql.NullString `db:"descript"`
}
