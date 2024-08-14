package main

import (
	"bhordesgame/dto"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

/* * * * * * * * * * * * * * * * * * * * * *
 *                   GET                   *
 * * * * * * * * * * * * * * * * * * * * * */
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

		if mhUser.ID > 0 {
			insertUser(mhUser)
			user.User = *mhUser
		}
	}

	ch := make(chan *dto.DetailedChallenge)

	currentUser, ok := sessions[key]
	go queryChallengesRelatedTo(ch, id, currentUser)

	c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_user.html", gin.H{"faq": wantFAQ(c.Cookie(NOFAQ)), "logged": cookieErr == nil && ok, "challenges": ch, "user": &user})
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                   POST                  *
 * * * * * * * * * * * * * * * * * * * * * */
func refreshHandle(c *gin.Context) {
	key := c.PostForm("key")
	if key == "" {
		var err error
		key, err = c.Cookie("user")
		if err != nil {
			fmt.Println(err)
			c.Status(http.StatusBadRequest)
			return
		}
	}
	if err := refreshData(key); err != nil {
		fmt.Println(err)
		if err.Error() == "too many request" {
			c.Status(http.StatusTooManyRequests)
			return
		}
	}
	c.Redirect(http.StatusFound, c.PostForm("redirect"))
}
