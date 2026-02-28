package services

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func BenchmarkHandleTestWebhook(b *testing.B) {
	gin.SetMode(gin.ReleaseMode)
	service := NewWebhookService(nil, nil, "whsec_test_secret")

	// Pre-allocate HTTP request
	req, _ := http.NewRequest("POST", "/test-webhook?company_id=comp_123", nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		service.HandleTestWebhook(c)
	}
}
