package main

import (
	"bhordesgame/dto"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

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

func userInfoActualizerHandle(c *gin.Context) {
	defer c.Status(http.StatusNoContent)

	userChan := make(chan *dto.User)
	go queryAllUsers(userChan)

	var ids strings.Builder
	users := make(map[int]*dto.User, 0)
	for user := range userChan {
		users[user.ID] = user
		ids.WriteString(strconv.Itoa(user.ID))
		ids.WriteRune(',')
	}
	ids.WriteRune('1')

	var wg sync.WaitGroup
	actualizedUsers := make(chan *dto.User)
	err := requestMultipleUsers(os.Getenv("USER_KEY"), ids.String(), func(u *dto.User) bool {
		oldUser, ok := users[u.ID]
		return ok && (oldUser.Name != u.Name || oldUser.Avatar != u.Avatar)
	}, actualizedUsers, &wg)
	if err != nil {
		logger.Println(err)
		return
	}

	_, err = insertMultipleUsers(actualizedUsers)
	if err != nil {
		logger.Println(err)
		return
	}
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
			logger.Println(err)
			c.Status(http.StatusBadRequest)
			return
		}
	}
	if err := refreshData(key); err != nil {
		logger.Println(err)
		if err.Error() == "too many request" {
			c.Status(http.StatusTooManyRequests)
			return
		}
	}
	c.Redirect(http.StatusFound, c.PostForm("redirect"))
}
