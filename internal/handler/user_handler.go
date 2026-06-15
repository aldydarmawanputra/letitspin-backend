package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"let-it-spin/internal/dto"
	"let-it-spin/internal/model"
	"let-it-spin/internal/repository"
)

type UserHandler struct {
	repo *repository.UserRepository
	resp *dto.ResponseHelper
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		repo: repo,
		resp: dto.NewResponseHelper(),
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.ErrorBadRequest(c, "invalid request payload", err)
		return
	}

	if req.Email == "" || req.Password == "" {
		h.resp.ErrorBadRequest(c, "email and password are required", nil)
		return
	}

	exists, err := h.repo.EmailExists(c.Request.Context(), req.Email)
	if err != nil {
		h.resp.ErrorInternal(c, "failed to check email", err)
		return
	}

	if exists {
		h.resp.ErrorConflict(c, "email already exists", nil)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.resp.ErrorInternal(c, "failed to hash password", err)
		return
	}

	tx, err := h.repo.BeginTx(c.Request.Context())
	if err != nil {
		h.resp.ErrorInternal(c, "failed to start transaction", err)
		return
	}

	defer func() {
		_ = h.repo.RollbackTx(tx)
	}()

	now := time.Now()

	user := &model.User{
		ID:            uuid.New(),
		Email:         req.Email,
		IsActive:      true,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := h.repo.CreateUser(c.Request.Context(), tx, user); err != nil {
		h.resp.ErrorInternal(c, "failed to create user", err)
		return
	}

	if err := h.repo.CreateCredential(c.Request.Context(), tx, user.ID, string(hash)); err != nil {
		h.resp.ErrorInternal(c, "failed to create credential", err)
		return
	}

	if err := h.repo.CommitTx(tx); err != nil {
		h.resp.ErrorInternal(c, "failed to commit transaction", err)
		return
	}

	h.resp.SuccessCreated(c, "user created successfully", gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"is_active":      user.IsActive,
		"email_verified": user.EmailVerified,
		"created_at":     user.CreatedAt,
	})
}

func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.resp.ErrorBadRequest(c, "invalid user id", err)
		return
	}

	user, err := h.repo.GetUserByID(c.Request.Context(), id)
	if err != nil {
		h.resp.ErrorNotFound(c, "user not found", err)
		return
	}

	h.resp.SuccessOK(c, "user retrieved successfully", user)
}

func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")

	if email == "" {
		h.resp.ErrorBadRequest(c, "email is required", nil)
		return
	}

	user, err := h.repo.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		h.resp.ErrorNotFound(c, "user not found", err)
		return
	}

	h.resp.SuccessOK(c, "user retrieved successfully", user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		IsActive      bool   `json:"is_active"`
		EmailVerified bool   `json:"email_verified"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.ErrorBadRequest(c, "invalid request payload", err)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		h.resp.ErrorBadRequest(c, "invalid user id", err)
		return
	}

	user := &model.User{
		ID:            id,
		Email:         req.Email,
		IsActive:      req.IsActive,
		EmailVerified: req.EmailVerified,
		UpdatedAt:     time.Now(),
	}

	if err := h.repo.UpdateUser(c.Request.Context(), user); err != nil {
		h.resp.ErrorInternal(c, "failed to update user", err)
		return
	}

	h.resp.SuccessOK(c, "user updated successfully", user)
}

func (h *UserHandler) PatchUser(c *gin.Context) {
	var req struct {
		Email         *string `json:"email"`
		IsActive      *bool   `json:"is_active"`
		EmailVerified *bool   `json:"email_verified"`
	}

	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.resp.ErrorBadRequest(c, "invalid user id", err)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.ErrorBadRequest(c, "invalid request payload", err)
		return
	}

	if req.Email == nil && req.IsActive == nil && req.EmailVerified == nil {
		h.resp.ErrorBadRequest(c, "no fields to update", nil)
		return
	}

	if err := h.repo.PatchUser(c.Request.Context(), id, req.Email, req.IsActive, req.EmailVerified); err != nil {
		h.resp.ErrorInternal(c, "failed to patch user", err)
		return
	}

	h.resp.SuccessOK(c, "user patched successfully", nil)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.resp.ErrorBadRequest(c, "invalid user id", err)
		return
	}

	if err := h.repo.DeleteUser(c.Request.Context(), id); err != nil {
		h.resp.ErrorInternal(c, "failed to delete user", err)
		return
	}

	h.resp.SuccessNoContent(c)
}
