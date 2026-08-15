ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS build
ARG SERVICE
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/service ./cmd/${SERVICE}

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=build /out/service /service
ENTRYPOINT ["/service"]
