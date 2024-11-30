package main

import (
	"bhordesgame/dto"
	"database/sql"
	"errors"
	"os"
	"sort"
	"strconv"
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
		dbHostName := os.Getenv("MYSQL_HOSTNAME")
		dbUser := os.Getenv("MYSQL_USER")
		dbPass := os.Getenv("MYSQL_PWD")
		dbName := os.Getenv("MYSQL_DB")
		instance, err = sql.Open(dbDriver, dbUser+":"+dbPass+"@tcp("+dbHostName+")/"+dbName)
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

	rows, err := dbConn().Query(`SELECT user.ID, user.name as cname, user.simplified_name, user.avatar
	, challenge.id, challenge.name, challenge.flags, challenge.start_date, challenge.end_date
	, COUNT(participant.user) AS participant_count
	, challenge.start_date <= UTC_TIMESTAMP() AS started
	, challenge.end_date < UTC_TIMESTAMP() AS ended
	 FROM challenge
	 LEFT JOIN user ON challenge.creator = user.id
	 LEFT JOIN participant ON challenge.id = participant.challenge
	 WHERE challenge.flags & 0x03 < 2
	 AND (challenge.flags & 0x04) = 0
	 AND (challenge.flags & 0x30) = 0x20
	 GROUP BY challenge.id, challenge.name, challenge.end_date
	 ORDER BY IFNULL(ended, 0), IFNULL(started,0)`)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var detailedChall dto.DetailedChallenge
		var Started sql.NullBool
		var Ended sql.NullBool
		if err := rows.Scan(
			&detailedChall.Creator.ID,
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
			logger.Println(err)
			return
		}
		detailedChall.UpdateDetailedProperties(Started.Bool, Ended.Bool)
		ch <- &detailedChall
	}
}

func queryChallenge(id, requestor int) (challenge dto.DetailedChallenge, err error) {
	var untilStart, untilEnd sql.NullString
	row := dbConn().QueryRow(`SELECT name, creator, flags, start_date, end_date
	, TIMEDIFF(start_date,UTC_TIMESTAMP()) as rem_start, TIMEDIFF(end_date,UTC_TIMESTAMP()) as rem_end
	 FROM challenge
	 WHERE id=?
	 AND (flags & 0x04 = 0 AND flags & 0x30 = 0x20
	   OR creator = ?
	   OR EXISTS (SELECT 1 FROM participant WHERE user = ? AND challenge = ?)
	   OR EXISTS (SELECT 1 FROM validator WHERE user = ? AND challenge = ?)
	   OR EXISTS (SELECT 1 FROM invitation WHERE user = ? AND challenge = ?)
	)`, id, requestor, requestor, id, requestor, id, requestor, id)

	if err = row.Scan(&challenge.Name, &challenge.Creator.ID, &challenge.Flags, &challenge.StartDate, &challenge.EndDate, &untilStart, &untilEnd); err != nil {
		return
	}
	challenge.ID = id
	challenge.UpdateDetailedProperties(untilStart.Valid && (untilStart.String[0] == '-' || untilStart.String == "00:00:00"), untilEnd.Valid && (untilEnd.String[0] == '-' || untilEnd.String == "00:00:00"))

	return
}

func insertUser(user *dto.User) error {
	return insertMultipleUsers([]dto.User{*user})
}
func insertMultipleUsers(user []dto.User) error {
	if len(user) < 1 {
		return nil
	}

	values := make([]any, 0)
	sqlValues := ""

	for _, u := range user {
		simplified, _, err := transform.String(*transformer(), u.Name)
		if err != nil {
			return err
		}
		u.SimplifiedName = strings.ToLower(simplified)
		values = append(values, u.ID, u.Name, u.SimplifiedName, u.Avatar)
		sqlValues += ",(?, ?, ?, ?)"
	}
	_, err := dbConn().Exec(`INSERT INTO user (id, name, simplified_name, avatar) VALUES `+sqlValues[1:]+
		`ON DUPLICATE KEY UPDATE name = VALUES(name), simplified_name = VALUES(simplified_name), avatar = VALUES(avatar)`, values...)

	return err
}

func queryAllUsers(ch chan<- *dto.User) {
	defer close(ch)

	rows, err := dbConn().Query("SELECT id, name, avatar FROM user")
	if err != nil {
		logger.Println(err)
		return
	}
	for rows.Next() {
		var user dto.User
		err = rows.Scan(&user.ID, &user.Name, &user.Avatar)
		if err != nil {
			logger.Println(err)
			return
		}
		ch <- &user
	}
}

