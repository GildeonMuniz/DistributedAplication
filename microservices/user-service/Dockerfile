FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
ENV GOWORK=off GONOSUMDB=* GOFLAGS=-mod=mod
WORKDIR /app
COPY shared/ ./shared/
COPY user-service/ ./user-service/
RUN cd user-service && go build -ldflags="-s -w" -o /bin/user-service ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /bin/user-service /bin/user-service
EXPOSE 8081
ENTRYPOINT ["/bin/user-service"]
