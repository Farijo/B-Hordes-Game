package main

import (
	"fmt"
	"net/http"
	"strconv"

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
	goalAmounts := make(map[int]int, 0)
	for key, val := range c.Request.PostForm {
		if len(val) > 0 {
			goalId, atoiErrGo := strconv.Atoi(key)
			amount, atoiErrAm := strconv.Atoi(val[0])
			if atoiErrAm == nil && atoiErrGo == nil {
				goalAmounts[goalId] = amount
			}
		}
	}
	insertErr := insertSuccesses(mileData.User, mileData.Dt, goalAmounts, c.GetInt("uid"))
	if insertErr != nil {
		fmt.Println(insertErr)
		c.Status(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, "/validation")
}
