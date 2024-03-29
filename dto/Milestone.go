package dto

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
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

func (incoming *Milestone) CheckFieldsDifference(previous *Milestone) bool {
	hasChanges := false

	if incoming.IsGhost.Valid {
		incoming.IsGhost.Valid = !previous.IsGhost.Valid || (incoming.IsGhost.Bool != previous.IsGhost.Bool)
		hasChanges = hasChanges || incoming.IsGhost.Valid
	}
	if incoming.PlayedMaps.Valid {
		incoming.PlayedMaps.Valid = !previous.PlayedMaps.Valid || (incoming.PlayedMaps.Int64 != previous.PlayedMaps.Int64)
		hasChanges = hasChanges || incoming.PlayedMaps.Valid
	}
	if incoming.Dead.Valid {
		incoming.Dead.Valid = !previous.Dead.Valid || (incoming.Dead.Bool != previous.Dead.Bool)
		hasChanges = hasChanges || incoming.Dead.Valid
	}
	if incoming.Ban.Valid {
		incoming.Ban.Valid = !previous.Ban.Valid || (incoming.Ban.Bool != previous.Ban.Bool)
		hasChanges = hasChanges || incoming.Ban.Valid
	}
	if incoming.BaseDef.Valid {
		incoming.BaseDef.Valid = !previous.BaseDef.Valid || (incoming.BaseDef.Byte != previous.BaseDef.Byte)
		hasChanges = hasChanges || incoming.BaseDef.Valid
	}
	if incoming.X.Valid {
		incoming.X.Valid = !previous.X.Valid || (incoming.X.Byte != previous.X.Byte)
		hasChanges = hasChanges || incoming.X.Valid
	}
	if incoming.Y.Valid {
		incoming.Y.Valid = !previous.Y.Valid || (incoming.Y.Byte != previous.Y.Byte)
		hasChanges = hasChanges || incoming.Y.Valid
	}
	if incoming.Job.Valid {
		incoming.Job.Valid = !previous.Job.Valid || (incoming.Job.Byte != previous.Job.Byte)
		hasChanges = hasChanges || incoming.Job.Valid
	}
	if incoming.Map.Wid.Valid {
		incoming.Map.Wid.Valid = !previous.Map.Wid.Valid || (incoming.Map.Wid.Byte != previous.Map.Wid.Byte)
		hasChanges = hasChanges || incoming.Map.Wid.Valid
	}
	if incoming.Map.Hei.Valid {
		incoming.Map.Hei.Valid = !previous.Map.Hei.Valid || (incoming.Map.Hei.Byte != previous.Map.Hei.Byte)
		hasChanges = hasChanges || incoming.Map.Hei.Valid
	}
	if incoming.Map.Days.Valid {
		incoming.Map.Days.Valid = !previous.Map.Days.Valid || (incoming.Map.Days.Byte != previous.Map.Days.Byte)
		hasChanges = hasChanges || incoming.Map.Days.Valid
	}
	if incoming.Map.Conspiracy.Valid {
		incoming.Map.Conspiracy.Valid = !previous.Map.Conspiracy.Valid || (incoming.Map.Conspiracy.Bool != previous.Map.Conspiracy.Bool)
		hasChanges = hasChanges || incoming.Map.Conspiracy.Valid
	}
	if incoming.Map.Custom.Valid {
		incoming.Map.Custom.Valid = !previous.Map.Custom.Valid || (incoming.Map.Custom.Bool != previous.Map.Custom.Bool)
		hasChanges = hasChanges || incoming.Map.Custom.Valid
	}

	changements := make([]byte, 0, 120)
	incoming.Rewards.Valid = false
	for id, number := range incoming.Rewards.Pictos {
		if number != previous.Rewards.Pictos[id] {
			changements = binary.LittleEndian.AppendUint16(changements, id)
			changements = binary.LittleEndian.AppendUint32(changements, number)

			incoming.Rewards.Valid = true
		}
	}
	// to test : raz
	for id := range previous.Rewards.Pictos {
		if _, ok := incoming.Rewards.Pictos[id]; !ok {
			changements = binary.LittleEndian.AppendUint16(changements, id)
			changements = binary.LittleEndian.AppendUint32(changements, 0)

			incoming.Rewards.Valid = true
		}
	}

	if incoming.Rewards.Valid {
		incoming.Rewards.String = string(changements)
		hasChanges = true
	}

	return hasChanges
}

// Override scan so the previous value is kept if the new value is invalid (ie: null)

func (n *jsonNullBool) Scan(value any) error {
	oldB, oldV := n.Bool, n.Valid
	err := n.NullBool.Scan(value)
	if !n.Valid {
		n.Bool, n.Valid = oldB, oldV
	}
	return err
}

func (n *jsonNullCounter) Scan(value any) error {
	oldB, oldV := n.Int64, n.Valid
	err := n.NullInt64.Scan(value)
	if !n.Valid {
		n.Int64, n.Valid = oldB, oldV
	}
	return err
}

func (n *jsonNullReward) Scan(value any) error {
	err := n.NullString.Scan(value)
	if n.Valid {
		if n.Pictos == nil {
			n.Pictos = make(map[uint16]uint32, 20)
		}
		nsLen := len(n.String)
		for i := 0; i < nsLen; i += 6 {
			id := binary.LittleEndian.Uint16([]byte(n.String[i : i+2]))
			number := binary.LittleEndian.Uint32([]byte(n.String[i+2 : i+6]))
			n.Pictos[id] = number
		}
	}
	return err
}

func (n *jsonNullByte) Scan(value any) error {
	oldB, oldV := n.Byte, n.Valid
	err := n.NullByte.Scan(value)
	if !n.Valid {
		n.Byte, n.Valid = oldB, oldV
	}
	return err
}

func (n *jsonNullJob) Scan(value any) error {
	oldB, oldV := n.Byte, n.Valid
	err := n.NullByte.Scan(value)
	if !n.Valid {
		n.Byte, n.Valid = oldB, oldV
	}
	return err
}

// custom JSON reading for reward and job

type jsonNullReward struct {
	sql.NullString
	Pictos map[uint16]uint32
}

func (v *jsonNullReward) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	v.Pictos = make(map[uint16]uint32, 20)
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
				v.Pictos[uint16(token.(float64))] = val
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
				v.Pictos[idx] = uint32(token.(float64))
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

		case "dig":
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

	if err == io.EOF {
		err = errors.New(string(data) + " = job not recognized")
	}

	return err
}
