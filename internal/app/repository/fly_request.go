package repository

import (
	"backend/internal/app/ds"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type CurrentRequestInfo struct {
	RequestID int
	RumbCount int
}

// GetCurrentRequestInfo возвращает информацию о текущей заявке пользователя

func (r *Repository) GetCurrentRequestInfo(userID int) (CurrentRequestInfo, error) {
	var info CurrentRequestInfo
	var request ds.FlyRequest
	err := r.db.Where("created_by_id = ? AND status = ?", userID, "draft").
		Order("created_at DESC").
		First(&request).Error
	if err == gorm.ErrRecordNotFound {
		return CurrentRequestInfo{0, 0}, nil
	} else if err != nil {
		return info, err
	}

	var count int64
	err = r.db.Model(&ds.FlyRequest_Rumb{}).
		Where("fly_request_id = ?", request.ID).
		Count(&count).Error
	if err != nil {
		return info, err
	}

	return CurrentRequestInfo{request.ID, int(count)}, nil
}

// GetFlyRequestsByUser возвращает заявки конкретного пользователя с фильтрацией
func (r *Repository) GetFlyRequestsByUser(userID int, status, formedAfter, formedBefore string) ([]ds.FlyRequest, error) {
	var reqs []ds.FlyRequest

	db := r.db.Model(&ds.FlyRequest{}).
		Where("created_by_id = ? AND status NOT IN ?", userID, []string{"deleted", "draft"})

	// Фильтр по статусу, если передан
	if status != "" {
		db = db.Where("status = ?", status)
	}

	// Фильтр по FormedAt
	if formedAfter != "" {
		db = db.Where("formed_at >= ?", formedAfter)
	}
	if formedBefore != "" {
		db = db.Where("formed_at <= ?", formedBefore)
	}

	err := db.Order("id ASC").Find(&reqs).Error
	return reqs, err
}

// GetFlyRequests возвращает список заявок на полет с фильтрацией

// GetFlyRequests возвращает все заявки (для модераторов) с фильтрацией
func (r *Repository) GetFlyRequests(status, formedAfter, formedBefore string) ([]ds.FlyRequest, error) {
	var reqs []ds.FlyRequest

	db := r.db.Model(&ds.FlyRequest{}).
		Where("status NOT IN ?", []string{"deleted", "draft"})

	// Фильтр по статусу, если передан
	if status != "" {
		db = db.Where("status = ?", status)
	}

	// Фильтр по FormedAt
	if formedAfter != "" {
		db = db.Where("formed_at >= ?", formedAfter)
	}
	if formedBefore != "" {
		db = db.Where("formed_at <= ?", formedBefore)
	}

	err := db.Order("id ASC").Find(&reqs).Error
	return reqs, err
}

// GetFlyRequestByID возвращает заявку на полет по ID

func (r *Repository) GetFlyRequestByID(id int) (*ds.FlyRequest, error) {
	var req ds.FlyRequest
	err := r.db.First(&req, id).Error
	return &req, err
}

// CreateFlyRequest создает новую заявку на полет

func (r *Repository) CreateFlyRequest(req *ds.FlyRequest) error {
	return r.db.Create(req).Error
}

// UpdateFlyRequest обновляет информацию о заявке на полет

func (r *Repository) UpdateFlyRequest(req *ds.FlyRequest) error {
	return r.db.Save(req).Error
}

// DeleteFlyRequest помечает заявку как удаленную (меняет статус на "deleted")
func (r *Repository) DeleteFlyRequest(id int) error {
	// Меняем статус заявки на "deleted"
	err := r.db.Model(&ds.FlyRequest{}).
		Where("id = ?", id).
		Update("status", "deleted").Error
	if err != nil {
		return fmt.Errorf("failed to delete fly request (set status to deleted): %v", err)
	}
	return nil
}

// AddRumbToRequest добавляет румб к заявке на полет
func (r *Repository) AddRumbToRequest(requestID, rumbID int) error {
	// Проверим, что заявка существует
	var request ds.FlyRequest
	if err := r.db.First(&request, requestID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("fly request not found")
		}
		return err
	}

	// Проверим, что румб существует
	var rumb ds.Rumb
	if err := r.db.First(&rumb, rumbID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("rumb not found")
		}
		return err
	}

	// Проверим, не добавлен ли уже этот румб в заявку
	var existing ds.FlyRequest_Rumb
	err := r.db.Where("fly_request_id = ? AND rumb_id = ?", requestID, rumbID).First(&existing).Error
	if err == nil {
		return errors.New("rumb already added to request")
	}

	// Добавляем с порядком (segment_order)
	var count int64
	r.db.Model(&ds.FlyRequest_Rumb{}).Where("fly_request_id = ?", requestID).Count(&count)

	frr := ds.FlyRequest_Rumb{
		FlyRequestID: uint(requestID),
		RumbID:       uint(rumbID),
		SegmentOrder: int(count + 1),
		DistanceKM:   0,
		WindSpeedKMH: 0,
		IsMain:       false,
	}

	return r.db.Create(&frr).Error
}

