package role

type Role int

const (
	User   Role = iota // 0 - обычный пользователь
	Moderator             // 1 - модератор  
	Admin               // 2 - администратор
)

// String returns string representation of role
func (r Role) String() string {
	switch r {
	case User:
		return "user"
	case Moderator:
		return "moderator"
	case Admin:
		return "admin"
	default:
		return "unknown"
	}
}

// IsModerator returns true if role has moderator privileges
func (r Role) IsModerator() bool {
	return r == Moderator || r == Admin
}