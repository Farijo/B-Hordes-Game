package main

import (
	"html"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

/* * * * * * * * * * * * * * * * * * * * * *
 *                   GET                   *
 * * * * * * * * * * * * * * * * * * * * * */

func validationHandle(c *gin.Context) {
	mustValidate, order, err := queryValidations(c.GetInt("uid"))
	if err != nil {
		logger.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}
	c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_validation.html", gin.H{
		"faq":         wantFAQ(c.Cookie(NOFAQ)),
		"logged":      true,
		"validations": mustValidate,
		"order":       order,
		"userkey":     c.GetString("key"),
	})
}

func milestoneHandle(c *gin.Context) {
	user, err := queryUser(c.GetInt("uid"))
	if err != nil {
		logger.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}
	c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_milestone.html", gin.H{
		"faq":        wantFAQ(c.Cookie(NOFAQ)),
		"logged":     true,
		"name":       user.Name,
		"avatar":     user.Avatar.String,
		"milestones": makeChannelFor(queryMilestone, c.GetInt("uid")),
		"userkey":    c.GetString("key"),
	})
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                   POST                  *
 * * * * * * * * * * * * * * * * * * * * * */

func validateGoalHandle(c *gin.Context) {
	var mileData struct {
		User int    `form:"user"`
		Dt   string `form:"dt"`
	}
	bindErr := c.Bind(&mileData)
	if bindErr != nil {
		logger.Println(bindErr)
		c.Status(http.StatusBadRequest)
		return
	}
	c.MultipartForm()
	delete(c.Request.PostForm, "user")
	delete(c.Request.PostForm, "dt")
	insertErr := insertSuccesses(mileData.User, mileData.Dt, c.Request.PostForm, c.GetInt("uid"))
	if insertErr != nil {
		switch errCasted := insertErr.(type) {
		case *mysql.MySQLError:
			if errCasted.Number == 1062 {
				c.Data(http.StatusBadRequest, "text/html", []byte(html.UnescapeString("Cannot specify already reached value")))
			}
		default:
			logger.Println(insertErr)
			c.Status(http.StatusBadRequest)
		}
		return
	}

	c.Redirect(http.StatusFound, "/validation")
}

func archiveValidationHandle(c *gin.Context) {
	challengeID, atoiErr := strconv.Atoi(c.PostForm("challenge"))
	if atoiErr != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := archiveChallengeValidation(challengeID, c.GetInt("uid")); err != nil {
		logger.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, "/validation")
}
