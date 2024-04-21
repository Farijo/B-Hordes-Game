package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

func getAccess() [][]string {
	return [][]string{
		{"Ouvert à tous", "R0lGODlhEAAQANU5AAAAAHx8fGRBMXtSQWpFNVw6K5poU4RYRqJuWLB6YpKSkmxGNmM/MbuAaFU1J5OTk59sV6NvWGhDM8SIbsGHbIqKirN7Y55rVo5fS2NjY29JOVg5Ko+Pj8zMzEcqHmZCMl89LphoU6ZwWkUqHsOHbqx2X6dyW72Ea3ZOPahzXK94YLyDanNzc08xJadzXLmAZ1EyJax3X6JvWGtra8eKb5NjT9TU1ISEhG5JN35NKgAAAAAAAAAAAAAAAAAAAAAAACH5BAEAADkALAAAAAAQABAAAAaFwJxwSCzmAEikMZk80RqA5YqEQE4iG+UQ+SoJkBQMzBYtAzqWwRfQOLQC0S3LhSIgTQQHnAi4xTRrKQIecGVHFSoEdgAiDHpxQgAcCR8DSDIMI3tmCgkCiwgSAoVHRwoGBZYABgujTEgPqKoQCweFTA8hqUgXOAcZrwAzNSAQwcdMRspDQQA7"},
		{"Sur demande", "R0lGODlhEAAQAOZlAAAAAGRBMWRkZGpFNXtSQYRYRppoU5OTk3x8fFw6K2BgYGtra2xGNi0tLZ9sV3Nzc5aWllU1J6enp5qamk1NTWxsbLB6YjExMZKSkjIyMqJuWDc3Ny4uLmM/MbuAaG9vb1paWkZGRjk5OadzXB8fH2hDM45fS5SUlB4eHq94YKhzXLGxsZiYmE8xJV89LkJCQr2Ea3ZOPdnZ2bN7Y7mAZ6urq8TExMSIbsGHbFEyJceKb9PT05hoU4SEhKdyW6NvWKJvWKZwWnt7e0cqHmNjY29JOaqqqmZCMm5JN7yDajYrGpNjT8zMzCMjI+Tk5MOHbpubm4qKiqampp5rVk9PT6x2X1g5KkUqHn5+fhoaGo6OjmZmZoKCgoiIiEVFRTU1NaSkpF9fX9TU1Kx3X4+Pj35NKgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACH5BAEAAGUALAAAAAAQABAAAAe8gGWCg4SFZQCIiIaJiTA6HgCLSU8aiDc/VoyMNFUBiDgmOV0rYk4yNjtMMwSeAB4FLSIgCFwPYFgPIzEDiD4DERcKQgcCUh89Y0WtKgFDHAoILAISH1EpA7wAQR0RDdATAkYVZBZHBIhAHVcNWwIQYTUgGBYB2RolARwLFFpeEi8YDCQ4B8AAgwAZKoSA0GRClgMCCTpgUGCDgA0nMkBBcYDHQERTkBRQ8oUElQtEKCxY4sKBppcvDckcFAgAOw=="},
		{"Sur invitation", "R0lGODlhEAAQAKIGAP/fkgAAAMOcTP/7ppZzLv//535NKgAAACH5BAEAAAYALAAAAAAQABAAAANIaLrc/jCyQKu1KojBB/CgEGTDBpxoCYxG4BFeSpysexIFV8w0iQqFggC18gGGAiSq5hkSj73WgPcEzGqEpHYrINQuYIpkTE4AADs="},
	}
}

func getStatus() [][]string {
	return [][]string{
		{"Création", "R0lGODlhEAAQALMAAAAAAP/Ge+SYW5tEIcBpO7qzlezjvXUzGdPLqZKNdem5ZbxmKn5NKgAAAAAAAAAAACH5BAEAAAwALAAAAAAQABAAAARWkMnJgL2AUpCMSYCRaRViIkUhkkDRpuCqYcOAkROwKCMuXTzfjyCw9ViCgBIgOP6YgWLzMlMGCINmwFZVCg6H28YaOAyUBCeATAhTNxbi4EB05jAWTQQAOw=="},
		{"Relecture", "R0lGODlhEAAQANUuAAAAABkNB3snGt25kIlGHt3VsqrC/zIIBf/s1nqMvqVlOU4YEk8ZDsSPYJKo5PDAVZpoU////5FPJty4j4RWRcONXXqPwsZpKGZCM1MTDKukiVQUDcmXZ9GjdanB/aehicSIbnZNPcS8n9myh6FhNXeKvmQbElISDGl9r86gcWgdE2x+rdrSscqYaX5NKgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACH5BAEAAC4ALAAAAAAQABAAAAaTQJdwSAQYiUKj0lhoAooDgEKgABQS2MLTBRg0AIwFgGUxmBNarpJBAIhK5ogjLWUowgDNyhMxzJ8ACAMTU20fKAl/SQ0CAgsLBG0PDxdbXIIDFQsHBwQBGAFFCI0pDScSniGhQwCMAwImGZ0BFKtJowMjHRycARC2l40CKhuzIMCBA8oCLSSeD8hLS77ASEQB2ENBADs="},
		{"Inscriptions", "R0lGODlhEAAQALMLAAAAAHZNMv+4ipJgQVo6JK10Uv/BkuKbcvyugv+3iseHYn5NKgAAAAAAAAAAAAAAACH5BAEAAAsALAAAAAAQABAAAARRcMlJq704T8C7/5xhJAaCHAdiDEOhhKRwomrrdkUAEASX7z3AAnDQDYJFwFE4FOgCQScAygRIBzqrsSfRKrNXXTfMIY91P+9vI+R020yNXB4BADs="},
		{"En cours", "R0lGODlhEAAQAMQAAAAAANbKvzENBNQ4FasnAVoYCe3o4TQrI/ZHIO2XFv+0QzUsI8EzEtjMwOvk3tzRx7otCrIrBefg2P99YLByFawnA35NKgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACH5BAEAABYALAAAAAAQABAAAAV5oCWOQmmOqCgQQ6IMg5CqrDuwk4yu7Uv8MJ1qQEkkYDcCQxhoBgxPKEwpOxQChSy2AA0MILoDYEwGGACFQSRcJp8LiIqg6zQ8DHh0XBCwiAEBZW83fH5jgWRYcQSFYoiJaT+NgGMmiggxAQsWfCUifSYyTqMSDjMWIQA7"},
		{"Terminé", "R0lGODlhEAARAMQUACglFWZXHX1tKYt7MaSRO0I6F8KUBu21CCwpGP/ba1dKF7ahQf///zc0IDEuHIpqBVxGAIx8MjQxHqWSPH5NKgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACH5BAEAABQALAAAAAAQABEAAAVpICWOQDmeKBAQAoC+Kuu+Z9zSZZmsLZO7scHgsCgOAgnhosQTJIXHgyDQsgkOQsF0FwgwC2ADGAww5ESIdPrQLRxmOIXhkPDRRCo5JAGHBeQPb3cUeQaGfSkrXYiJBBMRjDUACBINDighADs="},
	}
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