func insertMilestone(milestone *dto.Milestone) error {
	tx, err := dbConn().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`SELECT goal.id,typ,entity, challenge.flags & 0x08 = 0 as api, challenge.start_date
	FROM goal
	JOIN challenge ON goal.challenge = challenge.id
	JOIN participant ON challenge.id = participant.challenge AND participant.user = ?
	WHERE challenge.start_date <= UTC_TIMESTAMP()
	AND (UTC_TIMESTAMP() < challenge.end_date OR challenge.end_date IS NULL)
	AND (challenge.flags & 0x30) = 0x20`, milestone.User.ID)
	if err != nil {
		return err
	}
	successes := make([]dto.Success, 0)
	rowPresent := false
	lastStarted := ""
	for rows.Next() {
		var g dto.Goal
		var api bool
		var startDate sql.NullString
		if err = rows.Scan(&g.ID, &g.Typ, &g.Entity, &api, &startDate); err != nil {
			rows.Close()
			return err
		}
		rowPresent = true
		if startDate.Valid && startDate.String > lastStarted {
			lastStarted = startDate.String
		}
		if !api {
			continue
		}
		newSuccess := dto.Success{
			User:         milestone.User.ID,
			Goal:         g.ID,
			Accomplished: "",
			Amount:       0,
		}
		switch g.Typ {
		case 0: // picto
			newSuccess.Amount = milestone.Rewards.Data[g.Entity]
			successes = append(successes, newSuccess)
		case 2: // construire
			if milestone.Map.City.Buildings.Data[g.Entity] {
				newSuccess.Amount = 1
				successes = append(successes, newSuccess)
			}
		case 3: // en banque
			newSuccess.Amount = milestone.Map.City.Bank.Data[g.Entity]
			if newSuccess.Amount > 0 {
				successes = append(successes, newSuccess)
			}
		}
	}
	rows.Close()
	if !rowPresent {
		// not in a challenge, nothing to do
		// TODO delete last useless milestone (never delete milestone whose not the last)
		return nil
	}

	rows, err = tx.Query(`SELECT dt, isGhost, playedMaps, rewards, dead, isOut, ban, baseDef, x, y, job, mapWid, mapHei, mapDays, conspiracy, custom, buildings, bank, zoneItems
	FROM milestone WHERE user = ? ORDER BY dt ASC`, milestone.User.ID)
	if err != nil {
		return err
	}
	var previousMS dto.Milestone
	for rows.Next() {
		if err = rows.Scan(
			&previousMS.Dt,
			&previousMS.IsGhost,
			&previousMS.PlayedMaps,
			&previousMS.Rewards,
			&previousMS.Dead,
			&previousMS.Out,
			&previousMS.Ban,
			&previousMS.BaseDef,
			&previousMS.X,
			&previousMS.Y,
			&previousMS.Job,
			&previousMS.Map.Wid,
			&previousMS.Map.Hei,
			&previousMS.Map.Days,
			&previousMS.Map.Conspiracy,
			&previousMS.Map.Custom,
			&previousMS.Map.City.Buildings,
			&previousMS.Map.City.Bank,
			&previousMS.Map.Zones); err != nil {
			rows.Close()
			return err
		}
	}
	rows.Close()

	if !milestone.CheckFieldsDifference(&previousMS) && lastStarted < previousMS.Dt {
		// nothing has changed since last milestone
		// AND
		// no challenge started since last milestone
		return tx.Commit()
	}

	if err = tx.QueryRow(`SELECT UTC_TIMESTAMP(2)`).Scan(&milestone.Dt); err != nil {
		return err
	}

	if _, err = tx.Exec(`INSERT INTO milestone VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		milestone.User.ID,
		milestone.Dt,
		milestone.IsGhost,
		milestone.PlayedMaps,
		milestone.Rewards,
		milestone.Dead,
		milestone.Out,
		milestone.Ban,
		milestone.BaseDef,
		milestone.X,
		milestone.Y,
		milestone.Job,
		milestone.Map.Wid,
		milestone.Map.Hei,
		milestone.Map.Days,
		milestone.Map.Conspiracy,
		milestone.Map.Custom,
		milestone.Map.City.Buildings,
		milestone.Map.City.Bank,
		milestone.Map.Zones); err != nil {
		return err
	}

	// don't insert success without milestone
	var stmtBuilder strings.Builder
	stmtBuilder.Grow(355 + 21*len(successes) + 83)
	stmtBuilder.WriteString(`INSERT INTO success SELECT ?, goal.id, ?, IF(goal.amount IS NULL, 
		current,
		LEAST(
			current,
			IF(goal.typ = 0,
				IFNULL(
					goal.amount + (SELECT amount FROM success WHERE success.goal = goal.id AND success.user = ? ORDER BY accomplished LIMIT 1),
					current
				),
				goal.amount
			)
		)
	)
	FROM goal JOIN (SELECT 0 AS current, -1 AS gid`)
	successValues := make([]any, 0, 3+2*len(successes))
	successValues = append(successValues, milestone.User.ID, milestone.Dt, milestone.User.ID)
	for _, success := range successes {
		successValues = append(successValues, success.Amount, success.Goal)
		stmtBuilder.WriteString(` UNION ALL SELECT ?,?`)
	}
	stmtBuilder.WriteString(`) AS input ON goal.id = gid ON DUPLICATE KEY UPDATE success.amount = success.amount`)
	if _, err := tx.Exec(stmtBuilder.String(), successValues...); err != nil {
		return err
	}

	return tx.Commit()
}

func queryUser(id int) (user dto.DetailedUser, err error) {
	row := dbConn().QueryRow(`SELECT user.name, simplified_name, avatar, COUNT(DISTINCT challenge.id), COUNT(DISTINCT participant.challenge)
							  FROM user
							  LEFT JOIN challenge ON user.id = challenge.creator
							  LEFT JOIN participant ON participant.user = user.id
							  WHERE user.id = ?
							  GROUP BY user.id, user.name`, id)
	user.ID = id
	err = row.Scan(&user.Name, &user.SimplifiedName, &user.Avatar, &user.CreationCount, &user.ParticipationCount)
	return
}

func queryMultipleUsers(ch chan<- *dto.DetailedUser, idents []string) {
	defer close(ch)
	if len(idents) == 0 {
		return
	}

	var sqlStmt strings.Builder
	sqlStmt.Grow(293 + 36*len(idents) + 34)
	sqlStmt.WriteString(`SELECT user.id, user.name, simplified_name, avatar, COUNT(DISTINCT challenge.id), COUNT(DISTINCT participant.challenge)
				FROM user
				LEFT JOIN challenge ON user.id = challenge.creator
				LEFT JOIN participant ON participant.user = user.id
				WHERE user.id IN (SELECT id FROM user WHERE `)
	values := make([]any, 0, 2*len(idents))
	for _, ident := range idents {
		if len(ident) > 1 {
			sqlStmt.WriteString("id = ? OR simplified_name LIKE ? OR ")
			values = append(values, ident, "%"+ident+"%")
		}
	}
	if len(values) == 0 {
		return
	}
	sqlStmt.WriteString("FALSE) GROUP BY user.id, user.name")

	rows, err := dbConn().Query(sqlStmt.String(), values...)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user dto.DetailedUser
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.SimplifiedName,
			&user.Avatar,
			&user.CreationCount,
			&user.ParticipationCount); err != nil {
			logger.Println(err)
			return
		}
		ch <- &user
	}
}

func queryChallengesRelatedTo(ch chan<- *dto.DetailedChallenge, userId int, viewer int) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT user.id, user.name as cname, user.simplified_name, user.avatar
	, challenge.id, challenge.name, challenge.flags, challenge.start_date, challenge.end_date
	, (SELECT COUNT(*) FROM participant WHERE challenge = challenge.id) AS participant_count
	, challenge.start_date <= UTC_TIMESTAMP() AS started
	, challenge.end_date < UTC_TIMESTAMP() AS ended
	, challenge.creator=? AS created
	, participant.user IS NOT NULL as participate
	, validator.user IS NOT NULL as validate
	, invitation.user IS NOT NULL as invited
	 FROM challenge
	 LEFT JOIN user        ON challenge.creator = user.id
	 LEFT JOIN participant ON challenge.id = participant.challenge AND participant.user = ?
	 LEFT JOIN validator   ON challenge.id = validator.challenge AND validator.user = ?
	 LEFT JOIN invitation  ON challenge.id = invitation.challenge AND invitation.user = ? AND participant.user IS NULL
	 WHERE ? IN (challenge.creator, participant.user, validator.user, invitation.user)
	 AND (challenge.flags & 0x04 = 0 OR ? in (challenge.creator, participant.user, validator.user, invitation.user))
	 AND (?=? OR (challenge.flags & 0x30) = 0x20 AND challenge.flags & 0x03 < 2)
	 ORDER BY ended, started, challenge.flags & 0x30`, userId, userId, userId, userId, userId, viewer, userId, viewer)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var detailedChall dto.DetailedChallenge
		var Started sql.NullBool
		var Ended sql.NullBool
		var created, participate, validate, invited bool
		if err := rows.Scan(
			&detailedChall.Creator.ID,
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
			logger.Println(err)
			return
		}
		detailedChall.UpdateDetailedProperties(Started.Bool, Ended.Bool)

		detailedChall.Role = make([]int8, 0)
		if created {
			detailedChall.Role = append(detailedChall.Role, 0)
		}
		if participate {
			detailedChall.Role = append(detailedChall.Role, 1)
		} else if invited {
			if detailedChall.Access == 2 {
				detailedChall.Role = append(detailedChall.Role, 2)
			} else {
				detailedChall.Role = append(detailedChall.Role, 3)
			}
		}
		if validate {
			detailedChall.Role = append(detailedChall.Role, 4)
		}

		ch <- &detailedChall
	}
}

