package ds

// FlyRequest_Rumb представляет связь между заявкой на полет и направлением (румбом)
// @Description Модель связи заявки на полет и направления с порядком сегментов
type FlyRequest_Rumb struct {
	// ID заявки на полет (внешний ключ)
	FlyRequestID uint `gorm:"primaryKey;not null" json:"fly_request_id" example:"1"`
	// ID направления (румба) (внешний ключ)
	RumbID uint `gorm:"primaryKey;not null" json:"rumb_id" example:"2"`
	// Порядковый номер сегмента в маршруте
	SegmentOrder int `gorm:"primaryKey;not null;autoIncrement" json:"segment_order" example:"1"`

	// Расстояние сегмента в километрах
	DistanceKM float64 `gorm:"not null" json:"distance_km" example:"150.5"`
	// Скорость ветра в км/ч
	WindSpeedKMH float64 `gorm:"not null" json:"wind_speed_kmh" example:"25.3"`
	// Флаг основного сегмента маршрута
	IsMain bool `gorm:"default:false" json:"is_main" example:"true"`

	// Связанная заявка на полет
	FlyRequest FlyRequest `gorm:"foreignKey:FlyRequestID;references:ID;constraint:OnUpdate:CASCADE" json:"fly_request"`
	// Связанное направление (румб)
	Rumb Rumb `gorm:"foreignKey:RumbID;references:ID;constraint:OnUpdate:CASCADE" json:"rumb"`
}
