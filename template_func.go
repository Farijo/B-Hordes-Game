package main

import (
	"bhordesgame/dto"
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
	Amount  int
}

func decodeGoal(key string, goal dto.Goal, l map[int]GoalHeader) GoalHTML {
	splited := strings.Split(goal.Descript, ":")
	idxLast := len(splited) - 2
	goalMax, header := 0, ""
	if l != nil {
		defer func() { l[goal.ID] = GoalHeader{template.HTML(header), goalMax} }()
	}
	if idxLast >= 0 {
		if splited[idxLast] == "" {
			header = "+"
			splited[idxLast] = "le plus de"
		} else {
			header = splited[idxLast]
			goalMax, _ = strconv.Atoi(splited[idxLast])
		}
	}
	var out GoalHTML
	switch goal.Typ {
	case 0:
		if len(splited) < 2 {
			return out
		}
		out.Text = template.HTML(fmt.Sprintf("Accumuler <b>%s</b> pictos", splited[0]))
		if id, err := strconv.Atoi(splited[1]); err == nil {
			out.Icon, out.Label = getServerDataKey(id, "pictos", key)
			header += "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		}
	case 1:
		if len(splited) < 4 {
			return out
		}
		var txt string
		if splited[0] > "" {
			if splited[1] > "" {
				txt = "Etre sur la <b>case</b> [ <b>%s</b> / <b>%s</b> ] de l'OM avec <b>%s</b>"
			} else {
				txt = "Etre sur la <b>ligne %s%s</b> de l'OM avec <b>%s</b>"
			}
		} else {
			if splited[1] > "" {
				txt = "Etre sur la <b>colonne %s%s</b> de l'OM avec <b>%s</b>"
			} else {
				txt = "Etre dans l'OM avec <b>%s%s%s</b>"
			}
		}
		out.Text = template.HTML(fmt.Sprintf(txt, splited[0], splited[1], splited[2]))
		if id, err := strconv.Atoi(splited[3]); err == nil {
			out.Icon, out.Label = getServerDataKey(id, "items", key)
			header = fmt.Sprintf("[%s/%s] %s <img src=\"https://myhordes.eu/build/images/%s\">", splited[0], splited[1], header, out.Icon)
		}
	case 2:
		if len(splited) < 1 {
			return out
		}
		out.Text = "Construire"
		if id, err := strconv.Atoi(splited[0]); err == nil {
			out.Icon, out.Label = getServerDataKey(id, "buildings", key)
			header = "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
			goalMax = 1
		}
	case 3:
		if len(splited) < 2 {
			return out
		}
		out.Text = template.HTML(fmt.Sprintf("Avoir en banque <b>%s</b>", splited[0]))
		if id, err := strconv.Atoi(splited[1]); err == nil {
			out.Icon, out.Label = getServerDataKey(id, "items", key)
			header += "<img src=\"https://myhordes.eu/build/images/" + out.Icon + "\">"
		}
	}
	return out
}

func mkmap() map[int]GoalHeader {
	return make(map[int]GoalHeader, 0)
}
