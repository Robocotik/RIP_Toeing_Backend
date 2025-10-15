package repository

import (
	"backend/internal/app/ds"
	"errors"
	"fmt"

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

	fmt.Printf("Registering user: %s, password: %s\n", user.Login, user.Password)

	// Сохраняем пользователя БЕЗ хеширования пароля
	if err := r.db.Create(user).Error; err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		return err
	}

	fmt.Printf("User registered successfully: %s, ID: %d\n", user.Login, user.ID)
	return nil
}

// AuthenticateUser аутентифицирует пользователя по логину и паролю
func (r *Repository) AuthenticateUser(login, password string) (*ds.User, error) {
	fmt.Printf("Authenticating user: %s, password: %s\n", login, password)
	
	var user ds.User
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("User not found in database: %s\n", login)
			return nil, errors.New("invalid credentials")
		}
		fmt.Printf("Database error: %v\n", err)
		return nil, err
	}

	fmt.Printf("Found user: %s, ID: %d, password in DB: %s\n", user.Login, user.ID, user.Password)

	// Прямое сравнение паролей БЕЗ хеширования
	if user.Password != password {
		fmt.Printf("Password mismatch: expected %s, got %s\n", user.Password, password)
		return nil, errors.New("invalid credentials")
	}

	fmt.Printf("Authentication successful for user: %s\n", user.Login)
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
	return r.db.Model(&ds.User{}).Where("id = ?", user.ID).
		Updates(map[string]interface{}{
			"login":    user.Login,
			"password": user.Password, // Без хеширования
		}).Error
}