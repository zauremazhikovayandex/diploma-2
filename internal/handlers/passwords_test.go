package handlers

import (
	"bytes"
	"diploma-2/pkg/config"
	"diploma-2/pkg/db"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

// helper для создания API без реальной БД (db *SqlConnection можно оставить nil, если мы не доходим до репозитория).
func newTestAPI() *API {
	os.Setenv("JWT_SECRET_KEY", "secret_test")
	cfg := config.InitConfig()
	return &API{
		cfg: cfg,
		db:  &db.SqlConnection{}, // заглушка, в тестах ниже мы не будем до неё доходить
	}
}

// helper для gin-контекста
func newTestContext(method, target string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(method, target, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	return c, w
}

func TestPostPassword_UnauthorizedWhenNoLoginInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestAPI()

	c, w := newTestContext("POST", "/api/v1/passwords", []byte(`{}`))

	// В контекст пользователя логин не кладём — имитируем отсутствие auth.
	api.PostPassword(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPostPassword_BadRequestOnInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestAPI()

	c, w := newTestContext("POST", "/api/v1/passwords", []byte(`{invalid json`))

	// Подделываем логин пользователя в gin.Context (как будто прошли auth middleware).
	c.Set("user_login", "testuser")

	api.PostPassword(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestGetPasswordHandler_UnauthorizedWhenNoLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	api := newTestAPI()

	c, w := newTestContext("GET", "/api/v1/passwords/some-id", nil)

	api.GetPasswordHandler(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
