package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

/* * * * * * * * * * * * * * * * * * * * * *
 *                   GET                   *
 * * * * * * * * * * * * * * * * * * * * * */

func validationHandle(c *gin.Context) {
	mustValidate, order, err := queryValidations(c.GetInt("uid"))
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}
	c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_validation.html", gin.H{
		"logged":      true,
		"validations": mustValidate,
		"order":       order,
		"userkey":     c.GetString("key"),
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
		fmt.Println(bindErr)
		c.Status(http.StatusBadRequest)
		return
	}
	c.MultipartForm()
	delete(c.Request.PostForm, "user")
	delete(c.Request.PostForm, "dt")
	insertErr := insertSuccesses(mileData.User, mileData.Dt, c.Request.PostForm, c.GetInt("uid"))
	if insertErr != nil {
		fmt.Println(insertErr)
		c.Status(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, "/validation")
}
