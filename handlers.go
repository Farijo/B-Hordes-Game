package main

import (
	"bhordesgame/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* * * * * * * * * * * * * * * * * * * * * *
 *                   GET                   *
 * * * * * * * * * * * * * * * * * * * * * */
func indexHandle(c *gin.Context) {
	ch := make(chan *dto.DetailedChallenge)

	go queryPublicChallenges(ch)

	key, err := c.Cookie("user")
	_, ok := sessions[key]
	c.HTML(http.StatusOK, "index.html", gin.H{"logged": err == nil && ok, "challenges": ch, "ctx": c})
}
