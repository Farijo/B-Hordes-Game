package main

import (
	"embed"
	"html/template"
	"log"
	"os"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
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

var availableLangs = []language.Tag{language.English, language.French, language.Spanish, language.German}

//go:embed lang templates/*
var f embed.FS
var logger *log.Logger

func main() {
	logger = log.New(os.Stderr, "[BHG] ", log.Ldate|log.Ltime|log.Llongfile)

	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.SetHTMLTemplate(Must(template.New("").Funcs(template.FuncMap{
		"getAccess":  getAccess,
		"getStatus":  getStatus,
		"getRoles":   getRoles,
		"dumpStruct": dumpStruct,
		"dumpMile":   dumpMile,
		"incr":       func(i int) int { return i + 1 },
		"decodeGoal": decodeGoal,
		"mkmap":      mkmap,
	}).ParseFS(f, "templates/*.html")))
	r.Static("/style", "style")
	r.Static("/script", "script")
	r.StaticFile("/favicon.ico", "asset/favicon.ico")
	r.StaticFile("/question.svg", "asset/question.svg")
	r.StaticFile("/gear.svg", "asset/gear.svg")

	loadTranslations(f, availableLangs)
	lngHandler := func(ctx *gin.Context) { ctx.Set(LNG_KEY, "@@{ISO639-1}") }

	r.POST("/", connectionHandle)
	r.GET("/", lngHandler, indexHandle)
	r.POST("/settings", settingsHandle)
	r.GET("/logout", logoutHandle)
	r.POST("/user", refreshHandle)
	r.GET("/user/:id", lngHandler, userHandle)
	r.GET("/users/"+os.Getenv("USER_KEY"), userInfoActualizerHandle)

	restricted := r.Group("/challenge/:id")
	restricted.Use(restrictedChallenge)
	{
		restricted.GET("", lngHandler, challengeHandle)
		restricted.GET("/graph", lngHandler, challengeGraphHandle)
		restricted.GET("/history", lngHandler, challengeHistoryHandle)
		restricted.GET("/data", lngHandler, challengeRawHistoryHandle)
	}

	authorized := r.Group("/")
	authorized.Use(requireAuth)
	{
		authorized.GET("/user", lngHandler, selfHandle)
		authorized.POST("/challenge", createChallengeHandle)
		authorized.GET("/challenge", lngHandler, challengeCreationHandle)
		authorized.POST("/challenge/:id", updateChallengeHandle)
		authorized.POST("/challenge/:id/members", challengeMembersHandle)
		authorized.POST("/challenge/:id/date", challengeDateHandle)
		authorized.POST("/challenge/:id/scan", challengeScanHandle)
		authorized.POST("/challenge/:id/back", challengeCancelStartHandle)
		authorized.GET("/validation", lngHandler, validationHandle)
		authorized.POST("/validation", validateGoalHandle)
		authorized.POST("/validation/archive", archiveValidationHandle)
		authorized.GET("/milestone", lngHandler, milestoneHandle)
	}

	if gin.Mode() == gin.DebugMode {
		domain = "localhost"
	}

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
