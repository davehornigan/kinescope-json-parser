ARG GOLANG_VERSION=1.24
FROM golang:${GOLANG_VERSION}-alpine AS build

WORKDIR /app

COPY . .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /kinescope-json-parser

FROM scratch

COPY --from=builder /kinescope-json-parser /kinescope-json-parser

EXPOSE 8080

ENTRYPOINT ["/kinescope-json-parser"]