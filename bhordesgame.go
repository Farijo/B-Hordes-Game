package main

import (
	"bhordesgame/dto"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var sessions map[string]int = make(map[string]int, 10)

func refreshData(key string) error {
	milestone, err := requestMe(key)
	if err != nil {
		return err
	}
	sessions[key] = milestone.User.ID
	if err = insertUser(&milestone.User); err != nil {
		return err
	}
	if err = insertMilestone(milestone); err != nil {
		return err
	}
	return nil
}

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

var serverData template.JS

func getServerData(userkey string) template.JS {
	if serverData == "" {
		serverData = requestServerData(userkey)
	}
	return serverData
}

func pop[T any](a *[]T) T {
	rv := (*a)[0]
	*a = (*a)[1:]
	return rv
}

func createChallengeHandle(c *gin.Context) {
	var formChallenge struct {
		Name          string `form:"name"`
		Participation int8   `form:"participation"`
		Private       bool   `form:"privat"`
		ValidationApi bool   `form:"validation_api"`
		Act           string `form:"act"`
	}
	c.Bind(&formChallenge)

	var challenge dto.Challenge
	challenge.Name = formChallenge.Name
	challenge.Creator.ID = c.GetInt("uid")
	challenge.Flags = byte(formChallenge.Participation)
	if formChallenge.Private {
		challenge.Flags |= 0x04
	}
	if !formChallenge.ValidationApi {
		challenge.Flags |= 0x08
	}

	types := c.PostFormArray("type")
	goals := make([]dto.Goal, len(types))
	x := c.PostFormArray("x")
	y := c.PostFormArray("y")
	count := c.PostFormArray("count")
	val := c.PostFormArray("val")

	for i := range goals {
		v := &goals[i]
		v.Typ = types[i][0] - '0'
		switch v.Typ {
		case 0, 3:
			v.Descript = pop(&count) + ":" + pop(&val)
		case 1:
			v.Descript = pop(&x) + ":" + pop(&y) + ":" + pop(&count) + ":" + pop(&val)
		case 2:
			v.Descript = pop(&val)
		}
	}

	id, err := insertChallenge(&challenge, &goals)
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

func challengeHandle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	challenge, err := queryChallenge(id)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
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

	}
}

func main() {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"getAccess":    getAccess,
		"getStatus":    getStatus,
		"fDate":        fdate,
		"templateHTML": func(a string) template.HTML { return template.HTML(a) },
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/style", "style")
	r.Static("/script", "script")
	r.POST("/", connectionHandle)
	r.GET("/", indexHandle)
	r.GET("/logout", logoutHandle)
	r.GET("/user/:id", userHandle)
	r.GET("/challenge/:id", challengeHandle)

	authorized := r.Group("/")
	authorized.Use(requireAuth)
	{
		authorized.POST("/user", refreshHandle)
		authorized.GET("/user", selfHandle)
		authorized.POST("/challenge", createChallengeHandle)
		authorized.GET("/challenge", challengeCreationHandle)
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
