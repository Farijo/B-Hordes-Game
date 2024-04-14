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
	})
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                   POST                  *
 * * * * * * * * * * * * * * * * * * * * * */
