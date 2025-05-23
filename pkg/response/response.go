package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents the unified API response structure.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// NewResponse creates a new Response instance.
func NewResponse(code int, message string, data interface{}) *Response {
	return &Response{Code: code, Message: message, Data: data}
}

// Success sends a successful response.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, NewResponse(http.StatusOK, "Success", data))
}

// Error sends an error response.
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, NewResponse(code, message, nil))
}

// BadRequest sends a 400 Bad Request error response.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized sends a 401 Unauthorized error response.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden sends a 403 Forbidden error response.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound sends a 404 Not Found error response.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError sends a 500 Internal Server Error response.
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Conflict sends a 409 Conflict error response.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}
