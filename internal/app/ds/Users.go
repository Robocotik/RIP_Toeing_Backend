package ds

type User struct {
	ID          int    `gorm:"primaryKey;autoIncrement"`
	Login       string `gorm:"type:varchar(200);not null;unique"`
	Password    string `gorm:"type:varchar(200);not null"`
	IsModerator bool   `gorm:"type:boolean;default:false"`
}
