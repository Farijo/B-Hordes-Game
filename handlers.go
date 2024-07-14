package main

import (
	"bhordesgame/dto"
	"fmt"
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

	c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_index.html", gin.H{"logged": err == nil && ok, "challenges": ch, "faq": wantFAQ(c.Cookie(NOFAQ))})
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                   POST                  *
 * * * * * * * * * * * * * * * * * * * * * */
func settingsHandle(c *gin.Context) {
	var settings struct {
		Lang   string `form:"lang"`
		Faq    bool   `form:"faq"`
		Source string `form:"source"`
	}
	if err := c.Bind(&settings); err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}
	switch settings.Lang {
	case "🇫🇷":
		c.SetCookie(LNG_KEY, "fr", 0, "/", domain, false, true)
	case "🇬🇧":
		c.SetCookie(LNG_KEY, "en", 0, "/", domain, false, true)
	case "🇪🇸":
		c.SetCookie(LNG_KEY, "es", 0, "/", domain, false, true)
	case "🇩🇪":
		c.SetCookie(LNG_KEY, "de", 0, "/", domain, false, true)
	}
	if settings.Faq {
		c.SetCookie(NOFAQ, "0", -1, "/", domain, false, true)
	} else {
		c.SetCookie(NOFAQ, "1", 0, "/", domain, false, true)
	}
	c.Redirect(http.StatusFound, settings.Source)
}
