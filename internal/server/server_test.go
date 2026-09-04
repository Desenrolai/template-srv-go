package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/desenrolai/template-srv-go/internal/server"
)

func TestHealthRespondeOKEmJSON(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, server.HealthPath, nil)
	rec := httptest.NewRecorder()

	// Act
	server.NewMux().ServeHTTP(rec, req)

	// Assert
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, esperado %q", got, "application/json")
	}

	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("corpo não é JSON válido: %v (corpo: %q)", err, rec.Body.String())
	}

	if body["status"] != "ok" {
		t.Errorf(`corpo = %v, esperado {"status":"ok"}`, body)
	}
}

func TestHealthRecusaMetodoDiferenteDeGET(t *testing.T) {
	// O roteador usa o padrão com método do net/http: POST não pode cair no handler de saúde.
	req := httptest.NewRequest(http.MethodPost, server.HealthPath, nil)
	rec := httptest.NewRecorder()

	server.NewMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestRotaDesconhecidaRetorna404(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nao-existe", nil)
	rec := httptest.NewRecorder()

	server.NewMux().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, esperado %d", rec.Code, http.StatusNotFound)
	}
}

func TestPortUsaDefaultQuandoEnvVazia(t *testing.T) {
	t.Setenv("PORT", "")

	if got := server.Port(); got != server.DefaultPort {
		t.Errorf("Port() = %q, esperado %q", got, server.DefaultPort)
	}
}

func TestPortRespeitaEnv(t *testing.T) {
	t.Setenv("PORT", "9090")

	if got := server.Port(); got != "9090" {
		t.Errorf("Port() = %q, esperado %q", got, "9090")
	}
}
