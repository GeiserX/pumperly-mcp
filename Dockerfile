FROM golang:1.26 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/pumperly-mcp ./cmd/server

FROM alpine:3.23
COPY --from=builder /out/pumperly-mcp /usr/local/bin/pumperly-mcp
EXPOSE 8080
ENV TRANSPORT=stdio
ENV PUMPERLY_URL=https://pumperly.com
ENTRYPOINT ["/usr/local/bin/pumperly-mcp"]
