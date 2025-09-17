package repository

import (
	"backend/internal/app/ds"
	"fmt"
)

func (r *Repository) CreateFlyRequest(flyRequest *ds.FlyRequest) error {
	err := r.db.Create(flyRequest).Error
	if err != nil {
		return fmt.Errorf("failed to create fly request: %v", err)
	}
	return nil
}

func (r *Repository) GetFlyRequestByStatus(status string) (ds.FlyRequest, error) {
	var flyRequest ds.FlyRequest
	err := r.db.Where("status = ?", status).First(&flyRequest).Error
	if err != nil {
		return ds.FlyRequest{}, err
	}
	return flyRequest, nil
}

func (r *Repository) GetMaxSegmentOrder(flyRequestID int) (int, error) {
	var maxOrder int
	err := r.db.Model(&ds.FlyRequest_Rumb{}).
		Where("fly_request_id = ?", flyRequestID).
		Select("COALESCE(MAX(segment_order), 0)").
		Scan(&maxOrder).Error
	if err != nil {
		return 0, fmt.Errorf("failed to get max segment order: %v", err)
	}
	return maxOrder, nil
}

func (r *Repository) CreateFlyRequestRumb(flyRequestRumb ds.FlyRequest_Rumb) error {
	err := r.db.Create(&flyRequestRumb).Error
	if err != nil {
		return fmt.Errorf("failed to create fly request rumb link: %v", err)
	}
	return nil
}