FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
ENV GOWORK=off GONOSUMDB=* GOFLAGS=-mod=mod
WORKDIR /app
COPY shared/ ./shared/
COPY api-gateway/ ./api-gateway/
RUN cd api-gateway && go build -ldflags="-s -w" -o /bin/api-gateway ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/api-gateway /bin/api-gateway
EXPOSE 8080
ENTRYPOINT ["/bin/api-gateway"]
