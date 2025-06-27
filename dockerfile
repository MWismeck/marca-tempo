# Etapa 1: Build
FROM golang:1.23.3-alpine AS builder

# Habilita CGO para go-sqlite3
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Instala dependências de compilação C
RUN apk add --no-cache gcc g++ musl-dev sqlite-dev

# Prepara o diretório de build
WORKDIR /app
COPY . .

# Compila o binário principal
RUN go build -o app .

# Etapa 2: Execução (imagem mínima)
FROM alpine:latest

RUN apk --no-cache add sqlite

WORKDIR /app
COPY --from=builder /app/app .
COPY --from=builder /app/public ./public

# Expõe a porta usada pelo Echo (ex: 8080)
EXPOSE 8080

# Comando de execução
CMD ["./app"]
