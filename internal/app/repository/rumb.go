package repository

import (
	"backend/internal/app/ds"
	"backend/mock"
	"fmt"
)

func (r *Repository) GetRumbs() ([]ds.Rumb, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	var rumbs []ds.Rumb
	err := r.db.Find(&rumbs).Error

	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if err != nil {
		return nil, err
	}
	if len(rumbs) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return rumbs, nil
}

func (r *Repository) GetRumb(id int) (ds.Rumb, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
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
