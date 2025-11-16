FROM golang:1.25.4-alpine AS build

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata bash build-base && update-ca-certificates

RUN ln -snf /usr/share/zoneinfo/Europe/Moscow /etc/localtime
RUN echo "Europe/Moscow" > /etc/timezone
ENV TZ=Europe/Moscow

RUN go install github.com/air-verse/air@v1.62.0
RUN go install github.com/go-delve/delve/cmd/dlv@latest


# Устанавливаем GOPATH явно (по умолчанию в Go это /go, но явное указание повышает читаемость)
ENV GOPATH=/go

# Обновляем PATH, добавляя GOPATH/bin и стандартный Go bin
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

# Устанавливаем golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Проверяем установку golangci-lint
RUN golangci-lint --version

COPY . .
RUN go mod download

EXPOSE 40000

CMD ["air", "-c", ".air.toml"]