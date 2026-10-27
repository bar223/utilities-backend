package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"utilities-backend/internal/app/repository"
)

// registerRequest — тело запроса регистрации.
type registerRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// APIRegister — POST /api/users/register: регистрация нового
// пользователя.
func (h *Handler) APIRegister(ctx *gin.Context) {
	var req registerRequest
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

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": user})
}

// APILogin — POST /api/users/login: заглушка, настоящая авторизация
// (сессии/JWT) появится в лабораторной 4.
func (h *Handler) APILogin(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "авторизация будет реализована в лабораторной 4"})
}

// APILogout — POST /api/users/logout: заглушка, см. APILogin.
func (h *Handler) APILogout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "success", "message": "деавторизация будет реализована в лабораторной 4"})
}
