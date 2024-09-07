package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	domain = "bhordesgames.alwaysdata.net"
)

/* * * * * * * * * * * * * * * * * * * * * *
 *                   GET                   *
 * * * * * * * * * * * * * * * * * * * * * */
func logoutHandle(c *gin.Context) {
	if key, err := c.Cookie("user"); err != nil {
		delete(sessions, key)
	}
	c.SetCookie("user", "", -1, "/", domain, false, true)
	c.Redirect(http.StatusFound, "/")
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                   POST                  *
 * * * * * * * * * * * * * * * * * * * * * */
func connectionHandle(c *gin.Context) {
	if key := c.PostForm("key"); key > "" {
		c.SetCookie("user", key, 24*60*60, "/", domain, false, true)
		if err := refreshData(key); err != nil {
			logger.Println(err)
			if err.Error() == "too many request" {
				c.Status(http.StatusTooManyRequests)
				return
			}
		}
	}
	c.Redirect(http.StatusFound, "/")
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                MIDDLEWARE               *
 * * * * * * * * * * * * * * * * * * * * * */
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
