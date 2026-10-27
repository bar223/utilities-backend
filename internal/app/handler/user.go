package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"utilities-backend/internal/app/ds"
	"utilities-backend/internal/app/repository"
)

// RegisterUser godoc
// @Summary      Регистрация
// @Description  Создаёт пользователя-жильца. Пароль хранится в виде bcrypt-хэша.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      ds.RegisterRequest  true  "логин и пароль"
// @Success      201   {object}  ds.User
// @Failure      400   {object}  errorResponse
// @Failure      409   {object}  errorResponse  "логин занят"
// @Router       /users/register [post]
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

// LoginUser godoc
// @Summary      Аутентификация
// @Description  Проверяет логин и пароль и выдаёт JWT. Токен передаётся в остальных запросах в заголовке Authorization: Bearer <токен>.
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      ds.RegisterRequest  true  "логин и пароль"
// @Success      200   {object}  ds.LoginResponse
// @Failure      400   {object}  errorResponse
// @Failure      401   {object}  errorResponse  "неверный логин или пароль"
// @Router       /users/login [post]
func (h *Handler) LoginUser(ctx *gin.Context) {
	var req ds.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.CheckUser(req.Login, req.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(h.Config.JWT.ExpiresIn)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "utilities-backend",
			ID:        uuid.NewString(), // jti: токены одного пользователя не совпадают
		},
		UserID:      user.ID,
		IsModerator: user.IsModerator,
	})
	tokenStr, err := token.SignedString([]byte(h.Config.JWT.Secret))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, ds.LoginResponse{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
		ExpiresIn:   int64(h.Config.JWT.ExpiresIn.Seconds()),
		Login:       user.Login,
	})
}

// LogoutUser godoc
// @Summary      Деавторизация
// @Description  Добавляет текущий JWT в чёрный список Redis до конца его срока жизни.
// @Tags         users
// @Security     BearerAuth
// @Success      204
// @Failure      401  {object}  errorResponse
// @Router       /users/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	// токен уже проверен в WithAuthCheck; в чёрном списке он нужен только
	// до момента, когда истёк бы сам
	tokenStr := ctx.GetString(ctxToken)
	ttl := time.Until(ctx.MustGet(ctxTokenExp).(time.Time))

	if err := h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), tokenStr, ttl); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
