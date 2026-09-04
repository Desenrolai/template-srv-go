// Command healthcheck consulta o endpoint de saúde no próprio container e traduz o
// resultado em código de saída. Existe porque a imagem distroless não tem shell nem
// curl para o HEALTHCHECK do Dockerfile.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/desenrolai/template-srv-go/internal/server"
)

const probeTimeout = 3 * time.Second

var errUnhealthy = errors.New("endpoint de saúde respondeu status diferente de 200")

func main() {
	if err := probe(); err != nil {
		os.Exit(1)
	}
}

func probe() error {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	url := "http://127.0.0.1:" + server.Port() + server.HealthPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return errUnhealthy
	}

	return nil
}
