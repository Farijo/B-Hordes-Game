package main

import (
	"bhordesgame/dto"
	"database/sql"
)

var instance *sql.DB

func dbConn() (db *sql.DB) {
	if instance != nil {
		return instance
	}
	dbDriver := "mysql"
	dbUser := "tvallar"
	dbPass := ""
	dbName := "hordes_challenge"
	instance, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return instance
}

func queryPublicChallenges(ch chan<- dto.DetailedChallenge) {
	defer close(ch)
	db := dbConn()

	rows, err := db.Query(`SELECT user.name as cname, user.simplified_name, user.avatar
	, challenge.id, challenge.name, challenge.flags, challenge.start_date, challenge.end_date
	, COUNT(participant.user) AS participant_count
	, challenge.start_date <= NOW() AS started
	, challenge.end_date < NOW() AS ended
	 FROM challenge
	 LEFT JOIN user ON challenge.creator = user.id
	 LEFT JOIN participant ON challenge.id = participant.challenge
	 WHERE challenge.flags & 0x03 < 2
	 AND (challenge.flags & 0x04 = 0)
	 AND ((challenge.flags & 0x30) >> 4 = 2)
	 GROUP BY challenge.id, challenge.name, challenge.end_date
	 ORDER BY ended, started`)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var detailedChall dto.DetailedChallenge
		var Started sql.NullBool
		var Ended sql.NullBool
		if err := rows.Scan(
			&detailedChall.Creator.Name,
			&detailedChall.Creator.SimplifiedName,
			&detailedChall.Creator.Avatar,
			&detailedChall.ID,
			&detailedChall.Name,
			&detailedChall.Flags,
			&detailedChall.StartDate,
			&detailedChall.EndDate,
			&detailedChall.ParticipantCount,
			&Started,
			&Ended); err != nil {
			panic(err.Error())
		}
		detailedChall.Access = int8(detailedChall.Flags & 0x03)
		detailedChall.Private = detailedChall.Flags&0x04 == 0
		detailedChall.Status = int8((detailedChall.Flags & 0x30) >> 4)
		if Ended.Bool {
			detailedChall.Status += 2
		} else if Started.Bool {
			detailedChall.Status += 1
		}
		ch <- detailedChall
	}
}
