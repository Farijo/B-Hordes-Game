package dto

type Participant struct {
	User      int `db:"user"`
	Challenge int `db:"challenge"`
}
