FROM golang:1.21.1-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-crypto-alert

FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /docker-crypto-alert /docker-crypto-alert

USER nonroot:nonroot

ENTRYPOINT ["/docker-crypto-alert"]