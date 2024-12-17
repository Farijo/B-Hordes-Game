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
	}
	if mile.X.Valid {
		res["X"] = mile.X.Int16
	}
	if mile.Y.Valid {
		res["Y"] = mile.Y.Int16
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
	if mile.Map.Guide.Valid {
		res[lngData["guide"]] = mile.Map.Guide.Int16
	}
	if mile.Map.Shaman.Valid {
		res[lngData["shaman"]] = mile.Map.Shaman.Int16
	}
	if mile.Map.Custom.Valid {
		res[lngData["custom"]] = mile.Map.Custom.Bool
	}
	mapMap := make(map[string]any)
	if mile.Map.City.Door.Valid {
		mapMap[lngData["door"]] = mile.Map.City.Door.Bool
	}
	if mile.Map.City.Water.Valid {
		mapMap[lngData["water"]] = mile.Map.City.Water.Int16
	}
	if mile.Map.City.Chaos.Valid {
		mapMap[lngData["chaos"]] = mile.Map.City.Chaos.Bool
	}
	if mile.Map.City.Devast.Valid {
		mapMap[lngData["devast"]] = mile.Map.City.Devast.Bool
	}
	if mile.Map.City.Hard.Valid {
		mapMap[lngData["hard"]] = mile.Map.City.Hard.Bool
	}
	if mile.Map.City.X.Valid {
		mapMap["X"] = mile.Map.City.X.Int16
	}
	if mile.Map.City.Y.Valid {
		mapMap["Y"] = mile.Map.City.Y.Int16
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
	newsMap := make(map[string]any)
	if mile.Map.City.News.V.Z.Valid {
		newsMap[lngData["z"]] = mile.Map.City.News.V.Z.Int16
	}
	if mile.Map.City.News.V.Def.Valid {
		newsMap[lngData["def"]] = mile.Map.City.News.V.Def.Int16
	}
	if mile.Map.City.News.V.Water.Valid {
		newsMap[lngData["water"]] = mile.Map.City.News.V.Water.Int16
	}
	if mile.Map.City.News.V.RegenDir.Valid {
		newsMap[lngData["regenDir"]] = func(c1, c2 byte) string {
			if c1 > 3 || c2 > 2 {
				logger.Println("inconsistent direction", c1, c2)
				return ""
			}
			return []string{"est", "ouest", "nord", "sud"}[c1] + []string{"", "-est", "-ouest"}[c2]
		}(mile.Map.City.News.V.RegenDir.Byte&0xF0>>4, mile.Map.City.News.V.RegenDir.Byte&0x0F)
	}
	if len(newsMap) > 0 {
		mapMap[lngData["news"]] = newsMap
	}
	defMap := make(map[string]any)
	if mile.Map.City.Defense.Total.Valid {
		defMap[lngData["total"]] = mile.Map.City.Defense.Total.Int16
	}
	if mile.Map.City.Defense.Base.Valid {
		defMap[lngData["base"]] = mile.Map.City.Defense.Base.Int16
	}
	if mile.Map.City.Defense.Buildings.Valid {
		defMap[lngData["buildings"]] = mile.Map.City.Defense.Buildings.Int16
	}
	if mile.Map.City.Defense.Upgrades.Valid {
		defMap[lngData["upgrades"]] = mile.Map.City.Defense.Upgrades.Int16
	}
	if mile.Map.City.Defense.Items.Valid {
		defMap[lngData["items"]] = mile.Map.City.Defense.Items.Int16
	}
	if mile.Map.City.Defense.ItemsMul.Valid {
		defMap[lngData["itemsMul"]] = mile.Map.City.Defense.ItemsMul.V
	}
	if mile.Map.City.Defense.CitizenHomes.Valid {
		defMap[lngData["citizenHomes"]] = mile.Map.City.Defense.CitizenHomes.Int16
	}
	if mile.Map.City.Defense.CitizenGuardians.Valid {
		defMap[lngData["citizenGuardians"]] = mile.Map.City.Defense.CitizenGuardians.Int16
	}
	if mile.Map.City.Defense.Watchmen.Valid {
		defMap[lngData["watchmen"]] = mile.Map.City.Defense.Watchmen.Int16
	}
	if mile.Map.City.Defense.Souls.Valid {
		defMap[lngData["souls"]] = mile.Map.City.Defense.Souls.Int16
	}
	if mile.Map.City.Defense.Temp.Valid {
		defMap[lngData["temp"]] = mile.Map.City.Defense.Temp.Int16
	}
	if mile.Map.City.Defense.Cadavers.Valid {
		defMap[lngData["cadavers"]] = mile.Map.City.Defense.Cadavers.Int16
	}
	if mile.Map.City.Defense.Bonus.Valid {
		defMap[lngData["bonus"]] = mile.Map.City.Defense.Bonus.V
	}
	if len(defMap) > 0 {
		mapMap[lngData["def"]] = defMap
	}
	if mile.Map.City.Upgrades.V.List.Valid {
		buildingsMap := make(map[string]uint32)
		for buildingKey, v := range mile.Map.City.Upgrades.V.List.Data {
			_, name := getServerDataKey(buildingKey, "buildings", userkey, lng)
			buildingsMap[name] = v
		}
		if len(buildingsMap) > 0 {
			res[lngData["upgrades"]] = buildingsMap
		}
	}
	estimMap := make(map[string]any)
	if mile.Map.City.Estimations.V.Min.Valid {
		estimMap[lngData["min"]] = mile.Map.City.Estimations.V.Min.Int16
	}
	if mile.Map.City.Estimations.V.Max.Valid {
		estimMap[lngData["max"]] = mile.Map.City.Estimations.V.Max.Int16
	}
	if mile.Map.City.Estimations.V.Maxed.Valid {
		estimMap[lngData["maxed"]] = mile.Map.City.Estimations.V.Maxed.Bool
	}
	if len(estimMap) > 0 {
		res[lngData["estimation"]] = estimMap
	}
	estimNextMap := make(map[string]any)
	if mile.Map.City.EstimationsNext.V.Min.Valid {
		estimNextMap[lngData["min"]] = mile.Map.City.EstimationsNext.V.Min.Int16
	}
	if mile.Map.City.EstimationsNext.V.Max.Valid {
		estimNextMap[lngData["max"]] = mile.Map.City.EstimationsNext.V.Max.Int16
	}
	if mile.Map.City.EstimationsNext.V.Maxed.Valid {
		estimNextMap[lngData["maxed"]] = mile.Map.City.EstimationsNext.V.Maxed.Bool
	}
	if len(estimNextMap) > 0 {
		res[lngData["estimationNext"]] = estimNextMap
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
