package main

import (
	"bhordesgame/dto"
	"embed"
	"fmt"
	"html"
	"html/template"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
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

var acceptedLanguages = []language.Tag{language.French, language.English}

//go:embed gen/* lang/* templates/*
var f embed.FS

func main() {
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(i18n.Localize(i18n.WithBundle(&i18n.BundleCfg{
		RootPath:         "lang",
		AcceptLanguage:   acceptedLanguages,
		DefaultLanguage:  language.English,
		UnmarshalFunc:    yaml.Unmarshal,
		FormatBundleFile: "yaml",
		Loader:           &i18n.EmbedLoader{FS: f},
	}), i18n.WithGetLngHandle(genFindBestAcceptedLng(acceptedLanguages))))
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
		"translate":  i18n.GetMessage,
	}).ParseFS(f, "templates/*.html")))
	r.Static("/style", "style")
	r.Static("/script", "script")
	r.StaticFile("/question-mark.svg", "asset/question-mark.svg")
	r.StaticFile("/favicon.ico", "asset/favicon.ico")
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
