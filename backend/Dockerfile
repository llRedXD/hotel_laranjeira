FROM golang:1.24-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /hotel-api ./cmd/api

FROM alpine:3.21

RUN adduser -D -H appuser
USER appuser

COPY --from=build /hotel-api /hotel-api

EXPOSE 8080

CMD ["/hotel-api"]
