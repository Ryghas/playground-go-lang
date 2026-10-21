package response

import (
	"net/http"

	apperrors "pokemon-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Body — format response standar untuk semua endpoint.
type Body struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Meta    any        `json:"meta,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"pokemon tidak ditemukan"`
}

// ErrorResponse — bentuk response error; dipakai untuk dokumentasi Swagger.
type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Error   ErrorBody `json:"error"`
}

func OK(c *gin.Context, status int, data any) {
	c.JSON(status, Body{Success: true, Data: data})
}

func OKWithMeta(c *gin.Context, status int, data, meta any) {
	c.JSON(status, Body{Success: true, Data: data, Meta: meta})
}

func BadRequest(c *gin.Context, msg string) {
	abort(c, http.StatusBadRequest, apperrors.CodeBadRequest, msg)
}

// Error memetakan AppError ke HTTP status. Error yang tidak dikenal
// dianggap 500 dan detailnya hanya masuk log, tidak dikirim ke client.
func Error(c *gin.Context, err error) {
	_ = c.Error(err) // dicatat oleh middleware logger

	appErr, ok := apperrors.As(err)
	if !ok {
		abort(c, http.StatusInternalServerError, apperrors.CodeInternal, "terjadi kesalahan pada server")
		return
	}

	status := http.StatusInternalServerError
	switch appErr.Code {
	case apperrors.CodeNotFound:
		status = http.StatusNotFound
	case apperrors.CodeConflict:
		status = http.StatusConflict
	case apperrors.CodeBadRequest:
		status = http.StatusBadRequest
	}
	abort(c, status, appErr.Code, appErr.Message)
}

func abort(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, Body{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: msg},
	})
}
