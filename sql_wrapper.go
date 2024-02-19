package main

import (
	"bhordesgame/dto"
	"database/sql"
	"encoding/binary"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
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

	rows, err = dbConn().Query(`SELECT rewards, isGhost, playedMaps, dead, ban, baseDef, x, y, job, mapWid, mapHei, mapDays, conspiracy, custom FROM milestone WHERE user=? ORDER BY dt ASC`, milestone.User.ID)
	if err != nil {
		return err
	}
	var previousMS dto.Milestone
	for rows.Next() {
		if err = rows.Scan(
			&previousMS.Rewards,
			&previousMS.IsGhost,
			&previousMS.PlayedMaps,
			&previousMS.Dead,
			&previousMS.Ban,
			&previousMS.BaseDef,
			&previousMS.X,
			&previousMS.Y,
			&previousMS.Job,
			&previousMS.Map.Wid,
			&previousMS.Map.Hei,
			&previousMS.Map.Days,
			&previousMS.Map.Conspiracy,
			&previousMS.Map.Custom); err != nil {
			rows.Close()
			return err
		}
	}
	rows.Close()

	mustUpdate := milestone.InvalidateUnchangedFields(&previousMS)

	changements := make([]byte, 0, 120)
	milestone.Rewards.Valid = false
	for id, number := range milestone.Rewards.Pictos {
		if number != previousMS.Rewards.Pictos[id] {
			changements = binary.LittleEndian.AppendUint16(changements, id)
			changements = binary.LittleEndian.AppendUint32(changements, number)

			milestone.Rewards.Valid = true
		}
	}

	if milestone.Rewards.Valid {
		milestone.Rewards.String = string(changements)
	} else if !mustUpdate {
		// nothing has changed since last milestone
		return nil
	}

	if _, err = dbConn().Exec(`INSERT INTO milestone VALUES(?, NOW(2), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		milestone.User.ID,
		milestone.IsGhost,
		milestone.PlayedMaps,
		milestone.Rewards,
		milestone.Dead,
		milestone.Ban,
		milestone.BaseDef,
		milestone.X,
		milestone.Y,
		milestone.Job,
		milestone.Map.Wid,
		milestone.Map.Hei,
		milestone.Map.Days,
		milestone.Map.Conspiracy,
		milestone.Map.Custom); err != nil {
		return err
	}

	return nil
}

func queryUser(id int) (user dto.User, err error) {
	row := dbConn().QueryRow(`SELECT name, simplified_name, avatar FROM user WHERE id=?`, id)
	user.ID = id
	err = row.Scan(&user.Name, &user.SimplifiedName, &user.Avatar)
	return
}

func queryChallengesRelatedTo(userId int, viewer int, ch chan<- *dto.DetailedChallenge) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT user.name as cname, user.simplified_name, user.avatar
	, challenge.id, challenge.name, challenge.flags, challenge.start_date, challenge.end_date
	, COUNT(participant.user) AS participant_count
	, challenge.start_date <= NOW() AS started
	, challenge.end_date < NOW() AS ended
	, challenge.creator=? AS created
	, participant.user IS NOT NULL as participate
	, validator.user IS NOT NULL as validate
	, invitation.user IS NOT NULL as invited
	 FROM challenge
	 LEFT JOIN user        ON challenge.creator = user.id
	 LEFT JOIN participant ON challenge.id = participant.challenge AND participant.user = ?
	 LEFT JOIN validator   ON challenge.id = validator.challenge AND validator.user = ?
	 LEFT JOIN invitation  ON challenge.id = invitation.challenge AND invitation.user = ?
	 WHERE ? IN (challenge.creator, participant.user, validator.user, invitation.user)
	 AND (challenge.flags & 0x04 = 0 OR ? in (challenge.creator, participant.user, validator.user))
	 AND (?=? OR ((challenge.flags & 0x30) >> 4) = 2 AND challenge.flags & 0x03 < 2)
	 GROUP BY challenge.id, challenge.name, challenge.end_date, created, participate, validate, invited
	 ORDER BY ended, started, challenge.flags & 0x30`, userId, userId, userId, userId, userId, viewer, userId, viewer)
	if err != nil {
		panic(err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var detailedChall dto.DetailedChallenge
		var Started sql.NullBool
		var Ended sql.NullBool
		var created, participate, validate, invited bool
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
			&Ended,
			&created,
			&participate,
			&validate,
			&invited); err != nil {
			panic(err.Error())
		}
		detailedChall.UpdateDetailedProperties(Started.Bool, Ended.Bool)

		tmp := []string{}
		if created {
			tmp = append(tmp, "Créateur")
		}
		if participate {
			tmp = append(tmp, "Participant")
		} else if invited {
			if detailedChall.Access == 2 {
				tmp = append(tmp, "Invité")
			} else {
				tmp = append(tmp, "Postulant")
			}
		}
		if validate {
			tmp = append(tmp, "Approbateur")
		}

		detailedChall.Role = strings.Join(tmp, ", ")

		ch <- &detailedChall
	}
}
