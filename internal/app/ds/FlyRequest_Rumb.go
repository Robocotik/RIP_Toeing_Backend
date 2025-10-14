package ds

type FlyRequest_Rumb struct {
	FlyRequestID uint `gorm:"primaryKey;not null"`
	RumbID       uint `gorm:"primaryKey;not null"`
	SegmentOrder int  `gorm:"primaryKey;not null;autoIncrement"`

	DistanceKM   float64 `gorm:"not null"`
	WindSpeedKMH float64 `gorm:"not null"`
	IsMain       bool    `gorm:"default:false"`

	FlyRequest FlyRequest `gorm:"foreignKey:FlyRequestID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Rumb       Rumb       `gorm:"foreignKey:RumbID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
