package main

import (
	"bhordesgame/dto"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func connectionHandle(c *gin.Context) {
	if key := c.PostForm("key"); key > "" {
		c.SetCookie("user", key, 24*60*60, "/", "localhost", false, true)
		if err := refreshData(key); err != nil {
			fmt.Println(err)
		}
	}
	c.Redirect(http.StatusFound, "/")
}

func indexHandle(c *gin.Context) {
	ch := make(chan *dto.DetailedChallenge)

	go queryPublicChallenges(ch)

	key, err := c.Cookie("user")
	_, ok := sessions[key]
	c.HTML(http.StatusOK, "index.html", gin.H{"logged": err == nil && ok, "challenges": ch})
}

func logoutHandle(c *gin.Context) {
	if key, err := c.Cookie("user"); err != nil {
		delete(sessions, key)
	}
	c.SetCookie("user", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
}

func refreshHandle(c *gin.Context) {
	if err := refreshData(c.GetString("key")); err != nil {
		fmt.Println(err)
	}
	c.Redirect(http.StatusFound, "/user")
}

func selfHandle(c *gin.Context) {
	c.Redirect(http.StatusFound, fmt.Sprintf("/user/%d", c.GetInt("uid")))
}

func userHandle(c *gin.Context) {
	id, atoiErr := strconv.Atoi(c.Param("id"))
	if atoiErr != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	key, cookieErr := c.Cookie("user")
	user, queryErr := queryUser(id)
	if queryErr != nil {
		if cookieErr != nil {
			c.Redirect(http.StatusSeeOther, "https://myhordes.eu/jx/disclaimer/26")
			return
		}

		mhUser, mhApiErr := requestUser(key, id)
		if mhApiErr != nil {
			c.Status(http.StatusNotFound)
			return
		}

		insertUser(mhUser)
		user = *mhUser
	}

	ch := make(chan *dto.DetailedChallenge)

	currentUser, ok := sessions[key]
	go queryChallengesRelatedTo(id, currentUser, ch)

	c.HTML(http.StatusOK, "user.html", gin.H{"logged": cookieErr == nil && ok, "challenges": ch, "user": &user})
}

func createChallengeHandle(c *gin.Context) {
	var formChallenge FormChallenge
	c.Bind(&formChallenge)

	challenge := formChallenge.buildChallenge(c.GetInt("uid"))
	goals := buildGoalsFromForm(c.PostFormArray("type"), c.PostFormArray("x"), c.PostFormArray("y"), c.PostFormArray("count"), c.PostFormArray("val"))

	id, err := insertChallenge(challenge, goals)
	if err != nil {
		panic(err.Error())
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/challenge/%d", id))
}

func requireAuth(c *gin.Context) {
	key, cookieErr := c.Cookie("user")
	uid, ok := sessions[key]
	if cookieErr != nil || !ok {
		c.Redirect(http.StatusSeeOther, "https://myhordes.eu/jx/disclaimer/26")
		c.Abort()
	} else {
		c.Set("key", key)
		c.Set("uid", uid)
	}
}

func challengeCreationHandle(c *gin.Context) {
	c.HTML(http.StatusOK, "challenge-creation.html", gin.H{"logged": true, "challenge": nil, "srvData": getServerData(c.GetString("key"))})
}

func updateChallengeHandle(c *gin.Context) {
	id, atoiErr := strconv.Atoi(c.Param("id"))
	if atoiErr != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var formChallenge FormChallenge
	if bindErr := c.Bind(&formChallenge); bindErr != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var err error
	switch formChallenge.Act {
	case "Modifier":
		err = updateChallengeStatus(id, c.GetInt("uid"), 0)
	case "Ouvrir les inscriptions":
		err = updateChallengeStatus(id, c.GetInt("uid"), 2)
	default:
		challenge := formChallenge.buildChallenge(c.GetInt("uid"))
		challenge.ID = id
		goals := buildGoalsFromForm(c.PostFormArray("type"), c.PostFormArray("x"), c.PostFormArray("y"), c.PostFormArray("count"), c.PostFormArray("val"))
		err = updateChallenge(challenge, goals)
	}

	if err != nil {
		panic(err.Error())
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/challenge/%d", id))
}

func challengeMembersHandle(c *gin.Context) {
	fmt.Println(c.MultipartForm())
	fmt.Println(c.Request.PostForm)
}

func challengeHandle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	challenge, err := queryChallenge(id)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusNotFound)
		return
	}

	key, cookieErr := c.Cookie("user")
	uid, ok := sessions[key]

	selfChallenge := challenge.Creator.ID == uid && cookieErr == nil && ok

	switch challenge.Status {
	case 0, 1: // draft, review
		if selfChallenge {
			ch := make(chan *dto.Goal)
			go queryChallengeGoals(challenge.ID, ch)
			c.HTML(http.StatusOK, "challenge-creation.html", gin.H{"logged": true, "challenge": challenge, "goals": ch, "srvData": getServerData(key)})
		} else {
			c.Status(http.StatusForbidden)
			return
		}
	case 2: // invite
		challenge.Creator, err = queryUser(challenge.Creator.ID)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		goals := make(chan *dto.Goal)
		go queryChallengeGoals(challenge.ID, goals)

		searchResults := make(chan *dto.User)
		invitationResults := make(chan *dto.User)
		if selfChallenge {
			go func() {
				idents := strings.FieldsFunc(c.Query("ident"), func(r rune) bool { return r == ',' || r == ' ' })
				// Problem of this is making a request to MH on each reload with "ident" set
				// and the request would probably be useless (ie : we already have the info)
				//
				// if cookieErr == nil {
				// 	realIds := make([]string, 0)
				// 	for _, maybeId := range idents {
				// 		if _, err := strconv.Atoi(maybeId); err == nil {
				// 			realIds = append(realIds, maybeId)
				// 		}
				// 	}
				// 	if len(realIds) > 0 {
				// 		if users, err := requestMultipleUsers(key, realIds); err == nil {
				// 			if err := insertMultipleUsers(users); err != nil {
				// 				fmt.Println(err)
				// 			}
				// 		} else {
				// 			fmt.Println(err)
				// 		}
				// 	}
				// }
				queryMultipleUsers(idents, searchResults)
			}()
			go queryChallengeInvitations(challenge.ID, invitationResults)
		} else {
			close(searchResults)
			close(invitationResults)
		}
		validatorResults := make(chan *dto.User)
		go queryChallengeValidators(challenge.ID, validatorResults)

		participantResults := make(chan *dto.User)
		go queryChallengeParticipants(challenge.ID, participantResults)

		c.HTML(http.StatusOK, "challenge-recruit.html", gin.H{
			"logged":        cookieErr == nil && ok,
			"selfChallenge": selfChallenge,
			"selfID":        uid,
			"challenge":     challenge,
			"goals":         goals,
			"srvData":       getServerData(key),
			"searchResults": searchResults,
			"invitations":   invitationResults,
			"validators":    validatorResults,
			"participants":  participantResults,
		})
	case 3: // started
	case 4: // ended
	}
}
