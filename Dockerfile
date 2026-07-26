# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/app     ./cmd/app
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/migrate ./cmd/migrate

# ── Devtools image ────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS devtools

RUN apk --no-cache add git
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

WORKDIR /workspace
ENV PATH="/go/bin:${PATH}"

CMD ["sh"]

# ── App image ─────────────────────────────────────────────────────────────────
FROM alpine:3.19 AS app
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /out/app /app
COPY --from=builder /build/internal/migrations /internal/migrations
EXPOSE 8080
CMD ["/app"]

# ── Migrate image ─────────────────────────────────────────────────────────────
FROM alpine:3.19 AS migrate
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /out/migrate /migrate
COPY --from=builder /build/internal/migrations /internal/migrations
WORKDIR /
ENTRYPOINT ["/migrate"]
