FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
ENV GOWORK=off GONOSUMDB=* GOFLAGS=-mod=mod
WORKDIR /app
COPY shared/ ./shared/
COPY order-service/ ./order-service/
RUN cd order-service && go build -ldflags="-s -w" -o /bin/order-service ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/order-service /bin/order-service
EXPOSE 8082
ENTRYPOINT ["/bin/order-service"]
