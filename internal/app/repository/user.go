package repository

import (
	"errors"
	"strings"

	"utilities-backend/internal/app/ds"
)

// ErrLoginTaken возвращается при попытке зарегистрировать уже занятый
// логин.
var ErrLoginTaken = errors.New("логин уже занят")

// CreateUser регистрирует нового пользователя.
func (r *Repository) CreateUser(login, password string) (*ds.User, error) {
	user := ds.User{Login: login, Password: password}
	if err := r.db.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrLoginTaken
		}
		return nil, err
	}
	return &user, nil
}
