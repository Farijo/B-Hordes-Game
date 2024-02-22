package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	BASE_URL = "https://myhordes.eu/api/x/json/"

	ME    = "me?fields=id,name,avatar,isGhost,playedMaps,rewards,homeMessage,hero,dead,out,ban,baseDef,x,y,mapId,job,map"
	OTHER = "user?id=%d&fields=id,name,avatar"
)

func buildAuthQuery(userkey string) string {
	return "&appkey=" + os.Getenv("API_KEY") + "&userkey=" + userkey
}

func requestMe(userkey string) (*dto.Milestone, error) {
	resp, err := http.Get(BASE_URL + ME + buildAuthQuery(userkey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var flat struct {
		*dto.User
		dto.Milestone
	}
	flat.User = &flat.Milestone.User

	return &flat.Milestone, json.NewDecoder(resp.Body).Decode(&flat)
}

func requestUser(userkey string, id int) (*dto.User, error) {
	resp, err := http.Get(BASE_URL + fmt.Sprintf(OTHER, id) + buildAuthQuery(userkey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	var user dto.User
	return &user, json.NewDecoder(resp.Body).Decode(&user)
}

func requestServerData(userkey string) template.JS {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var builder strings.Builder
	queryParams := "?fields=id,img,name" + buildAuthQuery(userkey)

	for _, v := range []string{"pictos", "buildings", "items"} {
		wg.Add(1)
		go func(endpoint string) {
			defer wg.Done()
			resp, err := http.Get(BASE_URL + endpoint + queryParams)
			if err != nil {
				panic(err.Error())
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				panic(resp.Status)
			}

			mu.Lock()
			builder.WriteString("const ")
			builder.WriteString(endpoint)
			builder.WriteString("=")
			io.Copy(&builder, resp.Body)
			builder.WriteString(";")
			mu.Unlock()
		}(v)
	}

	wg.Wait()

	return template.JS(builder.String())
}
