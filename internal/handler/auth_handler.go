package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	jwtpkg "let-it-spin/internal/auth/jwt"
	"let-it-spin/internal/auth/repository"
	"let-it-spin/internal/dto"
)

type AuthHandler struct {
	authRepo *repository.AuthRepository
	jwt      *jwtpkg.JWTService
	resp     *dto.ResponseHelper
}

func NewAuthHandler(authRepo *repository.AuthRepository, jwt *jwtpkg.JWTService) *AuthHandler {
	return &AuthHandler{
		authRepo: authRepo,
		jwt:      jwt,
		resp:     dto.NewResponseHelper(),
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.ErrorBadRequest(c, "invalid request payload", err)
		return
	}

	user, err := h.authRepo.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		h.resp.ErrorUnauthorized(c, "invalid credentials", nil)
		return
	}

	hash, err := h.authRepo.GetUserCredentialByUserID(c.Request.Context(), user.ID)
	if err != nil {
		h.resp.ErrorUnauthorized(c, "invalid credentials", nil)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		h.resp.ErrorUnauthorized(c, "invalid credentials", nil)
		return
	}

	accessToken, err := h.jwt.GenerateAccessToken(user.ID.String(), user.Email, 15*time.Minute)
	if err != nil {
		h.resp.ErrorInternal(c, "failed to generate access token", err)
		return
	}

	refreshToken := uuid.NewString()

	if err := h.authRepo.StoreRefreshToken(c.Request.Context(), user.ID, refreshToken, time.Now().Add(7*24*time.Hour)); err != nil {
		h.resp.ErrorInternal(c, "failed to store refresh token", err)
		return
	}

	h.resp.SuccessOK(c, "login successful", dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.ErrorBadRequest(c, "invalid request payload", err)
		return
	}

	if req.RefreshToken == "" {
		h.resp.ErrorBadRequest(c, "refresh token required", nil)
		return
	}

	userID, err := h.authRepo.ValidateRefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		h.resp.ErrorUnauthorized(c, "invalid refresh token", err)
		return
	}

	user, err := h.authRepo.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		h.resp.ErrorUnauthorized(c, "user not found", err)
		return
	}

	newAccessToken, err := h.jwt.GenerateAccessToken(user.ID.String(), user.Email, 15*time.Minute)
	if err != nil {
		h.resp.ErrorInternal(c, "failed to generate access token", err)
		return
	}

	newRefreshToken := uuid.NewString()

	_ = h.authRepo.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)
	_ = h.authRepo.StoreRefreshToken(c.Request.Context(), user.ID, newRefreshToken, time.Now().Add(7*24*time.Hour))

	h.resp.SuccessOK(c, "token refreshed successfully", dto.LoginResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    900,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	_ = c.ShouldBindJSON(&req)

	if req.RefreshToken != "" {
		_ = h.authRepo.RevokeRefreshToken(c.Request.Context(), req.RefreshToken)
	}

	h.resp.SuccessOK(c, "logged out successfully", nil)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.resp.ErrorUnauthorized(c, "unauthorized", nil)
		return
	}

	id, err := uuid.Parse(userID.(string))
	if err != nil {
		h.resp.ErrorInternal(c, "invalid user id", err)
		return
	}

	user, err := h.authRepo.GetUserByID(c.Request.Context(), id)
	if err != nil {
		h.resp.ErrorNotFound(c, "user not found", err)
		return
	}

	h.resp.SuccessOK(c, "user retrieved successfully", gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"is_active":      user.IsActive,
		"email_verified": user.EmailVerified,
	})
}
