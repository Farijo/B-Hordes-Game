package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

func getAccess() []string {
	return []string{"open-to-all", "on-request", "on-invite"}
}

func getStatus() []string {
	return []string{"creation", "proofreading", "inscriptions", "running", "over"}
}

type GoalHTML struct {
	Text        template.HTML
	Icon, Label string
}

type GoalHeader struct {
	Content template.HTML
	Amount  uint32
}

func decodeGoal(key string, goal *dto.Goal, l map[int]GoalHeader) GoalHTML {
	amountStr, header := "le plus de", "+"
	if goal.Amount.Valid {
		amountStr = strconv.Itoa(int(goal.Amount.Int32))
		header = amountStr + " "
	}
	var out GoalHTML
	switch goal.Typ {
	case 0:
		out.Text = template.HTML(fmt.Sprintf("Accumuler <b>%s</b> pictos", amountStr))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "pictos", key)
		header += "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
	case 1:
		out.Icon, out.Label = getServerDataKey(goal.Entity, "items", key)
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
		out.Icon, out.Label = getServerDataKey(goal.Entity, "buildings", key)
		header = "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		goal.Amount.Int32 = 1
	case 3:
		out.Text = template.HTML(fmt.Sprintf("Avoir en banque <b>%s</b>", amountStr))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "items", key)
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

func dumpMile(mile *dto.Milestone, userkey string) template.HTML {
	res := make(map[string]any, 0)
	if mile.IsGhost.Valid {
		res["Fantôme"] = mile.IsGhost.Bool
	}
	if mile.PlayedMaps.Valid {
		res["Nombre d'incarnation"] = mile.PlayedMaps.Int64
	}
	if mile.Rewards.Valid {
		pictoMap := make(map[string]uint32)
		for k, v := range mile.Rewards.Data {
			_, n := getServerDataKey(k, "pictos", userkey)
			pictoMap[n] = v
		}
		if len(pictoMap) > 0 {
			res["Pictos"] = pictoMap
		}
	}
	if mile.Dead.Valid {
		res["Mort"] = mile.Dead.Bool
	}
	if mile.Out.Valid {
		res["Dehors"] = mile.Out.Bool
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
		res["Banni"] = mile.Ban.Bool
	}
	if mile.BaseDef.Valid {
		res["Défense de maison"] = mile.BaseDef.Byte
	}
	if mile.Job.Valid {
		res["Métier"] = []string{"Habitant", "Fouineur", "Éclaireur", "Gardien", "Ermite", "Apprivoiseur", "Technicien"}[mile.Job.Byte]
	}
	mapMap := make(map[string]any)
	if mile.Map.Wid.Valid {
		mapMap["Largeur"] = mile.Map.Wid.Byte
	}
	if mile.Map.Hei.Valid {
		mapMap["Hauteur"] = mile.Map.Hei.Byte
	}
	if mile.Map.Days.Valid {
		mapMap["Jours"] = mile.Map.Days.Byte
	}
	if mile.Map.Conspiracy.Valid {
		mapMap["Insurection"] = mile.Map.Conspiracy.Bool
	}
	if mile.Map.Custom.Valid {
		mapMap["Custom"] = mile.Map.Custom.Bool
	}
	if mile.Map.City.Buildings.Valid {
		buildingsMap := make([]string, len(mile.Map.City.Buildings.Data))
		i := 0
		for k := range mile.Map.City.Buildings.Data {
			_, buildingsMap[i] = getServerDataKey(k, "buildings", userkey)
			i++
		}
		if len(buildingsMap) > 0 {
			mapMap["Chantiers"] = buildingsMap
		}
	}
	if mile.Map.City.Bank.Valid {
		itemMap := make(map[string]uint32)
		for k, v := range mile.Map.City.Bank.Data {
			_, n := getServerDataKey(k, "items", userkey)
			itemMap[n] = v
		}
		if len(itemMap) > 0 {
			mapMap["Banque"] = itemMap
		}
	}
	if len(mapMap) > 0 {
		res["Ville"] = mapMap
	}
	if mile.Map.Zones.Valid {
		itemMap := make(map[string]uint32)
		for k, v := range mile.Map.Zones.Data {
			_, n := getServerDataKey(k, "items", userkey)
			itemMap[n] = v
		}
		if len(itemMap) > 0 {
			res["Objets au sol"] = itemMap
		}
	}
	s, _ := json.MarshalIndent(res, "", "&nbsp;&nbsp;")
	return template.HTML(strings.ReplaceAll(string(s), "\n", "<br>"))
}
