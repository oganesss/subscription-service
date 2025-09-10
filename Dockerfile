FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd

FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /bin/app /app
COPY ./configs /configs
COPY ./migrations /migrations
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app"]



