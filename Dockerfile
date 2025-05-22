ARG GOLANG_VERSION=1.24
FROM golang:${GOLANG_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o kinescope-json-parser

FROM scratch

COPY --from=builder /app/kinescope-json-parser /kinescope-json-parser
COPY --from=builder /app/static /static

EXPOSE 8080

ENTRYPOINT ["/kinescope-json-parser"]