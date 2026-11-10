package handler

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"utilities-backend/internal/app/config"
	"utilities-backend/internal/app/ds"
)

func TestParseTokenRejectsInvalidClaimsBeforeRedis(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "test-signing-key"
	h := &Handler{Config: &config.Config{JWT: config.JWTConfig{Secret: secret}}}
	for _, name := range []string{"missing expiration", "expired", "wrong issuer", "zero user", "future issue time"} {
		t.Run(name, func(t *testing.T) {
			now := time.Now()
			claims := ds.JWTClaims{UserID: 1, RegisteredClaims: jwt.RegisteredClaims{
				Issuer: "utilities-backend", ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
			}}
			switch name {
			case "missing expiration":
				claims.ExpiresAt = nil
			case "expired":
				claims.ExpiresAt = jwt.NewNumericDate(now.Add(-time.Hour))
			case "wrong issuer":
				claims.Issuer = "another-service"
			case "zero user":
				claims.UserID = 0
			case "future issue time":
				claims.IssuedAt = jwt.NewNumericDate(now.Add(time.Hour))
			}
			token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
			if err != nil {
				t.Fatal(err)
			}
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/api/utilities/draft", nil)
			ctx.Request.Header.Set("Authorization", "Bearer "+token)
			if _, _, err := h.parseToken(ctx); !errors.Is(err, errUnauthorized) {
				t.Fatalf("expected unauthorized, got %v", err)
			}
		})
	}
}

func TestGetUtilitiesRejectsNonFiniteTariff(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tariff := range []string{"NaN", "Inf", "-Inf", "abc"} {
		t.Run(tariff, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Request = httptest.NewRequest("GET", "/api/utilities?tariff="+tariff, nil)
			(&Handler{}).GetUtilities(ctx)
			if recorder.Code != 400 {
				t.Fatalf("expected 400, got %d", recorder.Code)
			}
		})
	}
}