func insertGoals(challengeId int, goals []dto.Goal) (sql.Result, error) {
	var sqlStmt strings.Builder
	sqlStmt.Grow(71 + 23*len(goals))
	sqlStmt.WriteString("INSERT INTO goal (challenge, typ, entity, amount, x, y, custom) VALUES ")
	values := make([]any, 0, 3*len(goals))
	for _, g := range goals {
		sqlStmt.WriteString("(?, ?, ?, ?, ?, ?, ?), ")
		values = append(values, challengeId, g.Typ, g.Entity, g.Amount, g.X, g.Y, g.Custom)
	}

	return dbConn().Exec(sqlStmt.String()[:sqlStmt.Len()-2], values...)
}

func insertChallenge(toInsert *dto.Challenge, associated *[]dto.Goal) (int, error) {
	if len(*associated) == 0 {
		return 0, errors.New("cannot create challenge without goals")
	}

	res, err := dbConn().Exec(`INSERT INTO challenge (name, creator, flags) VALUES (?, ?, ?)`, toInsert.Name, toInsert.Creator.ID, toInsert.Flags)
	if err != nil {
		return 0, err
	}

	challengeId64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	challengeId := int(challengeId64)

	_, err = insertGoals(challengeId, *associated)

	return challengeId, err
}

func queryChallengeGoals(ch chan<- *dto.Goal, challengeId int) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT id, typ, entity, amount, x, y, custom FROM goal WHERE challenge=?`, challengeId)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var goal dto.Goal
		goal.Challenge = challengeId
		if err := rows.Scan(&goal.ID, &goal.Typ, &goal.Entity, &goal.Amount, &goal.X, &goal.Y, &goal.Custom); err != nil {
			logger.Println(err)
			return
		}
		ch <- &goal
	}
}

func updateChallengeStatus(challengeId, creatorId int, newStatus byte) error {
	_, err := dbConn().Exec(`UPDATE challenge SET flags = (flags & 0x0F) | (? << 4) WHERE id = ? AND creator = ? AND (flags >> 4) < 2`, newStatus, challengeId, creatorId)
	return err
}

func updateChallenge(toUpdate *dto.Challenge, associated *[]dto.Goal) error {
	if len(*associated) == 0 {
		return errors.New("cannot update challenge without goals")
	}

	_, err := dbConn().Exec(`UPDATE challenge SET name = ?, flags = ? WHERE id = ? AND creator = ? AND (flags >> 4) < 2`, toUpdate.Name, toUpdate.Flags, toUpdate.ID, toUpdate.Creator.ID)
	if err != nil {
		return err
	}

	_, err = dbConn().Exec(`DELETE FROM goal WHERE challenge = ?`, toUpdate.ID)
	if err != nil {
		return err
	}

	_, err = insertGoals(toUpdate.ID, *associated)

	return err
}

func queryChallengeParticipants(ch chan<- *dto.DetailedUser, challengeId int) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT u.id, u.name, u.simplified_name, u.avatar, COUNT(DISTINCT c.id), COUNT(DISTINCT p.challenge)
								 FROM user u
								 LEFT JOIN challenge c ON u.id = c.creator
								 LEFT JOIN participant p ON u.id = p.user
								 WHERE u.id IN (SELECT user FROM participant WHERE challenge=?)
								 GROUP BY u.id, u.name`, challengeId)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user dto.DetailedUser
		if err := rows.Scan(&user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &user.CreationCount, &user.ParticipationCount); err != nil {
			logger.Println(err)
			return
		}
		ch <- &user
	}
}

