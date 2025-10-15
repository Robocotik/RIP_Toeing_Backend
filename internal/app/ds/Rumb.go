package ds

type Rumb struct {
	// Уникальный идентификатор направления
	ID int `gorm:"primaryKey;autoIncrement;not null" json:"id" example:"1"`
	// Название направления
	Title string `gorm:"type:varchar(200);not null;unique" json:"title" example:"Север-Восток"`
	// Ссылка на изображение направления
	Image string `gorm:"type:varchar(200)" json:"image" example:"/images/north-east.png"`
}
