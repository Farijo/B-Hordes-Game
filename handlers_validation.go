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
	mustValidate, err := queryValidations(c.GetInt("uid"))
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}
	c.HTML(http.StatusOK, "validation.html", gin.H{
		"logged":      true,
		"validations": mustValidate,
		"userkey":     c.GetString("key"),
	})
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                   POST                  *
 * * * * * * * * * * * * * * * * * * * * * */

func validateGoalHandle(c *gin.Context) {
	_ = c.GetInt("uid")
	c.MultipartForm()
	for id_status, action := range c.Request.PostForm {
		fmt.Println(id_status, action)
	}

	c.Redirect(http.StatusFound, "/validation")
}
