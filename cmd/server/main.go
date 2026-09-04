// Command server sobe o HTTP hello-world do template.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/desenrolai/template-srv-go/internal/server"
)

const (
	// readHeaderTimeout corta Slowloris antes de o servidor alocar o corpo da requisição.
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	// shutdownTimeout precisa caber na terminationGracePeriodSeconds do pod (30s por padrão).
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("servidor encerrado com erro", "erro", err)
		os.Exit(1)
	}
}

func run() error {
	addr := ":" + server.Port()

	srv := &http.Server{
		Addr:              addr,
		Handler:           server.NewMux(),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	listenErr := make(chan error, 1)

	go func() {
		slog.Info("servidor ouvindo", "addr", addr, "health", server.HealthPath)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenErr <- err

			return
		}

		listenErr <- nil
	}()

	select {
	case err := <-listenErr:
		return err
	case <-ctx.Done():
		slog.Info("sinal recebido, drenando conexões")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
