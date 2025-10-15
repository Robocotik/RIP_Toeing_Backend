package repository

import (
	"backend/internal/app/ds"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterUser регистрирует нового пользователя в системе
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

// AuthenticateUser ищет пользователя по логину и сверяет пароль
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
		return nil, errors.New("invalid credentials")
	}

	return &user, nil
}

// GetUserByID возвращает пользователя по ID
func (r *Repository) GetUserByID(id int) (ds.User, error) {
	var user ds.User
	if err := r.db.First(&user, id).Error; err != nil {
		return ds.User{}, fmt.Errorf("user not found")
	}
	return user, nil
}

// UpdateUser обновляет данные пользователя
func (r *Repository) UpdateUser(user *ds.User) error {
	// Если пароль передан — хэшируем
	if user.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %v", err)
		}
		user.Password = string(hash)
	}

	return r.db.Model(&ds.User{}).Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"login":    user.Login,
			"password": user.Password,
		}).Error
}