package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"strconv"
	"strings"

	"github.com/gin-contrib/i18n"
	"github.com/gin-gonic/gin"
)

func getAccess() []string {
	return []string{"open-to-all", "on-request", "on-invite"}
}

func getStatus() []string {
	return []string{"creation", "proofreading", "inscriptions", "running", "over"}
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

func decodeGoal(ctx *gin.Context, key string, goal *dto.Goal, l map[int]GoalHeader) GoalHTML {
	lang := i18n.MustGetMessage(ctx, "mh-lang")
	amountStr, header := "le plus de", "+"
	if goal.Amount.Valid {
		amountStr = strconv.Itoa(int(goal.Amount.Int32))
		header = amountStr + " "
	}
	var out GoalHTML
	switch goal.Typ {
	case 0:
		out.Text = template.HTML(fmt.Sprintf("Accumuler <b>%s</b> pictos", amountStr))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "pictos", key, lang)
		header += "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
	case 1:
		out.Icon, out.Label = getServerDataKey(goal.Entity, "items", key, lang)
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
		out.Text = template.HTML(i18n.MustGetMessage(ctx, "build"))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "buildings", key, lang)
		header = "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		goal.Amount.Int32 = 1
	case 3:
		out.Text = template.HTML(fmt.Sprintf(i18n.MustGetMessage(ctx, "have-in-bank")+" <b>%s</b>", amountStr))
		out.Icon, out.Label = getServerDataKey(goal.Entity, "items", key, lang)
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

type yn struct {
	yes, no string
}

func (i yn) boolToStr(b bool) string {
	if b {
		return i.yes
	} else {
		return i.no
	}
}

var jobList = []string{"basic", "dig", "vest", "shield", "book", "tamer", "tech"}

func dumpMile(ctx *gin.Context, mile *dto.Milestone, userkey string) template.HTML {
	lang := i18n.MustGetMessage(ctx, "mh-lang")
	conv := yn{i18n.MustGetMessage(ctx, "yes"), i18n.MustGetMessage(ctx, "no")}
	res := make(map[string]any, 0)
	if mile.IsGhost.Valid {
		res[i18n.MustGetMessage(ctx, "ghost")] = conv.boolToStr(mile.IsGhost.Bool)
	}
	if mile.PlayedMaps.Valid {
		res[i18n.MustGetMessage(ctx, "played-maps")] = mile.PlayedMaps.Int64
	}
	if mile.Rewards.Valid {
		pictoMap := make(map[string]uint32)
		for k, v := range mile.Rewards.Data {
			_, n := getServerDataKey(k, "pictos", userkey, lang)
			pictoMap[n] = v
		}
		if len(pictoMap) > 0 {
			res[i18n.MustGetMessage(ctx, "rewards")] = pictoMap
		}
	}
	if mile.Dead.Valid {
		res[i18n.MustGetMessage(ctx, "dead")] = conv.boolToStr(mile.Dead.Bool)
	}
	if mile.Out.Valid {
		res[i18n.MustGetMessage(ctx, "out")] = conv.boolToStr(mile.Out.Bool)
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
		res[i18n.MustGetMessage(ctx, "ban")] = conv.boolToStr(mile.Ban.Bool)
	}
	if mile.BaseDef.Valid {
		res[i18n.MustGetMessage(ctx, "base-def")] = mile.BaseDef.Byte
	}
	if mile.Job.Valid {
		res[i18n.MustGetMessage(ctx, "job")] = i18n.MustGetMessage(ctx, jobList[mile.Job.Byte])
	}
	mapMap := make(map[string]any)
	if mile.Map.Wid.Valid {
		mapMap[i18n.MustGetMessage(ctx, "wid")] = mile.Map.Wid.Byte
	}
	if mile.Map.Hei.Valid {
		mapMap[i18n.MustGetMessage(ctx, "hei")] = mile.Map.Hei.Byte
	}
	if mile.Map.Days.Valid {
		mapMap[i18n.MustGetMessage(ctx, "days")] = mile.Map.Days.Byte
	}
	if mile.Map.Conspiracy.Valid {
		mapMap[i18n.MustGetMessage(ctx, "conspiracy")] = conv.boolToStr(mile.Map.Conspiracy.Bool)
	}
	if mile.Map.Custom.Valid {
		mapMap[i18n.MustGetMessage(ctx, "custom")] = conv.boolToStr(mile.Map.Custom.Bool)
	}
	if mile.Map.City.Buildings.Valid {
		buildingsMap := make([]string, len(mile.Map.City.Buildings.Data))
		i := 0
		for k := range mile.Map.City.Buildings.Data {
			_, buildingsMap[i] = getServerDataKey(k, "buildings", userkey, lang)
			i++
		}
		if len(buildingsMap) > 0 {
			mapMap[i18n.MustGetMessage(ctx, "buildings")] = buildingsMap
		}
	}
	if mile.Map.City.Bank.Valid {
		itemMap := make(map[string]uint32)
		for k, v := range mile.Map.City.Bank.Data {
			_, n := getServerDataKey(k, "items", userkey, lang)
			itemMap[n] = v
		}
		if len(itemMap) > 0 {
			mapMap[i18n.MustGetMessage(ctx, "bank")] = itemMap
		}
	}
	if len(mapMap) > 0 {
		res[i18n.MustGetMessage(ctx, "map")] = mapMap
	}
	if mile.Map.Zones.Valid {
		itemMap := make(map[string]uint32)
		for k, v := range mile.Map.Zones.Data {
			_, n := getServerDataKey(k, "items", userkey, lang)
			itemMap[n] = v
		}
		if len(itemMap) > 0 {
			res[i18n.MustGetMessage(ctx, "items")] = itemMap
		}
	}
	s, _ := json.MarshalIndent(res, "", "&nbsp;&nbsp;")
	return template.HTML(strings.ReplaceAll(string(s), "\n", "<br>"))
}
