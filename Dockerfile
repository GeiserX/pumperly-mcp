# Base images are pinned by digest; Dependabot (docker) keeps them current.
FROM golang:1.27.4@sha256:08bcb99a38f85a265f9a93d1d3323e2440eec333380f3ea26ee0dec9264bb41c AS builder
ARG VERSION=dev
ARG COMMIT=none
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
      -ldflags "-s -w -X github.com/geiserx/pumperly-mcp/version.Version=${VERSION} -X github.com/geiserx/pumperly-mcp/version.Commit=${COMMIT}" \
      -o /out/pumperly-mcp ./cmd/server

FROM alpine:3.24.3@sha256:be2656061b536c1883c4034790e21d0a09c5aaaf51968b89368a3c0371ea2c9a
# The base image can lag the CA bundle; take the current one, which drops
# the compromised eMudhra (emSign) intermediates.
RUN apk upgrade --no-cache ca-certificates-bundle \
 && adduser -D -H -u 10001 app
LABEL io.modelcontextprotocol.server.name="io.github.GeiserX/pumperly-mcp"
COPY --from=builder /out/pumperly-mcp /usr/local/bin/pumperly-mcp
USER 10001:10001
# stdio by default: `docker run -i --rm drumsergio/pumperly-mcp:<tag>` is what
# MCP clients and the registry's OCI entry run. For HTTP set TRANSPORT=http;
# the 0.0.0.0 listen address below then requires MCP_AUTH_TOKEN.
ENV TRANSPORT=stdio
ENV LISTEN_ADDR=0.0.0.0:8080
ENV PUMPERLY_URL=https://pumperly.com
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pumperly-mcp"]
