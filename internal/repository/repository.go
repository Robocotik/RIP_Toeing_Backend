package repository

import (
	"backend/mock"
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

func (r *Repository) GetRumbs() ([]mock.Rumb, error) {
	// имитируем работу с БД. Типа мы выполнили sql запрос и получили эти строки из БД
	rumbs := mock.Rumbs
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	// тут я снова искусственно обработаю "ошибку" чисто чтобы показать вам как их передавать выше
	if len(rumbs) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return rumbs, nil
}

func (r *Repository) GetRumb(id int) (mock.Rumb, error) {
	// тут у вас будет логика получения нужной услуги, тоже наверное через цикл в первой лабе, и через запрос к БД начиная со второй
	rumbs, err := r.GetRumbs()
	if err != nil {
		return mock.Rumb{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, rumb := range rumbs {
		if rumb.ID == id {
			return rumb, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return mock.Rumb{}, fmt.Errorf("заказ не найден") // тут нужна кастомная ошибка, чтобы понимать на каком этапе возникла ошибка и что произошло
}

func (r *Repository) GetRumbsByTitle(title string) ([]mock.Rumb, error) {
	rumbs, err := r.GetRumbs()
	if err != nil {
		return []mock.Rumb{}, err
	}

	var result []mock.Rumb
	for _, rumb := range rumbs {
		if strings.Contains(strings.ToLower(rumb.Title), strings.ToLower(title)) {
			result = append(result, rumb)
		}
	}

	return result, nil
}