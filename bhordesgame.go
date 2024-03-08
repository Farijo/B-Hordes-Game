package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func pop[T any](a *[]T) T {
	rv := (*a)[0]
	*a = (*a)[1:]
	return rv
}

func Must[T any](t T, e error) T {
	if e != nil {
		panic(e.Error())
	}
	return t
}

//go:embed script style templates/* favicon.ico
var f embed.FS

func main() {
	r := gin.Default()
	t := template.Must(template.New("").Funcs(template.FuncMap{
		"getAccess":    getAccess,
		"getStatus":    getStatus,
		"fDate":        fdate,
		"templateHTML": func(a string) template.HTML { return template.HTML(a) },
	}).ParseFS(f, "templates/*.html"))
	r.SetHTMLTemplate(t)
	r.StaticFS("/style", http.FS(Must(fs.Sub(f, "style"))))
	r.StaticFS("/script", http.FS(Must(fs.Sub(f, "script"))))
	r.StaticFileFS("/favicon.ico", "favicon.ico", http.FS(f))
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
		authorized.POST("/challenge/:id/members", challengeMembersHandle)
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
