package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"utilities-backend/internal/app/repository"
)

// RegisterPage — GET /users/register: страница регистрации.
func (h *Handler) RegisterPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "register.html", gin.H{})
}

// RegisterSubmit — POST /users/register: создаёт пользователя по форме
// (обычная HTML-форма, без JS) и показывает результат на той же странице.
func (h *Handler) RegisterSubmit(ctx *gin.Context) {
	login := ctx.PostForm("login")
	password := ctx.PostForm("password")

	if login == "" || password == "" {
		ctx.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Error": "Заполните логин и пароль",
			"Login": login,
		})
		return
	}

	_, err := h.Repository.CreateUser(login, password)
	if err != nil {
		errText := "Не удалось зарегистрироваться"
		if errors.Is(err, repository.ErrLoginTaken) {
			errText = "Логин уже занят"
		}
		ctx.HTML(http.StatusOK, "register.html", gin.H{
			"Error": errText,
			"Login": login,
		})
		return
	}

	ctx.HTML(http.StatusOK, "register.html", gin.H{
		"Success": true,
		"Login":   login,
	})
}