func queryChallengeValidators(ch chan<- *dto.DetailedUser, challengeId int) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT u.id, u.name, u.simplified_name, u.avatar, COUNT(DISTINCT c.id), COUNT(DISTINCT p.challenge)
								 FROM user u
								 LEFT JOIN challenge c ON u.id = c.creator
								 LEFT JOIN participant p ON u.id = p.user
								 WHERE u.id IN (SELECT user FROM validator WHERE challenge=?)
								 GROUP BY u.id, u.name`, challengeId)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user dto.DetailedUser
		if err := rows.Scan(&user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &user.CreationCount, &user.ParticipationCount); err != nil {
			logger.Println(err)
			return
		}
		ch <- &user
	}
}

func queryChallengeInvitations(ch chan<- *dto.DetailedUser, challengeId int) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT u.id, u.name, u.simplified_name, u.avatar, COUNT(DISTINCT c.id), COUNT(DISTINCT p.challenge)
								 FROM user u
								 LEFT JOIN challenge c ON u.id = c.creator
								 LEFT JOIN participant p ON u.id = p.user
								 WHERE u.id IN (SELECT invitation.user FROM invitation 
												LEFT JOIN participant ON invitation.user=participant.user
												AND invitation.challenge=participant.challenge
												AND participant.user IS NULL
												WHERE invitation.challenge=?)
								 GROUP BY u.id, u.name`, challengeId)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user dto.DetailedUser
		if err := rows.Scan(&user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &user.CreationCount, &user.ParticipationCount); err != nil {
			logger.Println(err)
			return
		}
		ch <- &user
	}
}

func queryChallengeUserStatus(challengeId, userId int) (invited, participate bool) {
	rows, err := dbConn().Query(`SELECT 0 FROM invitation WHERE challenge=? AND user=? UNION SELECT 1 FROM participant WHERE challenge=? AND user=?`,
		challengeId, userId, challengeId, userId)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r int
		if err := rows.Scan(&r); err != nil {
			logger.Println(err)
			return
		}
		switch r {
		case 0:
			invited = true
		case 1:
			participate = true
		}

	}
	return invited, participate
}

