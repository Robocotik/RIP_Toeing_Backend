package repository

import (
	"backend/internal/app/ds"
	"backend/mock"
	"fmt"
)

func (r *Repository) GetRumbs() ([]ds.Rumb, error) {
	var rumbs []ds.Rumb
	err := r.db.Find(&rumbs).Error
	if err != nil {
		return nil, err
	}
	if len(rumbs) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return rumbs, nil
}

func (r *Repository) GetRumb(id int) (ds.Rumb, error) {
	rumb := ds.Rumb{}
	err := r.db.Where("id = ?", id).First(&rumb).Error
	if err != nil {
		return ds.Rumb{}, err
	}
	return rumb, nil
}

func (r *Repository) GetRumbsByTitle(title string) ([]ds.Rumb, error) {
	var rumbs []ds.Rumb
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&rumbs).Error
	if err != nil {
		return nil, err
	}
	return rumbs, nil
}

func (r *Repository) GetFlyRequest() (mock.Rumb, error) {
	rumbs := mock.Rumbs
	return rumbs[0], nil
}

func (r *Repository) GetRumbsByFlyRequestID(flyRequestID int) ([]ds.Rumb, error) {
	var rumbs []ds.Rumb
	
	err := r.db.Table("rumbs r").
		Select("r.id, r.title, r.image").
		Joins("INNER JOIN fly_request_rumbs frr ON r.id = frr.rumb_id").
		Where("frr.fly_request_id = ?", flyRequestID).
		Order("frr.segment_order ASC").
		Scan(&rumbs).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get rumbs by fly request ID: %v", err)
	}
	fmt.Print("из бд я вернул ", )
	return rumbs, nil
}