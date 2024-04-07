package dto

import "database/sql"

type Challenge struct {
	ID        int            `db:"id"`
	Name      string         `db:"name"`
	Creator   User           ``
	Flags     byte           `db:"flags"`
	StartDate sql.NullString `db:"start_date"`
	EndDate   sql.NullString `db:"end_date"`
}

type DetailedChallenge struct {
	Challenge
	ParticipantCount int
	Access           int8
	Private          bool
	API              bool
	Status           int8
	Role             string
}

func (challenge *DetailedChallenge) UpdateDetailedProperties(started, ended bool) {
	challenge.Access = int8(challenge.Flags & 0x03)
	challenge.Private = challenge.Flags&0x04 == 0x04
	challenge.API = challenge.Flags&0x08 == 0
	challenge.Status = int8((challenge.Flags & 0x30) >> 4)
	if ended {
		challenge.Status += 2
	} else if started {
		challenge.Status += 1
	}
}
