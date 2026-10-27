package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"utilities-backend/internal/app/ds"
	"utilities-backend/internal/app/repository"
)

// RegisterUser — POST /api/users/register: регистрация нового
// пользователя. Тело — JSON {login, password}, ответ 201 с созданным
// пользователем (пароль не сериализуется).
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var req ds.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if req.Login == "" || req.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, errors.New("login и password обязательны"))
		return
	}

	user, err := h.Repository.CreateUser(req.Login, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrLoginTaken) {
			h.errorHandler(ctx, http.StatusConflict, err)
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, user)
}

// LoginUser — POST /api/users/login: аутентификация, заглушка до
// лабораторной 4 (JWT).
func (h *Handler) LoginUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "аутентификация будет реализована в лабораторной 4"})
}

// LogoutUser — POST /api/users/logout: деавторизация, заглушка до
// лабораторной 4 (blacklist JWT).
func (h *Handler) LogoutUser(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "деавторизация будет реализована в лабораторной 4"})
}
