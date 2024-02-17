package main

import (
	"bhordesgame/dto"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var sessions map[string]int = make(map[string]int, 0)

func isLogged(c *gin.Context) bool {
	_, err := c.Cookie("user")
	return err == nil
}

func connectionHandle(c *gin.Context) {
	if key := c.PostForm("key"); key != "" {
		c.SetCookie("user", key, 24*60*60, "/", "localhost", false, true)
		if milestone, err := requestMe(key); err == nil {
			sessions[key] = milestone.User.ID
			insertUser(&milestone.User)
			insertMilestone(milestone)
		}
	}
	c.Redirect(http.StatusFound, "/")
}

func indexHandle(c *gin.Context) {
	ch := make(chan *dto.DetailedChallenge)

	go queryPublicChallenges(ch)

	c.HTML(http.StatusOK, "index.html", gin.H{"logged": isLogged(c), "challenges": ch})
}

func logoutHandle(c *gin.Context) {
	if key, err := c.Cookie("user"); err != nil {
		delete(sessions, key)
	}
	c.SetCookie("user", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/")
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
	r.GET("/challenge/:id", challengeHandle)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
