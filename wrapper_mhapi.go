package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

const (
	BASE_URL = "https://myhordes.eu/api/x/json/"

	ME     = "me?fields=id,name,avatar,isGhost,playedMaps.fields(mapId),rewards.fields(id,number),dead,out,ban,baseDef,x,y,job,map.fields(wid,hei,days,custom,conspiracy,city.fields(buildings.fields(id),bank.fields(id,count)),zones.fields(items.fields(id,count)))"
	OTHER  = "user?id=%d&fields=id,name,avatar"
	OTHERS = "users?ids=%s&fields=id,name,avatar"
	FULL   = ",isGhost,playedMaps.fields(mapId),rewards.fields(id,number),dead,out,ban,baseDef,x,y,job,map.fields(wid,hei,days,custom,conspiracy,city.fields(buildings.fields(id),bank.fields(id,count)),zones.fields(items.fields(id,count)))"
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
		Error string
	}
	flat.User = &flat.Milestone.User

	if err := json.NewDecoder(resp.Body).Decode(&flat); err != nil {
		return nil, err
	}

	if flat.Error > "" {
		return nil, errors.New(flat.Error)
	}

	return &flat.Milestone, nil
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

	var flat struct {
		dto.User
		Error string
	}

	if err := json.NewDecoder(resp.Body).Decode(&flat); err != nil {
		return nil, err
	}

	if flat.Error > "" {
		return nil, errors.New(flat.Error)
	}

	return &flat.User, nil
}

func requestMultipleInfo(userkey, ids string) ([]dto.User, error) {
	if err := registerCall(userkey); err != nil {
		return nil, err
	}
	resp, err := http.Get(BASE_URL + fmt.Sprintf(OTHERS, ids) + buildAuthQuery(userkey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}

	flat := make([]dto.User, len(ids))

	return flat, json.NewDecoder(resp.Body).Decode(&flat)
}

func requestMultipleUsers(userkey, ids string) ([]dto.Milestone, error) {
	if err := registerCall(userkey); err != nil {
		return nil, err
	}
	resp, err := http.Get(BASE_URL + fmt.Sprintf(OTHERS, ids) + FULL + buildAuthQuery(userkey))
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
		logger.Println(err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.Println(resp.StatusCode)
		return nil
	}
	datares := make(map[string]SrvData, 100)
	if err := json.NewDecoder(resp.Body).Decode(&datares); err != nil {
		logger.Println(err)
		return nil
	}

	return datares
}
