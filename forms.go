package main

import (
	"bhordesgame/dto"
	"strconv"
)

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
		v.Entity = uint16(Ignore(strconv.ParseUint(pop(&val), 10, 16)))
		switch v.Typ {
		case 1:
			v.X.Valid = true
			v.X.Byte = byte(Ignore(strconv.ParseInt(pop(&x), 10, 8)))
			v.Y.Valid = true
			v.Y.Byte = byte(Ignore(strconv.ParseInt(pop(&y), 10, 8)))
			fallthrough
		case 0, 3:
			v.Amount.Valid = true
			v.Amount.Int32 = int32(Ignore(strconv.ParseUint(pop(&count), 10, 32)))
		}
	}

	return &goals
}
