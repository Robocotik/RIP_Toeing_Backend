package repository

import (
	"backend/internal/app/ds"
	// "fmt"
)

// GetRumbs возвращает список всех румбов отсортированных по ID
func (r *Repository) GetRumbs() ([]ds.Rumb, error) {
	var rumbs []ds.Rumb
	err := r.db.Order("id ASC").Find(&rumbs).Error
	return rumbs, err
}

// GetRumbsByTitle возвращает румбы по частичному совпадению названия (регистронезависимо)
func (r *Repository) GetRumbsByTitle(title string) ([]ds.Rumb, error) {
	var rumbs []ds.Rumb
	err := r.db.Where("title ILIKE ?", "%"+title+"%").Find(&rumbs).Error
	return rumbs, err
}

// GetRumbByID возвращает румб по ID
func (r *Repository) GetRumbByID(id int) (*ds.Rumb, error) {
	var rumb ds.Rumb
	err := r.db.First(&rumb, id).Error
	return &rumb, err
}

// CreateRumb создает новый румб
func (r *Repository) CreateRumb(rumb *ds.Rumb) error {
	rumb.ID = 0
	return r.db.Create(rumb).Error
}

// UpdateRumb обновляет информацию о румбе
func (r *Repository) UpdateRumb(rumb *ds.Rumb) error {
	return r.db.Save(rumb).Error
}

// DeleteRumb удаляет румб из системы
func (r *Repository) DeleteRumb(id int) error {
	return r.db.Delete(&ds.Rumb{}, id).Error
}

// GetRumbsByFlyRequestID возвращает все румбы связанные с заявкой на полет
// Включает полную информацию о связи (расстояние, скорость ветра и т.д.)
func (r *Repository) GetRumbsByFlyRequestID(flyRequestID int) ([]ds.FlyRequest_Rumb, error) {
	var rumbs []ds.FlyRequest_Rumb
	err := r.db.
		Where("fly_request_id = ?", flyRequestID).
		Preload("Rumb").
		Find(&rumbs).Error
	return rumbs, err
}
