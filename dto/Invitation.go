package dto

type Invitation struct {
	User      int `db:"user"`
	Challenge int `db:"challenge"`
}
