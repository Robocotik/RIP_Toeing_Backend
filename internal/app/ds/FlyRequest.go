package ds

import "time"

// FlyRequest представляет заявку на полет
// @Description Модель заявки на полет с информацией о статусе и времени создания
type FlyRequest struct {
	// Уникальный идентификатор заявки
	ID int `gorm:"primaryKey;autoIncrement;not null" json:"id" example:"1"`
	// Статус заявки (например: "pending", "approved", "rejected")
	Status string `gorm:"type:varchar(20);not null" json:"status" example:"pending"`
	// ID пользователя, создавшего заявку
	CreatedByID int `gorm:"type:int;not null" json:"created_by_id" example:"123"`
	// Дата и время создания заявки
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:current_timestamp" json:"created_at" example:"2023-10-05T14:30:00Z"`
	// Дата и время формирования заявки
	FormedAt time.Time `gorm:"type:timestamp;null" json:"formed_at" example:"2023-10-05T15:00:00Z"`
	// ID модератора, обработавшего заявку
	ModeratorID int `gorm:"type:int;null" json:"moderator_id" example:"456"`
	// Имя/идентификатор системы или пользователя, выполнившего расчет
	CalculatedBy string `gorm:"type:varchar(100);null" json:"calculated_by" example:"auto_calc_system"`
}
