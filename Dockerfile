# syntax=docker/dockerfile:1

# --- build ---
FROM golang:1.27-alpine AS builder

WORKDIR /src

# go.su[m] casa de forma opcional: o template ainda não tem dependências externas.
COPY go.mod go.su[m] ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

# CGO desligado: a imagem final é distroless static, sem libc.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck

# --- runtime ---
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app

COPY --from=builder /out/server /app/server
COPY --from=builder /out/healthcheck /app/healthcheck

USER nonroot:nonroot

EXPOSE 8080

# A imagem distroless não tem shell nem curl: o probe é o binário healthcheck, que
# consulta o mesmo caminho de deploy.healthPath no forge.yaml.
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD ["/app/healthcheck"]

ENTRYPOINT ["/app/server"]
