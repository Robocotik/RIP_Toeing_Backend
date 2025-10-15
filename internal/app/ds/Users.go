package ds

type User struct {
	// Уникальный идентификатор пользователя
	ID int `gorm:"primaryKey;autoIncrement" json:"id" example:"1"`
	// Логин пользователя (уникальный)
	Login string `gorm:"type:varchar(200);not null;unique" json:"login" example:"user123"`
	// Пароль пользователя (захешированный)
	Password string `gorm:"type:varchar(200);not null" json:"password"`
	// Флаг, указывающий имеет ли пользователь права модератора
	IsModerator bool `gorm:"type:boolean;default:false" json:"is_moderator" example:"false"`
}
