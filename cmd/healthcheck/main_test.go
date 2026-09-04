package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// serveNoLoopback sobe um servidor de teste na porta que o probe vai consultar,
// para exercitar o caminho real (127.0.0.1:PORT + HealthPath) e não um mock.
func serveNoLoopback(t *testing.T, handler http.Handler) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	parsed, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("URL do servidor de teste inválida: %v", err)
	}

	t.Setenv("PORT", parsed.Port())
}

func TestProbeAceitaSaudeOK(t *testing.T) {
	serveNoLoopback(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	if err := probe(); err != nil {
		t.Errorf("probe() = %v, esperado nil", err)
	}
}

func TestProbeFalhaComStatusDeErro(t *testing.T) {
	serveNoLoopback(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))

	if err := probe(); err == nil {
		t.Error("probe() = nil, esperado erro para status 503")
	}
}

func TestProbeFalhaSemServidor(t *testing.T) {
	// Porta reservada pela IANA para "não use": ninguém escuta aqui.
	t.Setenv("PORT", "1")

	if err := probe(); err == nil {
		t.Error("probe() = nil, esperado erro de conexão")
	}
}