func insertOrDeleteChallengeMember(challengeId, requestorId, userId int, validator, add bool) error {
	var stmt string
	if validator {
		if add {
			stmt = "INSERT INTO validator SELECT ?,id,0 FROM challenge WHERE id=? AND creator=?"
		} else {
			stmt = "DELETE t FROM validator t LEFT JOIN challenge c ON t.challenge=c.id WHERE t.user=? AND c.id=? AND c.creator=?"
		}
	} else {
		stmt = "SET @userid=?, @challenge=?, @requestor=?;"
	}

	_, err := dbConn().Exec(stmt, userId, challengeId, requestorId)

	if validator || err != nil {
		return err
	}

	if add {
		_, err := dbConn().Exec(`INSERT INTO participant
		SELECT @userid, id
		FROM challenge
		WHERE
			id = @challenge AND (start_date IS NULL OR UTC_TIMESTAMP() < start_date)
			AND ((flags & 0x03 = 0 AND @userid = @requestor)
			OR   (flags & 0x03 = 1 AND creator = @requestor OR flags & 0x03 = 2 AND @requestor = @userid)
			AND EXISTS (SELECT 1 FROM invitation WHERE user = @userid AND challenge = @challenge))`)

		if err != nil {
			return err
		}
		_, err = dbConn().Exec(`INSERT INTO invitation
		SELECT @userid, id
		FROM challenge
		WHERE
			id = @challenge AND (start_date IS NULL OR UTC_TIMESTAMP() < start_date)
			AND ((flags & 0x03 = 1 AND @userid = @requestor)
			OR   (flags & 0x03 = 2 AND creator = @requestor))`)

		return err
	} else {
		_, err := dbConn().Exec(`DELETE p FROM participant p
		INNER JOIN challenge c ON p.challenge = c.id
		WHERE c.id = @challenge AND p.user = @userid AND (c.start_date IS NULL OR UTC_TIMESTAMP() < c.start_date)
		AND (@requestor = p.user OR @requestor = c.creator AND (c.flags & 0x03 > 0))`)

		if err != nil {
			return err
		}
		_, err = dbConn().Exec(`DELETE i FROM invitation i
		INNER JOIN challenge c ON i.challenge = c.id
		WHERE c.id = @challenge AND i.user = @userid AND (c.start_date IS NULL OR UTC_TIMESTAMP() < c.start_date)
		AND (@requestor = i.user AND (c.flags & 0x03 = 1) OR @requestor = c.creator AND (c.flags & 0x03 = 2))`)

		return err
	}
}

func updateChallengeDate(challengeID, requestorID int, date string, start bool) error {
	params := []any{challengeID, requestorID}
	var stmt string
	if start {
		if date > "" {
			stmt = `UPDATE challenge
					SET start_date = ?
					WHERE id = ?
					AND creator = ?
					AND (start_date IS NULL OR UTC_TIMESTAMP() < start_date)	# challenge pas commencé
					AND (  end_date IS NULL OR     ? <   end_date)	# date avant la fin`
			params = []any{date, challengeID, requestorID, date}
		} else {
			stmt = `UPDATE challenge
					SET start_date = UTC_TIMESTAMP()
					WHERE id = ?
					AND creator = ?
					AND (start_date IS NULL OR UTC_TIMESTAMP() < start_date)	# challenge pas commencé`
		}
	} else {
		if date > "" {
			stmt = `UPDATE challenge
					SET end_date = ?
					WHERE id = ?
					AND creator = ?
					AND UTC_TIMESTAMP() < ?								# date dans le futur
					AND start_date IS NOT NULL					# challenge doit avoir un début
					AND start_date < ?							# date après le début
					AND (end_date IS NULL OR UTC_TIMESTAMP() < end_date)  # challenge pas terminé`
			params = []any{date, challengeID, requestorID, date, date}
		} else {
			stmt = `UPDATE challenge
					SET end_date = UTC_TIMESTAMP()
					WHERE id = ?
					AND creator = ?
					AND start_date IS NOT NULL					# challenge doit avoir un début
					AND start_date < UTC_TIMESTAMP()						# challenge commencé
					AND (end_date IS NULL OR UTC_TIMESTAMP() < end_date)  # challenge pas terminé`
		}
	}
	_, err := dbConn().Exec(stmt, params...)
	return err
}

func queryChallengeAdvancements(ch chan<- *dto.UserAdvance, challengeID int) {
	defer close(ch)
	rows, err := dbConn().Query(`SELECT user.id, user.name, user.simplified_name, user.avatar, IFNULL(u.goal,0), IFNULL(u.accomplished,""), IFNULL(u.amount,0)
		FROM (
			SELECT user, challenge
			FROM participant
			WHERE participant.challenge = ?
		) as p
		LEFT JOIN (
			SELECT s1.user, s1.goal,  s1.accomplished,  IF(goal.typ = 0, GREATEST(COALESCE(s1.amount - s3.amount, 0), 0), s1.amount) AS amount
			FROM success s1
			JOIN goal ON goal.id = s1.goal AND goal.challenge = ?
			LEFT JOIN success s2 ON s1.user = s2.user AND s1.goal = s2.goal AND s1.accomplished < s2.accomplished
			LEFT JOIN success s3 ON s1.user = s3.user AND s1.goal = s3.goal AND s1.accomplished > s3.accomplished AND goal.typ = 0
			WHERE s2.user IS NULL
			AND  (s3.user IS NULL OR CONCAT(s1.accomplished, s3.accomplished) = (
				SELECT CONCAT(MAX(accomplished), MIN(accomplished))
				FROM success WHERE user = s3.user AND goal = s3.goal
			))
		) u ON p.user = u.user
		JOIN user on user.id = p.user
		ORDER BY id`, challengeID, challengeID)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	cuser := new(dto.UserAdvance)
	cuser.Successes = make(map[int]dto.Success)
	for rows.Next() {
		var user dto.User
		var success dto.Success
		if err := rows.Scan(&user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &success.Goal, &success.Accomplished, &success.Amount); err != nil {
			logger.Println(err)
			return
		}
		if user.ID != cuser.ID {
			if cuser.ID != 0 {
				ch <- cuser
				cuser = new(dto.UserAdvance)
				cuser.Successes = make(map[int]dto.Success)
			}
			cuser.User = user
		}
		cuser.Successes[success.Goal] = success
	}
	if cuser.ID != 0 {
		ch <- cuser
	}
}

