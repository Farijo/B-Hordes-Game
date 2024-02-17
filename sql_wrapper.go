package main

import (
	"bhordesgame/dto"
	"database/sql"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var instance *sql.DB

func dbConn() (db *sql.DB) {
	if instance == nil {
		var err error

		dbDriver := "mysql"
		dbUser := "tvallar"
		dbPass := ""
		dbName := "hordes_challenge"
		instance, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
		if err != nil {
			panic(err.Error())
		}
	}
	return instance
}

var trsf *transform.Transformer

func transformer() *transform.Transformer {
	if trsf == nil {
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		trsf = &t
	}
	return trsf
}

func queryPublicChallenges(ch chan<- *dto.DetailedChallenge) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT user.name as cname, user.simplified_name, user.avatar
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
		detailedChall.UpdateDetailedProperties(Started.Bool, Ended.Bool)
		ch <- &detailedChall
	}
}

func queryChallenge(id int) (challenge dto.DetailedChallenge, err error) {
	var untilStart, untilEnd sql.NullString
	row := dbConn().QueryRow(`SELECT name, creator, flags, start_date, end_date
	, TIMEDIFF(start_date,NOW()) as rem_start, TIMEDIFF(end_date,NOW()) as rem_end
	 FROM challenge
	 WHERE id=?`, id)

	if err = row.Scan(&challenge.Name, &challenge.Creator.ID, &challenge.Flags, &challenge.StartDate, &challenge.EndDate, &untilStart, &untilEnd); err != nil {
		return
	}
	challenge.UpdateDetailedProperties(untilStart.Valid && untilStart.String[0] == '-', untilEnd.Valid && untilEnd.String[0] == '-')

	return
}

func insertUser(user *dto.User) error {
	simplified, _, err := transform.String(*transformer(), user.Name)
	user.SimplifiedName = strings.ToLower(simplified)
	if err != nil {
		return err
	}
	_, err = dbConn().Exec(`INSERT INTO user (id, name, simplified_name, avatar) VALUES (?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE name=?, simplified_name=?, avatar=?`,
		user.ID, user.Name, user.SimplifiedName, user.Avatar, user.Name, user.SimplifiedName, user.Avatar)

	return err
}

func insertMilestone(milestone *dto.Milestone) error {
	rows, err := dbConn().Query(`SELECT typ,descript,goal.id
	FROM goal
	JOIN challenge ON goal.challenge = challenge.id
	JOIN participant ON challenge.id = participant.challenge AND participant.user = ?
	WHERE challenge.start_date <= NOW()
	AND (NOW() < challenge.end_date OR challenge.end_date IS NULL)
	AND ((challenge.flags & 0x30) >> 4) = 2`, milestone.User.ID)
	if err != nil {
		return err
	}
	rowPresent := rows.Next()
	rows.Close()
	if !rowPresent {
		// not in a challenge, nothing to do
		return nil
	}
	return nil
}
