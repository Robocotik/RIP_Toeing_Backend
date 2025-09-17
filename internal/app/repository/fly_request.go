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

type CurrentRequestInfo struct {
	RequestID int `json:"request_id"`
	RumbCount int `json:"rumb_count"`
}

func (r *Repository) GetCurrentRequestInfo(userID int) (CurrentRequestInfo, error) {
	var info CurrentRequestInfo

	err := r.db.Table("fly_requests fr").
		Select("fr.id as request_id, COUNT(frr.rumb_id) as rumb_count").
		Joins("LEFT JOIN fly_request_rumbs frr ON fr.id = frr.fly_request_id").
		Where("fr.status = ? AND fr.created_by_id = ?", "created", userID).
		Group("fr.id").
		Order("fr.id DESC").
		Limit(1).
		Scan(&info).Error

	if err != nil {
		return CurrentRequestInfo{}, fmt.Errorf("failed to get current request info: %v", err)
	}

	return info, nil
}

func (r *Repository) GetFlyRequestByID(id int) (ds.FlyRequest, error) {
	var flyRequest ds.FlyRequest
	err := r.db.First(&flyRequest, id).Error
	if err != nil {
		return ds.FlyRequest{}, fmt.Errorf("failed to get fly request by ID: %v", err)
	}
	return flyRequest, nil
}


func (r *Repository) UpdateFlyRequestStatus(id int, status string) error {
	err := r.db.Model(&ds.FlyRequest{}).
		Where("id = ?", id).
		Update("status", status).Error
	if err != nil {
		return fmt.Errorf("failed to update fly request status: %v", err)
	}
	return nil
}


func (r *Repository) GetAllFlyRequests() ([]ds.FlyRequest, error) {
	var flyRequests []ds.FlyRequest
	err := r.db.Find(&flyRequests).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get all fly requests: %v", err)
	}
	return flyRequests, nil
}
