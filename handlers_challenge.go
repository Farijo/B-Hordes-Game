package main

import (
	"bhordesgame/dto"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

/* * * * * * * * * * * * * * * * * * * * * *
 *                   GET                   *
 * * * * * * * * * * * * * * * * * * * * * */
func challengeCreationHandle(c *gin.Context) {
	c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_challenge-creation.html", gin.H{"faq": wantFAQ(c.Cookie(NOFAQ)), "logged": true, "challenge": nil, "srvData": getServerData(c.GetString("key"))})
}

func challengeHandle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	key, cookieErr := c.Cookie("user")
	uid, ok := sessions[key]
	logged := cookieErr == nil && ok

	challenge, err := queryChallenge(id, uid)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusNotFound)
		return
	}

	selfChallenge := challenge.Creator.ID == uid && logged

	switch challenge.Status {
	case 0, 1: // draft, review
		if selfChallenge {
			c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_challenge-creation.html", gin.H{
				"faq":       wantFAQ(c.Cookie(NOFAQ)),
				"logged":    true,
				"challenge": challenge,
				"goals":     makeChannelFor(queryChallengeGoals, challenge.ID),
				"srvData":   getServerData(key),
			})
		} else {
			c.Status(http.StatusForbidden)
		}
	case 2: // invite
		if duser, err := queryUser(challenge.Creator.ID); err != nil {
			c.Status(http.StatusNotFound)
			return
		} else {
			challenge.Creator = duser.User
		}

		searchResults := make(chan *dto.DetailedUser)
		invitationResults := make(chan *dto.DetailedUser)
		if selfChallenge {
			go func() {
				idents := strings.FieldsFunc(c.Query("ident"), func(r rune) bool { return r == ',' || r == ' ' })
				// Problem of this is making a request to MH on each reload with "ident" set
				// and the request would probably be useless (ie : we already have the info)
				//
				// if cookieErr == nil {
				// 	realIds := make([]string, 0)
				// 	for _, maybeId := range idents {
				// 		if _, err := strconv.Atoi(maybeId); err == nil {
				// 			realIds = append(realIds, maybeId)
				// 		}
				// 	}
				// 	if len(realIds) > 0 {
				// 		if users, err := requestMultipleUsers(key, realIds); err == nil {
				// 			if err := insertMultipleUsers(users); err != nil {
				// 				fmt.Println(err)
				// 			}
				// 		} else {
				// 			fmt.Println(err)
				// 		}
				// 	}
				// }
				queryMultipleUsers(searchResults, idents)
			}()
			go queryChallengeInvitations(invitationResults, challenge.ID)
		} else {
			close(searchResults)
			close(invitationResults)
		}

		ident := c.Query("ident")
		if ident > "" {
			ident = "?ident=" + ident
		}

		c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_challenge-recruit.html", gin.H{
			"faq":           wantFAQ(c.Cookie(NOFAQ)),
			"logged":        logged,
			"selfChallenge": selfChallenge,
			"selfID":        uid,
			"challenge":     challenge,
			"goals":         makeChannelFor(queryChallengeGoals, challenge.ID),
			"userkey":       key,
			"searchResults": searchResults,
			"invitations":   invitationResults,
			"validators":    makeChannelFor(queryChallengeValidators, challenge.ID),
			"participants":  makeChannelFor(queryChallengeParticipants, challenge.ID),
			"action":        makeChannelActionString(logged, challenge, uid, c.GetString(LNG_KEY)),
			"ident":         ident,
		})
	case 3: // started
		if duser, err := queryUser(challenge.Creator.ID); err != nil {
			c.Status(http.StatusNotFound)
			return
		} else {
			challenge.Creator = duser.User
		}

		c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_challenge-progress.html", gin.H{
			"faq":           wantFAQ(c.Cookie(NOFAQ)),
			"logged":        logged,
			"selfChallenge": selfChallenge,
			"challenge":     challenge,
			"goals":         makeChannelFor(queryChallengeGoals, challenge.ID),
			"userkey":       key,
			"advancement":   makeChannelFor(queryChallengeAdvancements, challenge.ID),
		})
	case 4: // ended
		if duser, err := queryUser(challenge.Creator.ID); err != nil {
			c.Status(http.StatusNotFound)
			return
		} else {
			challenge.Creator = duser.User
		}

		c.HTML(http.StatusOK, c.GetString(LNG_KEY)+"_challenge-progress.html", gin.H{
			"faq":           wantFAQ(c.Cookie(NOFAQ)),
			"logged":        logged,
			"selfChallenge": false, // disable challenge action
			"challenge":     challenge,
			"goals":         makeChannelFor(queryChallengeGoals, challenge.ID),
			"userkey":       key,
			"advancement":   makeChannelFor(queryChallengeAdvancements, challenge.ID),
		})
	}
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                   POST                  *
 * * * * * * * * * * * * * * * * * * * * * */
