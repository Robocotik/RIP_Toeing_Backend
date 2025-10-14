package repository

import (
	"backend/internal/app/ds"
	// "fmt"
)

func (r *Repository) GetRumbs() ([]ds.Rumb, error) {
	var rumbs []ds.Rumb
	err := r.db.Order("id ASC").Find(&rumbs).Error
	return rumbs, err
}

func (r *Repository) GetRumbsByTitle(title string) ([]ds.Rumb, error) {
	var rumbs []ds.Rumb
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&rumbs).Error
	return rumbs, err
}

func (r *Repository) GetRumbByID(id int) (*ds.Rumb, error) {
	var rumb ds.Rumb
	err := r.db.First(&rumb, id).Error
	return &rumb, err
}

func (r *Repository) CreateRumb(rumb *ds.Rumb) error {
	rumb.ID = 0
	return r.db.Create(rumb).Error
}

func (r *Repository) UpdateRumb(rumb *ds.Rumb) error {
	return r.db.Save(rumb).Error
}

func (r *Repository) DeleteRumb(id int) error {
	return r.db.Delete(&ds.Rumb{}, id).Error
}


// func (r *Repository) GetRumbsByFlyRequestID(flyRequestID int) ([]ds.Rumb, error) {
// 	var rumbs []ds.Rumb
	
// 	err := r.db.Table("rumbs r").
// 		Select("r.id, r.title, r.image").
// 		Joins("INNER JOIN fly_request_rumbs frr ON r.id = frr.rumb_id").
// 		Where("frr.fly_request_id = ?", flyRequestID).
// 		Order("frr.segment_order ASC").
// 		Scan(&rumbs).Error
	
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get rumbs by fly request ID: %v", err)
// 	}
// 	// fmt.Print("из бд я вернул ", rumbs)
// 	return rumbs, nil
// }


// Получаем все румбы для заявки с информацией о румбе
func (r *Repository) GetRumbsByFlyRequestID(flyRequestID int) ([]ds.FlyRequest_Rumb, error) {
	var rumbs []ds.FlyRequest_Rumb
	err := r.db.
		Where("fly_request_id = ?", flyRequestID).
		Preload("Rumb").
		Find(&rumbs).Error
	return rumbs, err
}