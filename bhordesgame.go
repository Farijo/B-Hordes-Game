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

	_, err := c.Cookie("user")
	c.HTML(http.StatusOK, "index.html", gin.H{"logged": err == nil, "challenges": ch})
}

func logoutHandle(c *gin.Context) {
	if key, err := c.Cookie("user"); err != nil {
		delete(sessions, key)
	}
	c.SetCookie("user", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
}

func refreshHandle(c *gin.Context) {
	if key, err := c.Cookie("user"); err != nil {
		fmt.Println(err)
	} else if err = refreshData(key); err != nil {
		fmt.Println(err)
	}
	c.Redirect(http.StatusFound, "/user")
}

func selfHandle(c *gin.Context) {
	key, cookieErr := c.Cookie("user")
	currentUser, ok := sessions[key]
	if cookieErr != nil || !ok {
		c.Redirect(http.StatusSeeOther, "https://myhordes.eu/jx/disclaimer/26")
		return
	}

	c.Redirect(http.StatusFound, "/user/"+strconv.Itoa(currentUser))
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

func challengeHandle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	queryChallenge(id)

	// TODO render right page
	c.JSON(200, gin.H{"e": c.Param("id")})
}

func main() {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"getAccess": getAccess,
		"getStatus": getStatus,
		"fDate":     fdate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/style", "style")
	r.Static("/script", "script")
	r.POST("/", connectionHandle)
	r.GET("/", indexHandle)
	r.GET("/logout", logoutHandle)
	r.POST("/user", refreshHandle)
	r.GET("/user", selfHandle)
	r.GET("/user/:id", userHandle)
	r.GET("/challenge/:id", challengeHandle)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
