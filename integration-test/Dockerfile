# Step 1: Modules caching
FROM golang:1.26.10-alpine3.23@sha256:34414731924cf2263160f0021f5cf6e1807382d85248f7cb419d4187b96d5e43 AS modules
ENV GOTOOLCHAIN=local
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.26.10-alpine3.23@sha256:34414731924cf2263160f0021f5cf6e1807382d85248f7cb419d4187b96d5e43
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

CMD ["go", "test", "-count=1", "-v", "./integration-test/..."]
