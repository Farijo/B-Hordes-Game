package dto

import (
	"database/sql"
	"time"
)

type Milestone struct {
	User       int            `db:"user"`
	Dt         time.Time      `db:"dt"`
	IsGhost    sql.NullInt64  `db:"isGhost"`
	PlayedMaps sql.NullInt64  `db:"playedMaps"`
	Rewards    sql.NullString `db:"rewards"`
	Dead       sql.NullInt64  `db:"dead"`
	Ban        sql.NullInt64  `db:"ban"`
	BaseDef    sql.NullInt64  `db:"baseDef"`
	X          sql.NullInt64  `db:"x"`
	Y          sql.NullInt64  `db:"y"`
	Job        sql.NullInt64  `db:"job"`
	MapWID     sql.NullInt64  `db:"mapWid"`
	MapHei     sql.NullInt64  `db:"mapHei"`
	MapDays    sql.NullInt64  `db:"mapDays"`
	Conspiracy sql.NullInt64  `db:"conspiracy"`
	Custom     sql.NullInt64  `db:"custom"`
}
