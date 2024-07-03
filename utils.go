package main

import (
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

func genFindBestAcceptedLng(languages []language.Tag) func(*gin.Context, string) string {
	return func(context *gin.Context, defaultLng string) string {
		if found := context.GetString("lng found"); found > "" {
			return found
		}
		lngSplited := strings.Split(context.GetHeader("Accept-Language"), ",")
		qualityValueList := make([][]string, len(lngSplited))
		for i, lngQValued := range lngSplited {
			qualityValueList[i] = strings.Split(lngQValued, ";q=")
		}

		// sort in descending order of quality values
		sort.Slice(qualityValueList, func(i, j int) bool {
			if len(qualityValueList[i]) < 2 { // no quality value, consider default value (ie 1)
				return true
			}
			if len(qualityValueList[j]) < 2 { // same for j
				return false
			}
			dotIdxI, dotIdxJ := strings.IndexByte(qualityValueList[i][1], '.'), strings.IndexByte(qualityValueList[j][1], '.')
			if dotIdxI < 0 { // no decimal, consider default value (ie 1)
				return true
			}
			if dotIdxJ < 0 { // same for j
				return false
			}
			lenI, lenJ := len(qualityValueList[i][1]), len(qualityValueList[j][1])
			for {
				dotIdxI++
				dotIdxJ++
				if dotIdxI >= lenI { // i ends so is either minus or equal to j
					return false
				}
				if dotIdxJ >= lenJ { // same for j
					return true
				}
				if qualityValueList[i][1][dotIdxI] != qualityValueList[j][1][dotIdxJ] {
					return qualityValueList[j][1][dotIdxJ] < qualityValueList[i][1][dotIdxI]
				}
			}
		})

		// parse the lng from the highest quality value to the lowest until we find an accepted lng
		for _, lng := range qualityValueList {
			lng := language.Make(strings.TrimSpace(lng[0]))
			for _, acceptedLang := range languages {
				if lng == acceptedLang {
					context.Set("lng found", defaultLng) //acceptedLang.String())
					return defaultLng                    //acceptedLang.String()
				}
			}
		}

		context.Set("lng found", defaultLng)
		return defaultLng
	}
}
