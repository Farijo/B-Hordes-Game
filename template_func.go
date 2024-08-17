package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var translations map[string]map[string]string

func loadTranslations(fs fs.FS, langs []language.Tag) {
	translations = make(map[string]map[string]string, len(langs))
	for _, lng := range langs {
		lng := lng.String()
		translations[lng] = make(map[string]string)
		f, err := fs.Open("lang/" + lng + ".yaml")
		if err != nil {
			logger.Println("loadTranslations ", err)
		} else {
			if err := yaml.NewDecoder(f).Decode(translations[lng]); err != nil {
				logger.Println("loadTranslations ", err)
			}
			if err := f.Close(); err != nil {
				logger.Println("loadTranslations ", err)
			}
		}
	}
}

func getTrad(lang string) map[string]string {
	lngData, ok := translations[lang]
	if ok {
		return lngData
	} else {
		return translations[availableLangs[0].String()]
	}
}

func getAccess(lang string) []string {
	lngData := getTrad(lang)
	return []string{
		lngData["open-to-all"],
		lngData["on-request"],
		lngData["on-invite"],
	}
}

func getStatus(lang string) []string {
	lngData := getTrad(lang)
	return []string{
		lngData["creation"],
		lngData["proofreading"],
		lngData["inscriptions"],
		lngData["running"],
		lngData["over"],
	}
}

func getRoles(lang string) []string {
	lngData := getTrad(lang)
	return []string{
		lngData["creator"],
		lngData["participant"],
		lngData["guest"],
		lngData["candidate"],
		lngData["approbator"],
	}
}

func dumpStruct(strct *dto.Goal) template.JS {
	var builder strings.Builder
	json.NewEncoder(&builder).Encode(strct)
	return template.JS(builder.String())
}

type GoalHTML struct {
	Text        template.HTML
	Icon, Label string
}

type GoalHeader struct {
	Content template.HTML
	Amount  uint32
}

