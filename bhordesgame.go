package main

import (
	"bhordesgame/dto"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func isLogged(c *gin.Context) bool {
	_, err := c.Cookie("user")
	return err == nil
}

func indexHandle(c *gin.Context) {
	logged := isLogged(c)

	if c.Query("logout") == "1" && logged {
		c.SetCookie("user", "", -1, "/", "localhost", false, true)
		logged = false
	}

	ch := make(chan dto.DetailedChallenge)

	go queryPublicChallenges(ch)

	c.HTML(http.StatusOK, "index.html", gin.H{"logged": logged, "challenges": ch})
}

func connectionHandle(c *gin.Context) {
	if key := c.PostForm("key"); key != "" {
		c.SetCookie("user", key, 24*60*60, "/", "localhost", false, true)
	}
	c.Redirect(http.StatusFound, "/")
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
	r.GET("/", indexHandle)
	r.POST("/", connectionHandle)
	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
