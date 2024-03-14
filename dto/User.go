package dto

type User struct {
	ID             int            `db:"id"`
	Name           string         `db:"name"`
	SimplifiedName string         `db:"simplified_name"`
	Avatar         jsonNullString `db:"avatar"`
}

type DetailedUser struct {
	User
	CreationCount      int
	ParticipationCount int
}
