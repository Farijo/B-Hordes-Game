package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	BASE_URL = "https://myhordes.eu/api/x/json/"

	ME     = "me?fields=id,name,avatar,isGhost,playedMaps.fields(mapId),rewards.fields(id,number),dead,out,ban,baseDef,x,y,job,map.fields(wid,hei,days,custom,conspiracy,city.fields(buildings.fields(id),bank.fields(id,count)),zones.fields(items.fields(id,count)))"
	OTHER  = "user?id=%d&fields=id,name,avatar"
	OTHERS = "users?ids=%s&fields=id,name,avatar,isGhost,playedMaps.fields(mapId),rewards.fields(id,number),dead,out,ban,baseDef,x,y,job,map.fields(wid,hei,days,custom,conspiracy,city.fields(buildings.fields(id),bank.fields(id,count)),zones.fields(items.fields(id,count)))"
)

func buildAuthQuery(userkey string) string {
	return "&appkey=" + os.Getenv("API_KEY") + "&userkey=" + userkey
}

func requestMe(userkey string) (*dto.Milestone, error) {
	if err := registerCall(userkey); err != nil {
		return nil, err
	}
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
	if err := registerCall(userkey); err != nil {
		return nil, err
	}
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

func requestMultipleUsers(userkey string, ids []int) ([]dto.Milestone, error) {
	if err := registerCall(userkey); err != nil {
		return nil, err
	}
	lastIdx := len(ids) - 1
	if lastIdx < 0 {
		return make([]dto.Milestone, 0), nil
	}
	var builder strings.Builder
	for i := 0; i < lastIdx; i++ {
		builder.WriteString(strconv.Itoa(ids[i]))
		builder.WriteRune(',')
	}
	builder.WriteString(strconv.Itoa(ids[lastIdx]))
	resp, err := http.Get(BASE_URL + fmt.Sprintf(OTHERS, builder.String()) + buildAuthQuery(userkey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	flat := make([]struct {
		*dto.User
		dto.Milestone
	}, len(ids))
	for i := range flat {
		flat[i].User = &flat[i].Milestone.User
	}

	if err := json.NewDecoder(resp.Body).Decode(&flat); err != nil {
		return nil, err
	}

	res := make([]dto.Milestone, len(flat))
	for i, v := range flat {
		res[i] = v.Milestone
	}

	return res, nil
}

func requestServerData(endpoint, userkey string) map[string]SrvData {
	resp, err := http.Get(BASE_URL + endpoint + "?fields=id,img,name" + buildAuthQuery(userkey))
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.StatusCode)
		return nil
	}
	datares := make(map[string]SrvData, 100)
	if err := json.NewDecoder(resp.Body).Decode(&datares); err != nil {
		fmt.Println(err)
		return nil
	}

	return datares
}
