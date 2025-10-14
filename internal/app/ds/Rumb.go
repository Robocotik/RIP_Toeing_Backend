package ds

type Rumb struct {
	ID    int    `gorm:"primaryKey;autoIncrement;not null"`
	Title string `gorm:"type:varchar(200);not null;unique"`
	Image string `gorm:"type:varchar(200)"`
}