// GetFlyRequestByStatus возвращает заявку по статусу
func (r *Repository) GetFlyRequestByStatus(status string) (ds.FlyRequest, error) {
	var flyRequest ds.FlyRequest
	err := r.db.Where("status = ?", status).First(&flyRequest).Error
	if err != nil {
		return ds.FlyRequest{}, err
	}
	return flyRequest, nil
}

// GetMaxSegmentOrder возвращает максимальный порядковый номер сегмента в заявке

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

// CreateFlyRequestRumb создает связь между заявкой и румбом
func (r *Repository) CreateFlyRequestRumb(flyRequestRumb ds.FlyRequest_Rumb) error {
	err := r.db.Create(&flyRequestRumb).Error
	if err != nil {
		return fmt.Errorf("failed to create fly request rumb link: %v", err)
	}
	return nil
}

// DeleteFlyRequestRumb удаляет связь между заявкой и румбом
func (r *Repository) DeleteFlyRequestRumb(flyRequestID, rumbID uint) error {
	return r.db.
		Where("fly_request_id = ? AND rumb_id = ?", flyRequestID, rumbID).
		Delete(&ds.FlyRequest_Rumb{}).Error
}

// UpdateFlyRequestRumb обновляет параметры связи заявки и румба

func (r *Repository) UpdateFlyRequestRumb(flyRequestID, rumbID uint, distanceKM, windSpeedKMH float64) error {
	return r.db.Model(&ds.FlyRequest_Rumb{}).
		Where("fly_request_id = ? AND rumb_id = ?", flyRequestID, rumbID).
		Updates(map[string]interface{}{
			"distance_km":    distanceKM,
			"wind_speed_kmh": windSpeedKMH,
		}).Error
}

// GetByIDWithRumbs возвращает заявку с полной информацией о связанных румбах
func (r *Repository) GetByIDWithRumbs(id uint) (*ds.FlyRequest, []ds.FlyRequest_Rumb, error) {
	var req ds.FlyRequest
	if err := r.db.First(&req, id).Error; err != nil {
		return nil, nil, err
	}

	var rumbs []ds.FlyRequest_Rumb
	if err := r.db.Where("fly_request_id = ?", id).Find(&rumbs).Error; err != nil {
		return nil, nil, err
	}

	return &req, rumbs, nil
}

// UpdateStatus обновляет статус заявки на полет
func (r *Repository) UpdateStatus(id uint, status string, moderatorID *int) error {
	updates := map[string]interface{}{"status": status}
	if moderatorID != nil {
		updates["moderator_id"] = *moderatorID
	}
	result := r.db.Model(&ds.FlyRequest{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("request not found")
	}
	return nil
}

// GetDraftByUser возвращает черновую заявку пользователя и количество румбов в ней
func (r *Repository) GetDraftByUser(userID int) (*ds.FlyRequest, int64, error) {
	var fr ds.FlyRequest
	// Ищем черновик
	if err := r.db.Where("created_by_id = ? AND status = ?", userID, "draft").First(&fr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	// Считаем количество услуг в заявке
	var count int64
	if err := r.db.Model(&ds.FlyRequest_Rumb{}).Where("fly_request_id = ?", fr.ID).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	return &fr, count, nil
}
