package dto

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

type jsonNullBool struct {
	sql.NullBool
}

type jsonNullByte struct {
	sql.NullByte
}

type jsonNullInt16 struct {
	sql.NullInt16
}

type jsonNullFloat struct {
	sql.Null[float64]
}

type jsonNullString struct {
	sql.NullString
}

type jsonNullCounter struct {
	sql.NullInt64
}

type jsonNullDict struct {
	sql.NullString
	Data map[uint16]uint32
}

type jsonNullList struct {
	sql.NullString
	Data map[uint16]bool
}

type jsonNullJob struct {
	sql.NullByte
}

type jsonNullRegen struct {
	sql.NullByte
}

type jsonNullAux[T any] struct {
	V T
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * *
 * * * * * * * * * * UnmarshalJSON * * * * * * * * * *
 * * * * * * * * * * * * * * * * * * * * * * * * * * */

func unmarshalGen[T any](validity *bool, value *T, data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *T
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		*validity = true
		*value = *x
	} else {
		*validity = false
	}
	return nil
}

func (v *jsonNullBool) UnmarshalJSON(data []byte) error {
	return unmarshalGen(&v.Valid, &v.Bool, data)
}

func (v *jsonNullByte) UnmarshalJSON(data []byte) error {
	return unmarshalGen(&v.Valid, &v.Byte, data)
}

func (v *jsonNullInt16) UnmarshalJSON(data []byte) error {
	return unmarshalGen(&v.Valid, &v.Int16, data)
}

func (v *jsonNullFloat) UnmarshalJSON(data []byte) error {
	return unmarshalGen(&v.Valid, &v.V, data)
}

func (v *jsonNullString) UnmarshalJSON(data []byte) error {
	return unmarshalGen(&v.Valid, &v.String, data)
}

func (v *jsonNullCounter) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *[]any
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Int64 = int64(len(*x))
	} else {
		v.Valid = false
	}
	return nil
}

func (v *jsonNullDict) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	v.Data = make(map[uint16]uint32, 20)
	decoder := json.NewDecoder(bytes.NewReader(data))
	idx, val := uint16(0), uint32(0)
	for token, err := decoder.Token(); err == nil; token, err = decoder.Token() {
		switch token {
		case "id", "buildingId":
			token, err = decoder.Token()
			if err != nil {
				return err
			}
			if val > 0 {
				v.Data[uint16(token.(float64))] = val
				val = 0
			} else {
				idx = uint16(token.(float64))
			}
		case "number", "count", "level":
			token, err = decoder.Token()
			if err != nil {
				return err
			}
			if idx > 0 {
				v.Data[idx] = uint32(token.(float64))
				idx = 0
			} else {
				val = uint32(token.(float64))
			}
		}
	}

	return nil
}

func (v *jsonNullList) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	v.Data = make(map[uint16]bool, 10)
	decoder := json.NewDecoder(bytes.NewReader(data))
	for token, err := decoder.Token(); err == nil; token, err = decoder.Token() {
		switch token {
		case "id":
			token, err = decoder.Token()
			if err != nil {
				return err
			}
			v.Data[uint16(token.(float64))] = true
		}
	}

	return nil
}

func (v *jsonNullJob) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))

	token, err := decoder.Token()

	for ; err == nil; token, err = decoder.Token() {
		switch token {
		case "basic":
			v.Byte = 0
			v.Valid = true
			return nil

		case "dig":
			v.Byte = 1
			v.Valid = true
			return nil

		case "vest":
			v.Byte = 2
			v.Valid = true
			return nil

		case "shield":
			v.Byte = 3
			v.Valid = true
			return nil

		case "book":
			v.Byte = 4
			v.Valid = true
			return nil

		case "tamer":
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

func (v *jsonNullRegen) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))

	token, err := decoder.Token()

	for ; err == nil; token, err = decoder.Token() {
		switch token {
		default:
			v.Byte = 0
			v.Valid = true
			return nil
		}
	}

	if err == io.EOF {
		err = errors.New(string(data) + " = direction not recognized")
	}

	return err
}

