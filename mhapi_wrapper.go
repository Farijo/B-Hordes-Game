package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const (
	BASE_URL = "https://myhordes.eu/api/x/json/"

	ME = "me?fields=id%2Cname%2Cavatar%2CisGhost%2CplayedMaps%2Crewards%2ChomeMessage%2Chero%2Cdead%2Cout%2Cban%2CbaseDef%2Cx%2Cy%2CmapId%2Cjob%2Cmap"
)

func buildAuthQuery(userkey string) string {
	return "&appkey=" + os.Getenv("API_KEY") + "&userkey=" + userkey
}

func requestMe(userkey string) (userdata *dto.Milestone, err error) {

	fmt.Println(userkey)
	resp, err := http.Get(BASE_URL + ME + buildAuthQuery(userkey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var flat struct {
		*dto.User
		dto.Milestone
	}
	flat.User = &flat.Milestone.User
	json.NewDecoder(resp.Body).Decode(&flat)

	fmt.Println(flat)

	return &flat.Milestone, nil
}
