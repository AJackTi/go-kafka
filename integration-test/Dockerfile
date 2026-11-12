# Step 1: Modules caching
FROM golang:1.26.11-alpine3.23@sha256:db99415e2625ce0ee03eed09edf7f32ae8c58dc399f28cc602a87fe7238119e4 AS modules
ENV GOTOOLCHAIN=local
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.26.11-alpine3.23@sha256:db99415e2625ce0ee03eed09edf7f32ae8c58dc399f28cc602a87fe7238119e4
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

CMD ["go", "test", "-count=1", "-v", "./integration-test/..."]
