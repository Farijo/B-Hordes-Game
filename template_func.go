package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

var translations map[string]map[string]string

func loadTranslations(fs fs.FS, langs []language.Tag) {
	translations = make(map[string]map[string]string)
	for _, lng := range langs {
		lng := lng.String()
		translations[lng] = make(map[string]string)
		f, err := fs.Open("lang/" + lng + ".yaml")
		if err != nil {
			fmt.Println("loadTranslations ", err)
		} else {
			if err := yaml.NewDecoder(f).Decode(translations[lng]); err != nil {
				fmt.Println("loadTranslations ", err)
			}
			if err := f.Close(); err != nil {
				fmt.Println("loadTranslations ", err)
			}
		}
	}
}

func getAccess(lang string) []string {
	res := []string{"open-to-all", "on-request", "on-invite"}
	for i, k := range res {
		res[i] = translations[lang][k]
	}
	return res
}

func getStatus(lang string) []string {
	res := []string{"creation", "proofreading", "inscriptions", "running", "over"}
	for i, k := range res {
		res[i] = translations[lang][k]
	}
	return res
}

func dumpStruct(strct *dto.Goal) template.JS {
	rep := `"`
	strct.Custom.String = rep + html.EscapeString(strct.Custom.String) + rep
	return template.JS(strings.ReplaceAll(fmt.Sprintf("%+v", *strct), " ", ","))
}

type GoalHTML struct {
	Text        template.HTML
	Icon, Label string
}

type GoalHeader struct {
	Content template.HTML
	Amount  uint32
}

func decodeGoal(key, lng string, goal *dto.Goal, l map[int]GoalHeader) GoalHTML {
	amountStr, header := "le plus de", "+"
	if goal.Amount.Valid {
		amountStr = strconv.Itoa(int(goal.Amount.Int32))
		header = amountStr + " "
	}
	var out GoalHTML
	switch goal.Typ {
	case 0:
		out.Text = template.HTML(fmt.Sprintf("Accumuler <b>%s</b> pictos", amountStr))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "pictos", key, lng)
		header += "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
	case 1:
		out.Icon, out.Label = getServerDataKey(goal.Entity, "items", key, lng)
		var txt string
		if goal.X.Valid {
			if goal.Y.Valid {
				txt = fmt.Sprintf("Etre sur la <b>case</b> [ <b>%d</b> / <b>%d</b> ] de l'OM avec <b>%s</b>", goal.X.Int16, goal.Y.Int16, amountStr)
				header = fmt.Sprintf("[%d/%d] %s<img src=\"https://myhordes.eu/build/images/%s\">", goal.X.Int16, goal.Y.Int16, header, out.Icon)
			} else {
				txt = fmt.Sprintf("Etre sur la <b>ligne %d</b> de l'OM avec <b>%s</b>", goal.X.Int16, amountStr)
				header = fmt.Sprintf("[%d/_] %s<img src=\"https://myhordes.eu/build/images/%s\">", goal.X.Int16, header, out.Icon)
			}
		} else {
			if goal.Y.Valid {
				txt = fmt.Sprintf("Etre sur la <b>colonne %d</b> de l'OM avec <b>%s</b>", goal.Y.Int16, amountStr)
				header = fmt.Sprintf("[_/%d] %s<img src=\"https://myhordes.eu/build/images/%s\">", goal.Y.Int16, header, out.Icon)
			} else {
				txt = fmt.Sprintf("Etre dans l'OM avec <b>%s</b>", amountStr)
				header = fmt.Sprintf("[_/_] %s<img src=\"https://myhordes.eu/build/images/%s\">", header, out.Icon)
			}
		}
		out.Text = template.HTML(txt)
	case 2:
		out.Text = "Construire"
		out.Icon, out.Label = getServerDataKey(goal.Entity, "buildings", key, lng)
		header = "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		goal.Amount.Int32 = 1
	case 3:
		out.Text = template.HTML(fmt.Sprintf("Avoir en banque <b>%s</b>", amountStr))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "items", key, lng)
		header += "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
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
	lngData := translations[lng]
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
		buildingsMap := make([]string, len(mile.Map.City.Buildings.Data))
		i := 0
		for k := range mile.Map.City.Buildings.Data {
			_, buildingsMap[i] = getServerDataKey(k, "buildings", userkey, lng)
			i++
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
	s, _ := json.MarshalIndent(res, "", "&nbsp;&nbsp;")
	return template.HTML(strings.ReplaceAll(string(s), "\n", "<br>"))
}
