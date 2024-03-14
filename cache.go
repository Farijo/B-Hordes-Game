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
var templateSrvData template.JS

var globalMtx, builderMtx sync.Mutex

func getServerData(userkey string) template.JS {
	if templateSrvData > "" {
		return templateSrvData
	}
	globalMtx.Lock()
	defer globalMtx.Unlock()
	if templateSrvData > "" {
		return templateSrvData
	}

	serverData = make(map[string]map[string]SrvData, 3)
	var builder strings.Builder
	var wg sync.WaitGroup
	for _, resource := range []string{"pictos", "buildings", "items"} {
		wg.Add(1)
		go func(rsc string) {
			defer wg.Done()
			serverData[rsc] = requestServerData(rsc, userkey)
			if serverData[rsc] != nil {
				if marsh, err := json.Marshal(serverData[rsc]); err == nil {
					builderMtx.Lock()
					builder.WriteString("const " + rsc + "=")
					builder.Write(marsh)
					builder.WriteByte(';')
					builderMtx.Unlock()
				}
			}
		}(resource)
	}
	wg.Wait()
	templateSrvData = template.JS(builder.String())
	return templateSrvData
}

func getServerDataKey(id int, datakey, userkey string) (img, name string) {
	if serverData == nil {
		getServerData(userkey)
	}
	for _, v := range serverData[datakey] {
		if v.Id == id {
			return v.Img, v.Name.Fr
		}
	}
	return
}
