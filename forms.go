package main

import "bhordesgame/dto"

type FormChallenge struct {
	Name          string `form:"name"`
	Participation int8   `form:"participation"`
	Private       bool   `form:"privat"`
	ValidationApi bool   `form:"validation_api"`
	Act           string `form:"act"`
}

func (form *FormChallenge) buildChallenge(creatorId int) *dto.Challenge {
	var result dto.Challenge
	result.Name = form.Name
	result.Creator.ID = creatorId
	result.Flags = byte(form.Participation)
	if form.Private {
		result.Flags |= 0x04
	}
	if !form.ValidationApi {
		result.Flags |= 0x08
	}
	if form.Act == "Valider" {
		result.Flags |= 1 << 4
	} else {
		result.Flags &= 0x0F
	}
	return &result
}

func buildGoalsFromForm(types, x, y, count, val []string) *[]dto.Goal {
	goals := make([]dto.Goal, len(types))

	for i := range goals {
		v := &goals[i]
		v.Typ = types[i][0] - '0'
		switch v.Typ {
		case 0, 3:
			v.Descript = pop(&count) + ":" + pop(&val)
		case 1:
			v.Descript = pop(&x) + ":" + pop(&y) + ":" + pop(&count) + ":" + pop(&val)
		case 2:
			v.Descript = pop(&val)
		}
	}

	return &goals
}
