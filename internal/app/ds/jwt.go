package ds

import "github.com/golang-jwt/jwt/v5"

// JWTClaims — содержимое токена: стандартные поля (срок жизни, издатель)
// и данные пользователя, по которым бэкенд определяет текущего
// пользователя и его роль без обращения к БД.
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      uint `json:"userId"`
	IsModerator bool `json:"isModerator"`
}
