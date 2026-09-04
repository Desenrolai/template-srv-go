// Package server monta o roteador HTTP e os handlers do template.
package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

// HealthPath espelha deploy.healthPath do forge.yaml. O HEALTHCHECK da imagem e as
// probes do cluster leem o mesmo caminho — mudou aqui, mude no forge.yaml.
const HealthPath = "/health"

// DefaultPort espelha deploy.port do forge.yaml.
const DefaultPort = "8080"

type healthResponse struct {
	Status string `json:"status"`
}

// NewMux devolve o roteador com as rotas do template registradas.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+HealthPath, handleHealth)

	return mux
}

// Port devolve a porta de escuta: PORT quando definida, DefaultPort caso contrário.
func Port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}

	return DefaultPort
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		slog.Error("falha ao serializar resposta", "erro", err)
		http.Error(w, `{"status":"error"}`, http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Escrita só falha quando o cliente já desconectou: registrar é tudo que resta.
	if _, err := w.Write(body); err != nil {
		slog.Warn("falha ao escrever resposta", "erro", err)
	}
}
