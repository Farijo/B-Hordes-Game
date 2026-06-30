package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	CALL_MAX_NB = 10
	CALL_DELAY  = 60
)

var sessions map[string]int = make(map[string]int, 10)
var calls map[string][]int64 = make(map[string][]int64, 10)

const (
	SEARCH_CACHE_TTL = 300 * time.Second
)

type searchCacheEntry struct {
	results   []any
	timestamp time.Time
}

var (
	searchCache   = make(map[string]searchCacheEntry, 64)
	searchCacheMu sync.Mutex
)

func searchCacheKey(idents []string) string {
	sorted := make([]string, len(idents))
	copy(sorted, idents)
	sort.Strings(sorted)
	return strings.Join(sorted, ",")
}

func searchCacheGet(idents []string) ([]any, bool) {
	key := searchCacheKey(idents)
	now := time.Now()
	searchCacheMu.Lock()
	defer searchCacheMu.Unlock()
	if entry, ok := searchCache[key]; ok && now.Sub(entry.timestamp) < SEARCH_CACHE_TTL {
		return entry.results, true
	}
	return nil, false
}

func searchCacheSet(idents []string, results []any) {
	key := searchCacheKey(idents)
	searchCacheMu.Lock()
	defer searchCacheMu.Unlock()
	searchCache[key] = searchCacheEntry{results: results, timestamp: time.Now()}
}

func registerCall(key string) error {
	now := time.Now().Unix()
	if userLastCalls, ok := calls[key]; ok {
		ago := now - CALL_DELAY
		for i, v := range userLastCalls {
			if v < ago {
				userLastCalls[i] = now
				return nil
			}
		}
		if len(userLastCalls) < CALL_MAX_NB {
			calls[key] = append(calls[key], now)
		} else {
			logger.Println(key, sessions[key])
			return errors.New("too many request")
		}
	} else {
		calls[key] = []int64{now}
	}
	return nil
}

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
	Id   uint16 `json:"id"`
	Img  string `json:"img"`
	Name struct {
		Fr string `json:"fr"`
		En string `json:"en"`
		Es string `json:"es"`
		De string `json:"de"`
	} `json:"name"`
}

var serverData map[string]map[string]SrvData
var templateSrvData template.JS

var globalMtx, builderMtx sync.Mutex

func getServerData(userkey string) template.JS {
	if templateSrvData > "" {
		return templateSrvData
	}
	if userkey == "" {
		return ""
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
		if serverData[resource] == nil {
			wg.Add(1)
			go func(rsc string) {
				defer wg.Done()
				serverData[rsc] = requestServerData(rsc, userkey)
				if serverData[rsc] != nil {
					if marsh, err := json.Marshal(serverData[rsc]); err == nil {
						builderMtx.Lock()
						builder.WriteString("const ")
						builder.WriteString(rsc)
						builder.WriteString("=")
						builder.Write(marsh)
						builder.WriteByte(';')
						builderMtx.Unlock()
					}
				}
			}(resource)
		}
	}
	wg.Wait()
	for _, resource := range []string{"pictos", "buildings", "items"} {
		if serverData[resource] == nil {
			// if at least one is nil, don't cache result
			return template.JS(builder.String())
		}
	}
	templateSrvData = template.JS(builder.String())
	return templateSrvData
}

func getServerDataKey(id uint16, datakey, userkey, lang string) (img, name string) {
	if serverData == nil || serverData[datakey] == nil {
		getServerData(userkey)
	}
	for _, v := range serverData[datakey] {
		if v.Id == id {
			switch lang {
			case "fr":
				return v.Img, v.Name.Fr
			case "es":
				return v.Img, v.Name.Es
			case "de":
				return v.Img, v.Name.De
			default:
				return v.Img, v.Name.En
			}
		}
	}
	return
}
