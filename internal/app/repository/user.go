package repository

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"utilities-backend/internal/app/ds"
)

// ErrLoginTaken возвращается при попытке зарегистрировать уже занятый
// логин.
var ErrLoginTaken = errors.New("логин уже занят")

// ErrBadCredentials — неверный логин или пароль.
var ErrBadCredentials = errors.New("неверный логин или пароль")

// CreateUser регистрирует нового пользователя. В БД хранится не пароль, а
// его bcrypt-хэш.
func (r *Repository) CreateUser(login, password string) (*ds.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := ds.User{Login: login, Password: string(hash)}
	if err := r.db.Create(&user).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, ErrLoginTaken
		}
		return nil, err
	}
	return &user, nil
}

// CheckUser находит пользователя по логину и сверяет пароль с хэшем.
func (r *Repository) CheckUser(login, password string) (*ds.User, error) {
	var user ds.User
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, ErrBadCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, ErrBadCredentials
	}
	return &user, nil
}
