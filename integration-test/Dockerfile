# Step 1: Modules caching
FROM golang:1.26.6-alpine3.23@sha256:e57c41c1d5864341031181b0db34b9a537bb5773eb6428e4e5bdaea0f9135406 AS modules
ENV GOTOOLCHAIN=local
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.26.6-alpine3.23@sha256:e57c41c1d5864341031181b0db34b9a537bb5773eb6428e4e5bdaea0f9135406
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

CMD ["go", "test", "-count=1", "-v", "./integration-test/..."]
