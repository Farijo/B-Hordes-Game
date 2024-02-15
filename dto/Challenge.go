package dto

import (
	"database/sql"
)

type Challenge struct {
	ID        int            `db:"id"`
	Name      string         `db:"name"`
	Creator   User           ``
	Flags     int            `db:"flags"`
	StartDate sql.NullString `db:"start_date"`
	EndDate   sql.NullString `db:"end_date"`
}

type DetailedChallenge struct {
	Challenge
	ParticipantCount int
	Access           int8
	Private          bool
	Status           int8
	Role             string
}
