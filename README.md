# template-srv-go

GitHub Template — backend Go (`net/http`, stdlib apenas).

## Stack

- **Go 1.27.1** — biblioteca padrão, sem dependências externas
- Roteamento com padrões de método do `net/http` (`GET /health`)
- Log estruturado com `log/slog`
- Encerramento gracioso em `SIGTERM` (o rollout do K8s drena as conexões)
- Imagem final `distroless/static:nonroot`, binário estático

## Requisitos

- Go 1.27.1+
- Docker (opcional, para a imagem)

## Rodar localmente

```bash
go run ./cmd/server
# GET http://localhost:8080/health → {"status":"ok"}
```

A porta vem de `PORT` (padrão `8080`, o mesmo `deploy.port` do `forge.yaml`).

## Testes e qualidade

```bash
go build ./...
go vet ./...
go test ./... -race -cover
golangci-lint run ./...      # v2.13.2, config em .golangci.yml
```

## Docker

```bash
docker build -t template-srv-go .
docker run --rm -p 8080:8080 template-srv-go
docker inspect --format '{{.State.Health.Status}}' <container>   # HEALTHCHECK
```

O `HEALTHCHECK` usa o binário `cmd/healthcheck`, que consulta `/health` no próprio
container — a imagem distroless não tem shell nem `curl`.

## Estrutura

```
cmd/server/            # entrypoint HTTP
cmd/healthcheck/       # probe do HEALTHCHECK (imagem sem shell)
internal/server/       # mux, handlers e configuração de porta
forge.yaml             # metadados do Forge (kind: srv, port 8080, healthPath /health)
.golangci.yml          # configuração do lint
```

## CI

- **lint** — `golangci-lint`
- **test** — `go build`, `go vet`, `go test ./... -race -cover`
- **docker** — constrói a imagem em todo PR; publica no GHCR só na branch default
