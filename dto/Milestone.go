package dto

import (
	"bytes"
	"database/sql"
	"encoding/json"
)

type Milestone struct {
	User       User            `db:"user"`
	Dt         string          `db:"dt"`
	IsGhost    jsonNullBool    `db:"isGhost"`
	PlayedMaps jsonNullCounter `db:"playedMaps"`
	Rewards    jsonNullReward  `db:"rewards"`
	Dead       jsonNullBool    `db:"dead"`
	Ban        jsonNullBool    `db:"ban"`
	BaseDef    jsonNullByte    `db:"baseDef"`
	X          jsonNullByte    `db:"x"`
	Y          jsonNullByte    `db:"y"`
	Job        jsonNullJob     `db:"job"`
	Map        struct {
		Wid        jsonNullByte `db:"mapWid"`
		Hei        jsonNullByte `db:"mapHei"`
		Days       jsonNullByte `db:"mapDays"`
		Conspiracy jsonNullBool `db:"conspiracy"`
		Custom     jsonNullBool `db:"custom"`
	}
}

type jsonNullReward struct {
	sql.NullString
	rewards map[uint16]uint32
}

func (v *jsonNullReward) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	v.rewards = make(map[uint16]uint32, 20)
	decoder := json.NewDecoder(bytes.NewReader(data))
	idx, val := uint16(0), uint32(0)
	for token, err := decoder.Token(); err == nil; token, err = decoder.Token() {
		switch token {
		case "id":
			token, err = decoder.Token()
			if err != nil {
				return err
			}
			if val > 0 {
				v.rewards[uint16(token.(float64))] = val
				val = 0
			} else {
				idx = uint16(token.(float64))
			}
		case "number":
			token, err = decoder.Token()
			if err != nil {
				return err
			}
			if idx > 0 {
				v.rewards[idx] = uint32(token.(float64))
				idx = 0
			} else {
				val = uint32(token.(float64))
			}
		}
	}
	return nil
}

type jsonNullJob struct {
	sql.NullByte
}

func (v *jsonNullJob) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))

	token, err := decoder.Token()

	for ; err == nil; token, err = decoder.Token() {
		switch token {
		case "hab":
			v.Byte = 0
			v.Valid = true
			return nil

		case "fouin":
			v.Byte = 1
			v.Valid = true
			return nil

		case "ecl":
			v.Byte = 2
			v.Valid = true
			return nil

		case "guard":
			v.Byte = 3
			v.Valid = true
			return nil

		case "mite":
			v.Byte = 4
			v.Valid = true
			return nil

		case "appri":
			v.Byte = 5
			v.Valid = true
			return nil

		case "tech":
			v.Byte = 6
			v.Valid = true
			return nil
		}
	}

	return err
}