func (v *jsonNullAux[any]) UnmarshalJSON(data []byte) error {
	if data[0] == '[' {
		return nil
	}
	return json.NewDecoder(bytes.NewReader(data)).Decode(&v.V)
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * *
 * * * * * * * * * * * * Scan  * * * * * * * * * * * *
 * * * * * * * * * * * * * * * * * * * * * * * * * * */

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

func (n *jsonNullDict) Scan(value any) error {
	wasValid := n.Valid
	err := n.NullString.Scan(value)
	if n.Valid {
		if n.Data == nil {
			n.Data = make(map[uint16]uint32, 20)
		}
		nsLen := len(n.String)
		for i := 0; i < nsLen; i += 6 {
			id := binary.LittleEndian.Uint16([]byte(n.String[i : i+2]))
			number := binary.LittleEndian.Uint32([]byte(n.String[i+2 : i+6]))
			n.Data[id] = number
		}
	} else {
		n.Valid = wasValid
	}
	return err
}

func (n *jsonNullList) Scan(value any) error {
	wasValid := n.Valid
	err := n.NullString.Scan(value)
	if n.Valid {
		if n.Data == nil {
			n.Data = make(map[uint16]bool, 10)
		}
		nsLen := len(n.String)
		for i := 0; i < nsLen; i += 3 {
			id := binary.LittleEndian.Uint16([]byte(n.String[i : i+2]))
			is := n.String[i+2] != '0'
			n.Data[id] = is
		}
	} else {
		n.Valid = wasValid
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

func (n *jsonNullInt16) Scan(value any) error {
	oldB, oldV := n.Int16, n.Valid
	err := n.NullInt16.Scan(value)
	if !n.Valid {
		n.Int16, n.Valid = oldB, oldV
	}
	return err
}

func (n *jsonNullFloat) Scan(value any) error {
	oldB, oldV := n.V, n.Valid
	err := n.Null.Scan(value)
	if !n.Valid {
		n.V, n.Valid = oldB, oldV
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

func (n *jsonNullRegen) Scan(value any) error {
	oldB, oldV := n.Byte, n.Valid
	err := n.NullByte.Scan(value)
	if !n.Valid {
		n.Byte, n.Valid = oldB, oldV
	}
	return err
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * *
 * * * * * * * * KeepDifferencesOnly * * * * * * * * *
 * * * * * * * * * * * * * * * * * * * * * * * * * * */

func (v *jsonNullDict) KeepDifferencesOnly(other jsonNullDict) bool {
	changements := make([]byte, 0, 120)
	v.Valid = false
	for id, number := range v.Data {
		if number != other.Data[id] {
			changements = binary.LittleEndian.AppendUint16(changements, id)
			changements = binary.LittleEndian.AppendUint32(changements, number)

			v.Valid = true
		}
	}
	// to test : raz
	for id, old := range other.Data {
		if _, ok := v.Data[id]; !ok && old > 0 {
			changements = binary.LittleEndian.AppendUint16(changements, id)
			changements = binary.LittleEndian.AppendUint32(changements, 0)

			v.Valid = true
		}
	}

	if v.Valid {
		v.String = string(changements)
		return true
	}

	return false
}

func (v *jsonNullList) KeepDifferencesOnly(other jsonNullList) bool {
	changements := make([]byte, 0, 120)
	v.Valid = false
	for id, is := range v.Data {
		if is != other.Data[id] {
			changements = binary.LittleEndian.AppendUint16(changements, id)
			changements = append(changements, '1')

			v.Valid = true
		}
	}
	// to test : raz
	for id, old := range other.Data {
		if _, ok := v.Data[id]; !ok && old {
			changements = binary.LittleEndian.AppendUint16(changements, id)
			changements = append(changements, '0')

			v.Valid = true
		}
	}

	if v.Valid {
		v.String = string(changements)
		return true
	}

	return false
}
