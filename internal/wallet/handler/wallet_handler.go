package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"let-it-spin/internal/dto"
	"let-it-spin/internal/wallet/service"
)

type WalletHandler struct {
	walletService *service.WalletService
	resp          *dto.ResponseHelper
}

func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
		resp:          dto.NewResponseHelper(),
	}
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
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

	wallet, err := h.walletService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		h.resp.ErrorInternal(c, "failed to get balance", err)
		return
	}

	h.resp.SuccessOK(c, "balance retrieved successfully", gin.H{
		"balance":  wallet.Balance,
		"currency": wallet.Currency,
	})
}

func (h *WalletHandler) Deposit(c *gin.Context) {
	var req struct {
		Amount      int64   `json:"amount" binding:"required,min=1"`
		ReferenceID *string `json:"reference_id,omitempty"`
		Description *string `json:"description,omitempty"`
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

	wallet, transaction, err := h.walletService.Deposit(c.Request.Context(), service.DepositRequest{
		UserID:      userID,
		Amount:      req.Amount,
		ReferenceID: req.ReferenceID,
		Description: req.Description,
	})
	if err != nil {
		h.resp.ErrorInternal(c, "failed to deposit", err)
		return
	}

	h.resp.SuccessOK(c, "deposit successful", gin.H{
		"balance":        wallet.Balance,
		"currency":       wallet.Currency,
		"transaction_id": transaction.ID,
	})
}

func (h *WalletHandler) Withdraw(c *gin.Context) {
	var req struct {
		Amount      int64   `json:"amount" binding:"required,min=1"`
		ReferenceID *string `json:"reference_id,omitempty"`
		Description *string `json:"description,omitempty"`
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

	wallet, transaction, err := h.walletService.Withdraw(c.Request.Context(), service.WithdrawRequest{
		UserID:      userID,
		Amount:      req.Amount,
		ReferenceID: req.ReferenceID,
		Description: req.Description,
	})
	if err != nil {
		if err.Error() == "insufficient balance" {
			h.resp.ErrorBadRequest(c, "insufficient balance", err)
			return
		}
		h.resp.ErrorInternal(c, "failed to withdraw", err)
		return
	}

	h.resp.SuccessOK(c, "withdraw successful", gin.H{
		"balance":        wallet.Balance,
		"currency":       wallet.Currency,
		"transaction_id": transaction.ID,
	})
}

func (h *WalletHandler) GetTransactions(c *gin.Context) {
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

	transactions, total, err := h.walletService.GetTransactions(c.Request.Context(), userID, page, limit)
	if err != nil {
		h.resp.ErrorInternal(c, "failed to get transactions", err)
		return
	}

	h.resp.SuccessOK(c, "transactions retrieved successfully", gin.H{
		"transactions": transactions,
		"total":        total,
		"page":         page,
		"limit":        limit,
	})
}
