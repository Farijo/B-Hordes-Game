package main

import "html/template"

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

var serverData template.JS

func getServerData(userkey string) template.JS {
	if serverData == "" {
		serverData = requestServerData(userkey)
	}
	return serverData
}
