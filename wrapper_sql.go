package main

import (
	"bhordesgame/dto"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
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
	 ORDER BY ended, started`)
	if err != nil {
		fmt.Println(err)
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
			fmt.Println(err)
			return
		}
		detailedChall.UpdateDetailedProperties(Started.Bool, Ended.Bool)
		ch <- &detailedChall
	}
}

func queryChallenge(id int) (challenge dto.DetailedChallenge, err error) {
	var untilStart, untilEnd sql.NullString
	row := dbConn().QueryRow(`SELECT name, creator, flags, start_date, end_date
	, TIMEDIFF(start_date,UTC_TIMESTAMP()) as rem_start, TIMEDIFF(end_date,UTC_TIMESTAMP()) as rem_end
	 FROM challenge
	 WHERE id=?`, id)

	if err = row.Scan(&challenge.Name, &challenge.Creator.ID, &challenge.Flags, &challenge.StartDate, &challenge.EndDate, &untilStart, &untilEnd); err != nil {
		return
	}
	challenge.ID = id
	challenge.UpdateDetailedProperties(untilStart.Valid && (untilStart.String[0] == '-' || untilStart.String == "00:00:00"), untilEnd.Valid && (untilEnd.String[0] == '-' || untilEnd.String == "00:00:00"))

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

// func insertMultipleUsers(user []dto.User) error {
// 	panic("not implem")
// }

func insertMilestone(milestone *dto.Milestone) error {
	tx, err := dbConn().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, err := tx.Query(`SELECT goal.id,typ,entity,amount,x,y, challenge.flags & 0x08 = 0 as api
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
	for rows.Next() {
		rowPresent = true
		var g dto.Goal
		var api bool
		if err = rows.Scan(&g.ID, &g.Typ, &g.Entity, &g.Amount, &g.X, &g.Y, &api); err != nil {
			rows.Close()
			return err
		}
		if !api || g.Typ != 0 {
			continue
		}
		count := milestone.Rewards.Pictos[g.Entity]
		if g.Amount.Valid && uint32(g.Amount.Int32) < count {
			count = uint32(g.Amount.Int32)
		}
		successes = append(successes, dto.Success{
			User:         milestone.User.ID,
			Goal:         g.ID,
			Accomplished: "",
			Amount:       count,
		})
	}
	rows.Close()
	if !rowPresent {
		// not in a challenge, nothing to do
		return nil
	}

	for _, success := range successes {
		if _, err := tx.Exec(`INSERT INTO success VALUES(?, ?, UTC_TIMESTAMP(2), ?)
		ON DUPLICATE KEY UPDATE amount=amount`, success.User, success.Goal, success.Amount); err != nil {
			return err
		}
	}

	rows, err = tx.Query(`SELECT rewards, isGhost, playedMaps, dead, ban, baseDef, x, y, job, mapWid, mapHei, mapDays, conspiracy, custom FROM milestone WHERE user=? ORDER BY dt ASC`, milestone.User.ID)
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
		return tx.Commit()
	}

	if _, err = tx.Exec(`INSERT INTO milestone VALUES(?, UTC_TIMESTAMP(2), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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

	sqlStmt := `SELECT user.id, user.name, simplified_name, avatar, COUNT(DISTINCT challenge.id), COUNT(DISTINCT participant.challenge)
				FROM user
				LEFT JOIN challenge ON user.id = challenge.creator
				LEFT JOIN participant ON participant.user = user.id
				WHERE user.id IN (SELECT id FROM user WHERE `
	values := make([]any, 0, 3)
	for _, ident := range idents {
		if len(ident) > 1 {
			sqlStmt += "id = ? OR simplified_name LIKE ? OR "
			values = append(values, ident, "%"+ident+"%")
		}
	}
	if len(values) == 0 {
		return
	}
	sqlStmt = sqlStmt[:len(sqlStmt)-3] + ") GROUP BY user.id, user.name"

	rows, err := dbConn().Query(sqlStmt, values...)
	if err != nil {
		fmt.Println(err)
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
			fmt.Println(err)
			return
		}
		ch <- &user
	}
}

func queryChallengesRelatedTo(ch chan<- *dto.DetailedChallenge, userId int, viewer int) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT user.id, user.name as cname, user.simplified_name, user.avatar
	, challenge.id, challenge.name, challenge.flags, challenge.start_date, challenge.end_date
	, COUNT(participant.user) AS participant_count
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
	 AND (challenge.flags & 0x04 = 0 OR ? in (challenge.creator, participant.user, validator.user))
	 AND (?=? OR (challenge.flags & 0x30) = 0x20 AND challenge.flags & 0x03 < 2)
	 GROUP BY challenge.id, challenge.name, challenge.end_date, created, participate, validate, invited
	 ORDER BY ended, started, challenge.flags & 0x30`, userId, userId, userId, userId, userId, viewer, userId, viewer)
	if err != nil {
		fmt.Println(err)
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
			fmt.Println(err)
			return
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

	sqlStmt := "INSERT INTO goal (challenge, typ, entity, amount, x, y) VALUES "
	values := make([]any, 0, 3*len(*associated))
	for _, g := range *associated {
		sqlStmt += "(?, ?, ?, ?, ?, ?), "
		values = append(values, challengeId, g.Typ, g.Entity, g.Amount, g.X, g.Y)
	}
	sqlStmt = sqlStmt[:len(sqlStmt)-2]

	_, err = dbConn().Exec(sqlStmt, values...)

	return challengeId, err
}

func queryChallengeGoals(ch chan<- *dto.Goal, challengeId int) {
	defer close(ch)

	rows, err := dbConn().Query(`SELECT id, typ, entity, amount, x, y FROM goal WHERE challenge=?`, challengeId)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var goal dto.Goal
		goal.Challenge = challengeId
		if err := rows.Scan(&goal.ID, &goal.Typ, &goal.Entity, &goal.Amount, &goal.X, &goal.Y); err != nil {
			fmt.Println(err)
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

	sqlStmt := "INSERT INTO goal (challenge, typ, entity, amount, x, y) VALUES "
	values := make([]any, 0, 3*len(*associated))
	for _, g := range *associated {
		sqlStmt += "(?, ?, ?, ?, ?, ?), "
		values = append(values, toUpdate.ID, g.Typ, g.Entity, g.Amount, g.X, g.Y)
	}
	sqlStmt = sqlStmt[:len(sqlStmt)-2]

	_, err = dbConn().Exec(sqlStmt, values...)

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
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user dto.DetailedUser
		if err := rows.Scan(&user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &user.CreationCount, &user.ParticipationCount); err != nil {
			fmt.Println(err)
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
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user dto.DetailedUser
		if err := rows.Scan(&user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &user.CreationCount, &user.ParticipationCount); err != nil {
			fmt.Println(err)
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
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var user dto.DetailedUser
		if err := rows.Scan(&user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &user.CreationCount, &user.ParticipationCount); err != nil {
			fmt.Println(err)
			return
		}
		ch <- &user
	}
}

func queryChallengeUserStatus(challengeId, userId int) (invited, participate bool) {
	rows, err := dbConn().Query(`select 0 from invitation where challenge=? and user=? union select 1 from participant where challenge=? and user=?`,
		challengeId, userId, challengeId, userId)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r int
		if err := rows.Scan(&r); err != nil {
			fmt.Println(err)
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
			stmt = "INSERT INTO validator SELECT ?,id FROM challenge WHERE id=? AND creator=?"
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
	rows, err := dbConn().Query(`SELECT user.id, name, simplified_name, avatar, s1.goal, s1.accomplished, s1.amount
				   FROM success s1
				   JOIN user on user.id = s1.user
				   LEFT JOIN success s2 ON s1.user = s2.user AND s1.goal = s2.goal AND s1.accomplished < s2.accomplished
				   JOIN	goal on goal.id = s1.goal AND goal.challenge = ?
				   WHERE s2.user IS NULL
				   ORDER BY s1.user`, challengeID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rows.Close()
	cuser := new(dto.UserAdvance)
	cuser.Successes = make(map[int]dto.Success)
	for rows.Next() {
		var user dto.User
		var success dto.Success
		if err := rows.Scan(&user.ID, &user.Name, &user.SimplifiedName, &user.Avatar, &success.Goal, &success.Accomplished, &success.Amount); err != nil {
			fmt.Println(err)
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
