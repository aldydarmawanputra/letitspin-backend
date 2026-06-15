package dto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type StandardResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

type PaginatedResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    interface{}     `json:"data"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalRows  int `json:"total_rows"`
	TotalPages int `json:"total_pages"`
}

type ResponseHelper struct{}

func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

func (h *ResponseHelper) Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func (h *ResponseHelper) SuccessOK(c *gin.Context, message string, data interface{}) {
	h.Success(c, http.StatusOK, message, data)
}

func (h *ResponseHelper) SuccessCreated(c *gin.Context, message string, data interface{}) {
	h.Success(c, http.StatusCreated, message, data)
}

func (h *ResponseHelper) SuccessNoContent(c *gin.Context) {
	c.JSON(http.StatusNoContent, nil)
}

func (h *ResponseHelper) Error(c *gin.Context, statusCode int, message string, err error) {
	response := StandardResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = &ErrorDetail{
			Details: err.Error(),
		}
	}

	c.JSON(statusCode, response)
}

func (h *ResponseHelper) ErrorBadRequest(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusBadRequest, message, err)
}

func (h *ResponseHelper) ErrorUnauthorized(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusUnauthorized, message, err)
}

func (h *ResponseHelper) ErrorForbidden(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusForbidden, message, err)
}

func (h *ResponseHelper) ErrorNotFound(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusNotFound, message, err)
}

func (h *ResponseHelper) ErrorConflict(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusConflict, message, err)
}

func (h *ResponseHelper) ErrorInternal(c *gin.Context, message string, err error) {
	h.Error(c, http.StatusInternalServerError, message, err)
}

func (h *ResponseHelper) ValidationError(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"success": false,
		"message": "validation failed",
		"errors":  errors,
	})
}
