package main

import (
	"encoding/json"
	"html/template"
	"strings"
	"sync"
)

var sessions map[string]int = make(map[string]int, 10)

func refreshData(key string) error {
	milestone, err := requestMe(key)
	if err != nil {
		return err
	}
	sessions[key] = milestone.User.ID
	if err = insertUser(&milestone.User); err != nil {
		return err
	}
	if err = insertMilestone(milestone); err != nil {
		return err
	}
	return nil
}

type SrvData struct {
	Id   int    `json:"id"`
	Img  string `json:"img"`
	Name struct {
		Fr string `json:"fr"`
	} `json:"name"`
}

var serverData map[string]map[string]SrvData

var global, perBuilder sync.Mutex

func getServerData(userkey string) template.JS {
	global.Lock()
	defer global.Unlock()

	if serverData == nil {
		serverData = make(map[string]map[string]SrvData, 3)
	}
	var wg sync.WaitGroup
	var builder strings.Builder
	for _, resource := range []string{"pictos", "buildings", "items"} {
		wg.Add(1)
		go func(rsc string) {
			defer wg.Done()
			if serverData[rsc] == nil {
				serverData[rsc] = requestServerData(rsc, userkey)
			}
			if marsh, err := json.Marshal(serverData[rsc]); err == nil {
				perBuilder.Lock()
				builder.WriteString("const " + rsc + "=")
				builder.Write(marsh)
				builder.WriteByte(';')
				perBuilder.Unlock()
			}
		}(resource)
	}
	wg.Wait()
	return template.JS(builder.String())
}
