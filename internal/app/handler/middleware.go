package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"utilities-backend/internal/app/ds"
)

const (
	bearerPrefix   = "Bearer "
	ctxUserID      = "userID"
	ctxIsModerator = "isModerator"
	ctxToken       = "token"
	ctxTokenExp    = "tokenExpiresAt"
)

var (
	errUnauthorized = errors.New("требуется авторизация")
	errTokenRevoked = errors.New("токен отозван (выполнен выход)")
)

// parseToken достаёт JWT из заголовка Authorization, проверяет подпись,
// срок жизни и отсутствие в чёрном списке Redis.
func (h *Handler) parseToken(ctx *gin.Context) (string, *ds.JWTClaims, error) {
	header := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(header, bearerPrefix) {
		return "", nil, errUnauthorized
	}
	tokenStr := strings.TrimPrefix(header, bearerPrefix)

	claims := &ds.JWTClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return []byte(h.Config.JWT.Secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return "", nil, errUnauthorized
	}

	revoked, err := h.Redis.IsJWTInBlacklist(ctx.Request.Context(), tokenStr)
	if err != nil {
		return "", nil, err
	}
	if revoked {
		return "", nil, errTokenRevoked
	}
	return tokenStr, claims, nil
}

// WithAuthCheck — middleware методов создателя: без действующего токена
// запрос прерывается с кодом 401.
func (h *Handler) WithAuthCheck(ctx *gin.Context) {
	tokenStr, claims, err := h.parseToken(ctx)
	if err != nil {
		code := http.StatusUnauthorized
		if !errors.Is(err, errUnauthorized) && !errors.Is(err, errTokenRevoked) {
			code = http.StatusInternalServerError
		}
		h.errorHandler(ctx, code, err)
		return
	}
	ctx.Set(ctxUserID, claims.UserID)
	ctx.Set(ctxIsModerator, claims.IsModerator)
	ctx.Set(ctxToken, tokenStr)
	ctx.Set(ctxTokenExp, claims.ExpiresAt.Time)
}

// WithOptionalAuth — middleware методов чтения: работают и для гостя, а
// при действующем токене дополнительно вычисляются признаки isMine/liked.
func (h *Handler) WithOptionalAuth(ctx *gin.Context) {
	if _, claims, err := h.parseToken(ctx); err == nil {
		ctx.Set(ctxUserID, claims.UserID)
		ctx.Set(ctxIsModerator, claims.IsModerator)
	}
}

// currentUserID — текущий пользователь из JWT (заменяет функцию-синглтон
// лабораторной 3). 0 — гость.
func currentUserID(ctx *gin.Context) uint {
	if id, ok := ctx.Get(ctxUserID); ok {
		return id.(uint)
	}
	return 0
}