func queryChallengeHistory(ch chan<- *dto.UserHistory, challengeID int) {
	defer close(ch)
	rows, err := dbConn().Query(`SELECT challenge.start_date, user.id, user.name, user.simplified_name, user.avatar, goal.typ = 0, success.goal, success.accomplished, success.amount FROM success
	JOIN goal ON success.goal = goal.id AND goal.challenge = ?
	JOIN challenge ON goal.challenge = challenge.ID
	JOIN user ON success.user = user.id
	ORDER BY user.id, goal.id, success.accomplished`, challengeID)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	cuser := new(dto.UserHistory)
	cuser.History = make(map[int][]dto.Success)
	firstPictAmount := uint32(0)
	for rows.Next() {
		picto := false
		var startDate string
		var user dto.User
		var success dto.Success
		if err := rows.Scan(&startDate, &user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &picto, &success.Goal, &success.Accomplished, &success.Amount); err != nil {
			logger.Println(err)
			return
		}
		if user.ID != cuser.ID {
			if cuser.ID != 0 {
				ch <- cuser
				cuser = new(dto.UserHistory)
				cuser.History = make(map[int][]dto.Success)
			}
			cuser.User = user
		}
		if cuser.History[success.Goal] == nil {
			if picto {
				firstPictAmount = success.Amount
				success.Amount = 0
			}
			cuser.History[success.Goal] = []dto.Success{{User: success.User, Goal: success.Goal, Accomplished: startDate, Amount: 0}, success}
		} else {
			if picto {
				success.Amount = success.Amount - firstPictAmount
			}
			cuser.History[success.Goal] = append(cuser.History[success.Goal], success)
		}
	}
	if cuser.ID != 0 {
		ch <- cuser
	}
}

func queryChallengeRawHistory(ch chan<- *dto.Success, challengeID int) {
	defer close(ch)
	rows, err := dbConn().Query(`SELECT success.user, success.goal, success.accomplished, success.amount FROM success
	JOIN goal ON success.goal = goal.id AND goal.challenge = ?
	ORDER BY success.accomplished`, challengeID)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var success dto.Success
		if err := rows.Scan(&success.User, &success.Goal, &success.Accomplished, &success.Amount); err != nil {
			logger.Println(err)
			return
		}
		ch <- &success
	}
}

