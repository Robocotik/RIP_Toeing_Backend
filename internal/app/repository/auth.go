package repository

import (
	"backend/internal/app/ds"
	"errors"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)


func (r *Repository) RegisterUser(user *ds.User) error {
	// Проверка существующего логина
	var existing ds.User
	if err := r.db.Where("login = ?", user.Login).First(&existing).Error; err == nil {
		return errors.New("user with this login already exists")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// Хешируем пароль
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)

	// Сохраняем пользователя
	if err := r.db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

// AuthenticateUser ищет пользователя по логину и сверяет пароль.
// Возвращает найденного пользователя (с заполнённым полем Password — хеш) при успехе.
// Если логин/пароль неверны — возвращает ошибку.
func (r *Repository) AuthenticateUser(login, password string) (*ds.User, error) {
	var user ds.User
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Сравниваем хеш пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		// неверный пароль
		return nil, errors.New("invalid credentials")
	}

	// Успешная аутентификация
	return &user, nil
}
