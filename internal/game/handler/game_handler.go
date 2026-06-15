package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"let-it-spin/internal/dto"
	"let-it-spin/internal/game/service"
)

type GameHandler struct {
	gameService *service.GameService
	resp        *dto.ResponseHelper
}

func NewGameHandler(gameService *service.GameService) *GameHandler {
	return &GameHandler{
		gameService: gameService,
		resp:        dto.NewResponseHelper(),
	}
}

func (h *GameHandler) GetGameConfig(c *gin.Context) {
	gameCode := c.Param("code")

	gameType, err := h.gameService.GetGameConfig(c.Request.Context(), gameCode)
	if err != nil {
		h.resp.ErrorNotFound(c, "game not found", err)
		return
	}

	h.resp.SuccessOK(c, "game config retrieved successfully", gameType)
}

func (h *GameHandler) Play(c *gin.Context) {
	gameCode := c.Param("code")

	var req struct {
		BetAmount int64                  `json:"bet_amount" binding:"required,min=1"`
		Options   map[string]interface{} `json:"options,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.resp.ErrorBadRequest(c, "invalid request payload", err)
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		h.resp.ErrorUnauthorized(c, "unauthorized", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.resp.ErrorInternal(c, "invalid user id", err)
		return
	}

	result, err := h.gameService.Play(c.Request.Context(), userID, gameCode, req.BetAmount, req.Options)
	if err != nil {
		if err.Error() == "insufficient balance" {
			h.resp.ErrorBadRequest(c, "insufficient balance", err)
			return
		}
		h.resp.ErrorInternal(c, "failed to play game", err)
		return
	}

	h.resp.SuccessOK(c, "game played successfully", result)
}

func (h *GameHandler) GetHistory(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		h.resp.ErrorUnauthorized(c, "unauthorized", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		h.resp.ErrorInternal(c, "invalid user id", err)
		return
	}

	gameCode := c.Query("game_code")

	page := 1
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	var codePtr *string
	if gameCode != "" {
		codePtr = &gameCode
	}

	sessions, total, err := h.gameService.GetHistory(c.Request.Context(), userID, codePtr, page, limit)
	if err != nil {
		h.resp.ErrorInternal(c, "failed to get history", err)
		return
	}

	h.resp.SuccessOK(c, "history retrieved successfully", gin.H{
		"sessions": sessions,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}
