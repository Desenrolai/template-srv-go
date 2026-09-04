# template-srv-go

GitHub Template — backend Go (`net/http`, stdlib apenas).

## Stack

- **Go 1.27.1** — biblioteca padrão, sem dependências externas
- Roteamento com padrões de método do `net/http` (`GET /health`)
- Log estruturado com `log/slog`
- Encerramento gracioso em `SIGTERM` (o rollout do K8s drena as conexões)
- Imagem final `distroless/static:nonroot`, binário estático
- Lint com `gosec` — a régua de segurança daqui se replica em todo serviço gerado

## Ao gerar um projeto a partir deste template

O Forge scaffolda com `repos.createUsingTemplate` do GitHub — **cópia literal dos
arquivos**, sem substituição de placeholder. Tudo que carrega o nome do template chega
no repo novo com o nome do template.

Em Go isso não é cosmético: o `module` é o prefixo de **todos os imports internos**.
Troque o module path antes do primeiro commit:

```bash
NOVO=github.com/desenrolai/<seu-repo>
ANTIGO=$(go list -m)

go mod edit -module "$NOVO"
grep -rl "$ANTIGO" --include='*.go' . | xargs perl -pi -e "s{\Q$ANTIGO\E}{$NOVO}g"

go build ./... && go vet ./... && go test ./... -race
```

São **4** pontos. Se for fazer à mão, são todos estes — nenhum a menos:

| # | Arquivo | O que é |
|---|---|---|
| 1 | `go.mod` | diretiva `module` |
| 2 | `cmd/server/main.go` | import de `internal/server` |
| 3 | `cmd/healthcheck/main.go` | import de `internal/server` |
| 4 | `internal/server/server_test.go` | import de `internal/server` **no teste** |

> ⚠️ **`go build ./...` não valida este rename.** Medido: trocando o `go.mod` e os dois
> `cmd/`, mas esquecendo o import do teste, `go build ./...` sai **0** — e só `go vet`
> (exit 1) e `go test` (`[setup failed]`) acusam. Valide com `go test ./...`, não com o
> build.

Também carregam o nome do template, sem quebrar nada: o título deste README e as tags
`docker build -t` dos exemplos abaixo.

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
golangci-lint run ./...      # v2.13.2, config em .golangci.yml (inclui gosec)
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

## Repo privado: o CI nasce morto sem estas variáveis

Este template é um repositório **público**, onde o GitHub Actions em runner hospedado é
gratuito e ilimitado — por isso o CI daqui está verde. O repo que você gera a partir dele
é **privado**, e lá a cota de minutos hospedados está esgotada. Antes do primeiro push,
defina duas *repository variables* (Settings → Secrets and variables → Actions →
Variables):

| Variável | Valor |
|---|---|
| `CI_RUNNER` | `["self-hosted","desenrolai"]` |
| `CI_RUNNER_DOCKER` | `["self-hosted","docker-builder"]` |

O `runs-on` lê essas variáveis e cai em `ubuntu-latest` quando elas não existem — é o que
mantém o CI deste template rodando em runner hospedado. O `fromJSON` não é enfeite: um
runner self-hosted da casa é um **conjunto de labels**, e `self-hosted,desenrolai` como
string simples viraria um único label com vírgula no nome, que não casa com runner nenhum.

> ⚠️ **O sintoma de não fazer isto não parece falta de runner.** O job morre em ~2
> segundos com **`steps: 0`** — nenhum step aparece, nenhum log de erro, nada que aponte
> para billing. Parece YAML quebrado, e a pessoa perde meia hora procurando erro de
> sintaxe. É cota.
>
> O que separa esse caso de um job legitimamente `skipped` — que **também** reporta zero
> steps — é a conclusão: aqui ela é `failure`, não `skipped`.

## CI

- **lint** — `golangci-lint`
- **test** — `go build`, `go vet`, `go test ./... -race -cover`
- **docker-build** — fora da branch padrão, constrói a imagem e descarta: sem login no
  GHCR e sem `packages: write`
- **docker** — na branch padrão, constrói e publica no GHCR
