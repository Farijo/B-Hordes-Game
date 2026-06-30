package main

import (
	"bhordesgame/dto"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
)

const (
	BASE_URL = "https://myhordes.eu/api/x/json/"

	FULL     = ",isGhost,playedMaps.fields(mapId),rewards.fields(id,number),dead,out,ban,baseDef,x,y,job,map.fields(wid,hei,days,guide,shaman,custom,conspiracy,city.fields(door,water,chaos,devast,hard,x,y,buildings.fields(id),news.fields(z,def,water,regenDir),defense.fields(total,base,buildings,upgrades,items,itemsMul,citizenHomes,citizenGuardians,watchmen,souls,temp,cadavers,bonus),upgrades.fields(list.fields(buildingId,level)),estimations.fields(min,max,maxed),estimationsNext.fields(min,max,maxed),bank.fields(id,count)),zones.fields(items.fields(id,count)))&languages=en"
	ME       = "me?fields=id,name,avatar" + FULL
	OTHER    = "user?id=%d&fields=id,name,avatar"
	OTHERS   = "users?ids=%s&fields=id,name,avatar"
	CITY     = "map?mapId=%d&fields=citizens.fields(id,name,avatar)"
	CITIZENS = "me?fields=map.fields(citizens.fields(id,name,avatar))"
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

func requestFellowCitizens(userkey string, ch chan<- *dto.User) error {
	defer close(ch)

	if err := registerCall(userkey); err != nil {
		logger.Println(err)
		return err
	}
	resp, err := http.Get(BASE_URL + CITIZENS + buildAuthQuery(userkey))
	if err != nil {
		logger.Println(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Println(resp.Status)
		return errors.New(resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)

	// read open bracket {
	if brace, err := decoder.Token(); err != nil {
		logger.Println(brace, err)
		return err
	}
	// read "map"
	if brace, err := decoder.Token(); err != nil {
		logger.Println(brace, err)
		return err
	}
	// read open bracket {
	if brace, err := decoder.Token(); err != nil {
		logger.Println(brace, err)
		return err
	}
	// read "citizens"
	if brace, err := decoder.Token(); err != nil {
		logger.Println(brace, err)
		return err
	}
	// read open bracket [
	if brace, err := decoder.Token(); err != nil {
		logger.Println(brace, err)
		return err
	}

	// while the array contains values
	for decoder.More() {
		var user dto.User

		if err := decoder.Decode(&user); err != nil {
			logger.Println(err)
			return err
		}
		ch <- &user
	}

	return nil
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

type FlattenMilestone struct {
	*dto.User
	dto.Milestone
}

func buildFlattenMilestone() *FlattenMilestone {
	fm := new(FlattenMilestone)
	fm.User = &fm.Milestone.User
	return fm
}
func requestMultipleMilestones(userkey, ids string, milestones chan<- *FlattenMilestone, wg *sync.WaitGroup) error {
	return requestMultiple(userkey, OTHERS+FULL, ids, buildFlattenMilestone, nil, milestones, wg)
}

func requestMultipleUsers(userkey, ids string, filter func(*dto.User) bool, actualizedUsers chan<- *dto.User, wg *sync.WaitGroup) error {
	return requestMultiple(userkey, OTHERS, ids, func() *dto.User { return &dto.User{} }, filter, actualizedUsers, wg)
}

func requestMultiple[S any](userkey, url, ids string, build func() *S, filter func(*S) bool, ch chan<- *S, wg *sync.WaitGroup) error {
	if err := registerCall(userkey); err != nil {
		wg.Wait()
		close(ch)
		return err
	}

	idsLen := len(ids)
	chunkSize := idsLen * 1000 / (idsLen + 1000)
	start := 0

	for start < idsLen {
		end := start + chunkSize
		for end < idsLen && ids[end] != ',' {
			end++
		}
		end = min(end, idsLen)

		wg.Add(1)
		go func(idsChunk string, ch chan<- *S) {
			defer wg.Done()
			resp, err := http.Get(BASE_URL + fmt.Sprintf(url, idsChunk) + buildAuthQuery(userkey)) // batch 1500char uri
			if err != nil {
				return
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logger.Println(resp.StatusCode)
				return
			}

			decoder := json.NewDecoder(resp.Body)

			// read open bracket
			if brace, err := decoder.Token(); err != nil {
				logger.Println(brace, err)
				return
			}

			// while the array contains values
			for decoder.More() {
				data := build()

				if err := decoder.Decode(data); err != nil {
					logger.Println(err)
					return
				}
				if filter != nil && filter(data) {
					ch <- data
				}
			}

			// read closing bracket
			if brace, err := decoder.Token(); err != nil {
				logger.Println(brace, err)
				return
			}
		}(ids[start:end], ch)

		start = end + 1
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	return nil
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

func requestCity(userkey string, id int, ch chan<- *dto.User) error {
	if err := registerCall(userkey); err != nil {
		logger.Println(err)
		return err
	}
	resp, err := http.Get(BASE_URL + fmt.Sprintf(CITY, id) + buildAuthQuery(userkey))
	if err != nil {
		logger.Println(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Println(resp.StatusCode)
		return errors.New(resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)

	// read open bracket
	if brace, err := decoder.Token(); err != nil {
		logger.Println(brace, err)
		return err
	}
	// read "citizens"
	if brace, err := decoder.Token(); err != nil {
		logger.Println(brace, err)
		return err
	}
	// read open bracket
	if brace, err := decoder.Token(); err != nil {
		logger.Println(brace, err)
		return err
	}

	// while the array contains values
	for decoder.More() {
		var user dto.User

		if err := decoder.Decode(&user); err != nil {
			logger.Println(err)
			return err
		}
		ch <- &user
	}

	return nil
}
