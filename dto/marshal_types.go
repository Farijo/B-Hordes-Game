package dto

import (
	"database/sql"
	"encoding/json"
)

type jsonNullBool struct {
	sql.NullBool
}

func (v *jsonNullBool) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *bool
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Bool = *x
	} else {
		v.Valid = false
	}
	return nil
}

type jsonNullByte struct {
	sql.NullByte
}

func (v *jsonNullByte) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *byte
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.Byte = *x
	} else {
		v.Valid = false
	}
	return nil
}

type jsonNullString struct {
	sql.NullString
}

func (v *jsonNullString) UnmarshalJSON(data []byte) error {
	// Unmarshalling into a pointer will let us detect null
	var x *string
	if err := json.Unmarshal(data, &x); err != nil {
		return err
	}
	if x != nil {
		v.Valid = true
		v.String = *x
	} else {
		v.Valid = false
	}
	return nil
}

type jsonNullCounter struct {
	sql.NullInt64
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
