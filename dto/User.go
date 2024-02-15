package dto

import (
	"database/sql"
)

type User struct {
	ID             int            `db:"id"`
	Name           sql.NullString `db:"name"`
	SimplifiedName sql.NullString `db:"simplified_name"`
	Avatar         sql.NullString `db:"avatar"`
}