func decodeGoal(key, lang string, goal *dto.Goal, l map[int]GoalHeader, withTag bool) GoalHTML {
	lngData := getTrad(lang)
	amountStr, header := lngData["the-most-of"], "+"
	if goal.Amount.Valid {
		amountStr = strconv.Itoa(int(goal.Amount.Int32))
		header = amountStr + " "
	}
	var out GoalHTML
	switch goal.Typ {
	case 0:
		out.Text = template.HTML(fmt.Sprintf(lngData["goal-stack-rewards"], amountStr))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "pictos", key, lang)
		if withTag {
			header += "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		}
	case 1:
		out.Icon, out.Label = getServerDataKey(goal.Entity, "items", key, lang)
		tagIcon := ""
		if withTag {
			tagIcon = "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		}
		var txt string
		if goal.X.Valid {
			if goal.Y.Valid {
				txt = fmt.Sprintf(lngData["goal-stand-x-y"], goal.X.Int16, goal.Y.Int16, amountStr)
				header = fmt.Sprintf("[%d/%d] %s%s", goal.X.Int16, goal.Y.Int16, header, tagIcon)
			} else {
				txt = fmt.Sprintf(lngData["goal-stand-x"], goal.X.Int16, amountStr)
				header = fmt.Sprintf("[%d/_] %s%s", goal.X.Int16, header, tagIcon)
			}
		} else {
			if goal.Y.Valid {
				txt = fmt.Sprintf(lngData["goal-stand-y"], goal.Y.Int16, amountStr)
				header = fmt.Sprintf("[_/%d] %s%s", goal.Y.Int16, header, tagIcon)
			} else {
				txt = fmt.Sprintf(lngData["goal-stand"], amountStr)
				header = fmt.Sprintf("[_/_] %s%s", header, tagIcon)
			}
		}
		out.Text = template.HTML(txt)
	case 2:
		out.Text = template.HTML(lngData["goal-build"])
		out.Icon, out.Label = getServerDataKey(goal.Entity, "buildings", key, lang)
		if withTag {
			header = "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		} else {
			header = ""
		}
		goal.Amount.Int32 = 1
	case 3:
		out.Text = template.HTML(fmt.Sprintf(lngData["goal-bank"], amountStr))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "items", key, lang)
		if withTag {
			header += "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		}
	case 4:
		out.Label = goal.Custom.String
		header = "-"
	}
	if l != nil {
		l[goal.ID] = GoalHeader{template.HTML(header), uint32(goal.Amount.Int32)}
	}
	return out
}

func mkmap() map[int]GoalHeader {
	return make(map[int]GoalHeader, 0)
}

func dumpMile(mile *dto.Milestone, userkey, lng string) template.HTML {
	lngData := getTrad(lng)
	res := make(map[string]any, 0)
	if mile.IsGhost.Valid {
		res[lngData["ghost"]] = mile.IsGhost.Bool
	}
	if mile.PlayedMaps.Valid {
		res[lngData["played-maps"]] = mile.PlayedMaps.Int64
	}
	if mile.Rewards.Valid {
		pictoMap := make(map[string]uint32)
		for k, v := range mile.Rewards.Data {
			_, n := getServerDataKey(k, "pictos", userkey, lng)
			pictoMap[n] = v
		}
		if len(pictoMap) > 0 {
			res[lngData["rewards"]] = pictoMap
		}
	}
	if mile.Dead.Valid {
		res[lngData["dead"]] = mile.Dead.Bool
	}
	if mile.Out.Valid {
		res[lngData["out"]] = mile.Out.Bool
		if mile.Out.Bool {
			if mile.X.Valid {
				res["X"] = mile.X.Int16
			}
			if mile.Y.Valid {
				res["Y"] = mile.Y.Int16
			}
		}
	}
	if mile.Ban.Valid {
		res[lngData["ban"]] = mile.Ban.Bool
	}
	if mile.BaseDef.Valid {
		res[lngData["base-def"]] = mile.BaseDef.Byte
	}
	if mile.Job.Valid {
		res[lngData["job"]] = lngData[[]string{"basic", "dig", "vest", "shield", "book", "tamer", "tech"}[mile.Job.Byte]]
	}
	mapMap := make(map[string]any)
	if mile.Map.Wid.Valid {
		res[lngData["wid"]] = mile.Map.Wid.Byte
	}
	if mile.Map.Hei.Valid {
		res[lngData["hei"]] = mile.Map.Hei.Byte
	}
	if mile.Map.Days.Valid {
		res[lngData["days"]] = mile.Map.Days.Byte
	}
	if mile.Map.Conspiracy.Valid {
		res[lngData["conspiracy"]] = mile.Map.Conspiracy.Bool
	}
	if mile.Map.Custom.Valid {
		res[lngData["custom"]] = mile.Map.Custom.Bool
	}
	if mile.Map.City.Buildings.Valid {
		buildingsMap := make(map[string]bool, len(mile.Map.City.Buildings.Data))
		for buildingKey, isBuilt := range mile.Map.City.Buildings.Data {
			_, name := getServerDataKey(buildingKey, "buildings", userkey, lng)
			buildingsMap[name] = isBuilt
		}
		if len(buildingsMap) > 0 {
			res[lngData["buildings"]] = buildingsMap
		}
	}
	if mile.Map.City.Bank.Valid {
		itemMap := make(map[string]uint32)
		for k, v := range mile.Map.City.Bank.Data {
			_, n := getServerDataKey(k, "items", userkey, lng)
			itemMap[n] = v
		}
		if len(itemMap) > 0 {
			res[lngData["bank"]] = itemMap
		}
	}
	if len(mapMap) > 0 {
		res[lngData["map"]] = mapMap
	}
	if mile.Map.Zones.Valid {
		itemMap := make(map[string]uint32)
		for k, v := range mile.Map.Zones.Data {
			_, n := getServerDataKey(k, "items", userkey, lng)
			itemMap[n] = v
		}
		if len(itemMap) > 0 {
			res[lngData["items"]] = itemMap
		}
	}
	var builder strings.Builder
	enc := json.NewEncoder(&builder)
	enc.SetIndent("</span><br><span>", "&nbsp;&nbsp;")
	enc.Encode(res)
	return template.HTML("<span>" + builder.String() + "</span>")
}
