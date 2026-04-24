FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
ENV GOWORK=off GONOSUMDB=* GOFLAGS=-mod=mod
WORKDIR /app
COPY shared/ ./shared/
COPY notification-service/ ./notification-service/
RUN cd notification-service && go build -ldflags="-s -w" -o /bin/notification-service ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/notification-service /bin/notification-service
ENTRYPOINT ["/bin/notification-service"]
