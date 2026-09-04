package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// portaLivre reserva e devolve uma porta efêmera. Há uma janela entre fechar o
// listener e o servidor subir, aceitável dentro de um teste.
func portaLivre(t *testing.T) string {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("não consegui reservar porta: %v", err)
	}

	porta := strconv.Itoa(l.Addr().(*net.TCPAddr).Port)

	if err := l.Close(); err != nil {
		t.Fatalf("não consegui liberar a porta reservada: %v", err)
	}

	return porta
}

func esperarSaude(t *testing.T, url string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	for ctx.Err() == nil {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			t.Fatalf("requisição inválida: %v", err)
		}

		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatal("servidor não respondeu /health dentro de 5s")
}

// Prova que o SIGTERM realmente drena o servidor: sem o signal.NotifyContext em run(),
// o processo morreria e run() nunca devolveria nil.
func TestRunEncerraGraciosamenteComSIGTERM(t *testing.T) {
	porta := portaLivre(t)
	t.Setenv("PORT", porta)

	fim := make(chan error, 1)
	go func() { fim <- run() }()

	esperarSaude(t, "http://127.0.0.1:"+porta+"/health")

	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("não achei o próprio processo: %v", err)
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("não consegui enviar SIGTERM: %v", err)
	}

	select {
	case err := <-fim:
		if err != nil {
			t.Errorf("run() = %v, esperado nil após SIGTERM", err)
		}
	case <-time.After(shutdownTimeout + 5*time.Second):
		t.Fatal("run() não retornou após SIGTERM")
	}
}

// Porta ocupada tem de virar erro de retorno, não pânico nem travamento.
func TestRunDevolveErroQuandoPortaEstaOcupada(t *testing.T) {
	// Curinga, igual ao que run() usa: em BSD/macOS o SO_REUSEADDR deixaria
	// 0.0.0.0:P conviver com 127.0.0.1:P, e o bind não falharia.
	//
	// O G102 do gosec ("binds to all network interfaces") é justamente o ponto do
	// teste: ele precisa ocupar a mesma faixa que o servidor ocupa para o conflito
	// existir. Listener de teste, fechado no Cleanup, nunca em código de produção.
	l, err := net.Listen("tcp", ":0") //nolint:gosec // G102 é o comportamento sob teste
	if err != nil {
		t.Fatalf("não consegui ocupar porta: %v", err)
	}

	t.Cleanup(func() { _ = l.Close() })
	t.Setenv("PORT", strconv.Itoa(l.Addr().(*net.TCPAddr).Port))

	fim := make(chan error, 1)
	go func() { fim <- run() }()

	select {
	case err := <-fim:
		if err == nil {
			t.Error("run() = nil, esperado erro de bind em porta ocupada")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run() não retornou erro de bind")
	}
}
