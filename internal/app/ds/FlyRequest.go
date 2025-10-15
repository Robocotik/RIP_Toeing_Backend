package ds

import "time"

type FlyRequest struct {
	ID           int       `gorm:"primaryKey;autoIncrement;not null"`
	Status       string    `gorm:"type:varchar(20);not null"`
	CreatedByID  int       `gorm:"type:int;not null"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:current_timestamp"`
	FormedAt     time.Time `gorm:"type:timestamp;null"`
	ModeratorID  int       `gorm:"type:int;null"`
	CalculatedBy string    `gorm:"type:varchar(100);null"`
}
