package ds

type Rumb struct {
	ID    int `gorm:"primaryKey;autoIncrement"`
	Title string `gorm:"type:varchar(200);not null"`
	Image string `gorm:"type:varchar(200)"`
}