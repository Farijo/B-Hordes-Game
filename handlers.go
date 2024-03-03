package main

import (
	"bhordesgame/dto"
	"fmt"
	"net/http"
	"strconv"

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
	c.Bind(&formChallenge)

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

	switch challenge.Status {
	case 0, 1: // draft, review
		if challenge.Creator.ID == uid && cookieErr == nil && ok {
			ch := make(chan *dto.Goal)
			go queryChallengeGoals(id, ch)
			c.HTML(http.StatusOK, "challenge-creation.html", gin.H{"logged": true, "challenge": challenge, "goals": ch, "srvData": getServerData(key)})
		} else {
			c.Status(http.StatusForbidden)
			return
		}
	case 2: // invite
	case 3: // started
	case 4: // ended
	}
}
