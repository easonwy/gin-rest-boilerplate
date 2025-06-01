package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingMiddleware(t *testing.T) {
	// Create a new Gin router
	r := gin.New()

	// Create a Zap test logger and an observer to capture logs
	core, observedLogs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	// Apply the logging middleware
	r.Use(LoggingMiddleware(logger))

	// Define a test route
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// Create a test request
	req, _ := http.NewRequest(http.MethodGet, "/test?query=test", nil)
	req.Header.Set("User-Agent", "TestAgent")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	r.ServeHTTP(w, req)

	// Assert the response status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert that one log entry was recorded at Info level
	assert.Equal(t, 1, observedLogs.Len())
	logEntry := observedLogs.All()[0]
	assert.Equal(t, zapcore.InfoLevel, logEntry.Level)
	assert.Equal(t, "Request", logEntry.Message)

	// Assert log fields
	fields := logEntry.ContextMap()
	assert.Equal(t, http.MethodGet, fields["method"])
	assert.Equal(t, "/test", fields["path"])
	assert.Equal(t, "query=test", fields["query"])
	assert.Equal(t, int64(http.StatusOK), fields["status"])
	// assert.Equal(t, "127.0.0.1", fields["ip"]) // ClientIP might vary in test environment
	assert.Equal(t, "TestAgent", fields["user-agent"])
	assert.Contains(t, fields, "duration")

	// Test with a different method and path
	r.POST("/another", func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	req, _ = http.NewRequest(http.MethodPost, "/another", bytes.NewBufferString("body"))
	req.Header.Set("User-Agent", "AnotherAgent")

	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Assert that another log entry was recorded
	assert.Equal(t, 2, observedLogs.Len())
	logEntry = observedLogs.All()[1]
	assert.Equal(t, zapcore.InfoLevel, logEntry.Level)
	assert.Equal(t, "Request", logEntry.Message)

	fields = logEntry.ContextMap()
	assert.Equal(t, http.MethodPost, fields["method"])
	assert.Equal(t, "/another", fields["path"])
	assert.Equal(t, "", fields["query"])
	assert.Equal(t, int64(http.StatusCreated), fields["status"])
	assert.Equal(t, "AnotherAgent", fields["user-agent"])
	assert.Contains(t, fields, "duration")
}
