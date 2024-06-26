package main

import (
	"bhordesgame/dto"
	"embed"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-contrib/gzip"
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

func Ignore[T any](t T, e error) T {
	return t
}

//go:embed gen/* favicon.ico
var f embed.FS

func main() {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.SetHTMLTemplate(Must(template.New("").Funcs(template.FuncMap{
		"getAccess": getAccess,
		"getStatus": getStatus,
		"dumpStruct": func(strct *dto.Goal) template.JS {
			rep := `"`
			strct.Custom.String = rep + html.EscapeString(strct.Custom.String) + rep
			return template.JS(strings.ReplaceAll(fmt.Sprintf("%+v", *strct), " ", ","))
		},
		"dumpMile": dumpMile,
		"incr": func(i int) int {
			return i + 1
		},
		"decodeGoal": decodeGoal,
		"mkmap":      mkmap,
	}).ParseFS(f, "gen/templates/*.html")))
	r.StaticFS("/style", http.FS(Must(fs.Sub(f, "gen/style"))))
	r.StaticFS("/script", http.FS(Must(fs.Sub(f, "gen/script"))))
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
		authorized.POST("/challenge/:id/date", challengeDateHandle)
		authorized.POST("/challenge/:id/scan", challengeScanHandle)
		authorized.GET("/validation", validationHandle)
		authorized.POST("/validation", validateGoalHandle)
	}

	if gin.Mode() == gin.DebugMode {
		domain = "localhost"
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