func queryChallengeParticipantsForScan(challengeID, requestorID int) (string, error) {
	rows, err := dbConn().Query(`SELECT user FROM participant
	WHERE challenge = ?
	AND EXISTS (SELECT * FROM challenge 
		WHERE id = ? AND CREATOR = ? 
		AND start_date <= UTC_TIMESTAMP()
		AND (UTC_TIMESTAMP() < end_date OR end_date IS NULL)
		AND flags & 0x30 = 0x20)`, challengeID, challengeID, requestorID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var builder strings.Builder

	if rows.Next() {
		for {
			var userID int
			if err := rows.Scan(&userID); err != nil {
				return "", err
			}
			builder.WriteString(strconv.Itoa(userID))
			if rows.Next() {
				builder.WriteRune(',')
			} else {
				break
			}
		}
	}

	return builder.String(), nil
}

type Acompletion struct {
	dto.Goal
	Previous sql.NullInt32
	Success  sql.NullInt32
}

type Verification struct {
	Milestone *dto.Milestone
	Goals     []Acompletion
}

type Verifications []Verification

func (v Verifications) Len() int {
	return len(v)
}
func (v Verifications) Less(i, j int) bool {
	return v[i].Milestone.Dt > v[j].Milestone.Dt
}
func (v Verifications) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func queryValidations(userID int) (map[int]Verifications, []*dto.Challenge, error) {
	rows, err := dbConn().Query(`SELECT m.*, user.name, user.avatar, goal.id, goal.typ, goal.entity, goal.x, goal.y, goal.custom, goal.amount, success.amount, DATEDIFF(UTC_TIMESTAMP(), end_date) rem
	FROM (
		SELECT m.*, challenge.id, challenge.name, challenge.end_date, (dt >= challenge.start_date) as bef FROM milestone m
		JOIN participant ON participant.user = m.user
		JOIN challenge ON challenge.id = participant.challenge
		WHERE challenge.start_date <= UTC_TIMESTAMP()
		AND (challenge.flags & 0x30) = 0x20
		UNION
		SELECT DISTINCT m.user, '9999-12-31', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, challenge.id, challenge.name, challenge.end_date, 2 FROM milestone m
		JOIN participant ON participant.user = m.user
		JOIN challenge ON challenge.id = participant.challenge
		WHERE challenge.start_date <= UTC_TIMESTAMP()
		AND (challenge.flags & 0x30) = 0x20
	) AS m
	JOIN user ON user.id = m.user
	JOIN goal ON goal.challenge = m.id
	JOIN validator ON validator.challenge = m.id AND validator.user = ? AND validator.archived = 0
	LEFT JOIN success ON success.user = m.user AND success.accomplished = m.dt AND success.goal = goal.id
	ORDER BY rem IS NULL OR rem < 0 DESC, rem IS NULL, ABS(rem), m.id, m.user, m.dt, m.bef, goal.id`, userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	result := make(map[int]Verifications)
	resOrder := make([]*dto.Challenge, 0)
	milestone := new(dto.Milestone)
	var previousAcompletion map[int]sql.NullInt32
	prevUser := -1

	for rows.Next() {
		var before int
		var challenge dto.Challenge
		var goal dto.Goal
		var successAmount sql.NullInt32
		var rem sql.NullInt64
		if err := rows.Scan(
			&milestone.User.ID,
			&milestone.Dt,
			&milestone.IsGhost,
			&milestone.PlayedMaps,
			&milestone.Rewards,
			&milestone.Dead,
			&milestone.Out,
			&milestone.Ban,
			&milestone.BaseDef,
			&milestone.X,
			&milestone.Y,
			&milestone.Job,
			&milestone.Map.Wid,
			&milestone.Map.Hei,
			&milestone.Map.Days,
			&milestone.Map.Conspiracy,
			&milestone.Map.Custom,
			&milestone.Map.City.Buildings,
			&milestone.Map.City.Bank,
			&milestone.Map.Zones,
			&challenge.ID,
			&challenge.Name,
			&challenge.EndDate,
			&before,
			&milestone.User.Name,
			&milestone.User.Avatar,
			&goal.ID,
			&goal.Typ,
			&goal.Entity,
			&goal.X,
			&goal.Y,
			&goal.Custom,
			&goal.Amount,
			&successAmount,
			&rem); err != nil {
			return nil, nil, err
		}
		if goal.Typ == 2 {
			goal.Amount.Int32 = 1
			goal.Amount.Valid = true
		}

		if prevUser != milestone.User.ID {
			previousAcompletion = make(map[int]sql.NullInt32, 0)
		}
		prevUser = milestone.User.ID

		addedChallenge := len(resOrder)
		if addedChallenge == 0 || resOrder[addedChallenge-1].ID != challenge.ID {
			if rem.Valid && rem.Int64 > 0 {
				challenge.Flags = 1
			}
			resOrder = append(resOrder, &challenge)
		}

		switch before {
		case 1:
			if _, ok := result[challenge.ID]; ok {
				last := result[challenge.ID][len(result[challenge.ID])-1]
				if last.Milestone.Dt == milestone.Dt && last.Milestone.User.ID == milestone.User.ID {
					result[challenge.ID][len(result[challenge.ID])-1].Goals = append(last.Goals, Acompletion{goal, previousAcompletion[goal.ID], successAmount})
				} else {
					if milestone.HasData() {
						result[challenge.ID] = append(result[challenge.ID], Verification{milestone, []Acompletion{{goal, previousAcompletion[goal.ID], successAmount}}})
					}
				}
			} else {
				if milestone.HasData() {
					result[challenge.ID] = Verifications{{milestone, []Acompletion{{goal, previousAcompletion[goal.ID], successAmount}}}}
				}
			}
			fallthrough
		case 2:
			milestone = new(dto.Milestone)
		}
		if successAmount.Valid {
			previousAcompletion[goal.ID] = successAmount
		}
	}

	for k := range result {
		sort.Sort(result[k])
	}

	return result, resOrder, nil
}

func queryMilestone(milestoneCh chan<- *dto.Milestone, requestor int) {
	defer close(milestoneCh)

	rows, err := dbConn().Query(`SELECT * FROM milestone WHERE user = ? ORDER BY dt DESC`, requestor)
	if err != nil {
		logger.Println(err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		milestone := new(dto.Milestone)
		if err := rows.Scan(
			&milestone.User.ID,
			&milestone.Dt,
			&milestone.IsGhost,
			&milestone.PlayedMaps,
			&milestone.Rewards,
			&milestone.Dead,
			&milestone.Out,
			&milestone.Ban,
			&milestone.BaseDef,
			&milestone.X,
			&milestone.Y,
			&milestone.Job,
			&milestone.Map.Wid,
			&milestone.Map.Hei,
			&milestone.Map.Days,
			&milestone.Map.Conspiracy,
			&milestone.Map.Custom,
			&milestone.Map.City.Buildings,
			&milestone.Map.City.Bank,
			&milestone.Map.Zones); err != nil {
			logger.Println(err)
			return
		}
		if milestone.HasData() {
			milestoneCh <- milestone
		}
	}
}

func insertSuccesses(user int, dt string, amounts map[string][]string, requestor int) error {
	if len(amounts) == 0 {
		return nil
	}

	tx, err := dbConn().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var stmtBuilder strings.Builder
	stmtBuilder.Grow(148 + 2*len(amounts) + 82)
	stmtBuilder.WriteString(`DELETE success FROM success
		JOIN goal ON goal.id = success.goal
		JOIN validator ON validator.challenge = goal.challenge
		WHERE success.goal IN (`)
	values := make([]any, 0, len(amounts)+3)
	for goal := range amounts {
		values = append(values, goal)
		stmtBuilder.WriteString("?,")
	}
	stmtBuilder.WriteString(`-1)
		AND validator.user = ?
		AND success.user = ?
		AND success.accomplished = ?`)
	values = append(values, requestor, user, dt)

	if _, err := tx.Exec(stmtBuilder.String(), values...); err != nil {
		return err
	}

	stmtBuilder.Reset()
	stmtBuilder.Grow(435 + 27)
	stmtBuilder.WriteString(`INSERT INTO success SELECT ?, goal.id, ?, IF(goal.amount IS NULL, 
		current,
		LEAST(
			current,
			IF(goal.typ = 0,
				IFNULL(
					goal.amount + (SELECT amount FROM success WHERE success.goal = goal.id AND success.user = ? ORDER BY accomplished LIMIT 1),
					current
				),
				goal.amount
			)
		)
	)
	FROM goal
	JOIN validator ON validator.challenge = goal.challenge AND validator.user = ?
	JOIN (SELECT 0 AS current, -1 AS gid`)
	successValues := make([]any, 0, 4+2*len(amounts))
	successValues = append(successValues, user, dt, user, requestor)
	for goal, amount := range amounts {
		if len(amount) > 0 && amount[0] > "" {
			successValues = append(successValues, amount[0], goal)
			stmtBuilder.WriteString(` UNION ALL SELECT ?,?`)
		}
	}
	stmtBuilder.WriteString(`) AS input ON goal.id = gid`)
	if _, err := tx.Exec(stmtBuilder.String(), successValues...); err != nil {
		return err
	}

	return tx.Commit()
}

func archiveChallengeValidation(challenge, requestor int) error {
	if _, err := dbConn().Exec(`UPDATE validator SET archived = 1 WHERE user = ? AND challenge = ?`, requestor, challenge); err != nil {
		return err
	}
	// remove participant milestone who became useless

	/*
		for all participants of challenge
		retrieve earliest non archived challenge start date
		rebuild milestone from begining to 1st milestone before challenge start_date
		update milestone
		delete milestones anterior to 1st milestone before challenge start_date
	*/

	rows, err := dbConn().Query(`SELECT p1.user, MIN(start_date)
	FROM challenge
	JOIN participant p1 ON p1.challenge = ?
	JOIN participant p2 ON p1.user = p2.user AND challenge.id = p2.challenge
	JOIN validator ON p2.challenge = validator.challenge AND validator.archived = 0
	GROUP BY p1.user`, challenge)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var milestone dto.Milestone
		if err := rows.Scan(&milestone.User.ID, &milestone.Dt); err != nil {
			return err
		}
		go mergeMilestonesOlderThan(&milestone, 5)
	}

	return nil
}

func mergeMilestonesOlderThan(milestone *dto.Milestone, threshold int) {
	tx, err := dbConn().Begin()
	if err != nil {
		logger.Println(err)
		return
	}
	defer tx.Rollback()

	rows, err := tx.Query(`SELECT dt, isGhost, playedMaps, rewards, dead, isOut, ban, baseDef, x, y, job, mapWid, mapHei, mapDays, conspiracy, custom, buildings, bank, zoneItems
	FROM milestone WHERE user = ? AND dt < ? ORDER BY dt ASC`, milestone.User.ID, milestone.Dt)
	if err != nil {
		logger.Println(err)
		return
	}

	i := 0
	for rows.Next() {
		if err = rows.Scan(
			&milestone.Dt,
			&milestone.IsGhost,
			&milestone.PlayedMaps,
			&milestone.Rewards,
			&milestone.Dead,
			&milestone.Out,
			&milestone.Ban,
			&milestone.BaseDef,
			&milestone.X,
			&milestone.Y,
			&milestone.Job,
			&milestone.Map.Wid,
			&milestone.Map.Hei,
			&milestone.Map.Days,
			&milestone.Map.Conspiracy,
			&milestone.Map.Custom,
			&milestone.Map.City.Buildings,
			&milestone.Map.City.Bank,
			&milestone.Map.Zones); err != nil {
			rows.Close()
			logger.Println(err)
			return
		}
		i += 1
	}
	rows.Close()

	if i < threshold {
		return
	}

	// populate sql fields
	milestone.CheckFieldsDifference(new(dto.Milestone))

	if _, err = tx.Exec(`UPDATE milestone SET isGhost = ?, playedMaps = ?, rewards = ?, dead = ?, isOut = ?, ban = ?, baseDef = ?, x = ?, y = ?, job = ?, mapWid = ?, mapHei = ?, mapDays = ?, conspiracy = ?, custom = ?, buildings = ?, bank = ?, zoneItems = ?
	WHERE user = ? AND dt = ?`,
		milestone.IsGhost,
		milestone.PlayedMaps,
		milestone.Rewards,
		milestone.Dead,
		milestone.Out,
		milestone.Ban,
		milestone.BaseDef,
		milestone.X,
		milestone.Y,
		milestone.Job,
		milestone.Map.Wid,
		milestone.Map.Hei,
		milestone.Map.Days,
		milestone.Map.Conspiracy,
		milestone.Map.Custom,
		milestone.Map.City.Buildings,
		milestone.Map.City.Bank,
		milestone.Map.Zones,
		milestone.User.ID,
		milestone.Dt); err != nil {
		logger.Println(err)
		return
	}

	if _, err := tx.Exec(`DELETE FROM milestone WHERE user = ? AND dt < ?`, milestone.User.ID, milestone.Dt); err != nil {
		logger.Println(err)
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Println(err)
	}
}

func removeChallengeStart(challenge, requestor int) error {
	_, err := dbConn().Exec(`UPDATE challenge SET start_date = NULL, end_date = NULL
	WHERE creator = ?
	AND id = ?
	AND NOT EXISTS (SELECT 1 FROM success JOIN goal ON success.goal = goal.id AND goal.challenge = ?)`, requestor, challenge, challenge)
	return err
}
