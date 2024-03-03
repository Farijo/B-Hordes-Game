package main

import (
	"html/template"

	"github.com/gin-gonic/gin"
)

func pop[T any](a *[]T) T {
	rv := (*a)[0]
	*a = (*a)[1:]
	return rv
}

func main() {
	r := gin.Default()
	r.SetFuncMap(template.FuncMap{
		"getAccess":    getAccess,
		"getStatus":    getStatus,
		"fDate":        fdate,
		"templateHTML": func(a string) template.HTML { return template.HTML(a) },
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/style", "style")
	r.Static("/script", "script")
	r.POST("/", connectionHandle)
	r.GET("/", indexHandle)
	r.GET("/logout", logoutHandle)
	r.GET("/user/:id", userHandle)
	r.GET("/challenge/:id", challengeHandle)

	authorized := r.Group("/")
	authorized.Use(requireAuth)
	{
		authorized.POST("/user", refreshHandle)
		authorized.GET("/user", selfHandle)
		authorized.POST("/challenge", createChallengeHandle)
		authorized.GET("/challenge", challengeCreationHandle)
		authorized.POST("/challenge/:id", updateChallengeHandle)
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