func createChallengeHandle(c *gin.Context) {
	var formChallenge FormChallenge
	c.Bind(&formChallenge)

	challenge := formChallenge.buildChallenge(c.GetInt("uid"))
	goals := buildGoalsFromForm(c.PostFormArray("type"), c.PostFormArray("x"), c.PostFormArray("y"), c.PostFormArray("count"), c.PostFormArray("val"), c.PostFormArray("custom"))

	id, err := insertChallenge(challenge, goals)
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/challenge/%d", id))
}

func updateChallengeHandle(c *gin.Context) {
	id, atoiErr := strconv.Atoi(c.Param("id"))
	if atoiErr != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	var formChallenge FormChallenge
	if bindErr := c.Bind(&formChallenge); bindErr != nil {
		fmt.Println(bindErr)
		c.Status(http.StatusBadRequest)
		return
	}

	var err error
	switch formChallenge.Act {
	case "modify":
		err = updateChallengeStatus(id, c.GetInt("uid"), 0)
	case "open-inscriptions":
		err = updateChallengeStatus(id, c.GetInt("uid"), 2)
	default:
		challenge := formChallenge.buildChallenge(c.GetInt("uid"))
		challenge.ID = id
		goals := buildGoalsFromForm(c.PostFormArray("type"), c.PostFormArray("x"), c.PostFormArray("y"), c.PostFormArray("count"), c.PostFormArray("val"), c.PostFormArray("custom"))
		err = updateChallenge(challenge, goals)
	}

	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusBadRequest)
		return
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/challenge/%d", id))
}

func challengeMembersHandle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	c.MultipartForm()
	for id_status, action := range c.Request.PostForm {
		idst := strings.Split(id_status, "-")
		if targetId, err := strconv.Atoi(idst[0]); err == nil {
			add := true
			switch action[0][0] {
			// case "✓", "+ approbator", "+ guest":
			case "x"[0]:
				add = false
			}
			insertOrDeleteChallengeMember(id, c.GetInt("uid"), targetId, idst[1] == "validator", add)
		}
	}

	ident := c.Query("ident")
	if ident > "" {
		ident = "?ident=" + ident
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/challenge/%d%s", id, ident))
}

func challengeDateHandle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	switch c.PostForm("validation") {
	case "validate":
		start := true
		date := c.PostForm("start_date")
		if date == "" {
			start = false
			date = c.PostForm("end_date")
		}
		if date > "" {
			updateChallengeDate(id, c.GetInt("uid"), date, start)
		}
	case "start-now":
		updateChallengeDate(id, c.GetInt("uid"), "", true)
	case "end-now":
		updateChallengeDate(id, c.GetInt("uid"), "", false)
	}

	ident := c.Query("ident")
	if ident > "" {
		ident = "?ident=" + ident
	}

	c.Redirect(http.StatusFound, fmt.Sprintf("/challenge/%d%s", id, ident))
}

func challengeScanHandle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	userIDs, err := queryChallengeParticipantsForScan(id, c.GetInt("uid"))
	if err != nil {
		fmt.Println(err)
		c.Status(http.StatusForbidden)
		return
	}
	if userIDs == "" {
		c.Status(http.StatusNoContent)
		return
	}
	milestones, err := requestMultipleUsers(c.GetString("key"), userIDs)
	if err != nil {
		fmt.Println(err)
		if err.Error() == "too many request" {
			c.Status(http.StatusTooManyRequests)
		} else {
			c.Status(http.StatusFailedDependency)
		}
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(milestones))
	for i := range milestones {
		go func(milestone *dto.Milestone, wg *sync.WaitGroup) {
			if err := insertMilestone(milestone); err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}(&milestones[i], &wg)
	}
	wg.Wait()

	c.Redirect(http.StatusFound, fmt.Sprintf("/challenge/%d", id))
}

/* * * * * * * * * * * * * * * * * * * * * *
 *                  OTHER                  *
 * * * * * * * * * * * * * * * * * * * * * */
func makeChannelActionString(logged bool, challenge dto.DetailedChallenge, uid int, lang string) <-chan []string {
	action := make(chan []string)
	if logged {
		go func() {
			defer close(action)
			lngData := translations[lang]
			invited, participate := queryChallengeUserStatus(challenge.ID, uid)
			if participate {
				action <- []string{"x", lngData["retire"]}
			} else {
				switch challenge.Access {
				case 0:
					action <- []string{"+", lngData["join"]}
				case 1:
					if invited {
						action <- []string{"x", lngData["cancel-request"]}
					} else {
						action <- []string{"+", lngData["send-request"]}
					}
				case 2:
					if invited {
						action <- []string{"+", lngData["accept-invite"]}
					}
				}
			}
		}()
	} else {
		close(action)
	}
	return action
}

func makeChannelFor[T any](act func(chan<- T, int), param int) <-chan T {
	ch := make(chan T)
	go act(ch, param)
	return ch
}
